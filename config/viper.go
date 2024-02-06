package config

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type viperConfig struct {
	values Values
}

func (c *viperConfig) load() Config {
	viper.SetConfigName("Config")

	viper.AddConfigPath("./conf")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("WS") // 设置环境变量前缀，Viper在自动绑定环境变量时会带上这个前缀
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	//viper.AutomaticEnv() // Viper从环境变量中读取配置

	err := viper.BindEnv("env")
	if err != nil {
		panic(err)
	}

	err = viper.BindEnv("node")
	if err != err {
		panic(err)
	}

	pflag.String("env", "prod", "set env,options:dev,prod,local")
	pflag.String("node", "100", "set node,usage snowflake node id and sentry")
	pflag.Int("ws_port", 8084, "set ws server port")
	pflag.Int("http_port", 8085, "set http server port")
	pflag.String("router", "", "set router address")

	pflag.Parse() // 解析命令行参数

	err = viper.BindPFlag("env", pflag.Lookup("env"))
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlag("node", pflag.Lookup("node"))
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlag("ws_server.port", pflag.Lookup("ws_port"))
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlag("http_server.port", pflag.Lookup("http_port"))
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlag("router.addr", pflag.Lookup("router"))
	if err != nil {
		panic(err)
	}

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&c.values)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *viperConfig) Values() *Values {
	return &c.values
}
