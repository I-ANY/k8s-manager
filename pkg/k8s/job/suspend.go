package job

import (
	"context"
	"encoding/json"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/global"
	"time"
)

// SetJobSuspend 设置 Job 的暂停状态（true=暂停，false=恢复）
func SetJobSuspend(ctx context.Context, namespace, name string, suspend bool) error {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1.构造要 patch 的数据
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"suspend": suspend, // 使用传入参数，而不是写死 true
		},
	}

	// 2序列化为 JSON
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return fmt.Errorf("failed to marshal patch data: %v", err)
	}

	// 3 发起 PATCH 请求（StrategicMergePatch）
	_, err = global.KubeClient.BatchV1().Jobs(namespace).Patch(
		c,
		name,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("Job %q not found in namespace %q", name, namespace)
		}
		return fmt.Errorf("failed to update Job suspend status: %v", err)
	}

	action := "suspended"
	if !suspend {
		action = "resumed"
	}
	global.Logger.Infof("Job %q in namespace %q %s successfully", name, namespace, action)
	return nil
}
