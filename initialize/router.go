// Package initialize 用于系统初始化相关逻辑（配置、路由、Swagger 等）
package initialize

import (
	"github.com/gin-gonic/gin"
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
	"k8soperation/pkg/k8s/k8s_cluster"

	// swagger-ui 静态文件（index.html + JS/CSS）
	swaggerFiles "github.com/swaggo/files"
	// gin-swagger 中间件（把 swagger-ui 注册为 Gin 的 handler）
	ginSwagger "github.com/swaggo/gin-swagger"

	// docs 包由 swag init 自动生成，包含 swagger.json / swagger.yaml / docs.go
	// 注意：必须用匿名导入（_），因为 init() 会在包加载时自动注册 SwaggerSpec
	_ "k8soperation/docs"

	// 项目内全局配置
	"k8soperation/global"
	// 项目业务路由
	"k8soperation/internal/app/routers"
)

type injector interface {
	Inject(router *gin.RouterGroup)
}

func (s *Engine) injectRouterGroup(root *gin.RouterGroup) {
	// Swagger（公共）
	root.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := root.Group("/api")
	v1 := api.Group("/v1")

	// 1) 公共分组：不鉴权（登录、hello、debug等）
	public := v1.Group("")
	publicRouters := []injector{
		routers.NewHelloWorldRouter(),
		routers.NewAuthLogoutRouter(),
		routers.NewRegistryUserRouter(),
		routers.NewAuthRouter(), // 如果有登录/刷新接口
	}
	if global.ServerSetting.RunMode != "release" {
		publicRouters = append(publicRouters, routers.NewDebugRouter())
	}
	for _, r := range publicRouters {
		r.Inject(public) // ✅ 只注入到 public
	}

	// 2) 受保护分组：需要 JWT
	auth := v1.Group("")
	auth.Use(middlewares.AuthJWT())
	protectedRouters := []injector{
		routers.NewUserRouterV1(), // ✅ 只注入到 auth（不要再放到 public）

		// 其他需要鉴权的路由……
	}
	for _, r := range protectedRouters {
		r.Inject(auth) // ✅ 用对切片变量
	}
	debug := v1.Group("")
	debug.Use(middlewares.AuthJWT()) // 需要携带 Authorization: Bearer <jwt>
	debugRouters := []injector{
		routers.NewDebugSessionRouter(), // ← 注意是 New
	}
	for _, r := range debugRouters {
		r.Inject(debug)
	}

	// k8s 集群路由
	k8s := v1.Group("/k8s")
	//k8s.Use(middlewares.AuthJWT())
	k8sRouters := []injector{
		k8s_cluster.NewK8sRouter(),
	}
	for _, r := range k8sRouters {
		r.Inject(k8s)
	}

	// k8s kube_pod 集群路由
	pod := k8s.Group("/pod")
	podRouters := []injector{
		kube_pod.NewkubePodRouter(),
	}

	for _, r := range podRouters {
		r.Inject(pod)
	}

	// k8s kube_deployment 集群路由
	deployment := k8s.Group("/deployment")
	deploymentRouters := []injector{
		kube_deployment.NewKubeDeploymentRouter(),
	}

	for _, r := range deploymentRouters {
		r.Inject(deployment)
	}

	// k8s kube_statefulset 集群路由
	sts := k8s.Group("/statefulset")
	StatefulsetRouters := []injector{
		kube_statefulset.NewKubeStatefulSetmentRouter(),
	}

	for _, r := range StatefulsetRouters {
		r.Inject(sts)
	}

	// k8s kube_daemonset 集群路由
	ds := k8s.Group("/daemonset")
	DaemonSetRouters := []injector{
		kube_daemonset.NewKubeDaemonSetRouter(),
	}

	for _, r := range DaemonSetRouters {
		r.Inject(ds)
	}

	// k8s kube_job 集群路由
	jb := k8s.Group("/job")
	JobRouters := []injector{
		kube_job.NewKubeJobRouter(),
	}

	for _, r := range JobRouters {
		r.Inject(jb)
	}

	// k8s kube_cronjob 集群路由
	cj := k8s.Group("/cronjob")
	CronJobRouters := []injector{
		kube_cronjob.NewKubeCronJobRouter(),
	}

	for _, r := range CronJobRouters {
		r.Inject(cj)
	}

	// k8s kube_service 集群路由
	svc := k8s.Group("/service")
	ServiceRouters := []injector{
		kube_service.NewKubeServiceRouter(),
	}

	for _, r := range ServiceRouters {
		r.Inject(svc)
	}

	// k8s kube_ingress 集群路由
	ingress := k8s.Group("/ingress")
	IngressRouters := []injector{
		kube_ingress.NewKubeIngressRouter(),
	}

	for _, r := range IngressRouters {
		r.Inject(ingress)
	}

	// k8s kube_secret 集群路由
	secret := k8s.Group("/secret")
	SecretRouters := []injector{
		kube_secret.NewKubeSecretRouter(),
	}

	for _, r := range SecretRouters {
		r.Inject(secret)
	}

	// k8s kube_configmap 集群路由
	configmap := k8s.Group("/configmap")
	ConfigMapRouters := []injector{
		kube_configmap.NewKubeConfigMapRouter(),
	}

	for _, r := range ConfigMapRouters {
		r.Inject(configmap)
	}

	// k8s kube_storageclass 集群路由
	storageclass := k8s.Group("/storageclass")
	StorageClassRouters := []injector{
		kube_storageclass.NewKubeStorageClassRouter(),
	}

	for _, r := range StorageClassRouters {
		r.Inject(storageclass)
	}

	// k8s kube_pv 集群路由
	pv := k8s.Group("/pv")
	PVRouters := []injector{
		kube_pv.NewKubePersistentVolumeRouter(),
	}

	for _, r := range PVRouters {
		r.Inject(pv)
	}

	// k8s kube_pvc 集群路由
	pvc := k8s.Group("/pvc")
	PVCRouters := []injector{
		kube_pvc.NewKubePersistentVolumeClaimRouter(),
	}

	for _, r := range PVCRouters {
		r.Inject(pvc)
	}

	// k8s kube_pvc 集群路由
	node := k8s.Group("/node")
	NodeRouters := []injector{
		kube_node.NewKubeNodeRouter(),
	}

	for _, r := range NodeRouters {
		r.Inject(node)
	}

	// k8s kube_naemspace 集群路由
	namespace := k8s.Group("/namespace")
	NamespaceRouters := []injector{
		kube_namespace.NewKubeNamespaceRouter(),
	}

	for _, r := range NamespaceRouters {
		r.Inject(namespace)
	}
}
