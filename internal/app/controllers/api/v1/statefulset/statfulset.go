package statefulset

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/utils"
	"k8soperation/pkg/valid"
	"time"
)

type KubeStatefulSetController struct{}

func NewKubeStatefulSetController() *KubeStatefulSetController {
	return &KubeStatefulSetController{}
}

// Create godoc
// @Summary 创建 StatefulSet（可选创建 Service）
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/create [post]
func (c *KubeStatefulSetController) Create(ctx *gin.Context) {
	req := requests.NewKubeStatefulSetCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeStatefulSetCreateRequest); !ok {
		return
	}
	service := services.NewServices(ctx)
	sts, svc, err := service.KubeStatefulSetCreateService(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetCreate error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "创建成功",
		"service": svc,
		"result":  sts,
	})
}

// List godoc
// @Summary 获取 StatefulSet 列表
// @Description 分页、模糊查询 StatefulSet 列表
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param name query string false "StatefulSet 名称关键字（模糊匹配）"
// @Param page query int false "页码（默认 1）"
// @Param limit query int false "每页数量（默认 10）"
// @Success 200 {object} string "查询成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/list [get]
func (c *KubeStatefulSetController) List(ctx *gin.Context) {
	req := requests.NewKubeStatefulSetListRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeStatefulSetListRequest); !ok {
		return
	}
	service := services.NewServices(ctx)
	sts, total, err := service.KubeStatefulSetList(ctx.Request.Context(), req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetList error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "查询成功",
		"result":  sts,
		"total":   total,
	})
}

// Detail godoc
// @Summary 获取 StatefulSet 详情
// @Description 根据命名空间和名称获取单个 StatefulSet 的详细信息
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "StatefulSet 名称"
// @Success 200 {object} string "获取详情成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/detail [get]
func (c *KubeStatefulSetController) Detail(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetDetailRequest); !ok {
		return
	}
	service := services.NewServices(ctx)
	sts, err := service.KubeStatefulSetDetail(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetDetail error", zap.Error(err))
	}

	r.Success(gin.H{
		"message": "获取详情成功",
		"result":  sts,
	})
}

// Scale godoc
// @Summary 扩缩容 StatefulSet（修改副本数）
// @Description 通过 Patch 局部更新 .spec.replicas，K8s 将按策略有序创建/删除 Pod
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetScaleRequest true "扩缩容参数（namespace、name、scale_num）"
// @Success 200 {object} map[string]interface{} "扩缩容成功，返回修改前后及当前副本信息"
// @Failure 400 {object} errorcode.Error
// @Failure 404 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/scale [put]
func (c *KubeStatefulSetController) Scale(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetScaleRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentScaleRequest); !ok {
		return
	}
	service := services.NewServices(ctx)
	sts, err := service.KubeStatefulSetPatchReplicas(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err) // 交给中间件
		return
	}

	// 扩缩容成功后的返回（sts 为 *appv1.StatefulSet）
	r.Success(gin.H{
		"namespace": sts.Namespace,
		"name":      sts.Name,
		"replicas":  utils.ValueOrZero(sts.Spec.Replicas),
		"ready":     sts.Status.ReadyReplicas,
		"updated":   sts.Status.UpdatedReplicas,
		"rv":        sts.ResourceVersion,
		"status": fmt.Sprintf(
			"扩缩容成功，目标副本数：%d，当前就绪：%d/%d",
			utils.ValueOrZero(sts.Spec.Replicas),
			sts.Status.ReadyReplicas,
			utils.ValueOrZero(sts.Spec.Replicas),
		),
	})

}

