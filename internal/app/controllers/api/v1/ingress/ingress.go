package v1

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

type KubeIngressController struct{}

func NewKubeIngressController() *KubeIngressController {
	return &KubeIngressController{}
}

// Create godoc
// @Summary 创建 Ingress
// @Tags K8s Ingress 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeIngressCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/create [post]
func (c *KubeIngressController) Create(ctx *gin.Context) {
	req := requests.NewKubeIngressCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeIngressCreateRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)

	// 调用 Ingress 创建逻辑
	ing, err := svc.KubeIngressCreate(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeIngressCreate error", zap.Error(err))
		return
	}

	// 和 Job 一样封装一下返回；如果你有统一的 Build*Response，也可替换为 ingress.BuildIngressResponse(…)
	r.Success(gin.H{
		"message":   "创建 Ingress 成功",
		"name":      ing.Name,
		"namespace": ing.Namespace,
	})
}

// List godoc
// @Summary 获取 K8s Ingress 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s Ingress 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "Ingress 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/ingress/list [get]
func (c *KubeIngressController) List(ctx *gin.Context) {
	// 1 构造请求参数结构体
	param := requests.NewKubeIngressListRequest()

	// 2.创建响应器
	r := response.NewResponse(ctx)

	// 3 参数校验（valid.Validate 内部自动绑定、返回错误）
	if ok := valid.Validate(ctx, param, requests.ValidKubeIngressListRequest); !ok {
		return // 校验失败直接返回
	}

	// 4.调用 Service 层逻辑
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	ingresses, total, err := svc.KubeIngressList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeIngressList error", zap.Error(err))
		return
	}

	// 5. 统一返回格式
	r.SuccessList(ingresses, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Ingress 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 Ingress 详情
// @Tags K8s Ingress 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/detail [get]
func (c *KubeIngressController) Detail(ctx *gin.Context) {
	// 1. 构造请求参数
	param := requests.NewKubeIngressDetailRequest()
	r := response.NewResponse(ctx)

	// 2. 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeIngressDetailRequest); !ok {
		return
	}

	// 3. 调用 Service 层
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	ing, err := svc.KubeIngressDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeIngressDetail error", zap.Error(err))
		return
	}

	// 4. 返回成功响应
	r.Success(gin.H{
		"message": "获取 Ingress 详情成功",
		"data":    ing,
	})
}

// @Summary Patch Ingress（StrategicMergePatch）
// @Tags K8s Ingress 管理
// @Accept application/strategic-merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/patch [patch]
func (c *KubeIngressController) Patch(ctx *gin.Context) {
	param := requests.NewKubeIngressUpdateRequest() // 如果你有 NewKubeIngressUpdateRequest() 也可用它
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	out, err := svc.KubeIngressPatch(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("KubeIngressPatch error", zap.Error(err))
		return
	}
	r.Success(gin.H{"message": "Ingress StrategicMergePatch 成功", "data": out})
}

// @Summary Patch Ingress（JSON Merge Patch – 覆盖式）
// @Tags K8s Ingress 管理
// @Accept application/merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/patch-json [post]
func (c *KubeIngressController) PatchJSON(ctx *gin.Context) {
	param := requests.NewKubeIngressUpdateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	out, err := svc.KubeIngressPatchJSON(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("KubeIngressPatchJSON error", zap.Error(err))
		return
	}
	r.Success(gin.H{"message": "Ingress JSON Merge Patch 成功", "data": out})
}

// Delete godoc
// @Summary 删除 Ingress
// @Tags K8s Ingress 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/delete [delete]
func (c *KubeIngressController) Delete(ctx *gin.Context) {
	param := requests.NewKubeIngressDeleteRequest()
	r := response.NewResponse(ctx)

	// 参数校验（通用 Valid）
	if ok := valid.Validate(ctx, param, requests.ValidKubeIngressDeleteRequest); !ok {
		return
	}

	// 调用服务层
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	if err := svc.KubeIngressDelete(ctx.Request.Context(), param); err != nil {
		a.Logger.Error("service.KubeIngressDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "Ingress 删除成功",
	})
}
