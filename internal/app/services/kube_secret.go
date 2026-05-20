package services

import (
	"context"
	"encoding/json"
	"fmt"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/configmap"
	"k8soperation/pkg/k8s/secret"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Services) KubeCreateSecret(ctx context.Context,
	req *requests.KubeSecretCreateRequest) (*corev1.Secret, error) {
	return secret.CreateSecret(ctx, req)
}

// KubeSecretList 获取 Secret 列表（支持名称过滤 + 分页）
func (s *Services) KubeSecretList(ctx context.Context, param *requests.KubeSecretListRequest) ([]corev1.Secret, int, error) {
	return secret.GetSecretList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeSecretDetail(ctx context.Context, param *requests.KubeSecretDetailRequest) (*corev1.Secret, error) {
	return secret.GetSecretDetail(ctx, param.Name, param.Namespace)
}

// 删除 Secret
func (s *Services) KubeSecretDelete(ctx context.Context, param *requests.KubeSecretDeleteRequest) error {
	return secret.DeleteSecret(ctx, param.Name, param.Namespace)
}

// Strategic Merge Patch（结构合并）
func (s *Services) KubeSecretPatch(ctx context.Context, param *requests.KubeSecretUpdateRequest) (*corev1.Secret, error) {
	return secret.PatchSecret(ctx, param.Namespace, param.Name, []byte(param.Content))
}

// JSON Merge Patch（覆盖式更新）
func (s *Services) KubeSecretUpdate(ctx context.Context, req *requests.KubeSecretUpdateRequest) (*corev1.Secret, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var sec corev1.Secret
	if err := json.Unmarshal([]byte(req.Content), &sec); err != nil {
		return nil, fmt.Errorf("解析 Secret JSON 失败: %w", err)
	}
	if sec.Name == "" {
		return nil, fmt.Errorf("metadata.name 不能为空")
	}
	sec.Namespace = req.Namespace

	// 为了 Update 成功：拿旧对象的 resourceVersion（以及必要时沿用 types）
	old, err := s.App().KubeClient.CoreV1().Secrets(req.Namespace).Get(c, sec.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取原 Secret 失败: %w", err)
	}
	if sec.ResourceVersion == "" {
		sec.ResourceVersion = old.ResourceVersion
	}
	if sec.Type == "" {
		sec.Type = old.Type
	}

	// 全量覆盖
	updated, err := s.App().KubeClient.CoreV1().Secrets(req.Namespace).Update(c, &sec, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Secret 失败: %w", err)
	}
	return updated, nil
}

func (s *Services) KubeSecretDecode(ctx context.Context, param *requests.KubeSecretDecodeRequest) (map[string]string, error) {
	return secret.DecodeSecretData(ctx, param)
}

// Strategic Merge Patch（结构合并）
func (s *Services) KubeConfigMapPatch(
	ctx context.Context,
	param *requests.KubeConfigMapUpdateRequest,
) (*corev1.ConfigMap, error) {
	return configmap.PatchConfigMap(ctx, param.Namespace, param.Name, []byte(param.Content))
}

// JSON Merge Patch（覆盖式更新）
func (s *Services) KubeConfigMapUpdate(
	ctx context.Context,
	req *requests.KubeConfigMapUpdateRequest,
) (*corev1.ConfigMap, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var cm corev1.ConfigMap
	if err := json.Unmarshal([]byte(req.Content), &cm); err != nil {
		return nil, fmt.Errorf("解析 ConfigMap JSON 失败: %w", err)
	}
	if cm.Name == "" {
		return nil, fmt.Errorf("metadata.name 不能为空")
	}
	cm.Namespace = req.Namespace

	// 为了 Update 成功：拿旧对象的 ResourceVersion（防止冲突）
	old, err := s.App().KubeClient.CoreV1().
		ConfigMaps(req.Namespace).
		Get(c, cm.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取原 ConfigMap 失败: %w", err)
	}
	if cm.ResourceVersion == "" {
		cm.ResourceVersion = old.ResourceVersion
	}

	// 执行全量覆盖
	updated, err := s.App().KubeClient.CoreV1().
		ConfigMaps(req.Namespace).
		Update(c, &cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}
	return updated, nil
}
