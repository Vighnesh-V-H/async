package router

import (
	"github.com/Vighnesh-V-H/async/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupWorkflowRoutes(router *gin.Engine , h *handler.WorkflowHandler ) {
	routes:= router.Group("/workflow")
	{
		routes.POST("/create" , h.CreateWorkflow)
	}
}