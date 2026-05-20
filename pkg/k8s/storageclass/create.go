package storageclass

import (
	"context"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	k8sclient "k8soperation/pkg/k8s"
)

func CreateStorageClass(client *k8sclient.Client, ctx context.Context, req *requests.KubeStorageClassCreateRequest) (*storagev1.StorageClass, error) {
	sc, err := buildStorageClassFromReq(req)
	if err != nil {
		return nil, err
	}

	created, err := global.KubeClient.StorageV1().
		StorageClasses().
		Create(ctx, sc, metav1.CreateOptions{})
	if err != nil {
		client.Log().Errorf("create storageclass failed: %v", err)
		return nil, err
	}

	client.Log().Infof("storageclass %q created successfully", created.Name)
	return created, nil
}
