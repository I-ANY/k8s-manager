package cronjob

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/pkg/k8s/dataselect"
)

func GetCronJobList(ctx context.Context, name, namespace string, page, limit int) ([]batchv1.CronJob, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	list, err := global.KubeClient.BatchV1().
		CronJobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	//使用你自己的 DataSelector 管道
	selector := dataselect.NewCronJobSelector(list.Items, name, page, limit)
	selector.Filter().Sort()
	total := selector.TotalCount()
	pageData := selector.Paginate()

	// 还原为 CronJob 列表返回
	return dataselect.FromCronJobCells(pageData.GenericDataList), total, nil
}
