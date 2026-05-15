package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/pvc"
	"time"
)

// KubeCreatePVC 创建 PersistentVolumeClaim
func (s *Services) KubeCreatePVC(ctx context.Context, req *requests.KubePVCCreateRequest) (*corev1.PersistentVolumeClaim, error) {
	// 1) 设置超时
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 2) 调用资源层进行构建 + 创建
	created, err := pvc.CreatePersistentVolumeClaim(c, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s already exists in namespace %s", req.Name, req.Namespace)
			return nil, fmt.Errorf("PersistentVolumeClaim %q already exists in namespace %q", req.Name, req.Namespace)
		}
		global.Logger.Errorf("create PersistentVolumeClaim failed: %v", err)
		return nil, err
	}

	// 3) 成功日志
	global.Logger.Infof("PersistentVolumeClaim %q created successfully in namespace %q", created.Name, req.Namespace)
	return created, nil
}

// KubePVCList 获取 PVC 列表（支持分页与名称模糊）
func (s *Services) KubePVCList(ctx context.Context, param *requests.KubePVCListRequest) ([]corev1.PersistentVolumeClaim, int, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	items, total, err := pvc.GetPVCList(c, param.Namespace, param.Name, param.Page, param.Limit)
	if err != nil {
		global.Logger.Errorf("list PVC failed: %v", err)
		return nil, 0, err
	}
	return items, total, nil
}

// KubePVCDetail 获取 PVC 详情
func (s *Services) KubePVCDetail(ctx context.Context, param *requests.KubePVCDetailRequest) (*corev1.PersistentVolumeClaim, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pvcDetail, err := pvc.GetPVCDetail(c, param.Namespace, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s/%s not found", param.Namespace, param.Name)
			return nil, fmt.Errorf("PersistentVolumeClaim %q not found in namespace %q", param.Name, param.Namespace)
		}
		global.Logger.Error("get PersistentVolumeClaim detail failed", zap.Error(err))
		return nil, err
	}

	return pvcDetail, nil
}

func (s *Services) KubePVCDelete(ctx context.Context, param *requests.KubePVCDeleteRequest) error {
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := pvc.DeletePersistentVolumeClaim(c, param.Namespace, param.Name); err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s/%s not found", param.Namespace, param.Name)
			return nil // 幂等
		}
		global.Logger.Errorf("delete PersistentVolumeClaim %s/%s failed: %v", param.Namespace, param.Name, err)
		return err
	}

	global.Logger.Infof("PersistentVolumeClaim %s/%s deleted successfully", param.Namespace, param.Name)
	return nil
}

// 扩容 PVC：仅允许修改 spec.resources.requests.storage
func (s *Services) KubePVCResize(ctx context.Context, req *requests.KubePVCResizeRequest) (*corev1.PersistentVolumeClaim, error) {
	c, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// 1) 读取当前 PVC
	curr, err := global.KubeClient.CoreV1().
		PersistentVolumeClaims(req.Namespace).
		Get(c, req.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("PersistentVolumeClaim %q not found in namespace %q", req.Name, req.Namespace)
		}
		return nil, err
	}

	// 2) 解析并比较容量（新值必须更大）
	currReq := curr.Spec.Resources.Requests[corev1.ResourceStorage]
	newQty, err := resource.ParseQuantity(req.Storage)
	if err != nil {
		return nil, fmt.Errorf("invalid storage quantity %q: %w", req.Storage, err)
	}
	if newQty.Cmp(currReq) <= 0 {
		return nil, fmt.Errorf("new storage %q must be greater than current %q", newQty.String(), currReq.String())
	}

	// 3) 校验 StorageClass 是否允许扩容（若有 SC）
	if curr.Spec.StorageClassName != nil && *curr.Spec.StorageClassName != "" {
		scName := *curr.Spec.StorageClassName
		sc, err := global.KubeClient.StorageV1().StorageClasses().Get(c, scName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("get StorageClass %q failed: %w", scName, err)
		}
		if sc.AllowVolumeExpansion == nil || !*sc.AllowVolumeExpansion {
			return nil, fmt.Errorf("StorageClass %q does not allow volume expansion", scName)
		}
	}

	// 4) 调用资源层 Patch（你已实现）
	updated, err := pvc.PatchPVC(c, req)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
