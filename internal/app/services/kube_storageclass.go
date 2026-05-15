package services

import (
	"context"
	"fmt"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/storageclass"
	"time"
)

func (s *Services) KubeCreateStorageClass(ctx context.Context, req *requests.KubeStorageClassCreateRequest) (*storagev1.StorageClass, error) {
	c, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	sc, err := storageclass.CreateStorageClass(c, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("storageclass %s already exists", req.Name)
			return nil, fmt.Errorf("storageclass %q already exists", req.Name)
		}
		return nil, fmt.Errorf("create storageclass failed: %w", err)
	}
	global.Logger.Infof("storageclass %q created successfully", sc.Name)
	return sc, nil
}

// internal/app/services/storageclass.go
func (s *Services) KubeStorageClassList(ctx context.Context, param *requests.KubeStorageClassListRequest) ([]storagev1.StorageClass, int, error) {
	return storageclass.GetStorageClassList(ctx, param.Name, param.Page, param.Limit)
}

func (s *Services) KubeStorageClassDetail(ctx context.Context, param *requests.KubeStorageClassDetailRequest) (*storagev1.StorageClass, error) {
	return storageclass.GetStorageClassDetail(ctx, param.Name)
}

func (s *Services) KubeStorageClassDelete(ctx context.Context, param *requests.KubeStorageClassDeleteRequest) error {
	return storageclass.DeleteStorageClass(ctx, param.Name)
}
