package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	server := InitWebServer()
	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}

func initViperV1() {
	// 配置文件的名字，但是不包含文件扩展名（不含含 .yaml, .toml 之类的后缀）
	viper.SetConfigName("dev")
	// 告诉 viper 我的配置用的是 yaml 格式
	// 现实中，有很多格式，JSON、TOML、XML、ini
	viper.SetConfigType("yaml")
	// 当前工作目录下的 config 子目录
	// 配置文件的路径，可以有多个，当多个的时候，会按照顺序依次查找
	viper.AddConfigPath("./config")
	//viper.AddConfigPath("/etc/webook")
	// 读取配置到 viper 里面去，或者可以理解为加载到内存里面
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	cfgFile := pflag.String("config", "config/config.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfgFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
