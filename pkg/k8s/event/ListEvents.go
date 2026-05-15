package event

import (
	"context"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/utils"
	"time"
)

func ListEvents(ctx context.Context, q *requests.KubeEventListRequest) (items []models.EventItem, next string, err error) {
	// 使用 context.WithTimeout 创建一个上下文，设置超时时间为30秒，并使用 defer 确保资源被正确释放
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 使用 utils.NormalizeNamespace 对命名空间进行规范化处理
	ns := utils.NormalizeNamespace(q.Namespace)

	// 优先新版，失败回退旧版
	if utils.TryEventsV1First() {
		if items, next, err = ListEventsV1(ctx, ns, q); err == nil {
			return items, next, nil
		}
	}
	return listEventsCoreV1(ctx, ns, q)
}
