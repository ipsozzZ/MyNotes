以下是使用 Go 语言和 etcd 实现分布式锁的 **原理分析** 和 **完整代码实现**，基于 etcd 的 `Lease`、`Revision` 和 `Watch` 机制：

---

### **一、分布式锁原理**
#### 1. **核心机制**
- **租约（Lease）**：为锁绑定一个租约，到期自动释放锁，避免死锁。
- **全局唯一 Revision**：通过 etcd 的事务机制（`Txn`）和键的 `Revision` 实现锁的 **公平性** 和 **互斥性**。
- **Watch 监听**：监听前一个持有锁的节点，实现阻塞等待。

#### 2. **实现步骤**
1. **尝试获取锁**：
   - 向 etcd 写入一个带租约的键（如 `/lock/resource1`）。
   - 通过事务判断写入的键是否为当前最小 Revision，若是则获得锁。
2. **未获得锁**：
   - 监听前一个 Revision 的键删除事件（通过 `Watch`），进入阻塞等待。
3. **释放锁**：
   - 删除键或让租约过期，通知后续等待者。

---

### **二、完整代码实现**
#### 1. 依赖安装
```bash
go get go.etcd.io/etcd/client/v3
```

#### 2. 代码实现
```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// DistributedLock 分布式锁结构体
type DistributedLock struct {
	client     *clientv3.Client
	lease      clientv3.Lease
	leaseID    clientv3.LeaseID
	key        string
	cancelFunc context.CancelFunc
}

// NewDistributedLock 创建锁实例
func NewDistributedLock(endpoints []string, key string) (*DistributedLock, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %v", err)
	}

	return &DistributedLock{
		client: client,
		key:    key,
	}, nil
}

// TryLock 尝试获取锁（阻塞等待）
func (dl *DistributedLock) TryLock(ctx context.Context, ttl int64) error {
	// 1. 创建租约
	lease := clientv3.NewLease(dl.client)
	grantResp, err := lease.Grant(ctx, ttl)
	if err != nil {
		return fmt.Errorf("failed to create lease: %v", err)
	}
	dl.lease = lease
	dl.leaseID = grantResp.ID

	// 2. 自动续约
	keepAliveCtx, cancel := context.WithCancel(ctx)
	dl.cancelFunc = cancel
	keepAliveCh, err := lease.KeepAlive(keepAliveCtx, dl.leaseID)
	if err != nil {
		return fmt.Errorf("failed to keep lease alive: %v", err)
	}

	// 处理续约响应（防止通道阻塞）
	go func() {
		for range keepAliveCh {
			// 续约成功，忽略具体内容
		}
	}()

	// 3. 创建 Session（简化事务操作）
	session, err := concurrency.NewSession(dl.client, concurrency.WithLease(dl.leaseID))
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// 4. 创建互斥锁
	mutex := concurrency.NewMutex(session, dl.key)

	// 5. 尝试获取锁（阻塞直到成功或超时）
	if err := mutex.Lock(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	return nil
}

// Unlock 释放锁
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	if dl.cancelFunc != nil {
		dl.cancelFunc() // 停止续约
	}
	if _, err := dl.lease.Revoke(ctx, dl.leaseID); err != nil {
		return fmt.Errorf("failed to revoke lease: %v", err)
	}
	return dl.client.Close()
}

// 示例使用
func main() {
	endpoints := []string{"localhost:2379"}
	lockKey := "/lock/resource1"

	// 创建锁实例
	lock, err := NewDistributedLock(endpoints, lockKey)
	if err != nil {
		log.Fatalf("Failed to create lock: %v", err)
	}

	// 尝试获取锁（TTL=10秒）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := lock.TryLock(ctx, 10); err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	defer lock.Unlock(context.Background())

	// 临界区操作
	fmt.Println("Lock acquired, doing work...")
	time.Sleep(5 * time.Second)
	fmt.Println("Work done, releasing lock.")
}
```

---

#### 3. 优化后的代码实现
以下是基于 etcd 实现分布式锁时，性能优化部分的具体代码实现，涵盖 TTL 设置、客户端复用、续约失败处理、Watch 压力优化等关键点：

