package daemonset

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s"
)

func CreateDaemonSet(client *k8s.Client, ctx context.Context, req *requests.KubeDaemonSetCreateRequest) (*appv1.DaemonSet, error) {
	// 1) 构造 DaemonSet（见下方 Build 函数）
	ds := BuildDaemonSetFromCreateReq(req)

	// 2) 创建 DaemonSet
	created, err := client.Interface.AppsV1().
		DaemonSets(req.Namespace).
		Create(ctx, ds, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("daemonset %q already exists in namespace %q", req.Name, req.Namespace)
		}
		client.Logger.Warnf("create daemonset failed: %v", err)
		return nil, err
	}

	client.Logger.Infof("daemonset %q created successfully", created.Name)
	return created, nil
}
