package config

var RuntimeConfig SvcConfig

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
