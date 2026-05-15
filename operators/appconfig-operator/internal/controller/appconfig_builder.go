package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
)

func buildDeployment(app *appv1alpha1.AppConfig) *appsv1.Deployment {
	replicas := int32(1)
	if app.Spec.Replicas != nil {
		replicas = *app.Spec.Replicas
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       app.Spec.AppName,
		"app.kubernetes.io/instance":   app.Name,
		"app.kubernetes.io/managed-by": "appconfig-operator",
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name, // 可以用 app.Spec.AppName 或 app.Name，看你策略
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  app.Spec.AppName,
							Image: app.Spec.Image,
							Env:   buildEnv(app), // 你也可以写一个小 helper 函数
						},
					},
				},
			},
		},
	}
}

// buildEnv 将 AppConfig.Spec.Env 转换为 Kubernetes Deployment 所需的 corev1.EnvVar 列表
func buildEnv(app *appv1alpha1.AppConfig) []corev1.EnvVar {
	// AppConfig.Spec.Env 是用户输入，类型是你自己定义的 EnvVar
	// 但 Deployment 需要的是 corev1.EnvVar（K8s 标准类型）

	result := make([]corev1.EnvVar, 0, len(app.Spec.Env))

	// 遍历用户配置的所有环境变量
	for _, e := range app.Spec.Env {
		// 转换成 K8s 的 EnvVar
		result = append(result, corev1.EnvVar{
			Name:  e.Name,
			Value: e.Value,
		})
	}

	return result
}
