package router

import (
	"github.com/gin-gonic/gin"
)

var engin = gin.Default()

func Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) {
	engin.Handle(httpMethod, relativePath, handlers...)
}

func Run() error {
	return engin.Run()
}
