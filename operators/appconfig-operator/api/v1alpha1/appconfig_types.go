package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//
// ================== Spec 相关 ==================
//

// EnvVarKV 用于表示一个 name/value 形式的环境变量
type EnvVarKV struct {
	// 环境变量名
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// 环境变量值
	// 如果你希望允许空字符串，可以去掉 MinLength 这行
	// +kubebuilder:validation:MinLength=1
	Value string `json:"value"`
}

// AppConfigSpec defines the desired state of AppConfig
type AppConfigSpec struct {
	// 应用名称，用于 Deployment / Pod / Label 等
	// +kubebuilder:validation:MinLength=1
	AppName string `json:"appName"`

	// 应用镜像
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// 副本数，可选；如果为空，controller 里默认 1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// 环境变量（name/value 列表）
	// +optional
	Env []EnvVarKV `json:"env,omitempty"`

	// 是否开启 metrics（比如决定是否加 sidecar、port 等）
	// +optional
	EnableMetrics bool `json:"enableMetrics,omitempty"`

	// 部署策略，比如 RollingUpdate / Recreate
	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	// +optional
	Strategy string `json:"strategy,omitempty"`
}

//
// ================== Status 相关 ==================
//

// AppConfigStatus defines the observed state of AppConfig.
type AppConfigStatus struct {
	// 当前阶段：Progressing / Available / Failed 等
	// +optional
	Phase string `json:"phase,omitempty"`

	// 最近一次状态变化的说明信息
	// +optional
	Message string `json:"message,omitempty"`

	// 最近一次状态更新的时间
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// 标准 Conditions，方便 kubectl / 控制器生态复用
	// 使用 listType=map + listMapKey=types 来支持 patch 合并
	// +listType=map
	// +listMapKey=types
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//
// ================== 顶层对象 & 列表 ==================
//

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AppConfig is the Schema for the appconfigs API
type AppConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of AppConfig
	// +required
	Spec AppConfigSpec `json:"spec"`

	// status defines the observed state of AppConfig
	// +optional
	Status AppConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppConfigList contains a list of AppConfig
type AppConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppConfig{}, &AppConfigList{})
}
