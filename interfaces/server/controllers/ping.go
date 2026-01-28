package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"infrastructure/logger"
	"interfaces/server/router"
)

func init() {
	router.Handle(http.MethodGet, `/ping`, Ping)
}

func Ping(c *gin.Context) {
	logger.Write(c, zap.InfoLevel, `controllers.pin`)
	c.JSON(http.StatusOK, gin.H{
		`message`: `pong`,
	})
}
