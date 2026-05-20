package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/pkg/k8s"
	"strings"
	"time"
)

// 通用 Patch：传入 StrategicMergePatch 的 bytes，返回最新的 Ingress
func PatchIngress(client *k8s.Client, ctx context.Context, namespace, name string, patch []byte) (*networkingv1.Ingress, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ing, err := client.Interface.
		NetworkingV1().
		Ingresses(namespace).
		Patch(c, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return ing, nil
}

// JSON MergePatch：传入 RFC 7386 的 bytes，返回最新的 Ingress
// PatchJsonIngress 全量更新（Update）Ingress
// 虽然名为 Patch，但实际行为为 PUT 覆盖更新：
// - 接收完整 Ingress JSON 对象
// - 自动补齐 ResourceVersion，避免 409 冲突
// - 命名空间由参数 namespace 指定
func PatchJsonIngress(client *k8s.Client, ctx context.Context, namespace, content string) (*networkingv1.Ingress, error) {
	// 统一上下文超时
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 参数校验
	if strings.TrimSpace(namespace) == "" {
		return nil, fmt.Errorf("namespace 不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content 不能为空")
	}

	// JSON → Ingress 对象
	var ing networkingv1.Ingress
	if err := json.Unmarshal([]byte(content), &ing); err != nil {
		return nil, fmt.Errorf("解析 Ingress JSON 失败: %w", err)
	}

	// 保证 namespace 一致
	ing.Namespace = namespace
	if ing.Name == "" {
		return nil, fmt.Errorf("Ingress 名称不能为空（JSON 中未包含 metadata.name）")
	}

	// 获取旧 Ingress，确保 ResourceVersion 一致（防止 409）
	old, err := client.Interface.NetworkingV1().
		Ingresses(namespace).
		Get(c, ing.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取原 Ingress 失败: %w", err)
	}

	if ing.ResourceVersion == "" {
		ing.ResourceVersion = old.ResourceVersion
	}

	// 防止 managedFields 引起冲突
	ing.ManagedFields = nil

	// 全量更新（PUT）
	updated, err := client.Interface.NetworkingV1().
		Ingresses(namespace).
		Update(c, &ing, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Ingress 失败: %w", err)
	}

	client.Logger.Infof("Ingress [%s] 在命名空间 [%s] 更新成功 (rv=%s)",
		updated.Name, namespace, updated.ResourceVersion)

	return updated, nil
}
