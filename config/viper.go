package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type viperConfig struct {
	values Values
}

func (c *viperConfig) load() Config {
	viper.SetEnvPrefix("WS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 添加命令行参数
	pflag.String("config", "", "set config file path")
	pflag.String("env", "", "set env,options:dev,prod,local")
	pflag.String("node", "", "set node,usage snowflake node id and sentry")
	pflag.Int("ws_port", 8084, "set ws server port")
	pflag.Int("http_port", 8085, "set http server port")
	pflag.String("router", "", "set router address")
	pflag.String("queue", "redis", "set queue type, options:redis,kafka")

	pflag.Parse()

	err := viper.BindEnv("env")
	if err != nil {
		panic(err)
	}
	err = viper.BindEnv("node")
	if err != nil {
		panic(err)
	}
	// 绑定环境变量（确保在 BindPFlag 之后）
	err = viper.BindEnv("config")
	if err != nil {
		panic(err)
	}

	// 绑定配置文件路径
	err = viper.BindPFlag("config", pflag.Lookup("config"))
	if err != nil {
		panic(err)
	}

	// 优先检查是否指定了配置文件路径
	configPath := viper.GetString("config")

	if configPath != "" {
		fmt.Println("使用指定的配置文件路径:", configPath)
		viper.SetConfigFile(configPath)
	} else {
		fmt.Println("使用默认配置文件路径")
		viper.SetConfigName("Config")
		viper.AddConfigPath("./conf")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
	}

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

	err = viper.BindPFlag("queue.use", pflag.Lookup("queue"))
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
