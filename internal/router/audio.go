package router

import (
	"github.com/Vighnesh-V-H/async/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupAudioRoutes(router *gin.Engine, h *handler.AudioHandler) {
	audio := router.Group("/audio")
	{
		audio.POST("/generate", h.GenerateAudio)
		audio.GET("/status/:execution_id", h.GetStatus)
	}
}
