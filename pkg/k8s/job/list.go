package job

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
	"k8soperation/pkg/k8s/dataselect"
	"time"
)

func GetJobList(client *k8s.Client, ctx context.Context, name, namespace string, page, limit int) ([]batchv1.Job, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	list, err := client.Interface.BatchV1().
		Jobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	selector := dataselect.NewJobSelector(list.Items, name, page, limit)
	selector.Filter().Sort()
	total := selector.TotalCount()
	data := selector.Paginate()

	return dataselect.FromJobCells(data.GenericDataList), total, nil
}
