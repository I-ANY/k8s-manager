package cronjob

import (
	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/job"
	"strings"
)

// 建议放在 cronjob 包顶部的默认值
var (
	defaultPolicy               = batchv1.ForbidConcurrent
	defaultSuccessHistory int32 = 3
	defaultFailedHistory  int32 = 1
)

func BuildCronJobFromCreateReq(req *requests.KubeCronJobCreateRequest) *batchv1.CronJob {
	// --- 基本参数 ---
	ns := strings.TrimSpace(req.Namespace)
	name := strings.TrimSpace(req.Name)
	schedule := strings.TrimSpace(req.Schedule)

	// --- RestartPolicy 默认值 ---
	rp := corev1.RestartPolicyOnFailure
	if req.RestartPolicy == string(corev1.RestartPolicyNever) {
		rp = corev1.RestartPolicyNever
	}

	// --- 构造容器 ---
	var containers []corev1.Container
	if len(req.Containers) > 0 {
		containers = req.Containers
	} else {
		cn := strings.TrimSpace(req.ContainerName)
		if cn == "" {
			cn = name
		}
		containers = []corev1.Container{{
			Name:    cn,
			Image:   strings.TrimSpace(req.ContainerImage),
			Command: buildCommand(req.ContainerCommand),
			Args:    buildArgs(req.ContainerCommandArgs),
		}}
	}

	// --- 构造 JobSpec ---
	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: job.BuildJobPodSpec(&requests.KubeJobCreateRequest{
				RestartPolicy:    string(rp),
				ImagePullSecrets: req.ImagePullSecrets,
				ServiceAccount:   req.ServiceAccount,
				NodeSelector:     req.NodeSelector,
				Tolerations:      req.Tolerations,
				Affinity:         req.Affinity,
			}, containers),
		},
	}

	// --- 构造 CronJob ---
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   schedule,
			Suspend:                    req.Suspend,
			JobTemplate:                batchv1.JobTemplateSpec{Spec: jobSpec},
			ConcurrencyPolicy:          batchv1.ConcurrencyPolicy(req.ConcurrencyPolicy),
			StartingDeadlineSeconds:    req.StartingDeadlineSeconds,
			SuccessfulJobsHistoryLimit: req.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     req.FailedJobsHistoryLimit,
		},
	}

	// --- 时区 ---
	if tz := strings.TrimSpace(req.TimeZone); tz != "" {
		cj.Spec.TimeZone = &tz
	}

	// --- Label ---
	if cj.Spec.JobTemplate.Spec.Template.Labels == nil {
		cj.Spec.JobTemplate.Spec.Template.Labels = make(map[string]string)
	}
	cj.Spec.JobTemplate.Spec.Template.Labels["system.k8soperation/app"] = name

	return cj
}

func buildCommand(cmd []string) []string {
	if len(cmd) == 0 {
		return nil
	}
	return cmd
}

func buildArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	return args
}

// 小工具
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// BuildCronJobResponse 构建 CronJob 的返回结构
func BuildCronJobResponse(cj *batchv1.CronJob, req *requests.KubeCronJobCreateRequest) gin.H {
	// 取第一个容器镜像（兼容旧前端单镜像展示）
	firstImage := ""
	cs := cj.Spec.JobTemplate.Spec.Template.Spec.Containers
	if len(cs) > 0 {
		firstImage = cs[0].Image
	}

	// 同时把所有容器返回给前端（推荐）
	outContainers := make([]gin.H, 0, len(cs))
	for _, c := range cs {
		outContainers = append(outContainers, gin.H{
			"name":    c.Name,
			"image":   c.Image,
			"command": c.Command,
			"args":    c.Args,
		})
	}

	spec := gin.H{
		"schedule":                   cj.Spec.Schedule,
		"concurrencyPolicy":          string(cj.Spec.ConcurrencyPolicy),
		"suspend":                    cj.Spec.Suspend != nil && *cj.Spec.Suspend,
		"successfulJobsHistoryLimit": cj.Spec.SuccessfulJobsHistoryLimit,
		"failedJobsHistoryLimit":     cj.Spec.FailedJobsHistoryLimit,
	}

	// 可选：把 JobTemplate 里的一些字段也回传
	js := cj.Spec.JobTemplate.Spec
	spec["jobSpec"] = gin.H{
		"restartPolicy":           string(cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy),
		"parallelism":             js.Parallelism,
		"completions":             js.Completions,
		"backoffLimit":            js.BackoffLimit,
		"ttlSecondsAfterFinished": js.TTLSecondsAfterFinished,
		"activeDeadlineSeconds":   js.ActiveDeadlineSeconds,
	}

	status := gin.H{
		"active": len(cj.Status.Active),
		"lastScheduleTime": func() interface{} {
			if cj.Status.LastScheduleTime != nil {
				return cj.Status.LastScheduleTime.Time
			}
			return nil
		}(),
		"lastSuccessfulTime": func() interface{} {
			if cj.Status.LastSuccessfulTime != nil {
				return cj.Status.LastSuccessfulTime.Time
			}
			return nil
		}(),
	}

	return gin.H{
		"cronjob": gin.H{
			"name":            cj.Name,
			"namespace":       cj.Namespace,
			"labels":          cj.Labels,
			"uid":             string(cj.UID),
			"resourceVersion": cj.ResourceVersion,
			"image":           firstImage,    // ★ 兼容老字段
			"containers":      outContainers, // ★ 新增容器数组
			"spec":            spec,
			"status":          status,
		},
	}
}
