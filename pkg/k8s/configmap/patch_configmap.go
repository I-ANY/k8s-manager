package configmap

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/pkg/k8s"
	"strings"
	"time"
)

// 通用 StrategicMergePatch（适合结构化对象）
func PatchConfigMap(client *k8s.Client, ctx context.Context, namespace, name string, patch []byte) (*corev1.ConfigMap, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cm, err := client.Interface.
		CoreV1().
		ConfigMaps(namespace).
		Patch(c, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func PatchConfigMapJson(client *k8s.Client, ctx context.Context, namespace, content string) (*corev1.ConfigMap, error) {
	// 设置超时（10s）
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 参数校验
	if strings.TrimSpace(namespace) == "" {
		return nil, fmt.Errorf("namespace 不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content 不能为空")
	}

	// 反序列化 JSON -> ConfigMap 对象
	var cm corev1.ConfigMap
	if err := json.Unmarshal([]byte(content), &cm); err != nil {
		return nil, fmt.Errorf("解析 ConfigMap JSON 失败: %w", err)
	}

	// 保证 namespace 一致
	cm.Namespace = namespace
	if cm.Name == "" {
		return nil, fmt.Errorf("ConfigMap 名称不能为空（JSON 中未包含 metadata.name）")
	}

	// 获取旧对象，用于继承 Labels / Annotations（可选）
	old, err := client.Interface.CoreV1().
		ConfigMaps(namespace).
		Get(c, cm.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取原 ConfigMap 失败: %w", err)
	}

	// 如果更新对象未提供 Labels/Annotations，则继承旧值
	if cm.Labels == nil {
		cm.Labels = old.Labels
	}
	if cm.Annotations == nil {
		cm.Annotations = old.Annotations
	}

	// 执行全量更新（PUT）
	updated, err := client.Interface.CoreV1().
		ConfigMaps(namespace).
		Update(c, &cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}

	client.Logger.Infof("ConfigMap [%s] 在命名空间 [%s] 更新成功", cm.Name, namespace)
	return updated, nil
}
