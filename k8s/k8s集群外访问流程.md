# k8s 集群外访问处理流程
以 Nginx 官方的 Ingress Controller 为例 

当外部请求通过 Nginx Ingress 进入 Kubernetes 集群时，整个流程涉及多个 Kubernetes 组件。以下是详细的流程解析：

---

### 1. 外部请求的发起
- 用户发起请求：客户端（如浏览器）通过域名（如 `example.com`）访问服务。
- DNS 解析：域名解析到 Kubernetes 集群外部的 负载均衡器 IP（如云厂商的 LoadBalancer）或直接解析到集群节点的公网 IP（如 NodePort 模式）。

---

### 2. 流量到达负载均衡器
- 负载均衡器类型：
  - 云厂商的 LoadBalancer（如 AWS ALB、GCP Cloud Load Balancer）：自动将流量转发到 Kubernetes 集群的 `NodePort` 或 `Ingress Controller` 的 Pod。
  - NodePort：如果未使用云厂商的 LB，流量直接到达集群节点的某个端口（如 30080）。
- 目标端口：流量被转发到 Nginx Ingress Controller Pod 的 `Service` 端口（通常是 80/HTTP 或 443/HTTPS）。

---

### 3. **Nginx Ingress Controller 处理请求**
- Ingress Controller 的作用：
  - 监听 Kubernetes API 中的 `Ingress` 资源变更。
  - 根据 `Ingress` 规则动态生成 Nginx 配置（如路由规则、TLS 证书等）。
  - 将外部请求按规则路由到对应的后端 Service。
- 处理流程：
  1. 接收请求：Nginx Ingress Pod 通过 `Service`（类型为 `LoadBalancer` 或 `NodePort`）接收外部请求。
  2. 匹配路由规则：根据 `Host` 和 `Path` 匹配 `Ingress` 资源中定义的规则。
  3. TLS 终止（如果配置 HTTPS）：
     - 使用 Kubernetes `Secret` 中存储的 TLS 证书解密 HTTPS 请求。
     - 将解密后的 HTTP 请求转发到后端服务。

---

### 4. 请求转发到后端 Service
- Service 的作用：通过标签选择器（`selector`）关联一组 Pod，并提供负载均衡。
- 流量路径：
  - Ingress Controller 根据 `Ingress` 规则找到目标 `Service`。
  - Service 通过 `kube-proxy` 维护的 iptables/IPVS 规则，将请求负载均衡到后端的 Pod。

---

### 5. Pod 处理请求
- Endpoints：Service 的实际后端是 `Endpoints`，由匹配 Service `selector` 的 Pod 的 IP 和端口组成。
- Pod 接收请求：请求最终到达运行实际业务的容器（如 Nginx、Spring Boot 等）。

---

### 6. 响应返回客户端
- 业务 Pod 处理完请求后，响应按原路径返回：  
  `Pod → Service → Ingress Controller → 负载均衡器 → 客户端`。

---

### 关键组件与资源
| 组件/资源                | 作用                                                                 |
|--------------------------|--------------------------------------------------------------------|
| Ingress Controller   | 监听 Ingress 规则，生成 Nginx 配置，处理外部请求路由和 TLS 终止。         |
| Ingress 资源          | 定义外部访问的路由规则（如域名、路径、后端 Service）。                    |
| Service              | 暴露 Pod 的访问入口，提供负载均衡。                                      |
| LoadBalancer         | 云厂商提供的负载均衡器，将外部流量引入集群。                              |
| kube-proxy           | 维护 Service 的 iptables/IPVS 规则，实现 Pod 的负载均衡。                |

---

### 详细配置示例
#### 1. 部署 Nginx Ingress Controller
```yaml
# 通过 Helm 或官方 YAML 部署
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx
```

#### 2. 定义 Ingress 资源
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /  # 重写路径
spec:
  tls:
  - hosts:
    - example.com
    secretName: example-tls  # TLS 证书 Secret
  rules:
  - host: example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: backend-service  # 后端 Service 名称
            port:
              number: 80
```

#### 3. 创建 Service 和 Deployment
```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  selector:
    app: backend  # 关联 Pod 标签
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080  # Pod 的实际端口
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: my-app:latest
        ports:
        - containerPort: 8080
```

---

### 调试与排查
1. 检查 Ingress 状态：
   ```bash
   kubectl get ingress
   kubectl describe ingress example-ingress
   ```
2. 查看 Nginx 配置：
   ```bash
   kubectl exec -it <ingress-controller-pod> -- cat /etc/nginx/nginx.conf
   ```
3. 检查 Service 和 Endpoints：
   ```bash
   kubectl get service, endpoints
   ```
4. 查看日志：
   ```bash
   kubectl logs <ingress-controller-pod>
   ```

---

### 注意事项
1. 防火墙规则：确保负载均衡器和节点端口对外开放。
2. TLS 证书：证书必须存储在 Kubernetes `Secret` 中，并在 `Ingress` 中引用。
3. 路径匹配：注意 `pathType`（`Exact`、`Prefix`）的配置是否与预期一致。
4. 性能优化：调整 Ingress Controller 的资源配置（如 CPU、内存）和 Nginx 参数（如 `worker_processes`）。

---

通过以上流程，外部请求能够高效、安全地通过 Nginx Ingress 进入 Kubernetes 集群，并按规则路由到后端服务。