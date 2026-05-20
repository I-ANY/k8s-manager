package pod

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
)

// DeletePod 删除Pod
func DeletePod(client *k8s.Client, namespace, name string) error {
	if err := client.Interface.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		client.Logger.Errorf("删除Pod失败，err：%v", err)
		return err
	}

	client.Logger.Info("删除Pod成功")
	return nil
}
