package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
)

var RuntimeConfig SvcConfig

func LoadConfig(config ...string) error {
	var configPath string
	var configInEnv string
	var err error

	if len(config) == 0 {
		var ok bool
		if configInEnv, ok = os.LookupEnv("SHOYA_CONFIG_JSON"); !ok {
			configPath = "./config.json"
		}
	}

	if configPath == "" && configInEnv == "" {
		for _, p := range config {
			p = path.Clean(p)
			if _, err := os.Stat(p); !os.IsNotExist(err) {
				configPath = p
				break
			}
		}

		if configPath == "" {
			return errors.New("no config found")
		}
	}

	if configInEnv != "" {
		err = json.Unmarshal([]byte(configInEnv), &RuntimeConfig)
		if err != nil {
			return err
		}

		return nil
	}

	if configPath != "" {
		var f []byte
		f, err = ioutil.ReadFile(configPath)
		if err != nil {
			return err
		}

		err = json.Unmarshal(f, &RuntimeConfig)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

// SvcConfig is the root configuration struct used for all core services.
type SvcConfig struct {
	Api       *ApiSvcConfig       `json:"api,omitempty"`
	Ws        *WsSvcConfig        `json:"ws,omitempty"`
	Discovery *DiscoverySvcConfig `json:"discovery,omitempty"`
	Files     *FilesSvcConfig     `json:"files,omitempty"`
}

// ApiSvcConfig is the configuration struct used by the `api` service.
type ApiSvcConfig struct {
	WebSvcConfig
	ApiConfigRefreshRateMs int `json:"apiConfigRefreshRateMs"` // The refresh rate of the dynamic configuration for the API.
}

// WsSvcConfig is the configuration struct used by the `ws` service.
type WsSvcConfig struct {
	WebSvcConfig
}

// DiscoverySvcConfig is the configuration struct used by the `discovery` service.
type DiscoverySvcConfig struct {
	WebSvcConfig
	DiscoveryApiKey string `json:"discoveryApiKey"` // The API key that is authorized to contact the Discovery service.
}

type FilesSvcConfig struct {
	GrpcSvcConfig
	Redis RedisSvcConfig `json:"redis"`
}

type WebSvcConfig struct {
	Fiber    FiberSvcConfig    `json:"fiber"`
	Redis    RedisSvcConfig    `json:"redis"`
	Postgres PostgresSvcConfig `json:"postgres"`
}

type FiberSvcConfig struct {
	ListenAddress string `json:"listen_address"`
	ProxyHeader   string `json:"proxy_header"`
	Prefork       bool   `json:"prefork"`
}

type RedisSvcConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Database int    `json:"db"`
}

type PostgresSvcConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"db"`
}

type GrpcSvcConfig struct {
	ListenAddress string `json:"listen_address"`
}
