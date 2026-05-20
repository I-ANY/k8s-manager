package secret

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
func PatchSecret(client *k8s.Client, ctx context.Context, namespace, name string, patch []byte) (*corev1.Secret, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sec, err := client.Interface.
		CoreV1().
		Secrets(namespace).
		Patch(c, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return sec, nil
}

func PatchSecretJson(client *k8s.Client, ctx context.Context, namespace, content string) (*corev1.Secret, error) {
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

	// 反序列化 JSON -> Secret 对象
	var sec corev1.Secret
	if err := json.Unmarshal([]byte(content), &sec); err != nil {
		return nil, fmt.Errorf("解析 Secret JSON 失败: %w", err)
	}

	// 保证 namespace 一致
	sec.Namespace = namespace
	if sec.Name == "" {
		return nil, fmt.Errorf("Secret 名称不能为空（JSON 中未包含 metadata.name）")
	}

	// 获取旧的 Secret，用于继承类型（防止覆盖为空）
	old, err := client.Interface.CoreV1().
		Secrets(namespace).
		Get(c, sec.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取原 Secret 失败: %w", err)
	}

	if sec.Type == "" {
		sec.Type = old.Type
	}

	// 执行全量更新（PUT）
	updated, err := client.Interface.CoreV1().
		Secrets(namespace).
		Update(c, &sec, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Secret 失败: %w", err)
	}

	client.Logger.Infof("Secret [%s] 在命名空间 [%s] 更新成功 (类型: %s)", sec.Name, namespace, sec.Type)
	return updated, nil
}
