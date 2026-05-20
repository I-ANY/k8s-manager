package k8s

import (
	"k8soperation/pkg/logger"
	"k8soperation/pkg/setting/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

var fallbackLogger = logger.NewBootstrapLogger(logger.AddCaller(), logger.AddCallerSkip(1))

// Client holds all dependencies needed by k8s operation functions,
// replacing the old global package.
type Client struct {
	Interface     kubernetes.Interface
	Logger        *logger.Logger
	PodLogSetting *types.PodLogSetting

	RestConfig       *rest.Config
	MetricsClient    *metricsclient.Clientset
	SupportsEventsV1 bool
}

func (c *Client) Log() *logger.Logger {
	if c == nil || c.Logger == nil {
		return fallbackLogger
	}
	return c.Logger
}
