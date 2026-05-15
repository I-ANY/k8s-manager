package services

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/job"
	"time"
)

// KubeJobCreate 仅创建 Job（不创建 Service）
func (s *Services) KubeJobCreate(ctx context.Context, req *requests.KubeJobCreateRequest) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	jobObj, err := job.CreateJob(ctx, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("job %s/%s already exists", req.Namespace, req.Name)
			return nil, fmt.Errorf("job %q already exists in namespace %q", req.Name, req.Namespace)
		}
		return nil, fmt.Errorf("create job failed: %w", err)
	}

	global.Logger.Infof("job %s/%s created successfully", req.Namespace, jobObj.Name)
	return jobObj, nil
}

// listJob 列出 Job
func (s *Services) KubeJobList(ctx context.Context, param *requests.KubeJobListRequest) ([]batchv1.Job, int, error) {
	return job.GetJobList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

// 获取 Job 详情
func (s *Services) KubeJobDetail(ctx context.Context, param *requests.KubeJobDetailRequest) (*batchv1.Job, error) {
	return job.GetJobDetail(ctx, param.Name, param.Namespace)
}

// KubeJobDelete 删除 Job
func (s *Services) KubeJobDelete(ctx context.Context, param *requests.KubeJobDeleteRequest) error {
	return job.DeleteJob(ctx, param.Name, param.Namespace)
}

// KubeSuspendJob 控制 Job 暂停或恢复
func (s *Services) KubeJobSuspend(ctx context.Context, param *requests.KubeJobSuspendRequest) error {
	return job.SetJobSuspend(ctx, param.Namespace, param.Name, param.Suspend)
}

// KubeJobRestart 重跑 Job（基于旧 Job 模板创建一个新名字的 Job）
func (s *Services) KubeJobRestart(ctx context.Context, param *requests.KubeJobRestartRequest) (*batchv1.Job, error) {
	return job.RestartJob(ctx, param.Namespace, param.Name)
}
