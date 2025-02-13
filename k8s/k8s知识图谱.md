Kubernetes（k8s）是一个分布式容器编排系统，其核心架构由**控制平面（Control Plane）**和**工作节点（Worker Nodes）**组成。以下是各主要组件及其功能的全景解析，帮助你建立全局知识图谱：

---

### **一、控制平面（Control Plane）**
控制平面负责集群的全局决策和状态管理，是集群的“大脑”。

#### 1. **API Server (kube-apiserver)**
- **核心功能**：集群的**唯一入口**，提供 RESTful API，处理所有操作请求（如创建 Pod、部署 Deployment）。
- **特点**：
  - 所有组件（如 `kubectl`、`kubelet`）都通过 API Server 交互。
  - 无状态设计，可通过水平扩展提高性能。
  - 负责鉴权、限流、审计等安全策略。

#### 2. **Scheduler (kube-scheduler)**
- **核心功能**：负责将新创建的 Pod **调度到合适的 Node** 上运行。
- **决策依据**：
  - 资源需求（CPU、内存）。
  - 节点负载、亲和性/反亲和性规则（如将 Pod 分散在不同区域）。
  - 硬件/软件约束（如 GPU 需求）。

#### 3. **Controller Manager (kube-controller-manager)**
- **核心功能**：运行一系列**控制器（Controllers）**，确保集群状态与用户声明的期望状态一致。
- **主要控制器**：
  - **Node Controller**：监控节点状态（如节点宕机时标记为不可用）。
  - **Deployment Controller**：管理 Deployment 的副本数和滚动更新。
  - **ReplicaSet Controller**：确保 Pod 副本数与声明一致。
  - **Service Controller**：管理 Service 和云提供商的负载均衡器。
  - **Job/CronJob Controller**：处理一次性任务和定时任务。

#### 4. **etcd**
- **核心功能**：分布式键值存储数据库，保存**集群的完整状态**（如 Pod、Service、ConfigMap 等资源对象）。
- **特点**：
  - 唯一有状态的控制平面组件。
  - 强一致性（基于 Raft 算法），需高可用部署。
  - 所有变更都通过 API Server 写入 etcd。

#### 5. **Cloud Controller Manager (可选)**
- **核心功能**：对接云服务商（如 AWS、GCP），管理云平台相关的资源。
- **职责**：
  - 节点（Node）的生命周期（如自动创建云主机）。
  - 负载均衡器（LoadBalancer）的配置。
  - 存储卷（Volume）的动态供给。

---

### **二、工作节点（Worker Nodes）**
工作节点负责运行容器化应用。

#### 1. **kubelet**
- **核心功能**：节点上的“代理”，负责管理 Pod 的生命周期。
- **职责**：
  - 从 API Server 接收 Pod 定义。
  - 调用容器运行时（如 Docker、containerd）启动/停止容器。
  - 监控容器状态并上报给 API Server。
  - 执行健康检查（Liveness/Readiness Probes）。

#### 2. **kube-proxy**
- **核心功能**：维护节点上的网络规则，实现 Service 的**负载均衡**和**服务发现**。
- **实现方式**：
  - 通过 iptables/IPVS 将流量转发到后端 Pod。
  - 确保每个 Service 的 IP 和端口可被访问。

#### 3. **容器运行时（Container Runtime）**
- **核心功能**：运行容器（如 Docker、containerd、CRI-O）。
- **职责**：
  - 拉取镜像、挂载存储卷。
  - 隔离容器进程（通过 Linux cgroups/namespaces）。

---

### **三、附加组件（Addons）**
这些组件增强集群功能，通常由用户自行部署。

#### 1. **CoreDNS**
- **功能**：为集群提供 DNS 服务，解析 Service 名称到 ClusterIP。

#### 2. **Ingress Controller**
- **功能**：管理外部流量路由（如 Nginx、Traefik），实现 HTTP/HTTPS 负载均衡。

#### 3. **Dashboard**
- **功能**：Web UI，可视化查看和管理集群资源。

#### 4. **Metrics Server**
- **功能**：收集资源指标（CPU/内存），供 HPA（自动扩缩容）使用。

#### 5. **CNI 插件（如 Calico、Flannel）**
- **功能**：实现 Pod 之间的网络通信和网络策略。

---

### **四、全局交互流程示例**
1. **用户提交 Deployment 配置**：通过 `kubectl` 发送到 API Server。
2. **写入 etcd**：API Server 将资源对象持久化到 etcd。
3. **调度决策**：Scheduler 根据资源需求选择合适 Node，更新 Pod 定义。
4. **kubelet 创建 Pod**：目标节点上的 kubelet 调用容器运行时启动容器。
5. **Controller 监控状态**：Deployment Controller 确保副本数符合预期，必要时创建 ReplicaSet。
6. **kube-proxy 配置网络**：为 Service 创建 iptables 规则，流量转发到 Pod。

---

### **五、知识图谱总结**
```plaintext
控制平面
├── API Server: 集群入口，处理请求
├── Scheduler: 调度 Pod 到 Node
├── Controller Manager: 确保资源状态
├── etcd: 存储集群状态
└── Cloud Controller Manager: 对接云平台

工作节点
├── kubelet: 管理 Pod 生命周期
├── kube-proxy: 网络规则和负载均衡
└── 容器运行时: 运行容器

附加组件
├── CoreDNS: 服务发现
├── Ingress Controller: 外部流量管理
├── Metrics Server: 资源监控
└── CNI 插件: 网络通信
```

通过理解这些组件的协作关系，你可以清晰地构建 Kubernetes 的全局视图，为后续深入学习（如 Service、ConfigMap、StatefulSet 等对象）打下基础。