package node

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	k8sclient "k8soperation/pkg/k8s"
	"time"
)

func GetNodeMetrics(client *k8sclient.Client, ctx context.Context, nodeName string) ([]models.NodeMetricItem, error) {
	// metrics client
	mc, err := metricsclient.NewForConfig(global.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("new metrics client: %w", err)
	}

	// 1) 拉 NodeMetrics
	nm, err := mc.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("metrics API error: %w (metrics-server installed?)", err)
	}

	// 2) 拉 Node（容量/可分配）
	node, err := global.KubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		// 容量拿不到不致命，给个告警但继续返回使用量
		client.Log().Warn("get node capacity failed", zap.String("nodeName", nodeName), zap.Error(err))
		node = &corev1.Node{}
	}

	// 3) 组装
	item := toMetricItem(nm, node)
	return []models.NodeMetricItem{item}, nil
}

func toMetricItem(nm *metricsv1beta1.NodeMetrics, node *corev1.Node) models.NodeMetricItem {
	// 获取时间戳和窗口时长（秒）
	ts := nm.Timestamp.Time
	windowSeconds := int64(nm.Window.Duration / time.Second)

	// 初始化CPU使用量（毫秒）和内存使用量（字节）为0
	cpuMilli := int64(0)
	memBytes := int64(0)
	// 检查CPU使用量是否存在，如果存在则获取其毫秒值
	if q := nm.Usage.Cpu(); q != nil {
		cpuMilli = q.MilliValue()
	}
	// 检查内存使用量是否存在，如果存在则获取其值
	if q := nm.Usage.Memory(); q != nil {
		memBytes = q.Value()
	}

	// 初始化容量相关的CPU和内存值为0
	capMilli, capMem := int64(0), int64(0)
	// 初始化分配相关的CPU和内存值为0
	allocMilli, allocMem := int64(0), int64(0)
	// 判断节点是否为空
	if node != nil {
		// 获取节点的CPU容量，如果存在则转换为毫秒值
		if q := node.Status.Capacity.Cpu(); q != nil {
			capMilli = q.MilliValue()
		}
		// 获取节点的内存容量，如果存在则获取其值
		if q := node.Status.Capacity.Memory(); q != nil {
			capMem = q.Value()
		}
		// 获取节点的可分配CPU资源，如果存在则转换为毫秒值
		if q := node.Status.Allocatable.Cpu(); q != nil {
			allocMilli = q.MilliValue()
		}
		// 获取节点的可分配内存资源，如果存在则获取其值
		if q := node.Status.Allocatable.Memory(); q != nil {
			allocMem = q.Value()
		}
	}

	/*
	 * 创建并返回一个NodeMetricItem结构体实例，该结构体包含了节点的各项指标数据
	 * 包括名称、时间戳、窗口时间以及CPU和内存的使用量和限制值
	 */
	m := models.NodeMetricItem{
		Name:          nm.Name,       // 节点名称
		Timestamp:     ts,            // 时间戳
		WindowSeconds: windowSeconds, // 窗口时间（秒）
		CPUUsageMilli: cpuMilli,      // CPU使用量（毫核）
		MemUsageBytes: memBytes,      // 内存使用量（字节）
		CPUCapMilli:   capMilli,      // CPU限制值（毫核）
		MemCapBytes:   capMem,        // 内存限制值（字节）
		CPUAllocMilli: allocMilli,    // CPU分配量（毫核）
		MemAllocBytes: allocMem,      // 内存分配量（字节）
	}
	// 计算CPU使用百分比
	// 首先确定分母，优先使用分配量，如果分配量为0则使用限制值
	denomCPU := allocMilli
	if denomCPU == 0 {
		denomCPU = capMilli
	}
	// 如果分母大于0，则计算CPU使用百分比
	if denomCPU > 0 {
		m.CPUUsagePercent = float64(cpuMilli) * 100 / float64(denomCPU)
	}
	// 计算内存使用百分比
	// 首先确定分母，优先使用分配量，如果分配量为0则使用限制值
	denomMem := allocMem
	if denomMem == 0 {
		denomMem = capMem
	}
	// 如果分母大于0，则计算内存使用百分比
	if denomMem > 0 {
		m.MemUsagePercent = float64(memBytes) * 100 / float64(denomMem)
	}
	// 返回填充好的NodeMetricItem结构体实例
	return m
}
