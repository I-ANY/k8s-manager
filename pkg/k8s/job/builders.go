package job

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/dataselect"
	"k8soperation/pkg/k8s/probe"
)

// 与 Deployment 共用的常量（保持一致）
const (
	DescriptionAnnotationKey = "description"
	SystemLabelKeyApp        = "system.k8soperation/app"
)

// ---------- 通用块（可与 Deployment 共享） ----------
func BuildObjectMetaJob(req *requests.KubeJobCreateRequest, labels map[string]string) metav1.ObjectMeta {
	ann := map[string]string{}
	if req.Description != nil && *req.Description != "" {
		ann[DescriptionAnnotationKey] = *req.Description
	}
	return metav1.ObjectMeta{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Labels:      labels,
		Annotations: ann,
	}
}

// 合并系统关键标签（与 Deployment 版一致）
//
//	末尾做一次 strip，避免误把系统控制器标签合入
func mergedLabels(raw map[string]string, appName string) map[string]string {
	out := map[string]string{SystemLabelKeyApp: appName}
	for k, v := range raw {
		if k == SystemLabelKeyApp {
			continue
		}
		out[k] = v
	}
	return stripJobControllerLabels(out)
}

// Job 不强制要求设置 selector（通常留空，让控制器用 template.labels 默认化）
// 如果你想与 Service/运维工具约定关键选择器，也可以保留：
func requiredSelector(appName string) map[string]string {
	return map[string]string{SystemLabelKeyApp: appName}
}

// ---------- 容器 / PodSpec 构建 ----------
func BuildJobContainer(req *requests.KubeJobCreateRequest) corev1.Container {
	c := corev1.Container{
		Name:  firstNonEmpty(req.ContainerName, req.Name),
		Image: req.ContainerImage, // 你 DTO 的字段；下方 BuildJobResponse 里用了 req.Image，保持你现有风格
		SecurityContext: &corev1.SecurityContext{
			Privileged: boolPtr(req.RunAsPrivileged),
		},
		Resources: corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{},
		},
		Env: dataselect.ConvertEnvVarSpec(req.Variables),
	}

	// command/args
	if s := strPtrVal(req.ContainerCommand); s != "" {
		c.Command = []string{s}
	}
	if s := strPtrVal(req.ContainerCommandArgs); s != "" {
		c.Args = strings.Fields(s)
	}

	// resources
	if s := strPtrVal(req.MemoryRequirement); s != "" {
		if q, err := resource.ParseQuantity(s); err == nil {
			c.Resources.Requests[corev1.ResourceMemory] = q
		}
	}
	if s := strPtrVal(req.CpuRequirement); s != "" {
		if q, err := resource.ParseQuantity(s); err == nil {
			c.Resources.Requests[corev1.ResourceCPU] = q
		}
	}

	// ports
	if len(req.PortMappings) > 0 {
		c.Ports = ConvertContainerPorts(req.PortMappings) // 你已有的端口转换
	}

	// probes（Job 也可以配，尤其是长任务）
	if req.IsReadinessEnable {
		c.ReadinessProbe = probe.BuildProbe(req.ReadinessProbe)
	}
	if req.IsLivenessEnable {
		c.LivenessProbe = probe.BuildProbe(req.LivenessProbe)
	}
	return c
}

func BuildJobPodSpec(req *requests.KubeJobCreateRequest, containers []corev1.Container) corev1.PodSpec {
	ps := corev1.PodSpec{
		Containers:    containers,
		RestartPolicy: corev1.RestartPolicyOnFailure,
	}
	if rp := strings.TrimSpace(req.RestartPolicy); rp != "" {
		ps.RestartPolicy = corev1.RestartPolicy(rp)
	}
	if len(req.ImagePullSecrets) > 0 {
		ps.ImagePullSecrets = req.ImagePullSecrets
	}
	if sa := strings.TrimSpace(req.ServiceAccount); sa != "" {
		ps.ServiceAccountName = sa
	}
	if len(req.NodeSelector) > 0 {
		ps.NodeSelector = req.NodeSelector
	}
	if len(req.Tolerations) > 0 {
		ps.Tolerations = req.Tolerations
	}
	if req.Affinity != nil {
		ps.Affinity = req.Affinity
	}
	return ps
}

