# 🚀 k8soperation · Kubernetes 后台管理系统（后端）

一个基于 **Go + Gin + Gorm + Zap** 的企业级 Kubernetes 后台管理系统后端，提供对 Kubernetes 各类资源的可视化运维能力，包括 Pod、Deployment、StatefulSet、Node、Ingress、Service、Job、CronJob、PVC、PV、Secret 等资源的完整生命周期管理。

> 📦 **配套 AppConfig Operator（Kubebuilder 项目）请访问：**
>  👉 https://gitee.com/jay-kim/appconfig-operator

系统支持多集群管理、事件聚合、滚动升级、镜像更新、扩缩容、Pod 日志流、节点驱逐/隔离、PVC 扩容等能力，是构建企业 K8s 管控平台的优秀后端基础设施。

------

## ✨ 核心特性

### 🧩 系统通用能力

- 配置化加载（YAML / ENV）
- JWT 鉴权 + 刷新机制
- Zap 双日志系统（系统日志 / 业务日志）
- Swagger 在线 API 文档（支持 Standalone）
- 健康检查与优雅关闭
- 标准化控制器 / 服务 / DAO 分层
- 全局异常拦截（中间件）

------

## ☸ Kubernetes 高级能力（全部已实现）

### Deployment 管理

- CRUD、扩缩容、镜像更新、滚动升级
- 滚动重启、基于 ReplicaSet 的版本回滚
- Pods 列表、事件聚合、历史版本查询

### Pod 管理

- 列表、详情、日志（流式/非流）
- 镜像 Patch、事件查询、强制删除

### StatefulSet / DaemonSet

- CRUD、扩缩容、镜像更新
- ControllerRevision 回滚

### Service / Ingress

- CRUD
- Strategic / JSON Merge Patch
- TLS 配置、事件聚合

### Job / CronJob

- Job：创建 / 删除 / 状态查询
- CronJob：启停、删除、历史 Job 查询

### Secret / PVC / PV / ConfigMap / StorageClass

- 全生命周期管理
- PVC 扩容、PV ReclaimPolicy 修改
- ConfigMap Patch、StorageClass CRUD

### Node 高级管理

- Cordon / Uncordon
- Drain（驱逐可驱逐 Pod）
- Pod Evict（支持 gracePeriod）
- 节点 Metrics、Pods 列表

### Event 事件聚合

- Pod / Deploy / StatefulSet / Node 等资源
- 支持排障快速定位（Backoff、PullError、Unschedulable）

------

## 🧩 多集群管理

- 保存 kubeconfig
- 连通性检测
- 多集群切换（clusterId）
- 自动创建对应 client-go

适合企业多集群统一管控场景。

------

## 📦 项目结构（真实仓库对应）

```bash
k8soperation/
├── cmd/
├── configs/
├── docs/
│   ├── 📄 K8sOperation 后台系统部署文档.md   <--（部署文档）
├── global/
├── initialize/
├── internal/
│   ├── app/
│   ├── errorcode/
│   ├── health/
│   └── k8soperation/
├── pkg/
├── deploy/
└── storage/
```

------

## ⚙️ 快速启动

```bash
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation
make deploy
./bin/k8soperation
```

访问 Swagger：

```bash
http://localhost:8080/swagger
http://localhost:8080/swagger-standalone
```

------

## 📄 部署文档（强烈推荐阅读）

官方部署说明文档（包括 Docker 部署、Containerd 部署、进程部署、K8s 运行方式）：

👉 **K8sOperation 后台系统部署文档**
[https://gitee.com/jay-kim/k8s_operation/blob/master/docs/%F0%9F%93%84%20K8sOperation%20%E5%90%8E%E5%8F%B0%E7%B3%BB%E7%BB%9F%E9%83%A8%E7%BD%B2%E6%96%87%E6%A1%A3.md](https://gitee.com/jay-kim/k8s_operation/blob/master/docs/📄 K8sOperation 后台系统部署文档.md)

内容包括：

- 构建二进制
- Docker / Containerd 镜像构建
- 使用 Systemd 管理服务
- k8s Deployment / Service 部署示例
- 参数说明与优化建议
- 生产环境目录规划

------

## 🔗 关联项目（推荐配套使用）

### 📘 AppConfig Operator

（Kubebuilder 开发，用于管理自定义资源 AppConfig）
👉 https://gitee.com/jay-kim/appconfig-operator

Operator → 管理 AppConfig CRD
k8soperation → 提供 HTTP API/Web 后台

两者解耦，便于独立演进。

------

## ⭐ Star / Watch / Fork

如果本项目对你有帮助，非常欢迎：

- ⭐ **Star**
- 👀 **Watch**
- 🍴 **Fork**

你的支持是我持续完善的最大动力！

------

## 📜 License

MIT License