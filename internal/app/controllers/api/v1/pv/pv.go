package pv

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type KubePVController struct{}

func NewKubePVController() *KubePVController {
	return &KubePVController{}
}

// @Summary     创建 PersistentVolume
// @Description 创建 PV（支持 HostPath / NFS）
// @Tags        K8s PV 管理
// @Accept      json
// @Produce     json
// @Param       body  body  requests.KubePVCreateRequest  true  "PV 创建参数"
// @Success     200   {object} response.Response
// @Failure     400   {object} errorcode.Error
// @Failure     500   {object} errorcode.Error
// @Router      /api/v1/k8s/pv/create [post]
func (ctl *KubePVController) Create(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	req := requests.NewKubePVCreateRequest()

	if ok := valid.Validate(ctx, req, requests.ValidKubePVCreateRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	pv, err := svc.KubeCreatePV(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCreatePV error", zap.Error(err))
		return
	}

	quantity := pv.Spec.Capacity[corev1.ResourceStorage]
	r.Success(gin.H{
		"name":         pv.Name,
		"capacity":     quantity.String(),
		"accessModes":  pv.Spec.AccessModes,
		"reclaim":      pv.Spec.PersistentVolumeReclaimPolicy,
		"storageClass": pv.Spec.StorageClassName,
		"volumeMode":   pv.Spec.VolumeMode,
		"created_at":   pv.CreationTimestamp,
	})
}

// @Summary 获取 PV 列表
// @Description 支持分页、名称模糊查询（PV 为集群级资源，无 namespace 参数）
// @Tags K8s PV 管理
// @Produce json
// @Param name  query string false "PV 名称(模糊匹配)" maxlength(100)
// @Param page  query int    true  "页码(从1开始)"
// @Param limit query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/pv/list [get]
func (ctl *KubePVController) List(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubePVListRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubePVListRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	items, total, err := svc.KubePVList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubePVList error", zap.Error(err))
		return
	}

	// 直接返回原生对象或做精简映射都可
	r.SuccessList(items, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 PV 列表成功，共 %d 条", total),
	})
}

// Detail godoc
// @Summary 获取 PersistentVolume 详情
// @Tags K8s PV 管理
// @Produce json
// @Param name query string true "PersistentVolume 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pv/detail [get]
func (c *KubePVController) Detail(ctx *gin.Context) {
	// 1构造参数结构体
	param := requests.NewKubePVDetailRequest()

	// 2 响应封装器
	r := response.NewResponse(ctx)

	// 3 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubePVDetailRequest); !ok {
		return
	}

	// 4 调用 Service
	svc := services.NewServices(ctx)
	pvDetail, err := svc.KubePVDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubePVDetail error", zap.Error(err))
		return
	}

	// 5 返回结果
	r.Success(gin.H{
		"message":          fmt.Sprintf("获取 PersistentVolume %s 详情成功", param.Name),
		"name":             pvDetail.Name,
		"capacity":         pvDetail.Spec.Capacity,
		"accessModes":      pvDetail.Spec.AccessModes,
		"reclaimPolicy":    pvDetail.Spec.PersistentVolumeReclaimPolicy,
		"storageClassName": pvDetail.Spec.StorageClassName,
		"volumeMode":       pvDetail.Spec.VolumeMode,
		"status":           pvDetail.Status.Phase,
		"created_at":       pvDetail.CreationTimestamp,
	})
}

// @Summary 删除 PersistentVolume
// @Tags K8s PV 管理
// @Produce json
// @Param name query string true "PersistentVolume 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pv/delete [delete]
func (ctl *KubePVController) Delete(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubePVDeleteRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubePVDeleteRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	if err := svc.KubePVDelete(ctx.Request.Context(), param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubePVDelete error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("PersistentVolume %s 删除成功", param.Name),
	})
}

// Reclaim godoc
// @Summary 修改 PersistentVolume 回收策略
// @Tags K8s PV 管理
// @Produce json
// @Param name query string true "PV 名称"
// @Param reclaimPolicy body string true "回收策略 (Delete / Retain)"
// @Success 200 {object} string "修改成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pv/reclaim [patch]
func (c *KubePVController) Reclaim(ctx *gin.Context) {
	param := requests.NewKubePVReclaimRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePVReclaimRequest); !ok {
		return
	}

	svc := services.NewServices(ctx)
	updated, err := svc.KubePVReclaim(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubePVReclaim error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("PV %s 回收策略修改为 %s", param.Name, param.ReclaimPolicy),
		"data":    updated,
	})
}
