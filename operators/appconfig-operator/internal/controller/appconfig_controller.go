package controller

import (
	"context"
	"time"

	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/util/retry" // ✅ 新增：用于 RetryOnConflict

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ConditionTypeAvailable   = "Available"
	ConditionTypeProgressing = "Progressing"
	ConditionTypeFailed      = "Failed"
)

// AppConfigReconciler reconciles a AppConfig object
type AppConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operation.operation.top,resources=appconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operation.operation.top,resources=appconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operation.operation.top,resources=appconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *AppConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. 读取 AppConfig
	var app appv1alpha1.AppConfig
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		if errors.IsNotFound(err) {
			// 被删了，什么也不做
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. 根据 AppConfig.Spec 构建期望的 Deployment
	desired := BuildDeployment(&app)
	if err := controllerutil.SetControllerReference(&app, desired, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 3. 获取现有 Deployment
	var deploy appsv1.Deployment
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      desired.Name,
	}, &deploy)

	if errors.IsNotFound(err) {
		// 3.1 不存在 → 创建
		if err := r.Create(ctx, desired); err != nil {
			logger.Error(err, "create deployment failed")
			_ = r.updateStatus(ctx, &app, "Failed", "create deployment failed: "+err.Error())
			return ctrl.Result{}, err
		}
		logger.Info("deployment created", "name", desired.Name)
		// 刚创建出来，先用 desired 作为本次 Reconcile 的 deploy 视图
		deploy = *desired
	} else if err != nil {
		// 3.2 读取 Deployment 出错
		return ctrl.Result{}, err
	} else {
		// 3.3 已存在 → 判断是否需要更新
		if needUpdateDeployment(&deploy, desired) {
			deploy.Spec = desired.Spec
			if err := r.Update(ctx, &deploy); err != nil {
				logger.Error(err, "update deployment failed")
				_ = r.updateStatus(ctx, &app, "Failed", "update deployment failed: "+err.Error())
				return ctrl.Result{}, err
			}
			logger.Info("deployment updated", "name", deploy.Name)
		}
	}

	// 4. 统一根据 Deployment 同步 AppConfig.Status（带冲突重试）
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// 每次重试都拿最新的 AppConfig，避免 resourceVersion 冲突
		var latest appv1alpha1.AppConfig
		if err := r.Get(ctx, req.NamespacedName, &latest); err != nil {
			return err
		}

		// 用最新对象计算状态
		r.syncStatusFromDeployment(&latest, &deploy)

		// 只更新 status 子资源
		return r.Status().Update(ctx, &latest)
	}); err != nil {
		logger.Error(err, "update appconfig status failed")
		return ctrl.Result{}, err
	}

	// 5. 再读一次当前 AppConfig，看 Phase 决定是否需要 RequeueAfter
	var cur appv1alpha1.AppConfig
	if err := r.Get(ctx, req.NamespacedName, &cur); err != nil {
		logger.Error(err, "get appconfig after status update failed")
		return ctrl.Result{}, err
	}

	// 只有在还没 Running 的时候才需要周期性 Requeue
	if cur.Status.Phase != "Running" {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	return ctrl.Result{}, nil
}

// updateStatus 只负责设置 Phase/Message/LastUpdateTime + Conditions，具体 Phase 由调用处传入
// ✅ 已加 RetryOnConflict，避免并发更新 status 时的冲突
func (r *AppConfigReconciler) updateStatus(
	ctx context.Context,
	app *appv1alpha1.AppConfig,
	phase, msg string,
) error {

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// 用 app 的 key 拿最新对象，避免旧 resourceVersion
		var latest appv1alpha1.AppConfig
		if err := r.Get(ctx, client.ObjectKeyFromObject(app), &latest); err != nil {
			return err
		}

		now := metav1.Now()

		latest.Status.Phase = phase
		latest.Status.Message = msg
		latest.Status.LastUpdateTime = now

		// 根据 Phase 简单维护 Conditions（操作 latest.Status）
		switch phase {
		case "Available":
			setCondition(&latest.Status, ConditionTypeAvailable, metav1.ConditionTrue, "AppAvailable", msg, now)
			setCondition(&latest.Status, ConditionTypeProgressing, metav1.ConditionFalse, "AppStable", "app is stable", now)
			setCondition(&latest.Status, ConditionTypeFailed, metav1.ConditionFalse, "NoError", "no error", now)
		case "Failed":
			setCondition(&latest.Status, ConditionTypeAvailable, metav1.ConditionFalse, "AppUnavailable", "app not available", now)
			setCondition(&latest.Status, ConditionTypeProgressing, metav1.ConditionFalse, "AppStopped", "app is not progressing", now)
			setCondition(&latest.Status, ConditionTypeFailed, metav1.ConditionTrue, "Error", msg, now)
		default: // Progressing / 其他
			setCondition(&latest.Status, ConditionTypeAvailable, metav1.ConditionFalse, "AppNotReady", "app not ready", now)
			setCondition(&latest.Status, ConditionTypeProgressing, metav1.ConditionTrue, "Reconciling", msg, now)
			setCondition(&latest.Status, ConditionTypeFailed, metav1.ConditionFalse, "NoError", "no error", now)
		}

		return r.Status().Update(ctx, &latest)
	})
}

