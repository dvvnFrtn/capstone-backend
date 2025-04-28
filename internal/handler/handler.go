package handler

import "github.com/gin-gonic/gin"

func Register(r *gin.Engine) {
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, "pong")
	})
}
