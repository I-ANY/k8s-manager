package deployment

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
	"k8soperation/pkg/k8s/dataselect"
	"time"
)

func GetDeploymentList(client *k8s.Client, ctx context.Context, name, namespace string, page, limit int) ([]appv1.Deployment, int, error) {
	// 超时上下文，避免 List 长时间阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 拉取列表
	list, err := client.Interface.AppsV1().
		Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	// 构造选择器：名称过滤 + 排序 + 分页
	selector := dataselect.NewDeploymentSelector(list.Items, name, page, limit)

	// 先过滤 + 排序
	selector.Filter().Sort()

	// 记录过滤后的总数（分页前）
	total := selector.TotalCount()

	// 再分页
	data := selector.Paginate()

	// 转回原始类型并返回
	return dataselect.FromDeploymentCells(data.GenericDataList), total, nil
}
