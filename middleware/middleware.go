package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Init(r *gin.Engine) {
	r.Use(gin.Logger())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
			log.Errorf("error recovery: %s")
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
}
