package controller

import (
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/middleware"
)

type AdminLoginController struct {
}

// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员接口
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [POST]
func (c *AdminLoginController) AdminLogin(context *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(context); err != nil {
		middleware.ResponseError(context, 1001, err)
		return
	}
	out := &dto.AdminLoginOutput{Token: params.UserName}
	middleware.ResponseSuccess(context, out)
}

func (c *AdminLoginController) AdminLogOut(context *gin.Context) {

}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.POST("/logout", adminLogin.AdminLogOut)
}
