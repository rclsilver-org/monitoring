package server

import (
	"fmt"
	"strings"

	"github.com/ovh/configstore"
)

const (
	KeyListenHost     = "HTTP_LISTEN_HOST"
	KeyListenPort     = "HTTP_LISTEN_PORT"
	KeyAllowedSources = "HTTP_ALLOWED_SOURCES"

	DefaultListenHost = "localhost"
	DefaultListenPort = 8080
)

type config struct {
	ListenHost string
	ListenPort int

	AllowedSources []string

	Title   string
	Version string

	Verbose bool
}

func loadConfig() (*config, error) {
	var cfg config

	host, err := configstore.GetItemValue(KeyListenHost)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the HTTP listen host: %w", err)
		}
		cfg.ListenHost = DefaultListenHost
	} else {
		cfg.ListenHost = host
	}

	port, err := configstore.GetItemValueInt(KeyListenPort)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the HTTP listen port: %w", err)
		}
		cfg.ListenPort = DefaultListenPort
	} else {
		cfg.ListenPort = int(port)
	}

	allowedSources, err := configstore.GetItemValue(KeyAllowedSources)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the HTTP allowed sources: %w", err)
		}
	} else {
		for _, v := range strings.Split(allowedSources, ",") {
			if vt := strings.TrimSpace(v); len(vt) > 0 {
				cfg.AllowedSources = append(cfg.AllowedSources, vt)
			}
		}
	}

	return &cfg, nil
}
