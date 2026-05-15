package deployment

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"time"
)

// GetPodByDeployment 获取某个 Deployment 对应的所有 Pod
func GetPodByDeployment(ctx context.Context, namespace, deploymentName string) ([]corev1.Pod, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 读取 Deployment
	deploy, err := global.KubeClient.AppsV1().Deployments(namespace).
		Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("deployment %s/%s not found", namespace, deploymentName)
		}
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to get deployment %s/%s", namespace, deploymentName)
		}
		return nil, fmt.Errorf("failed to get deployment %s/%s: %v", namespace, deploymentName, err)
	}

	// 构造 LabelSelector
	if deploy.Spec.Selector == nil {
		return nil, fmt.Errorf("deployment %s/%s has no selector", namespace, deploymentName)
	}
	selector := metav1.FormatLabelSelector(deploy.Spec.Selector)

	// 获取匹配的 Pod 列表
	podList, err := global.KubeClient.CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to list pods in namespace %s: %v", namespace, err)
		}
		return nil, fmt.Errorf("failed to list pods for deployment %s/%s: %v", namespace, deploymentName, err)
	}

	if len(podList.Items) == 0 {
		global.Logger.Infof("no pods found for deployment %s/%s", namespace, deploymentName)
	}

	return podList.Items, nil
}
