package services

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/k8s/event"
	"strings"
	"time"

	appv1 "k8s.io/api/apps/v1"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/deployment"
)

// 列表
func (s *Services) KubeDeploymentList(ctx context.Context, param *requests.KubeDeploymentListRequest) ([]appv1.Deployment, int, error) {
	return deployment.GetDeploymentList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

// 删除
func (s *Services) KubeDeploymentDelete(ctx context.Context, param *requests.KubeDeploymentDeleteRequest) error {
	return deployment.DeleteDeployment(ctx, param.Name, param.Namespace)
}

// 删除 Service
func (s *Services) KubeDeploymentDeleteService(ctx context.Context, param *requests.KubeDeploymentDeleteRequest) error {
	return deployment.DeleteService(ctx, param.Name, param.Namespace)
}

// 扩缩容（改副本数）
func (s *Services) KubeUpdateDeploymentReplicas(ctx context.Context, param *requests.KubeDeploymentScaleRequest) (*appv1.Deployment, error) {
	return deployment.PatchDeploymentReplicas(ctx, param.Namespace, param.Name, param.ScaleNum)
}

// 更新镜像
func (s *Services) KubeUpdateDeploymentImage(ctx context.Context, param *requests.KubeDeploymentUpdateImageRequest) (*appv1.Deployment, error) {
	return deployment.PatchDeploymentImage(ctx, param.Namespace, param.Name, param.Container, param.Image)
}

// Patch 模板（content 一般是 JSON Patch / StrategicMergePatch）
// 如果你传的是字符串，转成 []byte 再下发
func (s *Services) KubeUpdateDeploymentTemplate(ctx context.Context, param *requests.KubeDeploymentUpdateRequest) (*appv1.Deployment, error) {
	return deployment.PatchDeployment(ctx, param.Namespace, param.Name, []byte(param.Content))
}

// 回滚到指定 RS —— 和你的 DTO 对应
func (s *Services) KubeDeploymentRollback(ctx context.Context, param *requests.KubeDeploymentRollbackRequest) (*appv1.Deployment, error) {
	return deployment.RollbackDeployment(param.Name, param.Namespace, param.ReplicaSet)
}

// 重启 Deployment
func (s *Services) KubeDeploymentRestart(ctx context.Context, param *requests.KubeDeploymentRestartRequest) error {
	return deployment.RestartDeployment(param.Namespace, param.Name)
}

// 获取 Deployment 详情
func (s *Services) KubeDeploymentDetail(ctx context.Context, param *requests.KubeDeploymentDetailRequest) (*appv1.Deployment, error) {
	return deployment.GetDeploymentDetail(param.Name, param.Namespace)
}

// 创建 Deployment
func (s *Services) KubeDeploymentCreate(ctx context.Context, req *requests.KubeDeploymentCreateRequest) (*appv1.Deployment, *corev1.Service, error) {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1) 创建 Deployment
	dp, err := deployment.CreateDeployment(ctx, req)
	if err != nil {
		// 用 %w 包装，便于上层中间件用 errors.Is / apierrors.* 继续识别
		return nil, nil, fmt.Errorf("create deployment failed: %w", err)
	}

	// 2) 按需创建 Service
	var svcObj *corev1.Service
	if req.IsCreateService {
		svcObj, err = deployment.CreateServiceFromDeployment(ctx, req)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				svcName := strings.TrimSpace(req.ServiceName)
				if svcName == "" {
					svcName = req.Name
				}
				if exist, gerr := global.KubeClient.CoreV1().
					Services(req.Namespace).
					Get(ctx, svcName, metav1.GetOptions{}); gerr == nil {
					global.Logger.Infof("service %s/%s already exists, reuse it", req.Namespace, svcName)
					return dp, exist, nil
				}
				// Get 失败才回滚
			}

			// 真失败 → 回滚 Deployment（带传播策略，清理 RS/Pods）
			pol := metav1.DeletePropagationForeground // 或 Background
			if delErr := global.KubeClient.AppsV1().
				Deployments(req.Namespace).
				Delete(ctx, dp.Name, metav1.DeleteOptions{PropagationPolicy: &pol}); delErr != nil {
				global.Logger.Errorf("rollback delete deployment %s/%s failed: %v", req.Namespace, dp.Name, delErr)
			}
			return nil, nil, fmt.Errorf("create service failed: %w", err)
		}

	}

	return dp, svcObj, nil
}

// 获取 Deployment 对应的 Pod 列表（原始 Pod 对象）
func (s *Services) KubeDeploymentGetPod(ctx context.Context, param *requests.KubeCommonRequest) ([]corev1.Pod, error) {
	return deployment.GetPodByDeployment(ctx, param.Namespace, param.Name)
}

// 获取 Deployment 对应的 事件
func (s *Services) KubeEventList(ctx context.Context, param *requests.KubeEventListRequest) ([]models.EventItem, string, error) {
	return event.ListEvents(ctx, param)
}
