package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"infrastructure/logger"
	"interfaces/server/router"
)

func init() {
	router.Handle(http.MethodGet, `/redirect`, Redirect)
}

const defaultRedirectURL = `https://www.coinex.com`

func Redirect(c *gin.Context) {
	redirectURL := c.Query(`redirectURL`)
	if len(redirectURL) == 0 {
		redirectURL = defaultRedirectURL
		logger.Write(c, zap.ErrorLevel, `redirectURL type error or empty`)
	}
	logger.Write(c, zap.InfoLevel, `hit`,
		zap.String(`recipientName`, c.Query(`recipientName`)),
		zap.String(`redirectURL`, c.Query(`redirectURL`)),
	)
	c.Redirect(http.StatusFound, redirectURL)
}