// ---------- Job 构建主体 ----------
func BuildJobFromCreateReq(req *requests.KubeJobCreateRequest) *batchv1.Job {
	// 1) 用户 labels 规范化 + 合并关键标签（并 strip 掉系统控制器标签）
	userLabels := dataselect.GetLabelsMap(req.Labels)
	labels := mergedLabels(userLabels, req.Name)

	// 2)（可选）selector：Job 一般不强制设置；若设置必须与模板 labels 完全匹配
	var selector *metav1.LabelSelector
	if req.SetExplicitSelector {
		selector = &metav1.LabelSelector{MatchLabels: requiredSelector(req.Name)}
	}

	// 3) Meta
	meta := BuildObjectMetaJob(req, labels)

	// 4) PodTemplate
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,           // 已在 mergedLabels 里 strip 过系统标签
			Annotations: meta.Annotations, // 透传 description 等
		},
		Spec: BuildJobPodSpec(req, []corev1.Container{BuildJobContainer(req)}),
	}

	// 5) JobSpec
	spec := batchv1.JobSpec{
		Template: podTemplate,

		// 运行参数（均为可选指针，保持“未设置就不下发”的语义）
		Parallelism:             req.Parallelism,
		Completions:             req.Completions,
		BackoffLimit:            req.BackoffLimit,
		ActiveDeadlineSeconds:   req.ActiveDeadlineSeconds,
		TTLSecondsAfterFinished: req.TTLSecondsAfterFinished,
		Suspend:                 req.Suspend,

		Selector: selector, // 可为 nil
	}

	return &batchv1.Job{
		ObjectMeta: meta,
		Spec:       spec,
	}
}

// BuildJobResponse 构建 Job 的返回结构
func BuildJobResponse(job *batchv1.Job, req *requests.KubeJobCreateRequest) gin.H {
	// ---------------------- 基础信息 ----------------------
	resp := gin.H{
		"job": gin.H{
			"name":            job.Name,
			"namespace":       job.Namespace,
			"labels":          job.Labels,
			"uid":             string(job.UID),
			"resourceVersion": job.ResourceVersion,
			"image":           req.Image, // 你原代码如此保留
		},
	}

	// ---------------------- 规格信息 ----------------------
	spec := gin.H{}
	if job.Spec.Completions != nil {
		spec["completions"] = *job.Spec.Completions
	}
	if job.Spec.Parallelism != nil {
		spec["parallelism"] = *job.Spec.Parallelism
	}
	if job.Spec.BackoffLimit != nil {
		spec["backoffLimit"] = *job.Spec.BackoffLimit
	}
	if job.Spec.CompletionMode != nil {
		spec["completionMode"] = string(*job.Spec.CompletionMode)
	}
	if job.Spec.ActiveDeadlineSeconds != nil {
		spec["activeDeadlineSeconds"] = *job.Spec.ActiveDeadlineSeconds
	}
	if job.Spec.TTLSecondsAfterFinished != nil {
		spec["ttlSecondsAfterFinished"] = *job.Spec.TTLSecondsAfterFinished
	}
	if job.Spec.Suspend != nil {
		spec["suspend"] = *job.Spec.Suspend
	}
	if len(spec) > 0 {
		resp["job"].(gin.H)["spec"] = spec
	}

	// ---------------------- 状态信息 ----------------------
	status := gin.H{
		"active":    job.Status.Active,
		"succeeded": job.Status.Succeeded,
		"failed":    job.Status.Failed,
	}

	// 自动推导 Job 的阶段
	phase := "Pending"
	switch {
	case job.Status.Succeeded > 0:
		phase = "Complete"
	case job.Status.Failed > 0:
		phase = "Failed"
	case job.Status.Active > 0:
		phase = "Running"
	}
	status["phase"] = phase

	// 条件（如 Complete、Failed 的详细原因）
	if len(job.Status.Conditions) > 0 {
		conds := make([]gin.H, 0, len(job.Status.Conditions))
		for _, c := range job.Status.Conditions {
			conds = append(conds, gin.H{
				"type":               string(c.Type),
				"status":             string(c.Status),
				"lastTransitionTime": c.LastTransitionTime.Time,
				"reason":             c.Reason,
				"message":            c.Message,
			})
		}
		status["conditions"] = conds
	}
	resp["job"].(gin.H)["status"] = status
	return resp
}

