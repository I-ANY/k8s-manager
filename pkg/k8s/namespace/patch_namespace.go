package namespace

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/global"
)

func PatchNamespace(ctx context.Context, name string, patch []byte) (*corev1.Namespace, error) {
	return global.KubeClient.CoreV1().
		Namespaces().
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
}
