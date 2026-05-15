package requests

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ---------------------- Deployment 创建 ----------------------

func NewKubeDeploymentCreateRequest() *KubeDeploymentCreateRequest {
	return &KubeDeploymentCreateRequest{}
}

// KubeDeploymentCreateRequest 定义创建 Deployment 的请求结构
type KubeDeploymentCreateRequest struct {
	Name                 string                `json:"name" valid:"name"`                                     // Deployment 名称
	ContainerImage       string                `json:"container_image" valid:"container_image"`               // 容器镜像
	ImagePullSecret      *string               `json:"image_pull_secret" valid:"image_pull_secret"`           // 镜像拉取密钥（可选）
	ContainerCommand     *string               `json:"container_command" valid:"container_command"`           // 容器启动命令（可选）
	ContainerCommandArgs *string               `json:"container_command_args" valid:"container_command_args"` // 容器启动参数（可选）
	Replicas             int32                 `json:"replicas" valid:"replicas"`                             // 副本数
	PortMappings         []PortMapping         `json:"port_mappings" valid:"port_mappings"`                   // 容器端口映射
	Variables            []EnvironmentVariable `json:"variables" valid:"variables"`                           // 环境变量
	IsCreateService      bool                  `json:"is_create_service" valid:"is_create_service"`           // 是否同时创建 Service
	Description          *string               `json:"description" valid:"description"`                       // 描述信息（可选）
	Namespace            string                `json:"namespace" valid:"namespace"`                           // 命名空间
	MemoryRequirement    *string               `json:"memory_requirement" valid:"memory_requirement"`         // 内存需求（可选）
	CpuRequirement       *string               `json:"cpu_requirement" valid:"cpu_requirement"`               // CPU 需求（可选）
	Labels               []Label               `json:"labels" valid:"labels"`                                 // 标签
	RunAsPrivileged      bool                  `json:"run_as_privileged" valid:"run_as_privileged"`           // 是否以特权模式运行
	IsReadinessEnable    bool                  `json:"is_readiness_enable" valid:"is_readiness_enable"`       // 是否启用 Readiness 探针
	ReadinessProbe       HealthCheckDetail     `json:"readiness_probe" valid:"readiness_probe"`               // Readiness 探针配置
	IsLivenessEnable     bool                  `json:"is_liveness_enable" valid:"is_liveness_enable"`         // 是否启用 Liveness 探针
	LivenessProbe        HealthCheckDetail     `json:"liveness_probe" valid:"liveness_probe"`                 // Liveness 探针配置
	ServiceType          string                `json:"service_type" valid:"service_type"`                     // Service 类型
	ServiceName          string                `json:"service_name" valid:"service_name"`                     // Service 名称
}

// EnvironmentVariable 定义环境变量
type EnvironmentVariable struct {
	Name  string `json:"name" valid:"name"`   // 环境变量名
	Value string `json:"value" valid:"value"` // 环境变量值
}

// ---------------------- Deployment 更新 ----------------------

func NewKubeDeploymentUpdateRequest() *KubeDeploymentUpdateRequest {
	return &KubeDeploymentUpdateRequest{}
}

type KubeDeploymentUpdateRequest struct {
	Namespace string          `json:"namespace" valid:"namespace"` // 命名空间
	Content   json.RawMessage `json:"content" valid:"content"`     // 更新内容（一般是 YAML/JSON）
	Name      string          `json:"name" valid:"name"`           // Deployment 名称
}

func ValidKubeDeploymentUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"content":   []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"content":   []string{"required: content 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 列表 ----------------------

func NewKubeDeploymentListRequest() *KubeDeploymentListRequest {
	return &KubeDeploymentListRequest{}
}

type KubeDeploymentListRequest struct {
	KubeCommonRequest
	Page  int `json:"page" valid:"page"`   // 页码
	Limit int `json:"limit" valid:"limit"` // 每页条数
}

func ValidKubeDeploymentListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":      []string{""},         // 非必填
		"namespace": []string{"required"}, // 必填
		"page":      []string{},           // 非必填
		"limit":     []string{},           // 非必填
	}

	messages := govalidator.MapData{
		"name":      []string{"string_length: name长度应在1~64之间"},
		"namespace": []string{"required: namespace不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 扩缩容 ----------------------
func NewKubeDeploymentScaleRequest() *KubeDeploymentScaleRequest {
	return &KubeDeploymentScaleRequest{}
}

type KubeDeploymentScaleRequest struct {
	KubeCommonRequest
	ScaleNum int32 `json:"scale_num" valid:"scale_num"` // 副本数量
}

func ValidKubeDeploymentScaleRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":      []string{"required"},
		"namespace": []string{"required"},
		"scale_num": []string{"required"},
	}
	messages := govalidator.MapData{
		"name":      []string{"required: name不能为空"},
		"namespace": []string{"required: namespace不能为空"},
		"scale_num": []string{"required: scale_num不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 回滚 ----------------------

func NewKubeDeploymentRollbackRequest() *KubeDeploymentRollbackRequest {
	return &KubeDeploymentRollbackRequest{}
}

type KubeDeploymentRollbackRequest struct {
	KubeCommonRequest
	ReplicaSet string `json:"replica_set" valid:"replica_set"` // 指定回滚的 ReplicaSet
}

func ValidKubeDeploymentRollbackRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":   []string{"required"},
		"name":        []string{"required"},
		"replica_set": []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace":   []string{"required: namespace 不能为空"},
		"name":        []string{"required: name 不能为空"},
		"replica_set": []string{"required: replica_set 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 重启 ----------------------

func NewKubeDeploymentRestartRequest() *KubeDeploymentRestartRequest {
	return &KubeDeploymentRestartRequest{}
}

type KubeDeploymentRestartRequest struct {
	KubeCommonRequest
}

func ValidKubeDeploymentRestartRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}

	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}

	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------- 获取详情 ---------------- */

func NewKubeDeploymentDetailRequest() *KubeDeploymentDetailRequest {
	return &KubeDeploymentDetailRequest{}
}

type KubeDeploymentDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeDeploymentDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 删除 ----------------------

func NewKubeDeploymentDeleteRequest() *KubeDeploymentDeleteRequest {
	return &KubeDeploymentDeleteRequest{}
}

type KubeDeploymentDeleteRequest struct {
	KubeCommonRequest
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty" valid:"grace_period_seconds"` // 优雅终止时间（秒）
	Force              bool   `json:"force,omitempty" valid:"force"`                               // 是否强制删除
}

func ValidKubeDeploymentDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":      []string{"required"},
		"namespace": []string{"required"},
	}
	messages := govalidator.MapData{
		"name":      []string{"required: name不能为空"},
		"namespace": []string{"required: namespace不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 镜像更新 ----------------------
func NewKubeDeploymentUpdateImageRequest() *KubeDeploymentUpdateImageRequest {
	return &KubeDeploymentUpdateImageRequest{}
}

type KubeDeploymentUpdateImageRequest struct {
	KubeCommonRequest
	Container string `json:"container" valid:"container"` // 目标容器名称
	Image     string `json:"image" valid:"image"`         // 新镜像地址，例如 nginx:1.27
}

func ValidKubeDeploymentUpdateImageRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"container": []string{"required"},
		"image":     []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"container": []string{"required: container 不能为空"},
		"image":     []string{"required: image 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 创建和server服务 ----------------------

func NewKubeDeploymentCreateSvcRequest() *KubeDeploymentCreateSvcRequest {
	return &KubeDeploymentCreateSvcRequest{}
}

type KubeDeploymentCreateSvcRequest struct {
	KubeCommonRequest
}

func ValidKubeDeploymentCreateSvcRequest(data interface{}, ctx context.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}

	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}

	return valid.ValidateOptions(data, rules, messages)
}
