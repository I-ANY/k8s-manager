package cluster

import (
	"gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var scheme = runtime.NewScheme()

func init() {
	_ = v1alpha1.AddToScheme(scheme)
}

// NewAppConfigClient 返回只支持 AppConfig 的 K8s Client
func NewAppConfigClient(cfg *rest.Config) (client.Client, error) {
	return client.New(cfg, client.Options{
		Scheme: scheme,
	})
}
