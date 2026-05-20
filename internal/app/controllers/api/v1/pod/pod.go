package pod

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	appctx "k8soperation/pkg/app"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type PodController struct{}

func NewPodController() *PodController {
	return &PodController{}
}

// List godoc
// @Summary 列出K8s Pod
// @Description 列出K8s Pod
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/list [get]
func (c *PodController) List(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewKubePodListRequest()
	resp := response.NewResponse(ctx)

	// 1) 防御：K8s 客户端未初始化
	if a.KubeClient == nil {
		// 如果你有专门的错误码，建议用 ErrorClusterNotInitialized 之类的
		resp.ToErrorResponse(errorcode.ErrorK8sPodListFail.WithDetails("k8s client not initialized"))
		return
	}

	if ok := valid.Validate(ctx, param, requests.ValidKubePodListRequest); !ok {
		return
	}

	// 创建服务实例，传入上下文ctx
	svc := services.NewServices(ctx, a)
	// 调用服务实例的KubePodList方法获取Pod列表，传入参数param
	pods, err := svc.KubePodList(ctx, param)
	// 检查获取Pod列表时是否发生错误
	if err != nil {
		// 记录错误日志，包含错误信息
		a.Logger.Error("获取Pod列表失败", zap.String("error", err.Error()))
		// 返回错误响应，包含错误代码和详细信息
		resp.ToErrorResponse(errorcode.ErrorK8sPodListFail.WithDetails(err.Error()))
		// 终止当前函数执行
		return
	}

	resp.SuccessList(pods, len(pods))
}

