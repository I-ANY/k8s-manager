package daemonset

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/dataselect"
	"k8soperation/pkg/k8s/probe"
)

/* -------- 常量：与 Deployment 版本保持一致 -------- */

const (
	DescriptionAnnotationKey = "description"
	SystemLabelKeyApp        = "system.k8soperation/app"
)

/* -------- 公共元数据/标签工具：完全复用语义 -------- */

func BuildObjectMeta(req *requests.KubeDaemonSetCreateRequest, labels map[string]string) metav1.ObjectMeta {
	annotations := map[string]string{}
	if req.Description != nil && *req.Description != "" {
		annotations[DescriptionAnnotationKey] = *req.Description
	}
	return metav1.ObjectMeta{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}
}

// 写入系统关键标签，禁止覆盖
func mergedLabels(raw map[string]string, appName string) map[string]string {
	out := map[string]string{SystemLabelKeyApp: appName}
	for k, v := range raw {
		if k == SystemLabelKeyApp {
			continue
		}
		out[k] = v
	}
	return out
}

// DaemonSet/Service/PodTemplate 公用的最小 selector
func requiredSelector(appName string) map[string]string {
	return map[string]string{SystemLabelKeyApp: appName}
}

/* -------- 容器与 PodSpec：与 Deployment 基本一致 -------- */

func BuildContainerFromCreateReq(req *requests.KubeDaemonSetCreateRequest) corev1.Container {
	c := corev1.Container{
		Name:  req.Name,
		Image: req.ContainerImage,
		SecurityContext: &corev1.SecurityContext{
			Privileged: &req.RunAsPrivileged,
		},
		Resources: corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{},
		},
		Env: dataselect.ConvertEnvVarSpec(req.Variables),
	}

	if req.ContainerCommand != nil && *req.ContainerCommand != "" {
		c.Command = []string{*req.ContainerCommand}
	}
	if req.ContainerCommandArgs != nil && *req.ContainerCommandArgs != "" {
		c.Args = strings.Fields(*req.ContainerCommandArgs)
	}

	if req.MemoryRequirement != nil && *req.MemoryRequirement != "" {
		if q, err := resource.ParseQuantity(*req.MemoryRequirement); err == nil {
			c.Resources.Requests[corev1.ResourceMemory] = q
		}
	}
	if req.CpuRequirement != nil && *req.CpuRequirement != "" {
		if q, err := resource.ParseQuantity(*req.CpuRequirement); err == nil {
			c.Resources.Requests[corev1.ResourceCPU] = q
		}
	}

	if len(req.PortMappings) > 0 {
		c.Ports = probe.ConvertContainerPorts(req.PortMappings)
	}
	if req.IsReadinessEnable {
		c.ReadinessProbe = probe.BuildProbe(req.ReadinessProbe)
	}
	if req.IsLivenessEnable {
		c.LivenessProbe = probe.BuildProbe(req.LivenessProbe)
	}
	return c
}

func BuildPodSpec(req *requests.KubeDaemonSetCreateRequest, containers []corev1.Container) corev1.PodSpec {
	ps := corev1.PodSpec{Containers: containers}

	// 镜像拉取密钥
	if req.ImagePullSecret != nil && *req.ImagePullSecret != "" {
		ps.ImagePullSecrets = []corev1.LocalObjectReference{{Name: *req.ImagePullSecret}}
	}

	// DaemonSet 常见定制项（可选）：节点选择/容忍/亲和等
	if len(req.NodeSelector) > 0 {
		ps.NodeSelector = req.NodeSelector
	}
	if len(req.Tolerations) > 0 {
		ps.Tolerations = req.Tolerations
	}
	if req.Affinity != nil {
		ps.Affinity = req.Affinity
	}
	if req.HostNetwork {
		ps.HostNetwork = true
		// HostNetwork 时常见做法：把 DNSPolicy 置为 ClusterFirstWithHostNet
		ps.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	}
	return ps
}

/* -------- DaemonSet 构建 -------- */

