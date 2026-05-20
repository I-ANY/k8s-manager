package cronjob

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	k8sclient "k8soperation/pkg/k8s"
)

func CreateCronJob(client *k8sclient.Client, ctx context.Context, req *requests.KubeCronJobCreateRequest) (*batchv1.CronJob, error) {
	// 1. 构建 CronJob 对象（建议写个 BuildCronJobFromCreateReq 辅助函数）
	cronJob := BuildCronJobFromCreateReq(req)

	// 2. 调用 K8s API 创建 CronJob
	created, err := global.KubeClient.BatchV1().
		CronJobs(req.Namespace).
		Create(ctx, cronJob, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("cronjob %q already exists in namespace %q", req.Name, req.Namespace)
		}
		client.Log().Warnf("create cronjob failed: %v", err)
		return nil, err
	}

	client.Log().Infof("cronjob %q created successfully", created.Name)
	return created, nil
}
