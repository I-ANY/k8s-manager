package kube_pod

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/pod"
)

type kubePodRouter struct{}

func NewkubePodRouter() *kubePodRouter {
	return &kubePodRouter{}
}

func (r *kubePodRouter) Inject(router *gin.RouterGroup) {
	// 创建一个新的Pod控制器实例
	pod := v1.NewPodController()

	// 设置HTTP路由和处理函数
	router.GET("/list", pod.List)                                   // 获取Pod列表
	router.GET("/detail", pod.Detail)                               // 获取Pod详情
	router.PUT("/update", pod.Update)                               // 更新Pod
	router.DELETE("/grace_delete_pod", pod.DeletePod)               // 删除Pod
	router.GET("/container_name", pod.GetContainerName)             // 获取容器名称
	router.GET("/init_container_name", pod.GetInitContainerName)    // 获取初始化容器名称
	router.GET("/container_image", pod.GetContainerImages)          // 获取容器镜像
	router.GET("/init_container_image", pod.GetInitContainerImages) // 获取初始化容器镜像
	router.GET("/container_logs", pod.GetContainerLog)              // 获取容器日志
	router.PUT("/patch_image", pod.PatchImage)                      // 更新Pod容器镜像

}
