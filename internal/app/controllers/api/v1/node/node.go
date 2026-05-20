package node

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	appctx "k8soperation/pkg/app"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type KubeNodeController struct {
}

func NewKubeNodeController() *KubeNodeController {
	return &KubeNodeController{}
}

// @Summary 获取 Node 列表
// @Description 支持分页、名称模糊查询（Node 为集群级资源，无 namespace 参数）
// @Tags K8s Node 管理
// @Produce json
// @Param name  query string false "Node 名称(模糊匹配)" maxlength(100)
// @Param page  query int    true  "页码(从1开始)"
// @Param limit query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/node/list [get]
func (ctl *KubeNodeController) List(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNodeListRequest()

	// 若 valid.Validate 内部已完成 ShouldBindQuery，这里无需再绑定
	// 否则可手动绑定：_ = ctx.ShouldBindQuery(param)
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeListRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	items, total, err := svc.KubeNodeList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNodeList error", zap.Error(err))
		return
	}

	r.SuccessList(items, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Node 列表成功，共 %d 条", total),
	})
}

// Detail godoc
// @Summary 获取 Node 详情
// @Description 查询指定 Node 的详情（Node 为集群级资源，无 namespace 参数）
// @Tags K8s Node 管理
// @Produce json
// @Param name query string true "Node 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/detail [get]
func (c *KubeNodeController) Detail(ctx *gin.Context) {
	// 1) 构造参数
	param := requests.NewKubeNodeDetailRequest()

	// 2) 响应封装
	r := response.NewResponse(ctx)

	// 3) 参数校验（若 valid.Validate 不含绑定，请加上：_ = ctx.ShouldBindQuery(param)）
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeDetailRequest); !ok {
		return
	}

	// 4) 调用 Service
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	nodeObj, err := svc.KubeNodeDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNodeDetail error", zap.Error(err))
		return
	}

	// 5) 整理返回字段
	ready := "Unknown"
	for _, cond := range nodeObj.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				ready = "True"
			} else {
				ready = "False"
			}
			break
		}
	}

	r.Success(gin.H{
		"message":       fmt.Sprintf("获取 Node %s 详情成功", nodeObj.Name),
		"name":          nodeObj.Name,
		"labels":        nodeObj.Labels,
		"taints":        nodeObj.Spec.Taints,
		"unschedulable": nodeObj.Spec.Unschedulable,
		"capacity":      nodeObj.Status.Capacity,
		"allocatable":   nodeObj.Status.Allocatable,
		"addresses":     nodeObj.Status.Addresses,
		"ready":         ready,
		"created_at":    nodeObj.CreationTimestamp,
	})
}

// @Summary 获取指定 Node 上的 Pod 列表
// @Description 支持分页、名称模糊查询（Pod 属于命名空间级资源，但此处跨命名空间按 Node 过滤）
// @Tags K8s Node 管理
// @Produce json
// @Param nodeName query string true  "Node 名称"
// @Param name     query string false "Pod 名称(模糊匹配)" maxlength(100)
// @Param page     query int    true  "页码(从1开始)"
// @Param limit    query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/node/pods [get]
func (ctl *KubeNodeController) ListPods(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNodePodsRequest()

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodePodsRequest); !ok {
		return
	}

	// 调用 Service 层
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	items, err := svc.KubeNodePods(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNodePods error", zap.Error(err))
		return
	}

	// 成功响应
	r.SuccessList(items, gin.H{
		"total":   len(items),
		"message": fmt.Sprintf("获取 Node[%s] 上的 Pod 列表成功，共 %d 条", param.Name, len(items)),
	})
}

