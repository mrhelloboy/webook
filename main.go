package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initLogger()
	//initViperRemote()
	initViper()
	server := InitWebServer()
	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
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

	// 监听配置文件变更
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println(e.Name, e.Op)
	})

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperRemote() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}

	// 监听配置文件变更
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}

	// 该接口在远程配置变更是无效的
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println(e.Name, e.Op)
	//})

	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}
