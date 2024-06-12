package speedtest

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ovh/configstore"
)

const (
	KeyInterface     = "SPEEDTEST_INTERFACE"
	KeyInterval      = "SPEEDTEST_INTERVAL"
	KeyRetryInterval = "SPEEDTEST_RETRY_INTERVAL"
	KeyLatitude      = "SPEEDTEST_LATITUDE"
	KeyLongitude     = "SPEEDTEST_LONGITUDE"

	DefaultInterface     = "eth0"
	DefaultInterval      = "15m"
	DefaultRetryInterval = "5m"
	DefaultLatitude      = ""
	DefaultLongitude     = ""
)

type Config struct {
	Interface     string
	Interval      time.Duration
	RetryInterval time.Duration
	City          string
	Latitude      *float64
	Longitude     *float64
}

func LoadConfig() (*Config, error) {
	var cfg Config

	iface, err := configstore.GetItemValue(KeyInterface)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the speedtest interface: %w", err)
		}
		cfg.Interface = DefaultInterface
	} else {
		cfg.Interface = iface
	}

	interval, err := configstore.GetItemValue(KeyInterval)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the speedtest interval: %w", err)
		}
		interval = DefaultInterval
	}
	parsedInterval, err := time.ParseDuration(interval)
	if err != nil {
		return nil, fmt.Errorf("invalid interval format: %w", err)
	}
	cfg.Interval = parsedInterval

	retryInterval, err := configstore.GetItemValue(KeyRetryInterval)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the speedtest retry interval: %w", err)
		}
		retryInterval = DefaultRetryInterval
	}
	parsedRetryInterval, err := time.ParseDuration(retryInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid retry interval format: %w", err)
	}
	cfg.RetryInterval = parsedRetryInterval

	latitude, err := configstore.GetItemValue(KeyLatitude)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the speedtest latitude: %w", err)
		}
		latitude = DefaultLatitude
	}
	if latitude != "" {
		parsedLatitude, err := strconv.ParseFloat(latitude, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude format: %w", err)
		}
		cfg.Latitude = &parsedLatitude
	}

	longitude, err := configstore.GetItemValue(KeyLongitude)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the speedtest longitude: %w", err)
		}
		longitude = DefaultLongitude
	}
	if longitude != "" {
		parsedLongitude, err := strconv.ParseFloat(longitude, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude format: %w", err)
		}
		cfg.Longitude = &parsedLongitude
	}

	return &cfg, nil
}
