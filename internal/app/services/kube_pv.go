package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/pv"
	"time"
)

// KubeCreatePV 创建 PersistentVolume
func (s *Services) KubeCreatePV(ctx context.Context, req *requests.KubePVCreateRequest) (*corev1.PersistentVolume, error) {
	// 1) 设置超时
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 2) 调用资源层进行构建 + 创建
	created, err := pv.CreatePersistentVolume(s.App().K8sClient(), c, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			s.App().Logger.Warnf("PersistentVolume %s already exists", req.Name)
			return nil, fmt.Errorf("PersistentVolume %q already exists", req.Name)
		}
		s.App().Logger.Errorf("create PersistentVolume failed: %v", err)
		return nil, err
	}

	// 3) 成功日志
	s.App().Logger.Infof("PersistentVolume %q created successfully", created.Name)
	return created, nil
}
func (s *Services) KubePVList(ctx context.Context, param *requests.KubePVListRequest) ([]corev1.PersistentVolume, int, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	items, total, err := pv.GetPVList(c, param.Name, param.Page, param.Limit)
	if err != nil {
		s.App().Logger.Errorf("list PV failed: %v", err)
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Services) KubePVDetail(ctx context.Context, param *requests.KubePVDetailRequest) (*corev1.PersistentVolume, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pvDetail, err := pv.GetPVDetail(s.App().K8sClient(), c, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			s.App().Logger.Warnf("PersistentVolume %s not found", param.Name)
			return nil, fmt.Errorf("PersistentVolume %q not found", param.Name)
		}
		s.App().Logger.Error("get PersistentVolume detail failed", zap.Error(err))
		return nil, err
	}

	return pvDetail, nil
}

func (s *Services) KubePVDelete(ctx context.Context, param *requests.KubePVDeleteRequest) error {
	// 设置超时边界（整个删除过程不超过 30 秒）
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := pv.DeletePersistentVolume(c, param.Name); err != nil {
		if apierrors.IsNotFound(err) {
			s.App().Logger.Warnf("PersistentVolume %s not found", param.Name)
			return nil
		}
		s.App().Logger.Errorf("delete PersistentVolume %s failed: %v", param.Name, err)
		return err
	}

	s.App().Logger.Infof("PersistentVolume %s deleted successfully", param.Name)
	return nil
}

// 修改回收策略
func (s *Services) KubePVReclaim(ctx context.Context, req *requests.KubePVReclaimRequest) (*corev1.PersistentVolume, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return pv.ReclaimPersistentVolume(c, req)
}
