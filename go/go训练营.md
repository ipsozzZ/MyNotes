
## # 微服务-概念与治理

#### 微服务概念


#### 微服务设计


#### gRPC和服务发现


#### 多集群与多租户



## # 第二周-异常处理



## # 第三周-并行编程

#### Goroutine
合理的并发结构例举

```golang

// context控制go
func TestGoContext(t *testing.T) {
	tr := NewTracker()
	go tr.Run()

	_ = tr.Event(context.Background(), "test1")
	_ = tr.Event(context.Background(), "test2")
	_ = tr.Event(context.Background(), "test3")
	_ = tr.Event(context.Background(), "test4")
	_ = tr.Event(context.Background(), "test5")
	_ = tr.Event(context.Background(), "test6")
	_ = tr.Event(context.Background(), "test7")
	//time.Sleep(3 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()
	tr.Shutdown(ctx)
}

type Tracker struct {
	ch   chan string
	stop chan struct{}
}

func NewTracker() *Tracker {
	return &Tracker{
		ch:   make(chan string, 10),
		stop: make(chan struct{}, 1),
	}
}

func (t *Tracker) Event(ctx context.Context, data string) error {
	select {
	case t.ch <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *Tracker) Run() {
	for data := range t.ch {
		time.Sleep(1 * time.Second)
		fmt.Println(data)
	}
	fmt.Println("stop")
	t.stop <- struct{}{}
}

func (t *Tracker) Shutdown(ctx context.Context) {
	close(t.ch)
	select {
	case <-t.stop:
		fmt.Println("stop")
	case <-ctx.Done():
		fmt.Println("timeout")
	}
}

```

* WaitGroup控制go
```golang

// WaitGroup控制go
func TestGoSync(t *testing.T) {
	t1 := Tracker1{wg: sync.WaitGroup{}}
	t1.Event("test1")
	t1.Event("test2")
	t1.Event("test3")
	t1.Event("test4")
	t1.Event("test5")
	t1.Event("test6")

	t1.Shutdown()
}

type Tracker1 struct {
	wg sync.WaitGroup
}

func (t1 *Tracker1) Event(data string) {
	t1.wg.Add(1)

	go func() {
		defer t1.wg.Done()
		time.Sleep(1 * time.Millisecond)
		log.Println(data)
	}()
}

func (t1 *Tracker1) Shutdown() {
	t1.wg.Wait()
}


```

#### 内存模型


#### Package sync


##### Package context




## # 第四周-go工程化实践

#### 工程目录结构


#### API设计


#### 配置管理


#### 模块/单元测试




## # 第五周-微服务可用性设计

#### 隔离


#### 超时
1. 超时传递：通过元数据的形式将超时剩余时间在微服务之间传递


#### 过载保护和限流


#### 降级和重试


#### 重试和负载均衡



## # 第六周评论系统架构设计

#### 功能和架构设计


#### 存储和可用性设计




## # 第七周-历史记录架构设计

#### 功能和架构设计


#### 存储和可用性设计




## # 第八周-分布式缓存与分布式事务

#### 分布式缓存


#### 分布式事务




## # 第九周-网络编程

#### 网络通讯协议


#### Goim 长连接网关


#### IM私信系统




## # 第十周-日志&指标&链路追踪

#### 日志


#### 链路追踪


#### 指标



## # 第十一周-DNS & CDN & 多活架构

#### DNS和CDN


#### 多活




## # 第十二周-消息队列（kafka）

#### Topic & Partition


#### Producer & Consumer


#### Leader & Follower




## # 第十三周-Runtime

#### Goroutine原理


#### 内存分配原理


#### GC原理


#### Channel原理

