package handler

import (
	"net/http"

	"github.com/Vighnesh-V-H/async/internal/service"
	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
    svc *service.WorkflowService
}

func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
    return &WorkflowHandler{svc: svc}
}

func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
    var req service.CreateWorkflowRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx := c.Request.Context()
    wf, err := h.svc.CreateWorkflow(ctx, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, wf)
}
