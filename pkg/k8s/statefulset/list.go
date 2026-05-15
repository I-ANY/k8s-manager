package statefulset

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/pkg/k8s/dataselect"
	"time"
)

func GetSatefulSetList(ctx context.Context, name, namespace string, page, limit int) ([]appv1.StatefulSet, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	list, err := global.KubeClient.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	selector := dataselect.NewStatefulSetSelector(list.Items, name, page, limit)

	selector.Filter().Sort()

	total := selector.TotalCount()

	data := selector.Paginate()

	return dataselect.FromStatefulSetCells(data.GenericDataList), total, nil

}
