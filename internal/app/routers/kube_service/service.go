package kube_service

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/svc"
)

type KubeServiceRouter struct{}

func NewKubeServiceRouter() *KubeServiceRouter {
	return &KubeServiceRouter{}
}

// Inject 注册 Service 相关路由
func (r *KubeServiceRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	service := v1.NewKubeServiceController()

	// 注册路由
	{
		router.POST("/create", service.Create)         // 创建 Service
		router.GET("/list", service.List)              // 获取 Service 列表
		router.GET("/detail", service.Detail)          // 获取 Service 详情
		router.DELETE("/delete", service.Delete)       // 删除 Service
		router.PATCH("/patch", service.Patch)          // Patch 更新 Service
		router.POST("/patch_json", service.PatchJSON)  // JSON Patch
		router.GET("/endpoints", service.GetEndpoints) // JSON Patch
	}
}
