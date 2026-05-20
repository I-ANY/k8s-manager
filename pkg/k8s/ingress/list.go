package ingress

import (
	"context"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
	"k8soperation/pkg/k8s/dataselect"
	"time"
)

// GetIngressList 拉取 Ingress 列表，支持分页、名称过滤
func GetIngressList(client *k8s.Client, ctx context.Context, name, namespace string, page, limit int) ([]networkingv1.Ingress, int, error) {
	// 超时上下文
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 拉取 ingress 列表
	list, err := client.Interface.NetworkingV1().
		Ingresses(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, 0, fmt.Errorf("namespace %s not found", namespace)
		}
		return nil, 0, err
	}

	// 构造选择器（名称过滤 + 排序 + 分页）
	selector := dataselect.NewIngressSelector(list.Items, name, page, limit)

	// 过滤 + 排序
	selector.Filter().Sort()

	// 记录过滤后的总数
	total := selector.TotalCount()

	// 分页
	data := selector.Paginate()

	// 转回原始类型并返回
	return dataselect.FromIngressCells(data.GenericDataList), total, nil
}
