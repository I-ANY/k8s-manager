package daemonset

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/pkg/k8s"
	"k8soperation/pkg/k8s/deployment/patchbuilder"
)

// PatchDaemonSet 使用 JSONPatch 或 MergePatch 修改 DaemonSet
func PatchDaemonSet(client *k8s.Client, ctx context.Context, namespace, name string, patchBytes []byte) (*appv1.DaemonSet, error) {
	// 调用 K8s API 执行 patch
	patched, err := client.Interface.AppsV1().
		DaemonSets(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		client.Logger.Errorf("patch daemonset %s/%s failed: %v", namespace, name, err)
		return nil, err
	}

	client.Logger.Infof("daemonset %s/%s image patched successfully", namespace, name)
	return patched, nil
}

// PatchDeploymentImage 修改容器镜像
func PatchUpdateDaemonSetImage(client *k8s.Client, ctx context.Context, namespace, name, containerName, image string) (*appv1.DaemonSet, error) {
	patchImage, err := patchbuilder.BuildImagePatch(containerName, image)
	if err != nil {
		return nil, err
	}
	return PatchDaemonSet(client, ctx, namespace, name, patchImage)
}
