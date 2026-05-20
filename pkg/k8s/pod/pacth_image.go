package pod

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/pkg/k8s"
)

func PatchPodContainerImage(client *k8s.Client, ns, pod, containerName, newImage string) error {
	// patchObj 定义了要应用的补丁内容，使用map结构表示JSON
	patchObj := map[string]any{
		"spec": map[string]any{
			"containers": []map[string]string{
				{"name": containerName, "image": newImage}, // 指定要修改的容器名称和新的镜像
			},
		},
	}
	// 将 map对象序列化为JSON字节数组
	b, _ := json.Marshal(patchObj)

	// 使用Kubernetes客户端应用补丁，修改Pod中容器的镜像
	_, err := client.Interface.CoreV1().
		Pods(ns).                                                                           // 指定命名空间
		Patch(context.TODO(), pod, types.StrategicMergePatchType, b, metav1.PatchOptions{}) // 执行补丁操作
	return err // 返回操作结果
}
