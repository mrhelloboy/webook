package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	app := InitAPP()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	go func() {
		err := app.webAdmin.Start()
		log.Println(err)
	}()
	err := app.server.Serve()
	log.Println(err)
}

func initViper() {
	cfile := pflag.String("config", "config/config.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println(e.Name, e.Op)
		fmt.Println(viper.GetString("db.dsn"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
