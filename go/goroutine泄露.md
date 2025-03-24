## # Goroutine泄漏

在Go语言开发中，Goroutine泄漏是一个常见但隐蔽的问题。当Goroutine因未正确退出而长期驻留内存时，会导致内存占用持续增长，甚至引发程序崩溃。以下是Goroutine泄漏的常见原因及避免方法，结合代码示例详细说明。

---

### 一、Goroutine泄漏的常见原因

#### 1. 未关闭的Channel导致阻塞
```go
func leak1() {
    ch := make(chan int)
    go func() {
        val := <-ch // 永久阻塞，因无人往ch写入数据
        fmt.Println(val)
    }()
    // 函数退出后，Goroutine仍在等待
}
```

#### 2. 无限循环未设退出条件
```go
func leak2() {
    go func() {
        for { // 无限循环且无退出逻辑
            time.Sleep(time.Second)
        }
    }()
}
```

#### 3. 未处理Context取消
```go
func leak3(ctx context.Context) {
    go func() {
        select { // 未监听ctx.Done()
        case <-time.After(5 * time.Second):
            fmt.Println("Done")
        }
    }()
}
```

#### 4. WaitGroup使用错误
```go
func leak4() {
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        // 若此处panic，wg.Done()未执行
    }()
    wg.Wait()
}
```

---

### 二、检测Goroutine泄漏的方法

#### 1. 使用pprof监控
```bash
# 在代码中导入pprof
import _ "net/http/pprof"

# 运行后访问查看Goroutine数量
http://localhost:6060/debug/pprof/goroutine?debug=1
```

#### 2. runtime统计
```go
func checkGoroutines() {
    for {
        time.Sleep(5 * time.Second)
        fmt.Printf("当前Goroutine数: %d\n", runtime.NumGoroutine())
    }
}
```

#### 3. 单元测试检测
```go
func TestLeak(t *testing.T) {
    before := runtime.NumGoroutine()
    // 执行被测函数
    after := runtime.NumGoroutine()
    if after != before {
        t.Fatalf("Goroutine泄漏: 之前%d, 之后%d", before, after)
    }
}
```

---

### 三、避免泄漏的编码实践

#### 1. 明确Goroutine退出条件
```go
func safe1() {
    done := make(chan struct{})
    go func() {
        defer close(done)
        // 业务逻辑
    }()

    select {
    case <-done: // 等待完成
    case <-time.After(3 * time.Second): // 超时控制
    }
}
```

#### 2. 正确使用Context
```go
func safe2(ctx context.Context) {
    go func() {
        select {
        case <-ctx.Done(): // 监听取消信号
            return
        case <-time.After(5 * time.Second):
            fmt.Println("正常完成")
        }
    }()
}

// 调用方
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
safe2(ctx)
```

#### 3. Channel关闭策略
```go
func safe3() {
    ch := make(chan int, 10)
    done := make(chan struct{})

    go func() {
        defer close(done)
        for val := range ch { // 自动检测channel关闭
            fmt.Println(val)
        }
    }()

    // 生产数据
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch) // 明确关闭channel
    <-done    // 等待消费者退出
}
```

#### 4. 防御性WaitGroup使用
```go
func safe4() {
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer func() {
            if r := recover(); r != nil { // 捕获panic
                fmt.Println("Recovered:", r)
            }
        }()
        // 业务逻辑（可能panic）
    }()
    wg.Wait()
}
```

#### 5. 限制并发数量
```go
func safe5() {
    maxWorkers := 5
    sem := make(chan struct{}, maxWorkers)

    for i := 0; i < 100; i++ {
        sem <- struct{}{} // 获取信号量
        go func(n int) {
            defer func() { <-sem }()
            // 业务逻辑
        }(i)
    }

    // 等待所有任务完成
    for i := 0; i < maxWorkers; i++ {
        sem <- struct{}{}
    }
}
```

---

### 四、高级防护模式

#### 1. Goroutine生命周期管理
```go
type Worker struct {
    quit chan struct{}
}

func (w *Worker) Start() {
    go func() {
        for {
            select {
            case <-w.quit:
                return
            default:
                // 处理任务
            }
        }
    }()
}

func (w *Worker) Stop() {
    close(w.quit)
}
```

#### 2. 超时包装器
```go
func withTimeout(fn func(), timeout time.Duration) {
    done := make(chan struct{})
    go func() {
        defer close(done)
        fn()
    }()

    select {
    case <-done:
    case <-time.After(timeout):
        fmt.Println("超时强制退出")
    }
}
```

---

### 五、典型场景解决方案

#### 1. HTTP请求处理
```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()

    result := make(chan string)
    go func() {
        data := longOperation() // 可能阻塞的操作
        result <- data
    }()

    select {
    case data := <-result:
        fmt.Fprint(w, data)
    case <-ctx.Done():
        http.Error(w, "请求超时", http.StatusGatewayTimeout)
    }
}
```

#### 2. 数据库连接池
```go
func queryWithPool(ctx context.Context) {
    pool := make(chan *sql.DB, 5)
    // 初始化连接池...

    go func() {
        select {
        case conn := <-pool:
            defer func() { pool <- conn }()
            // 使用conn执行查询
        case <-ctx.Done():
            return
        }
    }()
}
```

---

### 六、总结与最佳实践

1. 核心原则：
   - 每个Goroutine必须有明确的退出路径
   - 资源创建与释放遵循谁创建谁释放原则

2. 防御性编程技巧：
   - 使用`defer`确保资源释放
   - 为长期运行的Goroutine添加心跳检测
   ```go
   go func() {
       ticker := time.NewTicker(30 * time.Second)
       defer ticker.Stop()
       for {
           select {
           case <-ticker.C:
               log.Println("Goroutine存活")
           // ...其他case
           }
       }
   }()
   ```

3. 监控告警：
   - 设置Goroutine数量阈值报警
   - 使用Prometheus+Grafana监控运行时指标

通过以上实践，可有效预防和解决Goroutine泄漏问题。建议在代码审查阶段重点关注Goroutine退出逻辑，并结合持续集成中的泄漏检测工具（如Go的`-race`检测器）构建完整防护体系。