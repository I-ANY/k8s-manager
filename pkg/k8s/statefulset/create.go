package statefulset

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s"
)

func CreateStatefulSet(client *k8s.Client, ctx context.Context, req *requests.KubeStatefulSetCreateRequest) (*appv1.StatefulSet, error) {
	// 1) 构造 StatefulSet
	sts := BuildStatefulSetFromCreateReq(req)

	// 2) 创建 StatefulSet
	createdSts, err := client.Interface.AppsV1().
		StatefulSets(req.Namespace).
		Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("statefulset %q already exists in namespace %q", req.Name, req.Namespace)
		}
		client.Logger.Warnf("create statefulset failed: %v", err)
		return nil, err
	}

	client.Logger.Infof("statefulset %q created successfully", createdSts.Name)
	return createdSts, nil
}
