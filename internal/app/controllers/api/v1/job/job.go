package job

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	appctx "k8soperation/pkg/app"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/k8s/job"
	"k8soperation/pkg/valid"
)

type KubeJobController struct {
}

func NewKubeJobController() *KubeJobController {
	return &KubeJobController{}
}

// Create godoc
// @Summary 创建 Job
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeJobCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/job/create [post]
func (c *KubeJobController) Create(ctx *gin.Context) {
	req := requests.NewKubeJobCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeJobCreateRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)

	// 调用 Job 创建逻辑（只有 Job，没有 Service）
	jobObj, err := svc.KubeJobCreate(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeJobCreate error", zap.Error(err))
		return
	}

	r.Success(job.BuildJobResponse(jobObj, req))
}

// List godoc
// @Summary 获取 K8s Job 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s Job 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "Job 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/list [get]
func (c *KubeJobController) List(ctx *gin.Context) {
	// 1) 构造请求参数结构体
	param := requests.NewKubeJobListRequest()

	// 2) 创建响应器
	r := response.NewResponse(ctx)

	// 3) 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeJobListRequest); !ok {
		return // 校验失败时 valid 已自动返回错误响应
	}

	// 4) 调用 Service 层
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	jobs, total, err := svc.KubeJobList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeJobList error", zap.Error(err))
		return
	}

	// 5) 返回成功响应（统一格式）
	r.SuccessList(jobs, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Job 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 Job 详情
// @Tags K8s Job 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Job 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/detail [get]
func (c *KubeJobController) Detail(ctx *gin.Context) {
	// 构造请求参数对象
	param := requests.NewKubeJobDetailRequest()
	r := response.NewResponse(ctx)

	// 参数校验（valid.Validate 内部负责 Bind + 校验 + 错误响应）
	if ok := valid.Validate(ctx, param, requests.ValidKubeJobDetailRequest); !ok {
		return
	}

	// 调用 Service 层逻辑
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	jobObj, err := svc.KubeJobDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeJobDetail error", zap.Error(err))
		return
	}

	// 成功返回
	r.Success(gin.H{
		"message": "获取 Job 详情成功",
		"data":    jobObj,
	})
}

// Delete godoc
// @Summary 删除 Job
// @Tags K8s Job 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Job 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/delete [delete]
func (c *KubeJobController) Delete(ctx *gin.Context) {
	// 参数解析
	param := requests.NewKubeJobDeleteRequest()
	r := response.NewResponse(ctx)

	// 参数校验（valid 内部自动 Bind + 校验 + 返回错误响应）
	if ok := valid.Validate(ctx, param, requests.ValidKubeJobDeleteRequest); !ok {
		return
	}

	// 调用 Service 层删除逻辑
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	if err := svc.KubeJobDelete(ctx.Request.Context(), param); err != nil {
		a.Logger.Error("service.KubeJobDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 删除成功响应
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "删除 Job 成功",
	})
}

// Suspend godoc
// @Summary 暂停或恢复 Job
// @Description 用于暂停或恢复指定命名空间下的 Job（通过设置 .spec.suspend 字段为 true/false）
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeJobSuspendRequest true "Job 暂停/恢复请求体"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/suspend [patch]
func (c *KubeJobController) Suspend(ctx *gin.Context) {
	req := requests.NewKubeJobSuspendRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeJobSuspendRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)

	// 调用 Service 层逻辑
	if err := svc.KubeJobSuspend(ctx.Request.Context(), req); err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeJobSuspend error", zap.Error(err))
		return
	}

	// 成功响应
	action := "暂停"
	if !req.Suspend {
		action = "恢复"
	}

	r.Success(gin.H{
		"namespace": req.Namespace,
		"name":      req.Name,
		"message":   fmt.Sprintf("Job 已成功%s", action),
	})
}

// Restart godoc
// @Summary 重启 Job
// @Description 基于已有 Job 模板重新创建一个新的 Job（清除状态、生成新名称）
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeJobRestartRequest true "Job 重启请求体"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/restart [post]
func (c *KubeJobController) Restart(ctx *gin.Context) {
	req := requests.NewKubeJobRestartRequest()
	r := response.NewResponse(ctx)

	// 参数校验（绑定 + 验证）
	if ok := valid.Validate(ctx, req, requests.ValidKubeJobRestartRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)

	// 调用 Service 层逻辑
	jobObj, err := svc.KubeJobRestart(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeJobRestart error", zap.Error(err))
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": req.Namespace,
		"name":      req.Name,
		"newJob":    jobObj.Name,
		"message":   fmt.Sprintf("Job %s 已成功重启为 %s", req.Name, jobObj.Name),
	})
}
