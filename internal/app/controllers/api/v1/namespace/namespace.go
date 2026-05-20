package namespace

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	appctx "k8soperation/pkg/app"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type KubeNamespaceController struct {
}

func NewKubeNamespaceController() *KubeNamespaceController {
	return &KubeNamespaceController{}
}

// @Summary     创建 Namespace
// @Description 创建命名空间，并可选设置 labels/annotations 和资源配额（CPU/Memory/Pods）
// @Tags        K8s Namespace 管理
// @Accept      json
// @Produce     json
// @Param       body body requests.KubeNamespaceCreateRequest true "Namespace 创建参数"
// @Success     200 {object} response.Response
// @Failure     400 {object} errorcode.Error
// @Failure     500 {object} errorcode.Error
// @Router      /api/v1/k8s/namespace/create [post]
func (ctl *KubeNamespaceController) Create(ctx *gin.Context) {
	// 标准响应对象
	r := response.NewResponse(ctx)

	// DTO
	req := requests.NewKubeNamespaceCreateRequest()

	// 参数绑定+校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeNamespaceCreateRequest); !ok {
		return
	}

	// Service 层
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)

	// 调用 Service 创建 Namespace
	ns, err := svc.KubeCreateNamespace(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeCreateNamespace error", zap.Error(err))
		return
	}

	// 返回创建结果
	r.Success(gin.H{
		"name":        ns.Name,
		"labels":      ns.Labels,
		"annotations": ns.Annotations,
		"status":      ns.Status.Phase,
		"createdAt":   ns.CreationTimestamp,
	})
}

// @Summary 获取 Namespace 列表
// @Description 支持分页、名称模糊查询（Namespace 属于集群级资源，无需 namespace 参数）
// @Tags K8s Namespace 管理
// @Produce json
// @Param name  query string false "命名空间名称(模糊匹配)" maxlength(100)
// @Param page  query int    true  "页码(从1开始)"
// @Param limit query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/namespace/list [get]
func (ctl *KubeNamespaceController) List(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNamespaceListRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceListRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	items, total, err := svc.KubeNamespaceList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNamespaceList error", zap.Error(err))
		return
	}

	r.SuccessList(items, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Namespace 列表成功，共 %d 条", total),
	})
}

// Detail godoc
// @Summary 获取 Namespace 详情
// @Description 查询指定 Namespace 的详细信息
// @Tags K8s Namespace 管理
// @Produce json
// @Param name query string true "Namespace 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/namespace/detail [get]
func (c *KubeNamespaceController) Detail(ctx *gin.Context) {
	param := requests.NewKubeNamespaceDetailRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceDetailRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	nsDetail, err := svc.KubeNamespaceDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNamespaceDetail error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":     fmt.Sprintf("获取 Namespace %s 详情成功", param.Name),
		"name":        nsDetail.Name,
		"status":      nsDetail.Status.Phase,
		"labels":      nsDetail.Labels,
		"annotations": nsDetail.Annotations,
		"created_at":  nsDetail.CreationTimestamp,
	})
}

// @Summary 删除 Namespace
// @Description 删除指定 Namespace（级联删除内部资源）
// @Tags K8s Namespace 管理
// @Produce json
// @Param name query string true "Namespace 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "删除失败"
// @Router /api/v1/k8s/namespace/delete [delete]
func (ctl *KubeNamespaceController) Delete(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNamespaceDeleteRequest()

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceDeleteRequest); !ok {
		return
	}

	// 调用 Service
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	if err := svc.KubeNamespaceDelete(ctx.Request.Context(), param); err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeNamespaceDelete error", zap.Error(err))
		return
	}

	// 返回成功
	r.Success(gin.H{
		"message": fmt.Sprintf("Namespace %s 删除成功", param.Name),
	})
}

// Patch godoc
// @Summary 修改 Namespace（labels / annotations）
// @Description 支持新增、更新、删除 labels 与 annotations
// @Tags K8s Namespace 管理
// @Produce json
// @Param name query string true "Namespace 名称"
// @Param patch body requests.KubeNamespaceUpdateRequest true "Patch 内容"
// @Success 200 {object} response.Response "修改成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/namespace/patch [patch]
func (c *KubeNamespaceController) Patch(ctx *gin.Context) {
	param := requests.NewKubeNamespaceUpdateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceUpdateRequest); !ok {
		return
	}
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	updated, err := svc.KubeNamespaceUpdate(ctx, param)
	if err != nil {
		ctx.Error(err)
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("Namespace %s 修改成功", param.Name),
		"data":    updated,
	})
}
