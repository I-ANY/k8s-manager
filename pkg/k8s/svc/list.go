package svc

import (
	"context"
	"k8soperation/pkg/k8s"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s/dataselect"
)

// GetServiceList 与 GetDeploymentList 逻辑一致，返回 Service 列表 + 总数（分页前）
func GetServiceList(client *k8s.Client, ctx context.Context, name, namespace string, page, limit int) ([]v1.Service, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 拉取 Service 列表（可在这里加 LabelSelector / FieldSelector）
	list, err := client.Interface.CoreV1().
		Services(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	// 用 dataselect 构造选择器（注意：函数名以你实现为准）
	selector := dataselect.NewServiceSelector(list.Items, name, page, limit)

	// 过滤 + 排序（跟 Deployment 一样调用）
	selector.Filter().Sort()

	total := selector.TotalCount()

	data := selector.Paginate()

	// 转回 corev1.Service 切片返回
	return dataselect.FromServiceCells(data.GenericDataList), total, nil
}
