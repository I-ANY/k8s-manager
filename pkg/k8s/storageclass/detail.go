package storageclass

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"time"
)

func GetStorageClassDetail(ctx context.Context, name string) (*storagev1.StorageClass, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sc, err := global.KubeClient.StorageV1().
		StorageClasses().
		Get(c, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Error("storageclass not found", zap.String("name", name))
			return nil, fmt.Errorf("storageclass %s not found", name)
		}
		global.Logger.Error("get storageclass failed", zap.String("name", name), zap.Error(err))
		return nil, err
	}
	return sc, nil
}
