package pod

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/pkg/k8s"
)

// init 容器同理
func PatchPodInitContainerImage(client *k8s.Client, ns, pod, initContainerName, newImage string) error {
	// 创建一个patch对象，用于更新Kubernetes Pod的配置
	patchObj := map[string]any{
		"spec": map[string]any{
			// 在Pod的规范中添加initContainers
			"initContainers": []map[string]string{
				// 定义一个init容器，包含名称和镜像信息
				{"name": initContainerName, "image": newImage},
			},
		},
	}
	// 将patch对象转换为JSON格式的字节切片
	b, _ := json.Marshal(patchObj)

	// 使用Kubernetes客户端对指定的Pod进行战略性合并补丁操作
	// 参数包括：命名空间、Pod名称、补丁类型、补丁数据和补丁选项
	_, err := client.Interface.CoreV1().
		Pods(ns).
		Patch(context.TODO(), pod, types.StrategicMergePatchType, b, metav1.PatchOptions{})
	// 返回操作过程中可能出现的错误
	return err
}
