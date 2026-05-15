package svc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app/response"
	svc_service "k8soperation/pkg/k8s/svc"
	"k8soperation/pkg/valid"
)

type KubeServiceController struct {
}

func NewKubeServiceController() *KubeServiceController {
	return &KubeServiceController{}
}

// @Summary     创建 Service
// @Description 独立创建 Kubernetes Service（支持 ClusterIP / NodePort / LoadBalancer / Headless）
// @Tags        K8s Service 管理
// @Accept      json
// @Produce     json
// @Param       body  body  requests.KubeServiceCreateRequest  true  "Service 创建参数"
// @Success     200   {object} response.Response "成功"
// @Failure     400   {object} errorcode.Error   "请求参数错误"
// @Failure     500   {object} errorcode.Error   "内部错误"
// @Router      /api/v1/k8s/service/create [post]
func (ctl *KubeServiceController) Create(ctx *gin.Context) {
	// 统一响应器
	r := response.NewResponse(ctx)

	// 构造请求参数并做统一校验（和 List 一样走 valid.Validate）
	req := requests.NewKubeServiceCreateRequest()
	if ok := valid.Validate(ctx, req, requests.ValidKubeServiceCreateRequest); !ok {
		return // 校验失败时，valid 已自动返回错误响应
	}

	// 调用 Services
	svc := services.NewServices(ctx)

	// 如果你项目里已经实现了 (s *Services) KubeCreateService：
	service, err := svc.KubeCreateService(ctx.Request.Context(), req)

	// 如果你用的是包级函数 services.CreateService，则替换为：
	// created, err := services.CreateService(ctx.Request.Context(), req)

	if err != nil {
		ctx.Error(err) // 交给全局中间件/Logger
		global.Logger.Error("service.KubeCreateService error", zap.Error(err))
		return
	}

	// 成功返回（和 List 一致用 r.Success / r.SuccessList）
	r.Success(gin.H{
		"name":        service.Name,              // Service 名称
		"namespace":   service.Namespace,         // 命名空间
		"type":        service.Spec.Type,         // ClusterIP / NodePort / LoadBalancer
		"cluster_ip":  service.Spec.ClusterIP,    // 集群 IP（Headless 时为 "None"）
		"external_ip": service.Spec.ExternalName, // ExternalName 类型时显示
		"ports":       service.Spec.Ports,        // 暴露的端口列表
		"selector":    service.Spec.Selector,     // Label 选择器
		"created_at":  service.CreationTimestamp, // 创建时间
	})
}

// List godoc
// @Summary 获取 K8s Service 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s Service 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "Service 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/service/list [get]
func (c *KubeServiceController) List(ctx *gin.Context) {
	// 构造请求参数结构体
	param := requests.NewKubeServiceListRequest()
	// 创建响应器
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeServiceListRequest); !ok {
		return // 校验失败时，valid 已自动返回错误响应
	}

	// 调用 Service 层
	svc := services.NewServices(ctx)
	servicesList, total, err := svc.KubeServiceList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeServiceList error", zap.Error(err))
		return
	}

	// 返回成功响应
	r.SuccessList(servicesList, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Service 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 Service 详情
// @Tags K8s Service 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Service 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/service/detail [get]
func (c *KubeServiceController) Detail(ctx *gin.Context) {
	// 构造请求参数
	param := requests.NewKubeServiceDetailRequest()

	// 构造统一响应器
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeServiceDetailRequest); !ok {
		return
	}

	// 调用业务逻辑层
	svc := services.NewServices(ctx)
	svcDetail, err := svc.KubeServiceDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeServiceDetail error", zap.Error(err))
		return
	}

	// 返回成功响应
	r.Success(gin.H{
		"message": "获取 Service 详情成功",
		"data":    svcDetail,
	})
}

// Delete godoc
// @Summary 删除 Service
// @Tags K8s Service 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Service 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/service/delete [delete]
func (c *KubeServiceController) Delete(ctx *gin.Context) {
	param := requests.NewKubeServiceDeleteRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeServiceDeleteRequest); !ok {
		return
	}

	// 调用 Service 层
	svc := services.NewServices(ctx)
	if err := svc.KubeServiceDelete(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("service.KubeServiceDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "删除成功",
	})
}

// @Summary Patch Service（StrategicMergePatch）
// @Tags K8s Service 管理
// @Accept application/strategic-merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Service 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/service/patch [patch]
func (c *KubeServiceController) Patch(ctx *gin.Context) {
	param := requests.KubeServiceUpdateRequest{}
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}
	svc := services.NewServices(ctx)
	out, err := svc.KubeServicePatch(ctx.Request.Context(), &param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeServicePatch error", zap.Error(err))
		return
	}
	r.Success(gin.H{"message": "Service StrategicMergePatch 成功", "data": out})
}

// @Summary Patch Service（JSON Merge Patch – 覆盖式）
// @Tags K8s Service 管理
// @Accept application/merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Service 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/service/patch-json [post]
func (c *KubeServiceController) PatchJSON(ctx *gin.Context) {
	param := requests.NewKubeServiceUpdateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}
	svc := services.NewServices(ctx)
	out, err := svc.KubeServicePatchJSON(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeServicePatchJSON error", zap.Error(err))
		return
	}
	r.Success(gin.H{"message": "Service JSON Merge Patch 成功", "data": out})
}

// @Summary 获取 Service Endpoints（core/v1）
// @Tags K8s Service 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Service 名称"
// @Success 200 {object} string
// @Router /api/v1/k8s/service/endpoints [get]
func (c *KubeServiceController) GetEndpoints(ctx *gin.Context) {
	param := requests.NewKubeServiceEndpointsRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeServiceDetailRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	ep, err := svc.KubeServiceEndpoints(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		return
	}
	r.Success(gin.H{
		"message":   "获取 Endpoints 成功",
		"endpoints": svc_service.BuildSimpleEndpointList(ep),
	})
}
