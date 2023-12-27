package main

import (
	"fmt"
	"gin-chat-demo/conf"
	"gin-chat-demo/router"
)

func main() {
	fmt.Println("gin-chat-demo")
	conf.Init()
	r := router.NewRouter()
	_ = r.Run(conf.HttpPort)
}
