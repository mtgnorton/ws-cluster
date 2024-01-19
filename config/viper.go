package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sync"
)

type viperConfig struct {
	values Values
}

var once sync.Once

func (c *viperConfig) load() Config {
	viper.SetConfigName("Config")
	viper.AddConfigPath("./conf")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("WS") // 设置环境变量前缀，Viper在自动绑定环境变量时会带上这个前缀
	viper.AutomaticEnv()     // Viper从环境变量中读取配置

	pflag.String("server.node", "1", "server node") // 设置一个命令行参数
	pflag.String("server.port", "8080", "server port")
	pflag.String("env", "prod", "env")

	pflag.Parse()                              // 解析命令行参数
	err := viper.BindPFlags(pflag.CommandLine) // 把命令行参数绑定到Viper上
	if err != nil {
		panic(err)
	}
	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.AutomaticEnv()
	err = viper.Unmarshal(&c.values)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *viperConfig) Values() *Values {
	return &c.values
}
