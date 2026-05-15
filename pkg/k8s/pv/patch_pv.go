package pv

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

func ReclaimPersistentVolume(ctx context.Context, req *requests.KubePVReclaimRequest) (*corev1.PersistentVolume, error) {
	patchBytes, err := BuildReclaimPolicyPatch(req.ReclaimPolicy)
	if err != nil {
		return nil, err
	}

	updated, err := global.KubeClient.CoreV1().
		PersistentVolumes().
		Patch(ctx, req.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("patch reclaim policy failed: %w", err)
	}
	return updated, nil
}
