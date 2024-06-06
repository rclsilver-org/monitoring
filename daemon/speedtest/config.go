package speedtest

import (
	"fmt"

	"github.com/ovh/configstore"
)

const (
	KeyInterface = "SPEEDTEST_INTERFACE"

	DefaultInterface = "eth0"
)

type Config struct {
	Interface string
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

	return &cfg, nil
}
