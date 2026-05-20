package pod

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
)

// GetPodDetail 获取Pod详情
// pkg/k8s/kube_pod/kube_pod.go
func GetPodDetail(client *k8s.Client, namespace, name string) (*corev1.Pod, error) {
	pod, err := client.Interface.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		client.Logger.Errorf("Failed to get kube_pod detail: %v", err)
		return nil, err
	}
	client.Logger.Infof("Pod detail: %v", pod)
	return pod, nil
}
