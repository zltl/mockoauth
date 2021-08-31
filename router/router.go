package router

import (
	"mockoauth/controller"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {

	ctrl := controller.NewController()

	gacg := r.Group("/oauth2")
	gacg.Use(ctrl.LogsMiddleware)
	{
		gacg.GET("/", ctrl.InfoPage)
		gacg.GET("/ws/:ID", ctrl.InfoPageWS)
		gacg.GET("/authorize/:ID", ctrl.Authorize)
		gacg.POST("/authorize/:ID", ctrl.AuthorizePost)
		gacg.POST("/token/:ID", ctrl.Token)
	}
}
