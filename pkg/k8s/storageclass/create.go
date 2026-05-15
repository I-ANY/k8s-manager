package storageclass

import (
	"context"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

func CreateStorageClass(ctx context.Context, req *requests.KubeStorageClassCreateRequest) (*storagev1.StorageClass, error) {
	sc, err := buildStorageClassFromReq(req)
	if err != nil {
		return nil, err
	}

	created, err := global.KubeClient.StorageV1().
		StorageClasses().
		Create(ctx, sc, metav1.CreateOptions{})
	if err != nil {
		global.Logger.Errorf("create storageclass failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("storageclass %q created successfully", created.Name)
	return created, nil
}
