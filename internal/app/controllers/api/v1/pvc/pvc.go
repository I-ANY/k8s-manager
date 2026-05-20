package pvc

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

// KubePVCController 负责 PersistentVolumeClaim 的 CRUD 与常用操作
type KubePVCController struct{}

func NewKubePVCController() *KubePVCController { return &KubePVCController{} }

// @Summary     创建 PersistentVolumeClaim
// @Description 创建 PVC（支持指定 StorageClass / AccessModes / 容量等）
// @Tags        K8s PVC 管理
// @Accept      json
// @Produce     json
// @Param       body  body  requests.KubePVCCreateRequest  true  "PVC 创建参数"
// @Success     200   {object} response.Response
// @Failure     400   {object} errorcode.Error
// @Failure     500   {object} errorcode.Error
// @Router      /api/v1/k8s/pvc/create [post]
func (ctl *KubePVCController) Create(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	req := requests.NewKubePVCCreateRequest()

	// 参数校验（根据你的 valid 体系）
	if ok := valid.Validate(ctx, req, requests.ValidKubePVCCreateRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	pvc, err := svc.KubeCreatePVC(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubeCreatePVC error", zap.Error(err))
		return
	}

	quantity := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	r.Success(gin.H{
		"name":             pvc.Name,
		"namespace":        pvc.Namespace,
		"storageSize":      quantity.String(),
		"accessModes":      pvc.Spec.AccessModes,
		"storageClassName": pvc.Spec.StorageClassName,
		"volumeMode":       pvc.Spec.VolumeMode,
		"status":           pvc.Status.Phase,
		"created_at":       pvc.CreationTimestamp,
	})
}

// @Summary 获取 PVC 列表
// @Description 支持分页、名称模糊查询（PVC 属于命名空间级资源，需要 namespace 参数）
// @Tags K8s PVC 管理
// @Produce json
// @Param namespace query string true  "命名空间"
// @Param name      query string false "PVC 名称(模糊匹配)" maxlength(100)
// @Param page      query int    true  "页码(从1开始)"
// @Param limit     query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/pvc/list [get]
func (ctl *KubePVCController) List(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubePVCListRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubePVCListRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	items, total, err := svc.KubePVCList(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubePVCList error", zap.Error(err))
		return
	}

	r.SuccessList(items, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 PVC 列表成功，共 %d 条", total),
	})
}

// Detail godoc
// @Summary 获取 PersistentVolumeClaim 详情
// @Description 查询指定命名空间下的 PVC 详情
// @Tags K8s PVC 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "PersistentVolumeClaim 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pvc/detail [get]
func (c *KubePVCController) Detail(ctx *gin.Context) {
	// 1) 构造参数结构体
	param := requests.NewKubePVCDetailRequest()

	// 2) 响应封装器
	r := response.NewResponse(ctx)

	// 3) 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubePVCDetailRequest); !ok {
		return
	}

	// 4) 调用 Service
	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	pvcDetail, err := svc.KubePVCDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubePVCDetail error", zap.Error(err))
		return
	}

	// 5) 返回结果
	quantity := pvcDetail.Spec.Resources.Requests[corev1.ResourceStorage]
	r.Success(gin.H{
		"message":          fmt.Sprintf("获取 PersistentVolumeClaim %s/%s 详情成功", param.Namespace, param.Name),
		"name":             pvcDetail.Name,
		"namespace":        pvcDetail.Namespace,
		"storage":          quantity.String(),
		"accessModes":      pvcDetail.Spec.AccessModes,
		"storageClassName": pvcDetail.Spec.StorageClassName,
		"volumeMode":       pvcDetail.Spec.VolumeMode,
		"status":           pvcDetail.Status.Phase,
		"boundVolume":      pvcDetail.Spec.VolumeName,
		"created_at":       pvcDetail.CreationTimestamp,
	})
}

// @Summary 删除 PersistentVolumeClaim
// @Tags K8s PVC 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name      query string true "PVC 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pvc/delete [delete]
func (ctl *KubePVCController) Delete(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubePVCDeleteRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubePVCDeleteRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	if err := svc.KubePVCDelete(ctx.Request.Context(), param); err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubePVCDelete error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("PersistentVolumeClaim %s/%s 删除成功", param.Namespace, param.Name),
	})
}

// @Summary 扩容 PersistentVolumeClaim（仅支持增大 storage）
// @Description 将 PVC 的 spec.resources.requests.storage 扩大为指定值（需 StorageClass 允许扩容）
// @Tags K8s PVC 管理
// @Accept json
// @Produce json
// @Param body body requests.KubePVCResizeRequest true "扩容参数：namespace/name/storage(如 10Gi)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 403 {object} errorcode.Error "StorageClass 不允许扩容"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pvc/resize [patch]
func (ctl *KubePVCController) Resize(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubePVCResizeRequest()

	// 与你的风格一致：直接走 valid.Validate（内部若已含绑定逻辑即可；否则在这里先 ShouldBindJSON）
	// _ = ctx.ShouldBindJSON(param)
	if ok := valid.Validate(ctx, param, requests.ValidKubePVCResizeRequest); !ok {
		return
	}

	a := appctx.FromContext(ctx)
	svc := services.NewServices(ctx, a)
	pvcObj, err := svc.KubePVCResize(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		a.Logger.Error("service.KubePVCResize error", zap.Error(err))
		return
	}

	qty := pvcObj.Spec.Resources.Requests[corev1.ResourceStorage]
	r.Success(gin.H{
		"message":    fmt.Sprintf("PersistentVolumeClaim %s/%s 扩容成功，storage=%s", param.Namespace, param.Name, qty.String()),
		"name":       pvcObj.Name,
		"namespace":  pvcObj.Namespace,
		"storage":    qty.String(),
		"status":     pvcObj.Status.Phase,
		"sc":         pvcObj.Spec.StorageClassName,
		"volumeMode": pvcObj.Spec.VolumeMode,
		"created_at": pvcObj.CreationTimestamp,
	})
}
