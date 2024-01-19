package config

var DefaultConfig = NewViperConfig()

type Config interface {
	Values() *Values
}

type Env string

const (
	Dev   Env = "dev"
	Prod  Env = "prod"
	Local Env = "local"
)

type Values struct {
	Env        Env        `mapstructure:"env"`
	Node       int64      `mapstructure:"node"`
	WsServer   WsServer   `mapstructure:"ws_server"`
	HttpServer HttpServer `mapstructure:"http_server"`
	Log        Log        `mapstructure:"log"`
	Redis      Redis      `mapstructure:"redis"`
	RedisQueue Redis      `mapstructure:"redis_queue"`
}

type WsServer struct {
	Port int `mapstructure:"port"`
}

type HttpServer struct {
	Port int `mapstructure:"port"`
}
type Log struct {
	Path       string `mapstructure:"path"`
	Print      bool   `mapstructure:"print"`
	Level      string `mapstructure:"level"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func NewViperConfig() Config {

	c := &viperConfig{}
	c.load()
	return c

}
