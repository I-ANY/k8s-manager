package common

import (
	"context"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/pkg/k8s"
)

// GetContainerNames 获取容器名称列表
func GetContainerNames(podTemplate *corev1.PodSpec) []string {
	var containerNames []string
	for _, container := range podTemplate.Containers {
		containerNames = append(containerNames, container.Name)
	}
	return containerNames
}

// GetInitContainerNames 获取Init容器名列表
func GetInitContainerNames(podTemplate *corev1.PodSpec) []string {
	var initContainerNames []string
	for _, container := range podTemplate.InitContainers {
		initContainerNames = append(initContainerNames, container.Name)
	}
	return initContainerNames
}

// GetInitContainerImages 获取Init容器的镜像列表
func GetInitContainerImages(podTemplate *corev1.PodSpec) []string {
	var initContainerImages []string
	for _, container := range podTemplate.InitContainers {
		initContainerImages = append(initContainerImages, container.Image)
	}
	return initContainerImages
}

// GetContainerImages 获取容器的镜像
func GetContainerImages(podTemplate *corev1.PodSpec) []string {
	var containerImages []string
	for _, container := range podTemplate.Containers {
		containerImages = append(containerImages, container.Image)
	}
	return containerImages
}

// GetAllContainerNames 返回 Pod 所有容器（常规 + Init）的名称
// 顺序就是：普通容器在前，Init 容器在后（因为你先遍历了 Containers，
// 再遍历 InitContainers）。返回值是 []string，可以直接当作字符串数组使用。
func GetAllContainerNames(podTemplate *corev1.PodSpec) []string {
	var names []string
	for _, c := range podTemplate.Containers {
		names = append(names, c.Name)
	}
	for _, c := range podTemplate.InitContainers {
		names = append(names, c.Name)
	}
	return names
}

// GetAllContainerImages 返回 Pod 所有容器（常规 + Init）的镜像
// 顺序：普通容器的镜像在前，Init 容器的镜像在后。
// 返回值是 []string，纯镜像列表。
func GetAllContainerImages(podTemplate *corev1.PodSpec) []string {
	var images []string
	for _, c := range podTemplate.Containers {
		images = append(images, c.Image)
	}
	for _, c := range podTemplate.InitContainers {
		images = append(images, c.Image)
	}
	return images
}

// kube_pod/log.go
func OpenPodLogStream(client *k8s.Client, ctx context.Context, name, ns, container string, tail int64) (io.ReadCloser, error) {
	// 统一默认 & 上限（用全局配置）
	if tail <= 0 {
		tail = client.PodLogSetting.TailDefault
	}
	if max := client.PodLogSetting.TailMax; max > 0 && tail > max {
		tail = max
	}

	follow := client.PodLogSetting.EnableStreaming // 全局决定是否跟随

	opts := &corev1.PodLogOptions{
		Container:  container,
		Follow:     follow,
		TailLines:  &tail,
		Timestamps: client.PodLogSetting.Timestamps,
		Previous:   client.PodLogSetting.Previous,
	}
	if lb := client.PodLogSetting.LimitBytes; lb > 0 && !follow {
		// 跟随时一般不设 LimitBytes；一次性时可作为上限
		opts.LimitBytes = &lb
	}

	return client.Interface.CoreV1().
		Pods(ns).
		GetLogs(name, opts).
		Stream(ctx) // 传入请求上下文，支持取消
}
