package services

import (
	"context"
	"k8soperation/global"
	"k8soperation/internal/app/dao"
)

type Services struct {
	ctx context.Context
	dao *dao.Dao
}

func NewServices(ctx context.Context) *Services {
	return &Services{
		ctx: ctx,
		dao: dao.NewDao(global.DB),
	}
}

// 启动期/后台任务使用（新增）
func NewBackgroundServices() *Services {
	return &Services{
		ctx: context.Background(),
		dao: dao.NewDao(global.DB),
	}
}