// Update godoc
// @Summary 更新Pod
// @Description 更新Pod
// @Tags K8s Pod管理
// @Produce json
// @Param body body requests.KubePodUpdateRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/update [post]
func (c *PodController) Update(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewKubePodUpdateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodUpdateRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices(ctx, a)
	if err := svc.KubePodUpdate(param); err != nil {
		a.Logger.Errorf("更新Pod失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sPodUpdateFail.WithDetails(err.Error()))
		return
	}
	resp.Success(gin.H{"msg": "Pod更新成功"})
}

// PatchImage godoc
// @Summary 更新 Pod 容器镜像
// @Description 基于 mergeKey=name 的 StrategicMergePatch 方式更新指定容器的镜像
// @Tags K8s Pod管理
// @Accept json
// @Produce json
// @Param body body requests.PatchPodImageRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/patch_image [put]
func (c *PodController) PatchImage(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewPatchPodImageRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodPatchContainerImageRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	if a.KubeClient == nil {
		resp.ToErrorResponse(errorcode.ErrorClusterInitFailed.WithDetails("k8s client 未初始化"))
		return
	}

	svc := services.NewServices(ctx, a)
	if err := svc.PatchPodImage(param); err != nil {
		a.Logger.Errorf("PatchPodImage 失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sPodPatchFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"msg":       "Pod 镜像更新成功",
		"namespace": param.Namespace,
		"name":      param.Name,
		"container": param.Container,
		"new_image": param.NewImage,
	})
}

// @Summary 删除 Pod
// @Description 删除指定命名空间下的 Pod（支持优雅终止/强制删除）
// @Tags K8s Pod管理
// @Produce json
// @Param namespace query string true  "命名空间"
// @Param name      query string true  "Pod 名称"
// @Param grace_seconds query int false "优雅终止秒数（默认30）"
// @Param force     query bool  false "是否强制删除（默认false）"
// @Success 200 {object} map[string]interface{} "删除请求已提交"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/grace_delete_pod [delete]
func (c *PodController) DeletePod(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	resp := response.NewResponse(ctx)

	param := requests.NewKubePodDeleteRequest()
	if ok := valid.Validate(ctx, param, requests.ValidKubePodDeleteRequest); !ok {
		return
	}

	if a.KubeClient == nil {
		resp.ToErrorResponse(errorcode.ErrorClusterInitFailed.WithDetails("k8s client 未初始化"))
		return
	}

	svc := services.NewServices(ctx, a)
	if err := svc.KubePodDelete(param); err != nil {
		resp.ToErrorResponse(errorcode.ErrorK8sPodDeleteFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"namespace":     param.Namespace,
		"name":          param.Name,
		"force":         param.Force,
		"grace_seconds": param.GraceSeconds,
		"message":       "删除成功",
	})
}

// Detail godoc
// @Summary 获取Pod的详情
// @Description 获取Pod的详情
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/detail [get]
func (c *PodController) Detail(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices(ctx, a)
	pod, err := svc.KubePodDetail(param)
	if err != nil {
		a.Logger.Errorf("获取Pod详情失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sPodDetailFail.WithDetails(err.Error()))
		return
	}

	resp.Success(pod)
}

// GetContainerName godoc
// @Summary 获取Pod的容器名
// @Description 获取Pod的容器名
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_name [get]
func (c *PodController) GetContainerName(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices(ctx, a)
	containerName, err := svc.GetContainerNames(param)
	if err != nil {
		a.Logger.Errorf("获取Pod容器名失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerName.WithDetails(err.Error()))
		return
	}

	resp.Success(containerName)
}

// GetInitContainerName godoc
// @Summary 获取Pod的Init容器名
// @Description 获取Pod的Init容器名
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/init_container_name [get]
func (c *PodController) GetInitContainerName(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewKubeCommonRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		a.Logger.Errorf("获取Pod的Init容器名失败: %v", ok)
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices(ctx, a)
	initContainerName, err := svc.GetInitContainerNames(param)
	if err != nil {
		a.Logger.Errorf("获取Pod的Init容器名失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetInitContainerName.WithDetails(err.Error()))
		return
	}

	resp.Success(initContainerName)
}

// GetContainerImages godoc
// @Summary 获取Pod的容器镜像
// @Description 获取Pod的容器镜像
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_image [get]
func (c *PodController) GetContainerImages(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		a.Logger.Errorf("校验失败：%v", ok)
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices(ctx, a)
	containerImages, err := svc.GetContainerImages(param)
	if err != nil {
		a.Logger.Errorf("获取Pod容器镜像失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerImage.WithDetails(err.Error()))
		return
	}

	a.Logger.Infof("获取Pod容器镜像成功，总数: %d", len(containerImages))
	resp.SuccessList(containerImages, len(containerImages))
}

// GetInitContainerImages godoc
// @Summary 获取Pod的Init容器镜像
// @Description 获取Pod的Init容器镜像
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/init_container_image [get]
func (c *PodController) GetInitContainerImages(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	param := requests.NewKubeCommonRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		a.Logger.Errorf("校验失败：%v", ok)
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices(ctx, a)
	initContainerImages, err := svc.GetInitContainerImages(param)
	if err != nil {
		a.Logger.Errorf("获取Pod的Init容器镜像失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetInitContainerImage.WithDetails(err.Error()))
		return
	}

	a.Logger.Infof("获取Pod的Init容器镜像成功: %d", len(initContainerImages))
	resp.SuccessList(initContainerImages, len(initContainerImages))
}

// GetContainerLog godoc
// @Summary 获取Pod的容器日志
// @Description 由全局开关 PodLog.EnableStreaming 控制：true=实时流式(text/plain)，false=一次性(JSON)
// @Tags K8s Pod管理
// @Produce json
// @Produce text/plain
// @Param name query string true "Pod名" maxlength(100)
// @Param namespace query string true "命名空间" maxlength(100)
// @Param container query string false "容器名(可选; 多容器建议指定)" maxlength(100)
// @Param tail query int false "仅返回最后N行(默认见配置)"
// @Success 200 {object} map[string]interface{} "成功(EnableStreaming=false)"
// @Success 200 {string} string "流式文本(EnableStreaming=true)"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_log [get]
func (c *PodController) GetContainerLogs(ctx *gin.Context) {
	a := appctx.FromContext(ctx)
	// 1) 绑定 + 校验
	param := requests.NewKubePodLogRequest()
	resp := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubePodLogRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("参数校验失败"))
		return
	}

	// 2) 检查 K8s 客户端
	if a.KubeClient == nil {
		resp.ToErrorResponse(errorcode.ErrorClusterInitFailed.WithDetails("k8s client 未初始化"))
		return
	}

	svc := services.NewServices(ctx, a)

	if param.Follow {
		// —— 流式 —— //
		rc, err := svc.KubePodLogStream(
			ctx.Request.Context(),
			param.Name,
			param.Namespace,
			param.Container,
			param.Tail,
		)
		if err != nil {
			resp.ToErrorResponse(errorcode.ErrorK8sGetContainerLog.WithDetails(err.Error()))
			return
		}
		defer func(rc io.ReadCloser) {
			err := rc.Close()
			if err != nil {
				a.Logger.Errorf("关闭流式日志失败 ns=%s pod=%s container=%s : %v",
					param.Namespace, param.Name, param.Container, err)
			}
		}(rc)
		ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if _, err := io.Copy(ctx.Writer, rc); err != nil {
			a.Logger.Errorf("stream copy err ns=%s pod=%s container=%s : %v",
				param.Namespace, param.Name, param.Container, err)
		}
		return // 这里要 return，避免继续往下走
	}

	// —— 一次性 —— //
	logStr, err := svc.KubePodLog(ctx.Request.Context(),
		param.Name, param.Namespace, param.Container, param.Tail)
	if err != nil {
		a.Logger.Errorf("get pod log failed ns=%s pod=%s container=%s tail=%d : %v",
			param.Namespace, param.Name, param.Container, param.Tail, err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerLog.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"namespace": param.Namespace,
		"pod":       param.Name,
		"container": param.Container,
		"tail":      param.Tail,
		"log":       logStr,
	})
}

// GetContainerLog godoc
// @Summary 获取Pod的容器日志
// @Description 获取Pod的容器日志
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Param container query string false "容器" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_logs [get]
func (k *PodController) GetContainerLog(ctx *gin.Context) {
	a := appctx.FromContext(ctx)

	// 临时调试：看看客户端到底传了什么
	a.Logger.Info("DEBUG RAW QUERY",
		zap.String("raw", ctx.Request.URL.RawQuery),
		zap.String("name", ctx.Query("name")),
		zap.String("namespace", ctx.Query("namespace")),
		zap.String("container", ctx.Query("container")),
		zap.String("tail", ctx.Query("tail")),
	)

	param := requests.NewKubePodLogRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodLogRequest); !ok {
		return
	}

	svc := services.NewServices(ctx, a)
	logs, err := svc.GetPodLog(param.Name, param.Namespace, param.Container, param.Tail)
	if err != nil {
		a.Logger.Error("获取 Pod 日志失败",
			zap.String("namespace", param.Namespace),
			zap.String("pod", param.Name),
			zap.String("container", param.Container),
			zap.Int("tail", int(param.Tail)),
			zap.Error(err),
		)
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerLog.WithDetails(err.Error()))
		return
	}

	resp.ToErrorResponse(errorcode.Success.WithDetails(logs))
}
