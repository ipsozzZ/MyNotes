# Kubernetes核心知识体系清单：

---

### 一、核心架构（Architecture）
1. 集群架构
   - Master节点组件：
     - API Server（集群操作入口）
     - etcd（分布式键值存储，保存集群状态）
     - Controller Manager（核心控制器）
     - Scheduler（资源调度决策）
   - Worker节点组件：
     - Kubelet（节点代理）
     - Kube-proxy（网络流量管理）
     - Container Runtime（Docker/Containerd/CRI-O）

2. 声明式API
   - 通过YAML/JSON描述期望状态（Desired State）
   - 控制器模式（Control Loop）驱动系统收敛到目标状态

---

### 二、核心对象模型（Core Objects）
1. 工作负载（Workloads）
   - Pod：最小调度单元（1或多个容器共享资源）
   - Deployment：无状态应用管理（滚动更新/回滚）
   - StatefulSet：有状态应用（稳定网络标识/持久存储）
   - DaemonSet：节点级守护进程（如日志收集）
   - Job/CronJob：批处理任务

2. 服务发现与网络
   - Service：
     - ClusterIP（内部访问）
     - NodePort（节点端口暴露）
     - LoadBalancer（云厂商负载均衡集成）
   - Ingress：七层流量路由（需配合Ingress Controller）
   - NetworkPolicy：网络策略（Pod间通信控制）

3. 存储管理
   - PersistentVolume (PV)：集群存储资源抽象
   - PersistentVolumeClaim (PVC)：用户存储请求
   - StorageClass：动态卷配置模板
   - ConfigMap/Secret：配置与敏感数据管理

4. 配置管理
   - ConfigMap：非敏感配置分离
   - Secret：敏感数据加密存储（Base64编码）
   - Env/Volume：配置注入方式

---

### 三、关键机制（Key Mechanisms）
1. 调度机制
   - 标签与选择器（Labels & Selectors）
   - 污点与容忍（Taints & Tolerations）
   - 节点亲和性/反亲和性（Node Affinity）
   - Pod亲和性/反亲和性（Pod Affinity）

2. 资源管理
   - Requests/Limits（CPU/Memory配额）
   - QoS等级（Guaranteed/Burstable/BestEffort）
   - Horizontal Pod Autoscaler（HPA，水平自动扩缩）
   - Vertical Pod Autoscaler（VPA，垂直自动扩缩）

3. 安全控制
   - RBAC（基于角色的访问控制）
   - ServiceAccount（Pod身份标识）
   - Pod Security Policies（PSP，已弃用）/Pod Security Admission（PSA）
   - Network Policies（网络隔离）

---

### 四、运维重点（Operations）
1. 集群运维
   - 日志收集方案（EFK/Loki）
   - 监控体系（Prometheus + Grafana）
   - 故障排查命令：
     - `kubectl describe` / `kubectl logs` / `kubectl exec`
     - `kubectl get events --watch`
   - 备份恢复（etcd快照/Velero工具）

2. 升级策略
   - 滚动更新（RollingUpdate）
   - 蓝绿部署（Blue-Green）
   - 金丝雀发布（Canary）

3. 扩展机制
   - Custom Resource Definitions (CRD)
   - Operator模式（自动化复杂应用管理）
   - Admission Controllers（请求拦截校验）

---

### 五、高级特性（Advanced Topics）
1. 服务网格
   - Istio/Linkerd的Sidecar模式集成
   - 流量管理/熔断/链路追踪

2. Serverless扩展
   - Knative（基于K8s的Serverless框架）
   - KEDA（事件驱动自动扩缩）

3. 多集群管理
   - Cluster API（声明式集群生命周期管理）
   - Kubefed（联邦集群管理）

---

### 六、生态工具（Ecosystem）
1. 包管理
   - Helm（应用模板化部署）
   - Kustomize（声明式资源配置覆盖）

2. CI/CD集成
   - Argo CD（GitOps持续交付）
   - Tekton（云原生CI/CD流水线）

3. 安全工具
   - Trivy（容器漏洞扫描）
   - Falco（运行时威胁检测）

---

### 七、学习路径建议
1. 官方文档：kubernetes.io/docs + interactive tutorials
2. 实践社区：Kubernetes Slack、CNCF项目生态
3. 书籍推荐：
   - 《Kubernetes in Action》
   - 《Programming Kubernetes》
4. 认证路径：CKA（认证管理员）→ CKAD（应用开发者）→ CKS（安全专家）

---

通过这个知识框架，你可以逐步深入每个模块。建议先掌握核心对象和调度机制，再通过实际部署应用（如部署一个带数据库的Web应用）将各组件串联理解，最后再探索高级特性。