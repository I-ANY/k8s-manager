package services

import (
	"context"

	"k8soperation/internal/app/dao"
	"k8soperation/pkg/app"
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
