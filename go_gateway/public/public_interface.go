package public

import "github.com/gin-gonic/gin"

// 参数绑定和校验
type BindValidParamInterface interface {
	BindValidParam(ctx *gin.Context) error
}