func BuildDaemonSetFromCreateReq(req *requests.KubeDaemonSetCreateRequest) *appv1.DaemonSet {
	// 1) labels 规范化 + 合并关键标签
	userLabels := dataselect.GetLabelsMap(req.Labels)
	labels := mergedLabels(userLabels, req.Name)

	// 2) 最小 selector（与 Pod 模版/Service 必须一致）
	selector := requiredSelector(req.Name)

	// 3) 元数据
	dsMeta := BuildObjectMeta(req, labels)

	// 4) PodTemplate（Namespace 继承自 DS）
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: dsMeta.Annotations,
		},
		Spec: BuildPodSpec(req, []corev1.Container{
			BuildContainerFromCreateReq(req),
		}),
	}

	// 5) 更新策略：DaemonSet 默认 RollingUpdate
	//    支持从请求读取 MaxUnavailable；没有就用 "1" 兜底
	var ru *appv1.RollingUpdateDaemonSet

	var ios intstr.IntOrString
	if req.MaxUnavailable != nil && strings.TrimSpace(*req.MaxUnavailable) != "" {
		s := strings.TrimSpace(*req.MaxUnavailable)
		ios = intstr.Parse(s) // "1" -> Int；"10%" -> String
		// 可选：若是 String，再做一次百分比合法校验
		if ios.Type == intstr.String && !regexp.MustCompile(`^\d+%$`).MatchString(ios.StrVal) {
			return nil // 或者 return error，新手段：直接拒绝非法值
		}
	} else {
		ios = intstr.FromInt(1) // 默认 1（注意：用 FromInt，不是 FromInt32）
	}

	ru = &appv1.RollingUpdateDaemonSet{MaxUnavailable: &ios}

	ds := &appv1.DaemonSet{
		ObjectMeta: dsMeta,
		Spec: appv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: podTemplate,
			UpdateStrategy: appv1.DaemonSetUpdateStrategy{
				Type:          appv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: ru,
			},
			RevisionHistoryLimit: req.RevisionHistoryLimit,
			MinReadySeconds:      req.MinReadySeconds,
		},
	}
	return ds
}

/* -------- 标准响应体（与 Deployment 风格一致） -------- */

func BuildDaemonSetResponse(ds *appv1.DaemonSet, svc *corev1.Service, req *requests.KubeDaemonSetCreateRequest) gin.H {
	resp := gin.H{
		"daemonset": gin.H{
			"name":            ds.Name,
			"namespace":       ds.Namespace,
			"image":           req.ContainerImage,
			"labels":          ds.Labels,
			"uid":             string(ds.UID),
			"resourceVersion": ds.ResourceVersion,
			"strategy":        ds.Spec.UpdateStrategy.Type,
		},
	}

	if svc != nil {
		svcData := gin.H{
			"created":  true,
			"name":     svc.Name,
			"types":    string(svc.Spec.Type),
			"ports":    svc.Spec.Ports,
			"selector": svc.Spec.Selector,
		}
		if ip := svc.Spec.ClusterIP; ip != "" {
			svcData["clusterIP"] = ip
		}
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			svcData["ingress"] = svc.Status.LoadBalancer.Ingress
		}
		resp["service"] = svcData
	}
	return resp
}

/* -------- Service：与 Deployment 版一致（selector 必须对齐） -------- */

func BuildServiceFromDaemonSetReq(req *requests.KubeDaemonSetCreateRequest) *corev1.Service {
	userLabels := dataselect.GetLabelsMap(req.Labels)
	labels := mergedLabels(userLabels, req.Name)

	selector := requiredSelector(req.Name)

	name := req.ServiceName
	if name == "" {
		name = req.Name
	}

	svcType := corev1.ServiceTypeClusterIP
	if req.ServiceType != "" {
		svcType = corev1.ServiceType(req.ServiceType)
	}

	ports := ConvertServicePorts(req.PortMappings)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     svcType,
			Selector: selector,
			Ports:    ports,
		},
	}
}

/* -------- 端口工具（与 Deployment 版一致） -------- */

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
