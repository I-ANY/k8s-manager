package node

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/global"
	k8sclient "k8soperation/pkg/k8s"
)

// CordonNode 通过 Patch 设置 spec.unschedulable（true: cordon, false: uncordon）
func CordonNode(client *k8sclient.Client, ctx context.Context, name string, unschedulable bool) error {
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"unschedulable": unschedulable,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch body failed: %w", err)
	}

	_, err = global.KubeClient.CoreV1().
		Nodes().
		Patch(ctx,
			name,
			types.StrategicMergePatchType,
			patchBytes,
			metav1.PatchOptions{},
		)
	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Log().Error("Node not found when cordon",
				zap.String("name", name),
				zap.Bool("unschedulable", unschedulable),
			)
			return fmt.Errorf("Node %q not found", name)
		}
		client.Log().Error("patch Node unschedulable failed",
			zap.String("name", name),
			zap.Bool("unschedulable", unschedulable),
			zap.Error(err),
		)
		return err
	}

	return nil
}
