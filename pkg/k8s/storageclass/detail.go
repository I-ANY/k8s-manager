package storageclass

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	k8sclient "k8soperation/pkg/k8s"
	"time"
)

func GetStorageClassDetail(client *k8sclient.Client, ctx context.Context, name string) (*storagev1.StorageClass, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sc, err := global.KubeClient.StorageV1().
		StorageClasses().
		Get(c, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Log().Error("storageclass not found", zap.String("name", name))
			return nil, fmt.Errorf("storageclass %s not found", name)
		}
		client.Log().Error("get storageclass failed", zap.String("name", name), zap.Error(err))
		return nil, err
	}
	return sc, nil
}