// ---------- 可复用：Service 端口转换（你已实现） ----------
func ConvertServicePorts(ports []requests.PortMapping) []corev1.ServicePort {
	out := make([]corev1.ServicePort, 0, len(ports))
	for _, p := range ports {
		if p.Port <= 0 || p.TargetPort <= 0 {
			continue
		}
		out = append(out, corev1.ServicePort{
			Name:       buildPortName(p.Protocol, p.Port),
			Port:       p.Port,
			TargetPort: intstr.FromInt32(p.TargetPort),
			Protocol:   parseProtocol(p.Protocol),
		})
	}
	return out
}

func ConvertContainerPorts(ports []requests.PortMapping) []corev1.ContainerPort {
	out := make([]corev1.ContainerPort, 0, len(ports))
	for _, p := range ports {
		if p.Port <= 0 {
			continue
		}
		out = append(out, corev1.ContainerPort{
			Name:          buildPortName(p.Protocol, p.Port),
			ContainerPort: p.Port,
			Protocol:      parseProtocol(p.Protocol),
		})
	}
	return out
}

func parseProtocol(s string) corev1.Protocol {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UDP":
		return corev1.ProtocolUDP
	case "SCTP":
		return corev1.ProtocolSCTP
	default:
		return corev1.ProtocolTCP
	}
}

func buildPortName(proto string, port int32) string {
	p := strings.ToLower(strings.TrimSpace(proto))
	if p == "" {
		p = "tcp"
	}
	return fmt.Sprintf("%s-%d", p, port)
}

// ---------- 小工具 ----------
func strPtrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func boolPtr(b *bool) *bool { return b }

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

// ==============================
// 旧 Job → 新 Job 的构造入口
// ==============================
func BuildJobFromOld(old *batchv1.Job, newName string) *batchv1.Job {
	// 1) Pod 模板（清理控制器标签）
	tpl := old.Spec.Template.DeepCopy()
	tpl.Labels = stripJobControllerLabels(tpl.Labels)
	if tpl.Spec.RestartPolicy == "" {
		tpl.Spec.RestartPolicy = corev1.RestartPolicyNever
	}

	// 2) Job 顶层标签与注解
	newLabels := stripJobControllerLabels(old.Labels)
	newAnnotations := map[string]string{}
	for k, v := range old.Annotations {
		newAnnotations[k] = v
	}
	newAnnotations["restartedFrom"] = old.Name
	newAnnotations["restartedAt"] = time.Now().Format(time.RFC3339)

	// 3) 构造新 Job（不设 selector/manualSelector）
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        newName,
			Namespace:   old.Namespace,
			Labels:      newLabels,
			Annotations: newAnnotations,
		},
		Spec: batchv1.JobSpec{
			Template:                *tpl,
			BackoffLimit:            old.Spec.BackoffLimit,
			Completions:             old.Spec.Completions,
			Parallelism:             old.Spec.Parallelism,
			TTLSecondsAfterFinished: old.Spec.TTLSecondsAfterFinished,
			ActiveDeadlineSeconds:   old.Spec.ActiveDeadlineSeconds,
			Selector:                nil, // 让控制器自动生成
			ManualSelector:          nil,
		},
	}

	// 4) 清理只读字段（否则 Create 会报错）
	job.ResourceVersion = ""
	job.UID = ""
	job.OwnerReferences = nil
	job.ManagedFields = nil
	job.CreationTimestamp = metav1.Time{}

	return job
}
