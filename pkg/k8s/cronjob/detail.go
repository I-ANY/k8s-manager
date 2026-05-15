package cronjob

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8soperation/global"
)

func GetCronJobDetail(ctx context.Context, name, namespace string) (*batchv1.CronJob, []batchv1.Job, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 1. 获取 CronJob
	cj, err := global.KubeClient.BatchV1().CronJobs(namespace).Get(c, name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	// 2. 获取 CronJob 对应的 Jobs
	labelSelector := fmt.Sprintf("batch.kubernetes.io/cronjob-name=%s", name)
	jobList, err := global.KubeClient.BatchV1().Jobs(namespace).List(c, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return cj, nil, err
	}

	return cj, jobList.Items, nil
}
