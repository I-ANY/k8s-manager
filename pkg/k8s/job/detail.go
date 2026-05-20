package job

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.uber.org/zap"
	"k8soperation/pkg/k8s"
)

// GetJobDetail 获取 Job 详情
func GetJobDetail(client *k8s.Client, ctx context.Context, name, namespace string) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job, err := client.Interface.BatchV1().
		Jobs(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Logger.Error("job not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("job %s/%s not found", namespace, name)
		}

		client.Logger.Error("get job failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return job, nil
}
