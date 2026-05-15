package services

import (
	"context"
	"encoding/json"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/ingress"
	"time"
)

// KubeIngressCreate 创建 Ingress
func (s *Services) KubeIngressCreate(ctx context.Context, req *requests.KubeIngressCreateRequest) (*networkingv1.Ingress, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ing, err := ingress.CreateIngress(c, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("ingress %s/%s already exists", req.Namespace, req.Name)
			return nil, fmt.Errorf("ingress %q already exists in namespace %q", req.Name, req.Namespace)
		}
		return nil, fmt.Errorf("create ingress failed: %w", err)
	}

	global.Logger.Infof("ingress %s/%s created successfully", ing.Namespace, ing.Name)
	return ing, nil
}

func (s *Services) KubeIngressList(ctx context.Context, param *requests.KubeIngressListRequest) ([]networkingv1.Ingress, int, error) {
	return ingress.GetIngressList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeIngressDetail(ctx context.Context, param *requests.KubeIngressDetailRequest) (*networkingv1.Ingress, error) {
	return ingress.GetIngressDetail(ctx, param.Name, param.Namespace)
}

// Strategic Merge Patch（结构合并）
func (s *Services) KubeIngressPatch(ctx context.Context, param *requests.KubeIngressUpdateRequest) (*networkingv1.Ingress, error) {
	return ingress.PatchIngress(ctx, param.Namespace, param.Name, []byte(param.Content))
}

// JSON Merge Patch（覆盖式更新）
func (s *Services) KubeIngressPatchJSON(ctx context.Context, req *requests.KubeIngressUpdateRequest) (*networkingv1.Ingress, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Step 1: JSON 解析为 Ingress 对象
	var ing networkingv1.Ingress
	if err := json.Unmarshal([]byte(req.Content), &ing); err != nil {
		return nil, fmt.Errorf("解析 Ingress JSON 失败: %w", err)
	}
	if ing.Name == "" {
		return nil, fmt.Errorf("metadata.name 不能为空")
	}
	ing.Namespace = req.Namespace

	// Step 2: 获取旧对象，继承 ResourceVersion
	old, err := global.KubeClient.NetworkingV1().
		Ingresses(req.Namespace).
		Get(c, ing.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取原 Ingress 失败: %w", err)
	}
	if ing.ResourceVersion == "" {
		ing.ResourceVersion = old.ResourceVersion
	}

	// Step 3: 移除 managedFields，防止 update 冲突
	ing.ManagedFields = nil

	// Step 4: 执行全量覆盖更新（PUT）
	updated, err := global.KubeClient.NetworkingV1().
		Ingresses(req.Namespace).
		Update(c, &ing, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Ingress 失败: %w", err)
	}

	global.Logger.Infof("Ingress [%s] 在命名空间 [%s] 更新成功 (rv=%s)",
		updated.Name, updated.Namespace, updated.ResourceVersion)

	return updated, nil
}

// 删除 Ingress
func (s *Services) KubeIngressDelete(ctx context.Context, param *requests.KubeIngressDeleteRequest) error {
	return ingress.DeleteIngress(ctx, param.Name, param.Namespace)
}
