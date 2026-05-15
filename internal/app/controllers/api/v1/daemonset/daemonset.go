package daemonset

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/k8s/daemonset"
	"k8soperation/pkg/valid"
)

type KubeDaemonSetController struct{}

func NewKubeDaemonSetController() *KubeDaemonSetController {
	return &KubeDaemonSetController{}
}

// Create godoc
// @Summary 创建 DaemonSet（可选创建 Service）
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDaemonSetCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/create [post]
func (c *KubeDaemonSetController) Create(ctx *gin.Context) {
	// 创建请求体结构
	req := requests.NewKubeDaemonSetCreateRequest()

	// 响应封装器
	r := response.NewResponse(ctx)

	// 参数校验（通用 valid.Validate）
	if ok := valid.Validate(ctx, req, requests.ValidKubeDaemonSetCreateRequest); !ok {
		return
	}

	// 调用 Service 层
	svc := services.NewServices(ctx)
	ds, svcObj, err := svc.KubeDaemonSetCreate(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetCreate error", zap.Error(err))
		return
	}

	// 成功响应（使用 daemonset 包的构建函数）
	r.Success(daemonset.BuildDaemonSetResponse(ds, svcObj, req))
}

// List godoc
// @Summary 获取 K8s DaemonSet 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "DaemonSet 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/daemonset/list [get]
func (c *KubeDaemonSetController) List(ctx *gin.Context) {
	// 1) 绑定与校验请求参数
	param := requests.NewKubeDaemonSetListRequest()

	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetListRequest); !ok {
		return
	}

	// 2) 调用 Service 层
	svc := services.NewServices(ctx)
	daemonsets, total, err := svc.KubeDaemonSetList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetList error", zap.Error(err))
		return
	}

	// 3) 返回统一列表响应
	r.SuccessList(daemonsets, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 DaemonSet 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 DaemonSet 详情
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/detail [get]
func (c *KubeDaemonSetController) Detail(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetDetailRequest); !ok {
		return
	}
	svc := services.NewServices(ctx)
	ds, err := svc.KubeDaemonSetDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetDetail error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message": "获取 DaemonSet 详情成功",
		"data":    ds,
	})
}

// Delete godoc
// @Summary 删除 DaemonSet
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/delete [delete]
func (c *KubeDaemonSetController) Delete(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetDeleteRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetDeleteRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	if err := svc.KubeDaemonSetDelete(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("service.KubeDaemonSetDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 成功返回
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet 删除成功",
	})
}

// DeleteService godoc
// @Summary 删除 DaemonSet 对应的 Service
// @Description 删除指定命名空间下，与 DaemonSet 同名的 Service 资源
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} response.Response "Service 删除成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "服务器内部错误"
// @Router /api/v1/k8s/daemonset/delete_service [delete]
func (c *KubeDaemonSetController) DeleteService(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetDeleteRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetDeleteRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	if err := svc.KubeDaemonSetDeleteService(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("service.KubeDaemonSetDeleteService error", zap.Error(err))
		ctx.Error(err)
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet Service 删除成功",
	})
}

// UpdateImage godoc
// @Summary 更新 DaemonSet 容器镜像
// @Description 修改指定命名空间下 DaemonSet 的容器镜像（支持滚动更新）
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDaemonSetUpdateImageRequest true "更新镜像参数"
// @Success 200 {object} string "更新成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "服务器内部错误"
// @Router /api/v1/k8s/daemonset/update_image [put]
func (c *KubeDaemonSetController) UpdateImage(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetUpdateImageRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetUpdateImageRequest); !ok {
		return
	}
	svc := services.NewServices(ctx)
	ds, err := svc.KubeDaemonSetUpdateImage(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetUpdateImage error", zap.Error(err))
		return

	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "更新 DaemonSet 镜像成功",
		"data":      ds,
	})
}

// Restart godoc
// @Summary 重启 DaemonSet
// @Description 触发指定命名空间下 DaemonSet 的滚动重启（等价于 kubectl rollout restart ds <name>）
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "DaemonSet 重启成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/restart [post]
func (c *KubeDaemonSetController) Restart(context *gin.Context) {
	param := requests.NewKubeDaemonSetRestartRequest()
	r := response.NewResponse(context)
	if ok := valid.Validate(context, param, requests.ValidKubeDaemonSetRestartRequest); !ok {
		return
	}
	svc := services.NewServices(context)
	err := svc.KubeDaemonSetRestart(context.Request.Context(), param)
	if err != nil {
		context.Error(err)
		global.Logger.Error("service.KubeDaemonSetRestart error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet 重启成功",
	})
}

// Rollback godoc
// @Summary 回滚 DaemonSet
// @Description 将 DaemonSet 回滚到指定的历史版本（ControllerRevision）。不传可在服务端实现“回滚到上一个版本”的兜底策略（可选）。
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param request body requests.KubeDaemonSetRollbackRequest true "回滚参数（namespace、name、revision_name）"
// @Success 200 {object} string "DaemonSet 回滚成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/rollback [post]
func (c *KubeDaemonSetController) Rollback(context *gin.Context) {
	param := requests.NewKubeDaemonSetRollbackRequest()
	r := response.NewResponse(context)
	if ok := valid.Validate(context, param, requests.ValidKubeDaemonSetRollbackRequest); !ok {
		return
	}
	svc := services.NewServices(context)
	_, err := svc.KubeDaemonSetRollback(context.Request.Context(), param)
	if err != nil {
		context.Error(err)
		global.Logger.Error("service.KubeDaemonSetRollback error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet 回滚成功",
	})
}
