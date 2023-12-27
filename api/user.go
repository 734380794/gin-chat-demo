package api

import (
	"fmt"
	"gin-chat-demo/service"
	"github.com/gin-gonic/gin"
	logging "github.com/sirupsen/logrus"
)

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	fmt.Println("api-user")
	var userRegisterService service.UserRegisterService
	if err := c.ShouldBind(&userRegisterService); err == nil {
		register := userRegisterService.Register()
		c.JSON(200, register)
	} else {
		c.JSON(400, ErrorResponse(err))
		logging.Info(err)
	}
}
