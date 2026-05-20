package services

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/cronjob"
	"time"
)

// KubeCronJobCreate 创建 CronJob
func (s *Services) KubeCronJobCreate(ctx context.Context, req *requests.KubeCronJobCreateRequest) (*batchv1.CronJob, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cronJobObj, err := cronjob.CreateCronJob(s.App().K8sClient(), ctx, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			s.App().Logger.Warnf("cronjob %s/%s already exists", req.Namespace, req.Name)
			return nil, fmt.Errorf("cronjob %q already exists in namespace %q", req.Name, req.Namespace)
		}
		return nil, fmt.Errorf("create cronjob failed: %w", err)
	}

	s.App().Logger.Infof("cronjob %s/%s created successfully", req.Namespace, cronJobObj.Name)
	return cronJobObj, nil
}

// KubeCronJobList 获取 CronJob 列表
func (s *Services) KubeCronJobList(ctx context.Context, param *requests.KubeCronJobListRequest) ([]batchv1.CronJob, int, error) {
	return cronjob.GetCronJobList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

// KubeCronJobDetail 获取 CronJob 详情
func (s *Services) KubeCronJobDetail(ctx context.Context, param *requests.KubeCronJobDetailRequest) (
	*batchv1.CronJob, []batchv1.Job, error) {
	return cronjob.GetCronJobDetail(ctx, param.Name, param.Namespace)
}

// KubeCronJobDelete 删除 cronjob 资源
func (s *Services) KubeCronJobDelete(ctx context.Context, param *requests.KubeCronJobDeleteRequest) error {
	return cronjob.DeleteCronJob(ctx, param.Name, param.Namespace)
}

// KubeCronJobSuspend 暂停和恢复
func (s *Services) KubeCronJobSuspend(ctx context.Context, param *requests.KubeCronJobSuspendRequest) error {
	return cronjob.SetCronJobSuspend(s.App().K8sClient(), ctx, param.Namespace, param.Name, param.Suspend)
}
