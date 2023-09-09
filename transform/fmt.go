package transform

import (
	"fmt"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type fmtBase64Config struct {
	Object iconfig.Object `json:"object"`
}

func (c *fmtBase64Config) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *fmtBase64Config) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type fmtFQDNConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *fmtFQDNConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *fmtFQDNConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}
