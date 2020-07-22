package controller

import (
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/middleware"
)

type AdminLoginController struct {
}

func (c *AdminLoginController) AdminLogin(context *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(context); err != nil {
		middleware.ResponseError(context, 1001, err)
		return
	}
	middleware.ResponseSuccess(context, "yes")
}

func (c *AdminLoginController) AdminLogOut(context *gin.Context) {

}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.POST("/logout", adminLogin.AdminLogOut)
}
