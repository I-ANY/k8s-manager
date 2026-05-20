package initialize

import (
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "k8soperation/docs"
	"k8soperation/internal/app/routers"
	"k8soperation/internal/app/routers/kube_configmap"
	"k8soperation/internal/app/routers/kube_cronjob"
	"k8soperation/internal/app/routers/kube_daemonset"
	"k8soperation/internal/app/routers/kube_deployment"
	"k8soperation/internal/app/routers/kube_ingress"
	"k8soperation/internal/app/routers/kube_job"
	"k8soperation/internal/app/routers/kube_namespace"
	"k8soperation/internal/app/routers/kube_node"
	"k8soperation/internal/app/routers/kube_pod"
	"k8soperation/internal/app/routers/kube_pv"
	"k8soperation/internal/app/routers/kube_pvc"
	"k8soperation/internal/app/routers/kube_secret"
	"k8soperation/internal/app/routers/kube_service"
	"k8soperation/internal/app/routers/kube_statefulset"
	"k8soperation/internal/app/routers/kube_storageclass"
	"k8soperation/middlewares"
	"k8soperation/pkg/app"
	"k8soperation/pkg/k8s/k8s_cluster"
)

type injector interface {
	Inject(router *gin.RouterGroup)
}

func (s *Engine) injectRouterGroup(root *gin.RouterGroup, a *app.App) {
	root.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := root.Group("/api")
	v1 := api.Group("/v1")

	// 1) public routes: no auth
	public := v1.Group("")
	publicRouters := []injector{
		routers.NewHelloWorldRouter(),
		routers.NewAuthLogoutRouter(),
		routers.NewRegistryUserRouter(),
		routers.NewAuthRouter(),
	}
	if a.ServerSetting.RunMode != "release" {
		publicRouters = append(publicRouters, routers.NewDebugRouter())
	}
	for _, r := range publicRouters {
		r.Inject(public)
	}

	// 2) protected routes: JWT auth
	auth := v1.Group("")
	auth.Use(middlewares.AuthJWT(a))
	protectedRouters := []injector{
		routers.NewUserRouterV1(),
	}
	for _, r := range protectedRouters {
		r.Inject(auth)
	}

	debug := v1.Group("")
	debug.Use(middlewares.AuthJWT())
	debugRouters := []injector{
		routers.NewDebugSessionRouter(),
	}
	for _, r := range debugRouters {
		r.Inject(debug)
	}

	// k8s cluster routes
	k8s := v1.Group("/k8s")
	k8sRouters := []injector{
		k8s_cluster.NewK8sRouter(),
	}
	for _, r := range k8sRouters {
		r.Inject(k8s)
	}

	// k8s sub-resource routes
	k8sSubs := []struct {
		path   string
		router injector
	}{
		{"/pod", kube_pod.NewkubePodRouter()},
		{"/deployment", kube_deployment.NewKubeDeploymentRouter()},
		{"/statefulset", kube_statefulset.NewKubeStatefulSetmentRouter()},
		{"/daemonset", kube_daemonset.NewKubeDaemonSetRouter()},
		{"/job", kube_job.NewKubeJobRouter()},
		{"/cronjob", kube_cronjob.NewKubeCronJobRouter()},
		{"/service", kube_service.NewKubeServiceRouter()},
		{"/ingress", kube_ingress.NewKubeIngressRouter()},
		{"/secret", kube_secret.NewKubeSecretRouter()},
		{"/configmap", kube_configmap.NewKubeConfigMapRouter()},
		{"/storageclass", kube_storageclass.NewKubeStorageClassRouter()},
		{"/pv", kube_pv.NewKubePersistentVolumeRouter()},
		{"/pvc", kube_pvc.NewKubePersistentVolumeClaimRouter()},
		{"/node", kube_node.NewKubeNodeRouter()},
		{"/namespace", kube_namespace.NewKubeNamespaceRouter()},
	}
	for _, sub := range k8sSubs {
		group := k8s.Group(sub.path)
		sub.router.Inject(group)
	}
}
