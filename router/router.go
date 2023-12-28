package router

import (
	"gin-chat-demo/api"
	"gin-chat-demo/service"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery(), gin.Logger())
	v1 := r.Group("/v1")
	{
		v1.GET("ping", func(c *gin.Context) {
			c.JSON(200, "success")
		})
		v1.POST("user/register", api.UserRegister)
		v1.GET("ws", service.Handler)
	}
	return r
}
