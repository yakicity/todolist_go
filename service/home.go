package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Home renders index.html
func Home(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", gin.H{"Title": "HOME"})
}
