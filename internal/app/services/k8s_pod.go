package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/common"
	pod2 "k8soperation/pkg/k8s/pod"
)

// PodList 获取Pod列表
func (s *Services) KubePodList(ctx context.Context, param *requests.KubePodListRequest) ([]corev1.Pod, error) {
	pods, err := pod2.GetPodList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
	if err != nil {
		global.Logger.Errorf("GetPodList error: %v", err)
		return nil, err
	}

	global.Logger.Infof("GetPodList success")
	return pods, nil
}

// PodDelete 从Pod列表中删除Pod
// services/pod_service.go
func (s *Services) KubePodDelete(param *requests.KubePodDeleteRequest) error {
	// 1) 组装 DeleteOptions
	opts := metav1.DeleteOptions{}

	// 先根据是否传入 grace_seconds 决定
	if param.GraceSeconds != nil {
		// 显式传了（可能是 0、>0）
		opts.GracePeriodSeconds = param.GraceSeconds
	} else {
		// 没传就用一个平台默认值（也可以保持 nil 让 K8s 用对象默认）
		def := int64(30)
		opts.GracePeriodSeconds = &def
	}

	// 2) 如果 Force=true，覆盖为 0，并采用后台级联删除
	if param.Force {
		zero := int64(0)
		opts.GracePeriodSeconds = &zero
		policy := metav1.DeletePropagationBackground
		opts.PropagationPolicy = &policy
	}

	// 3) 调 K8s API
	err := global.KubeClient.CoreV1().
		Pods(param.Namespace).
		Delete(context.TODO(), param.Name, opts)
	if err != nil {
		global.Logger.Errorf("删除 Pod 失败 ns=%s name=%s : %v", param.Namespace, param.Name, err)
		return err
	}

	// 4) 记录日志（指针打印要判空）
	var g int64 = -1
	if param.GraceSeconds != nil {
		g = *param.GraceSeconds
	}
	global.Logger.Infof("删除 Pod 已提交 ns=%s name=%s force=%v grace=%d", param.Namespace, param.Name, param.Force, g)
	return nil
}

// KubePodUpdate PodUpdate 更新Pod
func (s *Services) KubePodUpdate(param *requests.KubePodUpdateRequest) error {
	if err := pod2.UpdatePod(param.Namespace, param.Name, param.Content); err != nil {
		global.Logger.Errorf("UpdatePod error: %v", err)
		return err
	}
	global.Logger.Infof("UpdatePod success")
	return nil
}

func (s *Services) PatchPodImage(param *requests.PatchPodImageRequest) error {
	// 构造 patch JSON
	// patchObj 定义了一个用于 Kubernetes Pod 镜像更新的补丁对象
	// 使用 map[string]any 类型构建，符合 JSON 结构
	patchObj := map[string]any{
		"spec": map[string]any{
			// containers 是一个切片，包含一个或多个容器定义
			// 这里只更新指定容器的镜像
			"containers": []map[string]string{
				{"name": param.Container, "image": param.NewImage},
			},
		},
	}

	// 将补丁对象序列化为 JSON 格式
	// b 是序列化后的字节数组
	// _ 忽略可能的错误处理（在实际生产代码中应该处理）
	b, _ := json.Marshal(patchObj)

	// 调用 Kubernetes API 执行 Pod 更新操作
	// 使用 Strategic Merge Patch 类型进行部分更新
	_, err := global.KubeClient.CoreV1().
		Pods(param.Namespace). // 指定命名空间
		Patch(context.TODO(),  // 上下文
			param.Name,                    // Pod 名称
			types.StrategicMergePatchType, // 补丁类型
			b,                             // 补丁数据
			metav1.PatchOptions{},         // 补丁选项
		)
	// 检查更新操作是否出错
	if err != nil {
		// 记录错误日志并返回错误
		global.Logger.Errorf("PatchPodImage error: %v", err)
		return err
	}

	// 记录成功日志
	global.Logger.Infof("PatchPodImage success: ns=%s kube_pod=%s container=%s image=%s",
		param.Namespace, param.Name, param.Container, param.NewImage)
	// 返回 nil 表示操作成功
	return nil
}

// KubePodDetail PodDetail 获取单个 Pod
func (s *Services) KubePodDetail(param *requests.KubePodDetailRequest) (*corev1.Pod, error) {
	p, err := pod2.GetPodDetail(param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)
	return p, nil
}

