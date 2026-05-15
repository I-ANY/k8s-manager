# 📘 AppConfig Operator

基于 **Kubebuilder** 开发的 Kubernetes Operator，通过自定义资源 **AppConfig（operation.top/v1alpha1）** 来管理应用的全生命周期。该 Operator 能根据声明式配置自动创建、更新并维护 Deployment、Service 等 Kubernetes 原生对象，使应用交付与配置更加自动化与标准化。

------

## ✨ 功能特性

### 🚀 AppConfig 自定义资源（CRD）

- 通过 `AppConfig` 描述应用（镜像、变量、副本数、策略等）
- 自动生成与维护 Deployment、Service
- 支持 RollingUpdate / recreate 策略
- 配合 OwnerReference 自动级联删除相关资源

### 📊 状态管理

- 使用 **Conditions** 记录当前状态（Available / Progressing / Failed）
- Status 中包含 Phase、Message、LastUpdateTime
- 与 Deployment 实际状态同步

### 🏗 生产级基础能力

- 基于 **Kubebuilder** 的规范工程结构
- 内置：
    - Metrics（/metrics）
    - RBAC 权限控制
    - LeaderElection 高可用
    - Kustomize 部署（base/overlays）

------

## 📦 项目结构

```bash
appconfig-operator/
├── api/                 # CRD 定义
│   └── v1alpha1/
├── internal/controller/  # Reconcile 控制器逻辑
├── config/
│   ├── crd/              # CRD 生成目录
│   ├── default/          # 默认部署模板
│   ├── manager/          # Manager 部署
│   ├── rbac/             # 服务账号与角色
│   ├── prometheus/       # ServiceMonitor
│   └── samples/          # 示例 CR
└── cmd/main.go
```

------

## 🛠 本地开发

### 1. 安装依赖

```
make install   # 安装 CRD
make run       # 本地运行 Operator
```

### 2. 生成代码与 CRD

```bash
make generate
make manifests
```

------

## 🚀 部署 Operator

### 开发环境（不带 TLS）

```bash
kubectl apply -k config/default
```

### 生产环境（启用 TLS、Metrics）

```bash
kubectl apply -k config/overlays/prod
```

------

## 📝 示例 AppConfig

```go
apiVersion: operation.top/v1alpha1
kind: AppConfig
metadata:
  name: demo
  namespace: default
spec:
  appName: demo-app
  image: nginx:1.27
  replicas: 2
  env:
    LOG_LEVEL: info
    ENV: dev
  strategy: RollingUpdate
```

------

## 🔄 Reconcile 工作流

1. 监听 AppConfig 创建/更新/删除事件
2. 构建期望资源（Deployment / Service）
3. 比对差异并进行 Patch / Update
4. 更新 Status（Conditions / Phase / Message）
5. 保障 Deployment 按期望状态运行

------

## 🔮 Roadmap（规划）

- 支持 AppJob / AppCron 自定义资源
- 自动扩缩容（HPA 或自定义规则）
- Canary / Blue-Green 发布策略
- Helm Chart 打包发布
- 多版本 CRD（v1alpha2 → v1beta1）

------

## 🤝 贡献

欢迎提交 issue 或 PR 提出功能改进与代码优化。

------

## 📄 License

Apache License 2.0



## ⭐ 欢迎 Star & Watch！

你的 Star 和 Watch 是我持续更新的动力！
如果你对 Kubernetes Operator、Kubebuilder 或云原生研发感兴趣，欢迎加入一起学习交流！