// syncStatusFromDeployment 根据 Deployment 的实际状态推导 AppConfig.Status
func (r *AppConfigReconciler) syncStatusFromDeployment(app *appv1alpha1.AppConfig, deploy *appsv1.Deployment) {
	now := metav1.Now()

	ready := deploy.Status.ReadyReplicas
	desired := deploy.Spec.Replicas
	if desired == nil {
		var one int32 = 1
		desired = &one
	}

	var phase, msg string

	// 判断整体状态
	if ready >= *desired {
		phase = "Running"
		msg = "all pods are ready"

		setCondition(&app.Status, ConditionTypeAvailable, metav1.ConditionTrue, "AppReady", msg, now)
		setCondition(&app.Status, ConditionTypeProgressing, metav1.ConditionFalse, "Reconciled", "app is stable", now)
		setCondition(&app.Status, ConditionTypeFailed, metav1.ConditionFalse, "NoError", "no error", now)
	} else {
		phase = "Progressing"
		if ready == 0 {
			msg = "deployment pods creating"
		} else {
			msg = "some pods not ready"
		}

		setCondition(&app.Status, ConditionTypeAvailable, metav1.ConditionFalse, "AppNotReady", msg, now)
		setCondition(&app.Status, ConditionTypeProgressing, metav1.ConditionTrue, "Reconciling", msg, now)
		setCondition(&app.Status, ConditionTypeFailed, metav1.ConditionFalse, "NoError", "no error", now)
	}

	// 状态没变化就不要写 LastUpdateTime，避免没必要的 status 更新
	if app.Status.Phase != phase || app.Status.Message != msg {
		app.Status.Phase = phase
		app.Status.Message = msg
		app.Status.LastUpdateTime = now
	}
}

// needUpdateDeployment 只比较 spec（你可以根据需要更细化）
func needUpdateDeployment(existing, desired *appsv1.Deployment) bool {
	return !equality.Semantic.DeepEqual(existing.Spec, desired.Spec)
}

// setCondition 根据 types 更新/插入 Condition
func setCondition(status *appv1alpha1.AppConfigStatus, condType string, condStatus metav1.ConditionStatus,
	reason, message string, now metav1.Time) {

	newCond := metav1.Condition{
		Type:               condType,
		Status:             condStatus,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}

	// 如果已存在同 types，替换
	for i, c := range status.Conditions {
		if c.Type == condType {
			status.Conditions[i] = newCond
			return
		}
	}

	// 不存在则追加
	status.Conditions = append(status.Conditions, newCond)
}

// BuildDeployment 根据当前 AppConfig.Spec 构建 Deployment
func BuildDeployment(app *appv1alpha1.AppConfig) *appsv1.Deployment {
	labels := map[string]string{
		"app.kubernetes.io/name":     app.Spec.AppName,
		"app.kubernetes.io/instance": app.Name,
	}

	// 处理副本数，默认 1
	var replicas int32 = 1
	if app.Spec.Replicas != nil {
		replicas = *app.Spec.Replicas
	}

	// EnvVarKV → []corev1.EnvVar
	var envs []corev1.EnvVar
	for _, e := range app.Spec.Env {
		envs = append(envs, corev1.EnvVar{
			Name:  e.Name,
			Value: e.Value,
		})
	}

	// Deployment Strategy
	var strategy appsv1.DeploymentStrategy
	switch app.Spec.Strategy {
	case "Recreate":
		strategy = appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		}
	default: // 默认 RollingUpdate
		strategy = appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			// 可以根据需要补 RollingUpdate 的详细参数（MaxUnavailable/MaxSurge）
		}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Spec.AppName, // 也可以用 app.Name，看你的命名策略
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Strategy: strategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            app.Spec.AppName,
							Image:           app.Spec.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             envs,
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.AppConfig{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
