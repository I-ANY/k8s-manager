package services

import (
	"context"

	"k8soperation/internal/app/dao"
	"k8soperation/pkg/app"
	"k8soperation/pkg/cluster"
	"k8soperation/pkg/k8s"
)

type Services struct {
	ctx context.Context
	dao *dao.Dao
	app *app.App
}

func NewServices(ctx context.Context, a *app.App) *Services {
	return &Services{
		ctx: ctx,
		dao: dao.NewDao(a.DB),
		app: a,
	}
}

func NewServicesWithApp(a *app.App) *Services {
	return &Services{
		ctx: context.Background(),
		dao: dao.NewDao(a.DB),
		app: a,
	}
}

func (s *Services) App() *app.App {
	return s.app
}

func (s *Services) K8sClient(ctx context.Context, clusterID uint32) (*k8s.Client, error) {
	return cluster.GetClient(ctx, s.app, clusterID)
}
