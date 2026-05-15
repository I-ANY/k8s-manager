package kube_pv

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/pv"
)

type KubePersistentVolumeRouter struct{}

func NewKubePersistentVolumeRouter() *KubePersistentVolumeRouter {
	return &KubePersistentVolumeRouter{}
}

func (r *KubePersistentVolumeRouter) Inject(router *gin.RouterGroup) {
	pv := v1.NewKubePVController()

	{
		router.POST("/create", pv.Create)    // 创建 PV
		router.GET("/list", pv.List)         // 获取 PV 列表
		router.GET("/detail", pv.Detail)     // 获取 PV 详情
		router.DELETE("/delete", pv.Delete)  // 删除 PV
		router.PATCH("/reclaim", pv.Reclaim) // 修改回收策略
	}
}
