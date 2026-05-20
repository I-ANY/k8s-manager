package job

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s"
	"time"
)

func CreateJob(client *k8s.Client, ctx context.Context, req *requests.KubeJobCreateRequest) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	job := BuildJobFromCreateReq(req)

	created, err := client.Interface.BatchV1().
		Jobs(req.Namespace).
		Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("job %q already exists in namespace %q", req.Name, req.Namespace)
		}
		client.Logger.Warnf("create job failed: %v", err)
		return nil, err
	}

	client.Logger.Infof("job %q created successfully", created.Name)
	return created, nil
}
