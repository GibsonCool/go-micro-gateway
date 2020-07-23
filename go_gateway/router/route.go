package router

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go-micro-gateway/go_gateway/controller"
	"go-micro-gateway/go_gateway/docs"
	"go-micro-gateway/go_gateway/middleware"
	"log"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	// programatically set swagger info
	docs.SwaggerInfo.Title = lib.GetStringConf("base.swagger.title")
	docs.SwaggerInfo.Description = lib.GetStringConf("base.swagger.desc")
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = lib.GetStringConf("base.swagger.host")
	docs.SwaggerInfo.BasePath = lib.GetStringConf("base.swagger.base_path")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 登录、退出模块
	adminLoginRouter := router.Group("/admin_login")
	store, err := sessions.NewRedisStore(10,
		"tcp",
		lib.GetStringConf("base.session.redis_server"),
		lib.GetStringConf("base.session.redis_password"),
		[]byte("secret"))
	if err != nil {
		log.Fatalf("sessions.NewRedisStore err:%v", err)
	}

	adminLoginRouter.Use(
		sessions.Sessions("mysession", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.TranslationMiddleware())
	{
		controller.AdminLoginRegister(adminLoginRouter)
	}

	return router
}