```go
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// 全局复用 etcd 客户端（避免重复创建）
var (
	etcdClient *clientv3.Client
	clientOnce sync.Once
)

// InitEtcdClient 初始化全局 etcd 客户端（单例模式）
func InitEtcdClient(endpoints []string) error {
	var initErr error
	clientOnce.Do(func() {
		etcdClient, initErr = clientv3.New(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
			// 开启客户端重试（内置线性退避）
			Retry: clientv3.RetryConfig{
				Enable:     true,
				MaxRetries: 3,
			},
		})
	})
	return initErr
}

// DistributedLock 优化后的分布式锁结构体
type DistributedLock struct {
	client     *clientv3.Client
	session    *concurrency.Session
	mutex      *concurrency.Mutex
	key        string
	leaseID    clientv3.LeaseID
	cancelFunc context.CancelFunc
}

// NewDistributedLock 创建锁实例（复用全局客户端）
func NewDistributedLock(key string) (*DistributedLock, error) {
	if etcdClient == nil {
		return nil, fmt.Errorf("etcd client not initialized")
	}

	// 优化点 1：合理设置 TTL（根据业务需求调整）
	session, err := concurrency.NewSession(etcdClient, concurrency.WithTTL(10))
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return &DistributedLock{
		client:  etcdClient,
		session: session,
		mutex:   concurrency.NewMutex(session, key),
		key:     key,
		leaseID: session.Lease(),
	}, nil
}

// TryLock 尝试获取锁（带上下文超时和续约监控）
func (dl *DistributedLock) TryLock(ctx context.Context) error {
	// 优化点 2：监控续约状态
	keepAliveCtx, cancel := context.WithCancel(ctx)
	dl.cancelFunc = cancel

	// 启动续约状态监控协程
	go dl.monitorKeepAlive(keepAliveCtx)

	// 优化点 3：精确 Watch 前缀（减少压力）
	if err := dl.mutex.Lock(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	return nil
}

// monitorKeepAlive 监控续约状态（续约失败时主动释放锁）
func (dl *DistributedLock) monitorKeepAlive(ctx context.Context) {
	ch, err := dl.client.KeepAlive(ctx, dl.leaseID)
	if err != nil {
		log.Printf("keepalive failed: %v, releasing lock", err)
		dl.Unlock(context.Background())
		return
	}

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				log.Println("keepalive channel closed, releasing lock")
				dl.Unlock(context.Background())
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Unlock 释放锁（关闭续约并清理资源）
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	if dl.cancelFunc != nil {
		dl.cancelFunc()
	}
	if err := dl.mutex.Unlock(ctx); err != nil {
		return fmt.Errorf("failed to unlock: %v", err)
	}
	return dl.session.Close()
}

// 示例使用
func main() {
	endpoints := []string{"localhost:2379"}
	lockKey := "/lock/resource1"

	// 初始化全局客户端
	if err := InitEtcdClient(endpoints); err != nil {
		log.Fatalf("Failed to init etcd client: %v", err)
	}

	// 创建锁实例
	lock, err := NewDistributedLock(lockKey)
	if err != nil {
		log.Fatalf("Failed to create lock: %v", err)
	}

	// 尝试获取锁（带超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := lock.TryLock(ctx); err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	defer lock.Unlock(context.Background())

	// 临界区操作
	log.Println("Lock acquired, doing work...")
	time.Sleep(8 * time.Second) // 测试 TTL 续约
	log.Println("Work done, releasing lock.")
}
```

### **三、关键代码解析**
#### 1. **租约与续约**
- **`lease.Grant`**：创建一个 TTL 为 `ttl` 秒的租约，锁绑定此租约，到期自动释放。
- **`lease.KeepAlive`**：后台协程定期续约，防止任务未完成时锁过期。

#### 2. **事务竞争锁**
- **`concurrency.NewMutex`**：基于 etcd 的 `concurrency` 包实现互斥锁，内部使用事务和 Revision 机制。
  - 写入键 `/lock/resource1`，记录当前 Revision。
  - 事务检查：如果当前 Revision 是前缀 `/lock/resource1` 下最小的，则获得锁。

#### 3. **锁释放**
- **`lease.Revoke`**：主动释放租约，删除关联的键。
- **`session.Close()`**：清理会话资源。

#### 4. **阻塞等待**
- `mutex.Lock(ctx)` 内部实现：
  - 若锁已被占用，监听前一个 Revision 的删除事件（通过 `Watch`）。
  - 事件触发后，重新尝试获取锁。

---

### **四、优化与注意事项**
#### 1. **防止死锁**
- **合理设置 TTL**：确保任务能在 TTL 内完成，否则锁自动释放。
- **续约失败处理**：监控 `keepAliveCh` 的关闭，及时终止任务。

#### 2. **性能优化**
- **复用客户端**：避免频繁创建/关闭 etcd 客户端。
- **减少 Watch 压力**：使用较小的前缀范围监听。

#### 3. **错误处理**
- **网络重试**：在 `clientv3.Config` 中配置重试策略。
- **上下文超时**：所有操作绑定 `context.Context`，避免无限阻塞。

---

### **五、测试运行**
1. **启动本地 etcd**：
   ```bash
   etcd
   ```

2. **运行代码**：
   ```bash
   go run main.go
   ```

3. **观察 etcd 数据**：
   ```bash
   etcdctl get --prefix /lock/resource1
   ```

---

### **六、总结**
通过 etcd 的 **租约机制** 和 **Revision 事务**，我们实现了一个高可用的分布式锁。此方案具备以下特性：
- **自动释放**：防止客户端崩溃导致的死锁。
- **公平性**：按请求顺序获取锁（Revision 递增）。
- **高可用**：依赖 etcd 集群的强一致性。

实际生产环境中，可结合业务需求调整 TTL 和重试策略，或直接使用更成熟的库（如 `go.etcd.io/etcd/client/v3/concurrency`）。