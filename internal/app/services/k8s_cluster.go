package services

import (
	"encoding/base64"
	"fmt"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/errorcode"
	"os"
	"strings"
	"time"
)

func (s *Services) K8sClusterCreate(param *requests.K8sClusterCreateRequest) error {
	return s.dao.K8sClusterCreate(param.ClusterName, param.ClusterVersion, param.KubeConfig)
}

func (s *Services) K8sClusterUpdate(param *requests.K8sClusterUpdateRequest) error {
	return s.dao.K8sClusterUpdate(param.ID, param.ClusterName, param.ClusterVersion, param.KubeConfig, param.Status)
}

func (s *Services) K8sClusterDelete(param *requests.K8sClusterDeleteRequest) error {
	return s.dao.K8sClusterDelete(param.ID)
}

func (s *Services) K8sClusterList(param *requests.K8sClusterListRequest) ([]*models.K8sCluster, error) {
	return s.dao.K8sClusterList(param.ClusterName, param.Page, param.Limit)
}

// 初始化K8s集群
// 需要的 import：
// import (
//     "encoding/base64"
//     "os"
//     "strings"
//     "go.uber.org/zap"
//     "k8s.io/client-go/kubernetes"
//     "k8s.io/client-go/rest"
//     "k8s.io/client-go/tools/clientcmd"
// )

// internal/app/services/k8s_init.go
func (s *Services) K8sClusterInit(param *requests.K8sClusterInitRequest) (*K8sClients, error) {
	s.App().Logger.Info("K8sClusterInit begin", zap.Uint32("cluster_id", param.ID))

	// 1) 尝试 DB kubeconfig（先按原 YAML，失败再试 base64）
	if cfg, ok := s.tryFromDB(param.ID); ok {
		return s.buildClients(cfg)
	}
	// 2) 全局 kubeconfig 路径
	if cfg, ok := s.tryFromGlobalPath(); ok {
		return s.buildClients(cfg)
	}
	// 3) InCluster
	if cfg, ok := s.tryFromInCluster(); ok {
		return s.buildClients(cfg)
	}

	s.App().Logger.Error("K8sClusterInit failed: no valid kubeconfig source",
		zap.Uint32("cluster_id", param.ID),
		zap.String("global_path", strings.TrimSpace(s.App().AppSetting.GlobalKubeConfigPath)),
	)
	return nil, errorcode.ErrorClusterInitFailed
}

func (s *Services) buildClients(cfg *rest.Config) (*K8sClients, error) {
	tuneRESTConfig(cfg) // 统一调优(QPS/Burst/Timeout/UserAgent)

	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kube client: %w", err)
	}

	// metrics 不作为硬依赖，失败仅告警
	var mc *metricsclient.Clientset
	if m, mErr := metricsclient.NewForConfig(cfg); mErr != nil {
		s.App().Logger.Warn("init MetricsClient failed", zap.Error(mErr))
	} else {
		mc = m
	}

	// 探测 events.k8s.io/v1
	supports := false
	if _, err := kube.Discovery().ServerResourcesForGroupVersion("events.k8s.io/v1"); err == nil {
		supports = true
	}

	return &K8sClients{
		Config:       cfg,
		Kube:         kube,
		Metrics:      mc,
		SupportsEvV1: supports,
	}, nil
}

func tuneRESTConfig(cfg *rest.Config) {
	cfg.UserAgent = "k8soperation/1.0"
	cfg.QPS = 50
	cfg.Burst = 100
	cfg.Timeout = 30 * time.Second
}

func (s *Services) tryFromDB(clusterID uint32) (*rest.Config, bool) {
	kc, err := s.dao.K8sClusterGetByID(clusterID)
	if err != nil {
		s.App().Logger.Warn("get cluster by id failed, fallback to next",
			zap.Uint32("cluster_id", clusterID), zap.Error(err))
		return nil, false
	}
	raw := strings.TrimSpace(kc.KubeConfig)
	if raw == "" {
		s.App().Logger.Warn("empty kubeconfig in DB", zap.Uint32("cluster_id", clusterID))
		return nil, false
	}

	// 先当作原 YAML 解析
	if cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(raw)); err == nil {
		s.App().Logger.Info("init from DB kubeconfig (plain YAML)")
		return cfg, true
	}
	// 再试 base64
	decoded, decErr := base64.StdEncoding.DecodeString(raw)
	if decErr != nil {
		s.App().Logger.Error("decode DB kubeconfig base64 failed",
			zap.Uint32("cluster_id", clusterID), zap.Error(decErr))
		return nil, false
	}
	if cfg, err := clientcmd.RESTConfigFromKubeConfig(decoded); err == nil {
		s.App().Logger.Info("init from DB kubeconfig (base64)")
		return cfg, true
	}
	return nil, false
}

func (s *Services) tryFromGlobalPath() (*rest.Config, bool) {
	p := strings.TrimSpace(s.App().AppSetting.GlobalKubeConfigPath)
	if p == "" {
		s.App().Logger.Warn("global kubeconfig path is empty in config")
		return nil, false
	}
	if _, err := os.Stat(p); err != nil {
		s.App().Logger.Warn("global kubeconfig path not found", zap.String("global_path", p), zap.Error(err))
		return nil, false
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		s.App().Logger.Error("parse global kubeconfig failed", zap.String("global_path", p), zap.Error(err))
		return nil, false
	}
	s.App().Logger.Info("init from global kubeconfig success", zap.String("global_path", p))
	return cfg, true
}

func (s *Services) tryFromInCluster() (*rest.Config, bool) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		s.App().Logger.Warn("in-cluster config not available", zap.Error(err))
		return nil, false
	}
	s.App().Logger.Info("init from in-cluster config success")
	return cfg, true
}
