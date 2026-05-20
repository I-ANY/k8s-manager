package pod

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
	"k8soperation/pkg/k8s/cell"
	"k8soperation/pkg/k8s/dataselect"
)

// GetPodList 是一个获取Kubernetes Pod列表的函数
// 参数:
//
//	name: Pod名称（当前函数未使用此参数）
//	namespace: Pod所在的命名空间
//	page: 页码（当前函数未使用此参数）
//	limit: 每页限制数量（当前函数未使用此参数）
//
// 返回值:
//
//	[]corev1.Pod: Pod对象的切片
//	error: 错误信息，如果获取失败则返回错误

// GetPodList 是一个获取Pod列表的函数
// 参数:
//
//	name: Pod名称，用于过滤
//	namespace: Pod所在的命名空间
//	page: 页码，用于分页
//	limit: 每页显示的数量，用于分页
//
// 返回值:
//
//	[]corev1.Pod: Pod列表
//	error: 错误信息
func GetPodList(client *k8s.Client, ctx context.Context, name, namespace string, page, limit int) ([]corev1.Pod, error) {
	// 1) 列表
	list, err := client.Interface.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 2) 转成 []DataCell
	cells := dataselect.ToCells(list.Items, func(p corev1.Pod) dataselect.DataCell {
		return cell.PodCell(p) // ⭐ 把 Pod 适配为实现 DataCell 的 PodCell
	})

	// 3) 组装选择器（按你的 NewDataSelector 实现来）
	sel := dataselect.NewDataSelector(cells, name, limit, page)

	// 4) 过滤/排序/分页
	data := sel.Filter().Sort().Paginate()

	// 5) 再转回 []corev1.Pod 返回
	pods := dataselect.FromCells[corev1.Pod](data.GenericDataList, func(c dataselect.DataCell) corev1.Pod {
		// 将 cell.PodCell 类型的变量 c 转换为 corev1.Pod 类型并返回
		// 这是一个类型断言，假设 c 实际上就是 cell.PodCell 类型
		return corev1.Pod(c.(cell.PodCell))
	})

	return pods, nil
}
