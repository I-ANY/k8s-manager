package pod

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
)

// kube_pod/update.go
func UpdatePod(namespace, name string, content json.RawMessage) error {
	pod, err := global.KubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &pod.Spec); err != nil {
		return fmt.Errorf("反序列化 Pod.Spec 失败: %w", err)
	}

	_, err = global.KubeClient.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}
