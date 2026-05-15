package cronjob

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/k8s/cronjob"
	"k8soperation/pkg/valid"
)

type KubeCronJobController struct{}

func NewKubeCronJobController() *KubeCronJobController {
	return &KubeCronJobController{}
}

// Create godoc
// @Summary 创建 CronJob
// @Tags K8s CronJob 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeCronJobCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/cronjob/create [post]
func (c *KubeCronJobController) Create(ctx *gin.Context) {
	req := requests.NewKubeCronJobCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeCronJobCreateRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)

	// 调用 CronJob 创建逻辑
	cronJobObj, err := svc.KubeCronJobCreate(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobCreate error", zap.Error(err))
		return
	}

	// 返回结果（BuildCronJobResponse 是你类似 job.BuildJobResponse 的构造函数）
	r.Success(cronjob.BuildCronJobResponse(cronJobObj, req))
}

// List godoc
// @Summary 获取 K8s CronJob 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s CronJob 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "CronJob 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cronjob/list [get]
func (c *KubeCronJobController) List(ctx *gin.Context) {
	// 1) 构造请求参数
	param := requests.NewKubeCronJobListRequest()

	// 2) 响应器
	r := response.NewResponse(ctx)

	// 3) 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeCronJobListRequest); !ok {
		return
	}

	// 4) 调用 Service 层
	svc := services.NewServices(ctx)
	cjs, total, err := svc.KubeCronJobList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobList error", zap.Error(err))
		return
	}

	// 5) 返回
	r.SuccessList(cjs, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 CronJob 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 CronJob 详情（含历史 Job 列表）
// @Tags K8s CronJob 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "CronJob 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cronjob/detail [get]
func (c *KubeCronJobController) Detail(ctx *gin.Context) {
	param := requests.NewKubeCronJobDetailRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeCronJobDetailRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	cj, jobs, err := svc.KubeCronJobDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobDetail error", zap.Error(err))
		return
	}

	// 构造统一返回格式
	jobSummaries := make([]gin.H, 0, len(jobs))
	for _, j := range jobs {
		phase := "Pending"
		switch {
		case j.Status.Succeeded > 0:
			phase = "Complete"
		case j.Status.Failed > 0:
			phase = "Failed"
		case j.Status.Active > 0:
			phase = "Running"
		}
		jobSummaries = append(jobSummaries, gin.H{
			"name":           j.Name,
			"startTime":      j.Status.StartTime,
			"completionTime": j.Status.CompletionTime,
			"active":         j.Status.Active,
			"succeeded":      j.Status.Succeeded,
			"failed":         j.Status.Failed,
			"phase":          phase,
		})
	}

	r.Success(gin.H{
		"message": "获取 CronJob 详情成功",
		"cronjob": gin.H{
			"name":              cj.Name,
			"namespace":         cj.Namespace,
			"schedule":          cj.Spec.Schedule,
			"suspend":           cj.Spec.Suspend != nil && *cj.Spec.Suspend,
			"concurrencyPolicy": string(cj.Spec.ConcurrencyPolicy),
			"lastScheduleTime":  cj.Status.LastScheduleTime,
			"lastSuccessfulTime": func() interface{} {
				if cj.Status.LastSuccessfulTime == nil {
					return nil
				}
				return cj.Status.LastSuccessfulTime.Time
			}(),
		},
		"jobs": jobSummaries,
	})
}

// Delete godoc
// @Summary 删除 CronJob
// @Tags K8s CronJob 管理
// @Produce json
// @Param body body requests.KubeCronJobDeleteRequest true "删除参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/cronjob/delete [delete]
func (c *KubeCronJobController) Delete(ctx *gin.Context) {
	req := requests.NewKubeCronJobDeleteRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, req, requests.ValidKubeCronJobDeleteRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	if err := svc.KubeCronJobDelete(ctx.Request.Context(), req); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobDelete error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("CronJob %s/%s 删除成功", req.Namespace, req.Name),
	})
}

// Suspend godoc
// @Summary 暂停或恢复 CronJob
// @Description 用于暂停或恢复指定命名空间下的 CronJob（通过设置 .spec.suspend 字段为 true/false）
// @Tags K8s CronJob 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeCronJobSuspendRequest true "CronJob 暂停/恢复请求体"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cronjob/suspend [patch]
func (c *KubeCronJobController) Suspend(ctx *gin.Context) {
	req := requests.NewKubeCronJobSuspendRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeCronJobSuspendRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)

	// 调用 Service 层逻辑
	if err := svc.KubeCronJobSuspend(ctx.Request.Context(), req); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobSuspend error", zap.Error(err))
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
		"message":   fmt.Sprintf("CronJob 已成功%s", action),
	})
}
