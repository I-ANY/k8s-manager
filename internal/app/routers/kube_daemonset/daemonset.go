package kube_daemonset

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/daemonset"
)

type KubeDaemonSetRouter struct{}

func NewKubeDaemonSetRouter() *KubeDaemonSetRouter {
	return &KubeDaemonSetRouter{}
}

func (r *KubeDaemonSetRouter) Inject(router *gin.RouterGroup) {
	ds := v1.NewKubeDaemonSetController()
	{
		router.POST("/create", ds.Create)
		router.GET("/list", ds.List)
		router.GET("/detail", ds.Detail)
		router.DELETE("/delete", ds.Delete)
		router.DELETE("/delete_service", ds.DeleteService)
		router.PUT("/update_image", ds.UpdateImage)
		router.POST("/restart", ds.Restart)
		router.POST("/rollback", ds.Rollback) // 需基于 ControllerRevision 实现
	}
}