// UpdateImage godoc
// @Summary 更新 StatefulSet 容器镜像（Patch 局部更新）
// @Description 仅修改 .spec.template.spec.containers[*].image，不影响其它字段；根据 UpdateStrategy 触发滚动更新
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetUpdateImageRequest true "更新镜像参数（namespace、name、container、image）"
// @Success 200 {object} map[string]interface{} "更新成功，返回资源版本与副本进度"
// @Failure 400 {object} errorcode.Error
// @Failure 404 {object} errorcode.Error
// @Failure 409 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/update_image [put]
func (c *KubeStatefulSetController) UpdateImage(ctx *gin.Context) {
	// 1) 绑定请求体
	param := new(requests.KubeStatefulSetUpdateImageRequest)
	r := response.NewResponse(ctx)

	// 2) 参数校验（StatefulSet 的校验器）
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetUpdateImageRequest); !ok {
		return
	}

	// 3) 调用服务
	svc := services.NewServices(ctx)
	sts, err := svc.KubeStatefulSetPatchImage(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err) // 交给中间件统一处理
		return
	}

	// 4) 成功返回（回显关键信息）
	r.Success(gin.H{
		"namespace": sts.Namespace,
		"name":      sts.Name,
		"container": param.Container,
		"image":     param.Image,
		"ready":     sts.Status.ReadyReplicas,             // 就绪副本
		"replicas":  utils.ValueOrZero(sts.Spec.Replicas), // 期望副本
		"rv":        sts.ResourceVersion,                  // 资源版本
		"status": fmt.Sprintf("修改镜像成功，当前就绪：%d/%d",
			sts.Status.ReadyReplicas, utils.ValueOrZero(sts.Spec.Replicas)),
	})
}

// Restart godoc
// @Summary 重启 StatefulSet（触发滚动重启）
// @Description 在 .spec.template.metadata.annotations 写入 `kubectl.kubernetes.io/restartedAt` 时间戳，从而触发 StatefulSet 的滚动更新；等价于 `kubectl rollout restart sts <name>`。
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetRestartRequest true "重启参数（namespace、name）"
// @Success 200 {object} map[string]interface{} "重启成功，返回命名空间、名称与触发时间等信息"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 404 {object} errorcode.Error "StatefulSet 未找到"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/statefulset/restart [post]
func (c *KubeStatefulSetController) Restart(ctx *gin.Context) {
	ts := time.Now().Format(time.RFC3339)

	param := requests.NewKubeStatefulSetRestartRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetRestartRequest); !ok {
		return
	}
	service := services.NewServices(ctx)
	sts, err := service.KubeStatefulSetRestart(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetRestart error", zap.Error(err))
	}
	r.Success(gin.H{
		"namespace":   sts.Namespace,
		"name":        sts.Name,
		"restartedAt": ts,
		"status":      "StatefulSet 滚动重启已触发",
	})
}

// Delete godoc
// @Summary     删除 StatefulSet
// @Description 前台级联删除 StatefulSet（先删 Pod/ControllerRevision，再删 StatefulSet 本体）；成功返回命名空间、名称与状态信息
// @Tags        K8s StatefulSet 管理
// @Accept      json
// @Produce     json
// @Param       namespace query string true  "命名空间"
// @Param       name      query string true  "StatefulSet 名称"
// @Success     200 {object} map[string]interface{} "示例: {\"namespace\":\"default\",\"name\":\"web\",\"status\":\"StatefulSet 删除成功\"}"
// @Failure     400 {object} errorcode.Error "参数错误"
// @Failure     404 {object} errorcode.Error "StatefulSet 未找到"
// @Failure     500 {object} errorcode.Error "内部错误"
// @Router      /api/v1/k8s/statefulset/delete [delete]
// @Security    BearerAuth
func (c *KubeStatefulSetController) Delete(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetDeleteRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetDeleteRequest); !ok {
		return
	}
	service := services.NewServices(ctx)
	err := service.KubeStatefulSetDelete(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"status":    "StatefulSet 删除成功",
	})
}

// Delete godoc
// @Summary     删除 Service
// @Description 根据命名空间与名称删除 Service；成功仅表示删除请求已受理。
// @Tags        K8s StatefulSet 管理
// @Accept      json
// @Produce     json
// @Param       body  body  map[string]string  true  "删除参数（namespace、name）"
// @Success     200 {object} map[string]interface{} "删除成功"
// @Failure     400 {object} errorcode.Error "参数错误"
// @Failure     404 {object} errorcode.Error "Service 未找到"
// @Failure     500 {object} errorcode.Error "内部错误"
// @Router      /api/v1/k8s/statefulset/delete_svc [delete]
func (c *KubeStatefulSetController) DeleteService(context *gin.Context) {
	param := requests.NewKubeStatefulSetDeleteRequest()
	r := response.NewResponse(context)
	if ok := valid.Validate(context, param, requests.ValidKubeStatefulSetDeleteRequest); !ok {
		return
	}
	service := services.NewServices(context)
	err := service.KubeStatefulSetDeleteService(context.Request.Context(), param)
	if err != nil {
		context.Error(err)
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"status":    "Service 删除成功",
	})
}
