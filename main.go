package main

import (
	"gin-chat-demo/conf"
	"gin-chat-demo/router"
	"gin-chat-demo/service"
)

func main() {
	conf.Init()
	go service.Manager.Start()
	r := router.NewRouter()
	_ = r.Run(conf.HttpPort)
}
