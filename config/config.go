package config

import (
	"github.com/gogf/gf/v2/frame/g"
)

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
	Router     Router     `mapstructure:"router"`
	WsServer   WsServer   `mapstructure:"ws_server"`
	HttpServer HttpServer `mapstructure:"http_server"`
	Queue      Queue      `mapstructure:"queue"`
	Log        Log        `mapstructure:"log"`
	Redis      Redis      `mapstructure:"redis"`
	Kafka      Kafka      `mapstructure:"kafka"`
	Jwt        Jwt        `mapstructure:"jwt"`
	Sentry     Sentry     `mapstructure:"sentry"`
	Prometheus Prometheus `mapstructure:"prometheus"`
	Pprof      Pprof      `mapstructure:"pprof"`
	Swagger    Swagger    `mapstructure:"swagger"`
}

type Router struct {
	Enable  bool   `mapstructure:"enable"`
	Addr    string `mapstructure:"addr"`
	OutHost string `mapstructure:"out_host"`
}

type WsServer struct {
	Port int `mapstructure:"port"`
}

type HttpServer struct {
	Port int `mapstructure:"port"`
}
type Queue struct {
	Use   string `mapstructure:"use"`
	Redis Redis  `mapstructure:"redis"`
	Kafka Kafka  `mapstructure:"kafka"`
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

type Kafka struct {
	Broker  string `mapstructure:"broker"`
	Version string `mapstructure:"version"`
}

type Jwt struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}
type Sentry struct {
	DSN              string  `mapstructure:"dsn"`
	TracesSampleRate float64 `mapstructure:"traces_sample_rate"`
}

type Prometheus struct {
	Path   string `mapstructure:"path"`
	Addr   string `mapstructure:"addr"`
	Enable bool   `mapstructure:"enable"`
}

type Pprof struct {
	Enable bool `mapstructure:"enable"`
	Port   int  `mapstructure:"port"`
}

type Swagger struct {
	Enable bool   `mapstructure:"enable"`
	Path   string `mapstructure:"path"`
	Port   int    `mapstructure:"port"`
}

func NewViperConfig(configFullPath ...string) Config {

	c := &viperConfig{}
	c.load()

	g.Dump("configs:", c.Values())

	return c

}
