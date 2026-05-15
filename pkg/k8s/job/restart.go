package job

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"time"
)

func RestartJob(ctx context.Context, namespace, name string) (*batchv1.Job, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1. 获取旧 Job
	oldJob, err := global.KubeClient.BatchV1().Jobs(namespace).Get(c, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("job %s/%s not found", namespace, name)
		}
		return nil, fmt.Errorf("failed to get job %s/%s: %v", namespace, name, err)
	}

	// 2. 生成新 Job
	newName := fmt.Sprintf("%s-restart-%d", name, time.Now().Unix())

	// 去除了系统自动生成标签
	newJob := BuildJobFromOld(oldJob, newName)

	// 3. 创建新 Job
	created, err := global.KubeClient.BatchV1().Jobs(namespace).Create(c, newJob, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create new job: %v", err)
	}

	global.Logger.Infof("Job %s restarted as %s", name, created.Name)
	return created, nil
}
