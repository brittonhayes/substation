package transform

import (
	"fmt"
	"time"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

const (
	DefaultFmt = "2006-01-02T15:04:05.000Z"
)

type timeUnixConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *timeUnixConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *timeUnixConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type timePatternConfig struct {
	Object iconfig.Object `json:"object"`

	Location string `json:"location"`
	Format   string `json:"format"`
}

func (c *timePatternConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *timePatternConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Format == "" {
		return fmt.Errorf("format: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func timeUnixToBytes(t time.Time) []byte {
	return []byte(fmt.Sprintf("%d", t.UnixMilli()))
}

// timeUnixToStr converts a UnixMilli timestamp to a string.
func timeUnixToStr(timeInt64 int64, timeFmt string, loc string) (string, error) {
	timeDate := time.UnixMilli(timeInt64)

	if loc != "" {
		ll, err := time.LoadLocation(loc)
		if err != nil {
			return "", fmt.Errorf("location %s: %v", loc, err)
		}

		timeDate = timeDate.In(ll)
	}

	return timeDate.Format(timeFmt), nil
}

func timeStrToUnix(timeStr, timeFmt string, loc string) (time.Time, error) {
	var timeDate time.Time
	if loc != "" {
		ll, err := time.LoadLocation(loc)
		if err != nil {
			return timeDate, fmt.Errorf("location %s: %v", loc, err)
		}

		pil, err := time.ParseInLocation(timeFmt, timeStr, ll)
		if err != nil {
			return timeDate, fmt.Errorf("format %s location %s: %v", timeFmt, loc, err)
		}

		timeDate = pil
	} else {
		p, err := time.Parse(timeFmt, timeStr)
		if err != nil {
			return timeDate, fmt.Errorf("format %s: %v", timeFmt, err)
		}

		timeDate = p
	}

	return timeDate, nil
}
