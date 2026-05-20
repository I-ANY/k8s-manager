package ingress

import (
	"context"
	"fmt"
	"time"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateIngress 创建 Ingress
func CreateIngress(client *k8s.Client, ctx context.Context, req *requests.KubeIngressCreateRequest) (*networkingv1.Ingress, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ing := BuildIngressFromReq(req)

	created, err := client.Interface.NetworkingV1().
		Ingresses(req.Namespace).
		Create(ctx, ing, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("ingress %q already exists in namespace %q", ing.Name, ing.Namespace)
		}
		client.Logger.Errorf("create ingress failed: %v", err)
		return nil, err
	}

	client.Logger.Infof("ingress %q created successfully", created.Name)
	return created, nil
}