// @Summary 获取 Node 指标（CPU/内存使用率）
// @Description name 为空则返回全量节点指标；填写则返回指定节点。
// @Tags K8s Node 管理
// @Produce json
// @Param name query string false "Node 名称（为空=全局）"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/node/metrics [get]
func (ctl *KubeNodeController) Metrics(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := &requests.KubeNodeMetricsRequest{}

	// 参数绑定与校验
	if err := ctx.ShouldBindQuery(param); err != nil {
		return
	}
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeMetricsRequest); !ok {
		return
	}

	// 调用 service
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	items, total, err := svc.KubeNodeMetricsList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Errorf("获取 Node 指标失败")
		a.Logger.Error("service.KubeNodeMetricsList error", zap.Error(err))
		return
	}

	// 响应信息
	msg := "获取全量 Node 指标成功"
	if param.Name != "" {
		msg = fmt.Sprintf("获取 Node[%s] 指标成功", param.Name)
	}

	r.SuccessList(items, gin.H{
		"total":   total,
		"message": msg,
	})
}

// Cordon godoc
// @Summary 标记 Node 是否可调度（cordon / uncordon）
// @Description 通过设置 spec.unschedulable 实现 cordon / uncordon
// @Tags K8s Node 管理
// @Accept json
// @Produce json
// @Param data body requests.KubeNodeCordonRequest true "Node 调度控制参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/cordon [post]
func (c *KubeNodeController) Cordon(ctx *gin.Context) {
	// 1) 构造参数
	param := requests.NewKubeNodeCordonRequest()

	// 2) 响应封装
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeCordonRequest); !ok {
		return
	}

	// 4) 调用 Service
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	if err := svc.KubeNodeCordon(ctx.Request.Context(), param); err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNodeCordon error", zap.Error(err))
		return
	}

	// 5) 成功返回
	r.Success(gin.H{
		"message":       fmt.Sprintf("设置 Node %s unschedulable=%v 成功", param.NodeName, param.Unschedulable),
		"nodeName":      param.NodeName,
		"unschedulable": param.Unschedulable,
	})
}

// Drain godoc
// @Summary 驱逐节点上的可驱逐 Pod（drain）
// @Description cordon 节点并驱逐其上非 DaemonSet/非静态 Pod（维护/下线常用）
// @Tags K8s Node 管理
// @Accept json
// @Produce json
// @Param data body requests.KubeNodeUncordonRequest true "Node 驱逐参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/drain [post]
func (c *KubeNodeController) Drain(ctx *gin.Context) {
	// 1) 构造参数
	param := requests.NewKubeNodeUncordonRequest()

	// 2) 响应封装
	r := response.NewResponse(ctx)

	// 3) 绑定 + 校验
	_ = ctx.ShouldBindJSON(param)
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeDrainRequest); !ok {
		return
	}

	// 4) 调用 Service
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	if err := svc.KubeNodeDrain(ctx.Request.Context(), param); err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNodeDrain error", zap.Error(err))
		return
	}

	// 5) 成功返回
	r.Success(gin.H{
		"message":  fmt.Sprintf("节点 %s drain 成功（已 cordon 并驱逐可驱逐 Pod）", param.NodeName),
		"nodeName": param.NodeName,
	})
}

// Evict godoc
// @Summary 驱逐指定 Pod
// @Description 通过 Eviction API 驱逐某个 Pod（受 PDB 约束）
// @Tags K8s Pod 管理
// @Accept json
// @Produce json
// @Param data body requests.KubePodEvictRequest true "Pod 驱逐参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "Pod 不存在"
// @Failure 429 {object} errorcode.Error "被 PDB 限制"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/evict [post]
func (c *KubeNodeController) Evict(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubePodEvictRequest()

	// 2) 绑定 + 校验
	if ok := valid.Validate(ctx, param, requests.ValidKubePodEvictRequest); !ok {
		return
	}

	// 3) 调 Service
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	if err := svc.KubePodEvict(ctx.Request.Context(), param); err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubePodEvict error", zap.Error(err))
		return
	}

	// 4) 返回
	r.Success(gin.H{
		"message":   "Pod 驱逐成功",
		"namespace": param.Namespace,
		"podName":   param.PodName,
	})
}