// GetContainerNames 获取容器名称列表
func (s *Services) GetContainerNames(param *requests.KubePodDetailRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod2.GetPodDetail(param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取容器名称列表
	containersNames := common.GetContainerNames(&p.Spec)
	return containersNames, nil
}

// GetInitContainerNames 获取Init容器名称列表
func (s *Services) GetInitContainerNames(param *requests.KubeCommonRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod2.GetPodDetail(param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取Init容器
	initContainerNames := common.GetInitContainerNames(&p.Spec)
	return initContainerNames, nil
}

// GetContainerImages 获取容器镜像名称列表
func (s *Services) GetContainerImages(param *requests.KubePodDetailRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod2.GetPodDetail(param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取容器镜像名称列表
	containerImages := common.GetContainerImages(&p.Spec)
	return containerImages, nil
}

// GetInitContainerImages 获取Init容器镜像名称列表
func (s *Services) GetInitContainerImages(param *requests.KubeCommonRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod2.GetPodDetail(param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取Init容器镜像名称列表
	initContainerImages := common.GetInitContainerImages(&p.Spec)
	return initContainerImages, nil
}

// 获取所有容器名称（常规 + Init）
func (s *Services) KubePodAllContainerNames(param *requests.KubePodDetailRequest) ([]string, error) {
	// 1. 获取 Pod 对象
	p, err := pod2.GetPodDetail(param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}

	// 2. 从 PodSpec 中提取容器名称
	names := common.GetAllContainerNames(&p.Spec)

	global.Logger.Infof("GetAllContainerNames success: %s/%s -> %v", param.Namespace, param.Name, names)
	return names, nil
}

// 获取所有容器镜像（常规 + Init）
func (s *Services) KubePodAllContainerImages(param *requests.KubePodDetailRequest) ([]string, error) {
	// 1. 获取 Pod 对象
	p, err := pod2.GetPodDetail(param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}

	// 2. 从 PodSpec 中提取容器镜像
	images := common.GetAllContainerImages(&p.Spec)

	global.Logger.Infof("GetAllContainerImages success: %s/%s -> %v", param.Namespace, param.Name, images)
	return images, nil
}

func (s *Services) KubePodLog(ctx context.Context, name, namespace, container string, tail int64) (string, error) {
	// 统一默认/上限
	// 设置tail行数，如果未指定则使用默认值
	t := tail
	if t <= 0 {
		t = global.PodLogSetting.TailDefault
	}
	// 确保tail行数不超过最大限制
	if max := global.PodLogSetting.TailMax; max > 0 && t > max {
		t = max
	}

	// 创建Pod日志选项配置
	opts := &corev1.PodLogOptions{
		Container:  container,                       // 指定容器名称
		TailLines:  &t,                              // 设置从日志末尾开始的行数
		Timestamps: global.PodLogSetting.Timestamps, // 是否显示时间戳
		Previous:   global.PodLogSetting.Previous,   // 是否显示 previous 容器的日志
		Follow:     false,                           // 不启用流式日志模式
	}
	// 如果设置了字节限制，则添加到选项中（仅在一次性模式下生效）
	if lb := global.PodLogSetting.LimitBytes; lb > 0 {
		opts.LimitBytes = &lb // 仅一次性模式生效
	}

	// 获取Pod日志流
	rc, err := global.KubeClient.CoreV1().Pods(namespace).GetLogs(name, opts).Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("open log stream: %w", err)
	}
	defer rc.Close() // 确保资源被正确关闭

	// 将日志流读取到缓冲区
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return "", fmt.Errorf("read log: %w", err)
	}
	// 记录成功获取日志的信息，包括日志长度
	global.Logger.Infof("get kube_pod log success, len=%d", buf.Len())
	return buf.String(), nil
}

func (s *Services) KubePodLogStream(ctx context.Context, name, namespace, container string, tail int64) (io.ReadCloser, error) {
	// 设置日志获取的tail行数
	t := tail
	// 如果tail小于0，则设置为0
	if t < 0 {
		t = 0
	}
	// 如果tail为0，则使用默认的tail行数
	if t == 0 {
		t = global.PodLogSetting.TailDefault
	}
	// 检查并确保tail行数不超过最大限制
	if max := global.PodLogSetting.TailMax; max > 0 && t > max {
		t = max
	}

	// 创建Pod日志选项
	opts := &corev1.PodLogOptions{
		Container:  container,                            // 指定容器名称
		TailLines:  &t,                                   // 设置要获取的日志行数
		Timestamps: global.PodLogSetting.Timestamps,      // 是否显示时间戳
		Previous:   global.PodLogSetting.Previous,        // 是否获取之前容器的日志
		Follow:     global.PodLogSetting.EnableStreaming, // 关键
		// Follow 模式不要设置 LimitBytes，否则会被截断
	}
	return global.KubeClient.CoreV1().Pods(namespace).GetLogs(name, opts).Stream(ctx)
}

// GetPodLog 用于获取指定Pod的日志内容
// 参数:
//
//	name: Pod的名称
//	namespace: Pod所在的命名空间
//	container: Pod中容器的名称
//	tailLine: 要获取的日志行数，从最新日志开始计算
//
// 返回值:
//
//	string: Pod的日志内容
//	error: 错误信息，如果获取日志失败则返回错误
func (s *Services) GetPodLog(name, namespace, container string, tailLine int64) (string, error) {
	// 创建Pod日志选项，指定容器和要获取的日志行数
	options := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLine,
	}
	// 创建获取Pod日志的请求
	req := global.KubeClient.CoreV1().Pods(namespace).GetLogs(name, options)
	// 通过流式方式获取Pod日志
	podLog, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}

	// 确保在函数返回前关闭Pod日志流
	defer podLog.Close()

	// 将 response body 写入到缓冲区，目的是为了转换成可读的string类型
	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, podLog)
	if err != nil {
		return "", err
	}
	// 将缓冲区内容转换为字符串格式
	log := buff.String()

	return log, nil
}
