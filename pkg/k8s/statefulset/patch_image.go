package statefulset

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/pkg/k8s"
	"k8soperation/pkg/k8s/deployment/patchbuilder"
	"time"
)

func PatchImageStatefulSet(client *k8s.Client, ctx context.Context, namespace, name, container, image string) (*appv1.StatefulSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	patchBytes, err := patchbuilder.BuildImagePatch(container, image)
	if err != nil {
		return nil, err
	}

	sts, err := client.Interface.AppsV1().
		StatefulSets(namespace).
		Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 StatefulSet 镜像失败: %w", err)
	}
	return sts, nil
}
