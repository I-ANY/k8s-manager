package errorcode

// ===== Pod 相关（201xxx）=====
var (
	ErrorPodNotFound     *Error
	ErrorPodCreateFail   *Error
	ErrorPodDeleteFail   *Error
	ErrorPodUpdateFail   *Error
	ErrorPodQueryFail    *Error // 列表 / 单查失败
	ErrorPodLogFail      *Error // 获取日志失败
	ErrorK8sPodPatchFail *Error // 更新镜像失败
)

func registerPod() {
	ErrorPodNotFound = NewError(201001, "Pod 不存在")
	ErrorPodCreateFail = NewError(201002, "创建 Pod 失败")
	ErrorPodDeleteFail = NewError(201003, "删除 Pod 失败")
	ErrorPodUpdateFail = NewError(201004, "更新 Pod 失败")
	ErrorPodQueryFail = NewError(201005, "查询 Pod 失败")
	ErrorPodLogFail = NewError(201006, "获取 Pod 日志失败")
	ErrorK8sPodPatchFail = NewError(200201, "更新 Pod 镜像失败")
}
