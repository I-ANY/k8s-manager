package configmap

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"time"
)

// CreateConfigMap 创建 ConfigMap（对应 CreateSecret 的风格）
func CreateConfigMap(ctx context.Context, req *requests.KubeConfigMapCreateRequest) (*corev1.ConfigMap, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cm, err := BuildConfigMapFromReq(req)
	if err != nil {
		return nil, err
	}
	created, err := global.KubeClient.CoreV1().
		ConfigMaps(req.Namespace).
		Create(c, cm, metav1.CreateOptions{})
	if err != nil {
		// 按需处理 AlreadyExists / 其他错误
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("configmap %q already exists in namespace %q", cm.Name, cm.Namespace)
		}
		global.Logger.Errorf("create configmap failed: %v", err)
		return nil, err
	}
	global.Logger.Infof("configmap %q created successfully", created.Name)
	return created, nil
}
