package secret

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

// CreateSecret 创建 Secret（支持 Opaque / TLS / dockerconfigjson）
// - req.Data 传“明文”，这里用 StringData 让 APIServer 负责编码
func CreateSecret(ctx context.Context, req *requests.KubeSecretCreateRequest) (*corev1.Secret, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sec, err := BuildSecretFromReq(req)
	if err != nil {
		return nil, err
	}

	created, err := global.KubeClient.CoreV1().
		Secrets(req.Namespace).
		Create(ctx, sec, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("secret %q already exists in namespace %q", sec.Name, sec.Namespace)
		}
		global.Logger.Errorf("create secret failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("secret %q created successfully", created.Name)
	return created, nil
}
