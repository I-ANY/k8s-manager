package storageclass

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
)

// —— 辅助：从请求构造 SC ——

// 根据你的 requests.KubeStorageClassCreateRequest 组装 StorageClass
func buildStorageClassFromReq(req *requests.KubeStorageClassCreateRequest) (*storagev1.StorageClass, error) {
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Labels:      nil,
			Annotations: nil,
		},
		Provisioner:          req.Provisioner,
		Parameters:           req.Parameters,           // map[string]string
		AllowVolumeExpansion: req.AllowVolumeExpansion, // *bool
		MountOptions:         req.MountOptions,         // []string
	}

	if req.ReclaimPolicy != "" {
		p, err := parseReclaimPolicy(req.ReclaimPolicy)
		if err != nil {
			return nil, err
		}
		sc.ReclaimPolicy = &p
	}

	if req.VolumeBindingMode != "" {
		m, err := parseVolumeBindingMode(req.VolumeBindingMode)
		if err != nil {
			return nil, err
		}
		sc.VolumeBindingMode = &m
	}

	return sc, nil
}

// —— 字符串枚举转换 ——

// "Delete" / "Retain"
func parseReclaimPolicy(s string) (corev1.PersistentVolumeReclaimPolicy, error) {
	switch s {
	case "Delete", "delete":
		return corev1.PersistentVolumeReclaimDelete, nil
	case "Retain", "retain":
		return corev1.PersistentVolumeReclaimRetain, nil
	default:
		return "", fmt.Errorf("不支持的 ReclaimPolicy: %q（仅支持 Delete/Retain）", s)
	}
}

// "Immediate" / "WaitForFirstConsumer"
func parseVolumeBindingMode(s string) (storagev1.VolumeBindingMode, error) {
	switch s {
	case "Immediate", "immediate":
		return storagev1.VolumeBindingImmediate, nil
	case "WaitForFirstConsumer", "waitforfirstconsumer", "WaitForFirst":
		return storagev1.VolumeBindingWaitForFirstConsumer, nil
	default:
		return "", fmt.Errorf("不支持的 VolumeBindingMode: %q（仅支持 Immediate/WaitForFirstConsumer）", s)
	}
}
