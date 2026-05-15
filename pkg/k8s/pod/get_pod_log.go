package pod

import (
	"bytes"
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/global"
	"k8soperation/pkg/k8s/dataselect"
)

func GetPodLog(ctx context.Context, name, namespace, container string, tail int64, follow bool) (string, error) {
	// 创建Pod日志选项，包括容器名称、行数限制和是否跟随日志
	options := dataselect.NewPodLogOptions(container, tail, follow)

	// 获取指定Pod的日志流
	// 参数：命名空间、Pod名称和日志选项
	rc, err := global.KubeClient.CoreV1().Pods(namespace).GetLogs(name, options).Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("open log stream: %w", err)
	}
	// 确保日志流在使用后被关闭
	defer rc.Close()

	// 使用缓冲区存储日志内容
	var buf bytes.Buffer
	// 将日志流复制到缓冲区
	if _, err := io.Copy(&buf, rc); err != nil {
		return "", fmt.Errorf("read log: %w", err)
	}

	// 记录获取日志成功的日志，包括日志长度
	global.Logger.Infof("get kube_pod log success, len=%d", buf.Len())
	// 返回日志内容
	return buf.String(), nil
}

func NewPodLogOptions(container string, tailLines int64, follow bool) *corev1.PodLogOptions {
	// 统一默认与上限（来自全局配置）
	// 设置日志行数参数，如果输入值无效则使用默认值，同时考虑最大限制
	t := tailLines
	if t <= 0 {
		t = global.PodLogSetting.TailDefault // 如果tailLines小于等于0，则使用默认的行数设置
	}
	if max := global.PodLogSetting.TailMax; max > 0 && t > max { // 检查是否有最大行数限制
		t = max // 如果请求的行数超过最大限制，则使用最大限制值
	}

	// 创建Pod日志选项结构体，包含容器名称、行数限制、时间戳和是否获取历史日志等配置
	opts := &corev1.PodLogOptions{
		Container:  container,                       // 指定要获取日志的容器名称
		TailLines:  &t,                              // 设置从日志末尾开始显示的行数
		Timestamps: global.PodLogSetting.Timestamps, // 是否显示时间戳
		Previous:   global.PodLogSetting.Previous,   // 是否获取容器之前的日志（如果容器已重启）
	}

	// 根据是否跟随日志流来设置不同的选项
	if follow {
		opts.Follow = true // 跟随日志，实时获取日志输出
		// 跟随时不要设置 LimitBytes，避免被截断
	} else if lb := global.PodLogSetting.LimitBytes; lb > 0 { // 如果不是跟随模式，且设置了字节限制
		opts.LimitBytes = &lb // 一次性返回时可限制返回体大小，避免响应过大
	}

	return opts // 返回配置好的Pod日志选项
}
