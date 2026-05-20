package initialize

import (
	"fmt"

	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8soperation/pkg/app"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AppConfigScheme = runtime.NewScheme()

func init() {
	_ = appv1alpha1.AddToScheme(AppConfigScheme)
}

func NewAppConfigRuntimeClient(cfg *rest.Config) (client.Client, error) {
	return client.New(cfg, client.Options{
		Scheme: AppConfigScheme,
	})
}

func SetupAppConfigClient(a *app.App) error {
	if a.KubeConfig == nil {
		return fmt.Errorf("KubeConfig is nil, should run SetupK8sBootstrap first")
	}

	cli, err := NewAppConfigRuntimeClient(a.KubeConfig)
	if err != nil {
		return fmt.Errorf("init AppConfig runtime client failed: %w", err)
	}

	a.AppConfigClient = cli
	return nil
}
