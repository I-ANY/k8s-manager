package pod

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
)

// DeletePod 删除Pod
func DeletePod(namespace, name string) error {
	if err := global.KubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		global.Logger.Errorf("删除Pod失败，err：%v", err)
		return err
	}

	global.Logger.Info("删除Pod成功")
	return nil
}
