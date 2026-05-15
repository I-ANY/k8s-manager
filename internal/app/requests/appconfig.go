package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

//
// ================== DTO Struct 定义 ==================
//

// EnvVarKV 用于表示请求里的环境变量（与 CRD 中的 EnvVarKV 结构对应）
type EnvVarKV struct {
	Name  string `json:"name"  valid:"required~环境变量name不能为空"`
	Value string `json:"value" valid:"required~环境变量value不能为空"`
}

// 创建 AppConfig
type KubeAppConfigCreateRequest struct {
	ClusterID     uint32     `json:"cluster_id"      valid:"required~cluster_id不能为空"`
	Namespace     string     `json:"namespace"       valid:"required~namespace不能为空"`
	AppName       string     `json:"app_name"        valid:"required~app_name不能为空"`
	Image         string     `json:"image"           valid:"required~image不能为空"`
	Replicas      *int32     `json:"replicas"`       // 可选；不传由 Controller 默认 1
	Env           []EnvVarKV `json:"env"`            // 与 CRD 中的 Env 类型保持一致
	EnableMetrics bool       `json:"enable_metrics"` // 对应 CRD 的 enableMetrics
	Strategy      string     `json:"strategy"`       // RollingUpdate / Recreate，可选
}

func NewKubeAppConfigCreateRequest() *KubeAppConfigCreateRequest {
	return &KubeAppConfigCreateRequest{}
}

// 更新 AppConfig
type KubeAppConfigUpdateRequest struct {
	ClusterID     uint32     `json:"cluster_id"      valid:"required~cluster_id不能为空"`
	Namespace     string     `json:"namespace"       valid:"required~namespace不能为空"`
	AppName       string     `json:"app_name"        valid:"required~app_name不能为空"`
	Image         string     `json:"image"`          // 可选：如果为空就不更新 image
	Replicas      *int32     `json:"replicas"`       // 可选
	Env           []EnvVarKV `json:"env"`            // 可选：不传则不修改 env
	EnableMetrics bool       `json:"enable_metrics"` // 可选：按你的业务决定怎么更新
	Strategy      string     `json:"strategy"`       // RollingUpdate / Recreate，可选
}

func NewKubeAppConfigUpdateRequest() *KubeAppConfigUpdateRequest {
	return &KubeAppConfigUpdateRequest{}
}

// 用于 detail / delete
type KubeAppConfigNameRequest struct {
	ClusterID uint32 `form:"cluster_id" valid:"required~cluster_id不能为空"`
	Namespace string `form:"namespace"  valid:"required~namespace不能为空"`
	AppName   string `form:"app_name"   valid:"required~app_name不能为空"`
}

func NewKubeAppConfigNameRequest() *KubeAppConfigNameRequest {
	return &KubeAppConfigNameRequest{}
}

// 列表查询
type KubeAppConfigListRequest struct {
	ClusterID uint32 `form:"cluster_id" valid:"required~cluster_id不能为空"`
	Namespace string `form:"namespace"` // 可选：空字符串表示查询所有 ns
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

func NewKubeAppConfigListRequest() *KubeAppConfigListRequest {
	return &KubeAppConfigListRequest{}
}

//
// ================== 校验方法 ==================
//

// Create 校验
func ValidKubeAppConfigCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required", "numeric", "min:1"},
		"namespace":  []string{"required"},
		"app_name":   []string{"required"},
		"image":      []string{"required"},
		// strategy 可选，但如果传了必须是 RollingUpdate 或 Recreate
		"strategy": []string{"omitempty", "in:RollingUpdate,Recreate"},
		// env 为数组时的校验（thedevsaddam/govalidator 支持这种写法）
		"env.*.name":  []string{"required"},
		"env.*.value": []string{"required"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"namespace": {"required:namespace不能为空"},
		"app_name":  {"required:app_name不能为空"},
		"image":     {"required:image不能为空"},

		"strategy": {
			"in:strategy必须是RollingUpdate或Recreate",
		},

		"env.*.name": {
			"required:环境变量name不能为空",
		},
		"env.*.value": {
			"required:环境变量value不能为空",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// Update 校验（比 Create 宽松一点：image / replicas / env 可选）
func ValidKubeAppConfigUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required", "numeric", "min:1"},
		"namespace":  []string{"required"},
		"app_name":   []string{"required"},
		// image / replicas / env 都是可选；如果你想强制 image 必须传，可以加 required
		"strategy":    []string{"omitempty", "in:RollingUpdate,Recreate"},
		"env.*.name":  []string{"omitempty", "required"},
		"env.*.value": []string{"omitempty", "required"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"namespace": {"required:namespace不能为空"},
		"app_name":  {"required:app_name不能为空"},

		"strategy": {
			"in:strategy必须是RollingUpdate或Recreate",
		},

		"env.*.name": {
			"required:环境变量name不能为空",
		},
		"env.*.value": {
			"required:环境变量value不能为空",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// detail/delete 使用
func ValidKubeAppConfigNameRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required", "numeric", "min:1"},
		"namespace":  []string{"required"},
		"app_name":   []string{"required"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"namespace": {"required:namespace不能为空"},
		"app_name":  {"required:app_name不能为空"},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// 列表查询使用
func ValidKubeAppConfigListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required", "numeric", "min:1"},
		// namespace 可选
		"page":  []string{"numeric", "min:1"},
		"limit": []string{"numeric", "min:1"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"page": {
			"numeric:page必须是数字",
			"min:page必须大于0",
		},
		"limit": {
			"numeric:limit必须是数字",
			"min:limit必须大于0",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}
