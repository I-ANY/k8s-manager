package daemonset

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/pkg/k8s/dataselect"
)

func GetDaemonSetList(ctx context.Context, name, namespace string, page, limit int) ([]appv1.DaemonSet, int, error) {
	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 拉取 DS 列表
	list, err := global.KubeClient.AppsV1().
		DaemonSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	// 名称过滤 + 排序 + 分页（复用你的 dataselect 体系）
	selector := dataselect.NewDaemonSetSelector(list.Items, name, page, limit)
	selector.Filter().Sort()
	total := selector.TotalCount()
	data := selector.Paginate()

	return dataselect.FromDaemonSetCells(data.GenericDataList), total, nil
}
