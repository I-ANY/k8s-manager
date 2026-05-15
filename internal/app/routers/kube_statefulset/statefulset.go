package kube_statefulset

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/statefulset"
)

type KubeStatefulSetmentRouter struct{}

func NewKubeStatefulSetmentRouter() *KubeStatefulSetmentRouter {
	return &KubeStatefulSetmentRouter{}
}

// Inject 注册 Deployment 相关路由
func (r *KubeStatefulSetmentRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	statefulset := v1.NewKubeStatefulSetController()

	// 注册路由
	{

		router.POST("/create", statefulset.Create)
		router.GET("/list", statefulset.List)
		router.GET("/detail", statefulset.Detail)
		router.PUT("/scale", statefulset.Scale)
		router.PUT("/update_image", statefulset.UpdateImage)
		router.POST("/restart", statefulset.Restart)
		router.DELETE("/delete", statefulset.Delete)
		router.DELETE("/delete_svc", statefulset.DeleteService)
		//router.GET("/pods", statefulset.PodList)
		//router.POST("/events", statefulset.EventList)
	}
}
