//go:build !wasm

package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"time"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
	mess "github.com/brexhq/substation/message"
)

type procKVStoreConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// IgnoreClose determines whether the KV store is closed when the transform is closed.
	IgnoreClose bool `json:"ignore_close"`
	// Type determines the action applied to the KV store.
	//
	// Must be one of:
	//
	// - get: value is retrieved from the store
	//
	// - set: value is put into the store
	Type string `json:"type"`
	// Prefix is prepended to either the Key (in the case of get)
	// or the SetKey (in the case of set) and is intended to simplify
	// data management within a KV store.
	//
	// This is optional and defaults to an empty string.
	Prefix string `json:"prefix"`
	// TTLKey retrieves a value from an object that is used as the time-to-live (TTL)
	// of the item set into the KV store. This value must be an integer that represents
	// the epoch time (in seconds) when the item will be evicted from the store.
	//
	// This is optional and defaults to using no TTL when setting items into the store.
	TTLKey string `json:"ttl_key"`
	// TTLOffset is an offset (in seconds) used to determine the time-to-live (TTL)
	// of the item set into the KV store. If TTLKey is configured, then this value is
	// added to the TTL value retrieved from the object. If TTLKey is not used, then this
	// value is added to the current time.
	//
	// For example, if TTLKey is not set and the offset is 86400 (1 day), then the value
	// will be evicted from the store if more than 1 day has elapsed.
	//
	// This is optional and defaults to using no TTL when setting values into the store.
	TTLOffset int64 `json:"ttl_offset"`
	// KVOptions determine the type of KV store used by the transform. Refer to internal/kv
	// for more information.
	KVStore config.Config `json:"kv_store"`
}

type procKVStore struct {
	conf    procKVStoreConfig
	kvStore kv.Storer
}

func newProcKVStore(ctx context.Context, cfg config.Config) (*procKVStore, error) {
	conf := procKVStoreConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("new_proc_kv_store: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("new_proc_kv_store: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"get",
			"set",
		},
		conf.Type) {
		return nil, fmt.Errorf("new_proc_kv_store: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	kvStore, err := kv.Get(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("new_proc_kv_store: %v", err)
	}

	proc := procKVStore{
		conf:    conf,
		kvStore: kvStore,
	}

	if !proc.kvStore.IsEnabled() {
		if err := proc.kvStore.Setup(ctx); err != nil {
			return nil, fmt.Errorf("new_proc_kv_store: %v", err)
		}
	}

	return &proc, nil
}

func (t *procKVStore) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (t *procKVStore) Close(ctx context.Context) error {
	if t.conf.IgnoreClose {
		return nil
	}

	if t.kvStore.IsEnabled() {
		if err := t.kvStore.Close(); err != nil {
			return fmt.Errorf("close: proc_kv_store: %v", err)
		}
	}

	return nil
}

func (t *procKVStore) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.conf.Type {
		case "get":
			key := message.Get(t.conf.Key).String()
			if t.conf.Prefix != "" {
				key = fmt.Sprint(t.conf.Prefix, ":", key)
			}

			v, err := t.kvStore.Get(ctx, key)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_kv_store: %v", err)
			}

			if err := message.Set(t.conf.SetKey, v); err != nil {
				return nil, fmt.Errorf("transform: proc_kv_store: %v", err)
			}

			output = append(output, message)
		case "set":
			key := message.Get(t.conf.SetKey).String()
			if t.conf.Prefix != "" {
				key = fmt.Sprint(t.conf.Prefix, ":", key)
			}

			//nolint: nestif // ignore nesting complexity
			if t.conf.TTLKey != "" && t.conf.TTLOffset != 0 {
				ttl := message.Get(t.conf.TTLKey).Int() + t.conf.TTLOffset
				if err := t.kvStore.SetWithTTL(ctx, key, message.Get(t.conf.Key).String(), ttl); err != nil {
					return nil, fmt.Errorf("transform: proc_kv_store: %v", err)
				}
			} else if t.conf.TTLKey != "" {
				ttl := message.Get(t.conf.TTLKey).Int()
				if err := t.kvStore.SetWithTTL(ctx, key, message.Get(t.conf.Key).String(), ttl); err != nil {
					return nil, fmt.Errorf("transform: proc_kv_store: %v", err)
				}
			} else if t.conf.TTLOffset != 0 {
				ttl := time.Now().Add(time.Duration(t.conf.TTLOffset) * time.Second).Unix()
				if err := t.kvStore.SetWithTTL(ctx, key, message.Get(t.conf.Key).String(), ttl); err != nil {
					return nil, fmt.Errorf("transform: proc_kv_store: %v", err)
				}
			} else {
				if err := t.kvStore.Set(ctx, key, message.Get(t.conf.Key).String()); err != nil {
					return nil, fmt.Errorf("transform: proc_kv_store: %v", err)
				}
			}

			output = append(output, message)
		}
	}

	return output, nil
}
