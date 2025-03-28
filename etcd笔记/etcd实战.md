# etcd 相关笔记

## # etcd
在使用 Kubernetes、etcd 的过程中，很可能也会遇到下面这些典型问题：
- etcd Watch 机制能保证事件不丢吗？（原理类）
- 哪些因素会导致你的集群 Leader 发生切换? （稳定性类）
- 为什么基于 Raft 实现的 etcd 还可能会出现数据不一致？（一致性类）
- 为什么你删除了大量数据，db 大小不减少？为何 etcd 社区建议 db 大小不要超过 8G？（db 大小类）
- 为什么集群各节点磁盘 I/O 延时很低，写请求也会超时？（延时类）
- 为什么你只存储了 1 个几百 KB 的 key/value， etcd 进程却可能耗费数 G 内存? （内存类）
- 当你在一个 namespace 下创建了数万个 Pod/CRD 资源时，同时频繁通过标签去查询指定 Pod/CRD 资源时，APIServer 和 etcd 为什么扛不住?（最佳实践类）


---

## # etcd的前世今生：为什么Kubernetes使用etcd？

#### etcd v1 和 v2 诞生
首先我们来看服务高可用及数据一致性。前面我们提到单副本存在单点故障，而多副本又引入数据一致性问题。

因此为了解决数据一致性问题，需要引入一个共识算法，确保各节点数据一致性，并可容忍一定节点故障。常见的共识算法有 Paxos、ZAB、Raft 等。CoreOS 团队选择了易理解实现的 Raft 算法，它将复杂的一致性问题分解成 Leader 选举、日志同步、安全性三个相对独立的子问题，只要集群一半以上节点存活就可提供服务，具备良好的可用性。

其次我们再来看数据模型（Data Model）和 API。数据模型参考了 ZooKeeper，**使用的是基于目录的层次模式。API 相比 ZooKeeper 来说，使用了简单、易用的 REST API，提供了常用的 Get/Set/Delete/Watch 等 API，实现对 key-value 数据的查询、更新、删除、监听等操作。**

key-value 存储引擎上，ZooKeeper 使用的是 Concurrent HashMap，而 etcd 使用的是则是简单内存树，它的节点数据结构精简后如下，含节点路径、值、孩子节点信息。这是一个典型的低容量设计，数据全放在内存，无需考虑数据分片，只能保存 key 的最新版本，简单易实现。

下面我分别从功能局限性、Watch 事件的可靠性、性能、内存开销来分别给你剖析 etcd v2 的问题。

首先是功能局限性问题。它主要是指 etcd v2 不支持范围和分页查询、不支持多 key 事务。第一，etcd v2 不支持范围查询和分页。分页对于数据较多的场景是必不可少的。在 Kubernetes 中，在集群规模增大后，Pod、Event 等资源可能会出现数千个以上，但是 etcd v2 不支持分页，不支持范围查询，大包等 expensive request 会导致严重的性能乃至雪崩问题。第二，etcd v2 不支持多 key 事务。在实际转账等业务场景中，往往我们需要在一个事务中同时更新多个 key。然后是 Watch 机制可靠性问题。Kubernetes 项目严重依赖 etcd Watch 机制，然而 etcd v2 是内存型、不支持保存 key 历史版本的数据库，只在内存中使用滑动窗口保存了最近的 1000 条变更事件，当 etcd server 写请求较多、网络波动时等场景，很容易出现事件丢失问题，进而又触发 client 数据全量拉取，产生大量 expensive request，甚至导致 etcd 雪崩。其次是性能瓶颈问题。etcd v2 早期使用了简单、易调试的 HTTP/1.x API，但是随着 Kubernetes 支撑的集群规模越来越大，HTTP/1.x 协议的瓶颈逐渐暴露出来。比如集群规模大时，由于 HTTP/1.x 协议没有压缩机制，批量拉取较多 Pod 时容易导致 APIServer 和 etcd 出现 CPU 高负载、OOM、丢包等问题。

另一方面，etcd v2 client 会通过 HTTP 长连接轮询 Watch 事件，当 watcher 较多的时候，因 HTTP/1.x 不支持多路复用，会创建大量的连接，消耗 server 端过多的 socket 和内存资源。

同时 etcd v2 支持为每个 key 设置 TTL 过期时间，client 为了防止 key 的 TTL 过期后被删除，需要周期性刷新 key 的 TTL。实际业务中很有可能若干 key 拥有相同的 TTL，可是在 etcd v2 中，即使大量 key TTL 一样，你也需要分别为每个 key 发起续期操作，当 key 较多的时候，这会显著增加集群负载、导致集群性能显著下降。

最后是内存开销问题。etcd v2 在内存维护了一颗树来保存所有节点 key 及 value。在数据量场景略大的场景，如配置项较多、存储了大量 Kubernetes Events， 它会导致较大的内存开销，同时 etcd 需要定时把全量内存树持久化到磁盘。这会消耗大量的 CPU 和磁盘 I/O 资源，对系统的稳定性造成一定影响。

#### etcd v3
etcd v3 就是为了解决以上稳定性、扩展性、性能问题而诞生的。

在内存开销、Watch 事件可靠性、功能局限上，它通过引入 B-tree、boltdb 实现一个 MVCC 数据库，数据模型从层次型目录结构改成扁平的 key-value，提供稳定可靠的事件通知，实现了事务，支持多 key 原子更新，同时基于 boltdb 的持久化存储，显著降低了 etcd 的内存占用、避免了 etcd v2 定期生成快照时的昂贵的资源开销。

性能上，首先 etcd v3 使用了 gRPC API，使用 protobuf 定义消息，消息编解码性能相比 JSON 超过 2 倍以上，并通过 HTTP/2.0 多路复用机制，减少了大量 watcher 等场景下的连接数。

其次使用 Lease 优化 TTL 机制，每个 Lease 具有一个 TTL，相同的 TTL 的 key 关联一个 Lease，Lease 过期的时候自动删除相关联的所有 key，不再需要为每个 key 单独续期。

最后是 etcd v3 支持范围、分页查询，可避免大包等 expensive request。

2016 年 6 月，etcd 3.0 诞生，随后 Kubernetes 1.6 发布，默认启用 etcd v3，助力 Kubernetes 支撑 5000 节点集群规模。

从 2013 年发布第一个版本 v0.1 到今天的 3.5.0-pre，从 v2 到 v3，etcd 走过了 7 年的历程，etcd 的稳定性、扩展性、性能不断提升。

发展到今天，在 GitHub 上 star 数超过 34K。在 Kubernetes 的业务场景磨炼下它不断成长，走向稳定和成熟，成为技术圈众所周知的开源产品，而 v3 方案的发布，也标志着 etcd 进入了技术成熟期，成为云原生时代的首选元数据存储产品。


---

## # 基础架构：etcd一个读请求是如何执行的？
今天，介绍一下 etcd v3 的基础架构，从整体上对 etcd 有一个初步的了解，心中能构筑起一幅 etcd 模块全景图。这样，在你遇到诸如“Kubernetes 在执行 kubectl get pod 时，etcd 如何获取到最新的数据返回给 APIServer？”等流程架构问题时，就能知道各个模块由上至下是如何紧密协作的。

即便是遇到请求报错，你也能通过顶层的模块全景图，推测出请求流程究竟在什么模块出现了问题。

#### 基础架构

下面是一张 etcd 的简要基础架构图，我们先从宏观上了解一下 etcd 都有哪些功能模块。p1:
![p1](http://cdn.ipso.live/notes/etcd/etcd001.png)

你可以看到，按照分层模型，etcd 可分为 Client 层、API 网络层、Raft 算法层、逻辑层和存储层。这些层的功能如下：
- Client 层：Client 层包括 client v2 和 v3 两个大版本 API 客户端库，提供了简洁易用的 API，同时支持负载均衡、节点间故障自动转移，可极大降低业务使用 etcd 复杂度，提升开发效率、服务可用性。

- API 网络层：API 网络层主要包括 client 访问 server 和 server 节点之间的通信协议。一方面，client 访问 etcd server 的 API 分为 v2 和 v3 两个大版本。v2 API 使用 HTTP/1.x 协议，v3 API 使用 gRPC 协议。同时 v3 通过 etcd grpc-gateway 组件也支持 HTTP/1.x 协议，便于各种语言的服务调用。另一方面，server 之间通信协议，是指节点间通过 Raft 算法实现数据复制和 Leader 选举等功能时使用的 HTTP 协议。

- Raft 算法层：Raft 算法层实现了 Leader 选举、日志复制、ReadIndex 等核心算法特性，用于保障 etcd 多个节点间的数据一致性、提升服务可用性等，是 etcd 的基石和亮点。

- 功能逻辑层：etcd 核心特性实现层，如典型的 KVServer 模块、MVCC 模块、Auth 鉴权模块、Lease 租约模块、Compactor 压缩模块等，其中 MVCC 模块主要由 treeIndex 模块和 boltdb 模块组成。

- 存储层：存储层包含预写日志 (WAL) 模块、快照 (Snapshot) 模块、boltdb 模块。其中 WAL 可保障 etcd crash 后数据不丢失，boltdb 则保存了集群元数据和用户写入的数据。

etcd 是典型的读多写少存储，在我们实际业务场景中，读一般占据 2/3 以上的请求。为了对 etcd 有一个深入的理解，接下来先分析一个读请求是如何执行的，来了解 etcd 的核心模块，进而由点及线、由线到面地帮助你构建 etcd 的全景知识脉络。

在下面这张架构图中，我用序号标识了 etcd 默认读模式（线性读）的执行流程，接下来，我们就按照这个执行流程从头开始说。p2:
![p2](http://cdn.ipso.live/notes/etcd/etcd002.png)

#### 环境准备
首先介绍一个好用的进程管理工具[goreman](https://github.com/mattn/goreman)，基于它，我们可快速创建、停止本地的多节点 etcd 集群。

你可以通过如下go get命令快速安装 goreman，然后从etcd release页下载 etcd，再从etcd 源码中下载 [goreman Procfile](https://github.com/etcd-io/etcd/blob/v3.4.9/Procfile) 文件，它描述了 etcd 进程名、节点数、参数等信息。最后通过goreman -f Procfile start命令就可以快速启动一个 3 节点的本地集群了。
```sh
go get github.com/mattn/goreman
```

#### client
启动完 etcd 集群后，当你用 etcd 的客户端工具 etcdctl 执行一个 get hello 命令（如下）时，对应到图2中流程一，etcdctl 是如何工作的呢？
```sh
etcdctl get hello --endpoints http://127.0.0.1:2379  
hello  
world  
```

首先，etcdctl 会对命令中的参数进行解析。我们来看下这些参数的含义，其中，参数“get”是请求的方法，它是 KVServer 模块的 API；“hello”是我们查询的 key 名；“endpoints”是我们后端的 etcd 地址，通常，生产环境下中需要配置多个 endpoints，这样在 etcd 节点出现故障后，client 就可以自动重连到其它正常的节点，从而保证请求的正常执行。

在 etcd v3.4.9 版本中，etcdctl 是通过 clientv3 库来访问 etcd server 的，clientv3 库基于 gRPC client API 封装了操作 etcd KVServer、Cluster、Auth、Lease、Watch 等模块的 API，同时还包含了负载均衡、健康探测和故障切换等特性。

在解析完请求中的参数后，etcdctl 会创建一个 clientv3 库对象，使用 KVServer 模块的 API 来访问 etcd server。

接下来，就需要为这个 get hello 请求选择一个合适的 etcd server 节点了，这里得用到负载均衡算法。在 etcd 3.4 中，clientv3 库采用的负载均衡算法为 Round-robin。针对每一个请求，Round-robin 算法通过轮询的方式依次从 endpoint 列表中选择一个 endpoint 访问 (长连接)，使 etcd server 负载尽量均衡。

关于负载均衡算法，你需要特别注意以下两点:
1. 如果你的 client 版本 <= 3.3，那么当你配置多个 endpoint 时，负载均衡算法仅会从中选择一个 IP 并创建一个连接（Pinned endpoint），这样可以节省服务器总连接数。但在这我要给你一个小提醒，在 heavy usage 场景，这可能会造成 server 负载不均衡。

2. 在 client 3.4 之前的版本中，负载均衡算法有一个严重的 Bug：如果第一个节点异常了，可能会导致你的 client 访问 etcd server 异常，特别是在 Kubernetes 场景中会导致 APIServer 不可用。不过，该 Bug 已在 Kubernetes 1.16 版本后被修复。

为请求选择好 etcd server 节点，client 就可调用 etcd server 的 KVServer 模块的 Range RPC 方法，把请求发送给 etcd server。

这里我说明一点，client 和 server 之间的通信，使用的是基于 HTTP/2 的 gRPC 协议。相比 etcd v2 的 HTTP/1.x，HTTP/2 是基于二进制而不是文本、支持多路复用而不再有序且阻塞、支持数据压缩以减少包大小、支持 server push 等特性。因此，基于 HTTP/2 的 gRPC 协议具有低延迟、高性能的特点，有效解决了我们在上一讲中提到的 etcd v2 中 HTTP/1.x 性能问题。

#### KVServer
client 发送 Range RPC 请求到了 server 后，就开始进入我们架构图中的流程二，也就是 KVServer 模块了。

etcd 提供了丰富的 metrics、日志、请求行为检查等机制，可记录所有请求的执行耗时及错误码、来源 IP 等，也可控制请求是否允许通过，比如 etcd Learner 节点只允许指定接口和参数的访问，帮助大家定位问题、提高服务可观测性等，而这些特性是怎么非侵入式的实现呢？

答案就是拦截器。

#### 拦截器
etcd server 定义了如下的 Service KV 和 Range 方法，启动的时候它会将实现 KV 各方法的对象注册到 gRPC Server，并在其上注册对应的拦截器。下面的代码中的 Range 接口就是负责读取 etcd key-value 的的 RPC 接口。\
```proto
service KV {  
  // Range gets the keys in the range from the key-value store.  
  rpc Range(RangeRequest) returns (RangeResponse) {  
      option (google.api.http) = {  
        post: "/v3/kv/range"  
        body: "*"  
      };  
  }  
  ....
}  
```

拦截器提供了在执行一个请求前后的 hook 能力，除了我们上面提到的 debug 日志、metrics 统计、对 etcd Learner 节点请求接口和参数限制等能力，etcd 还基于它实现了以下特性:
- 要求执行一个操作前集群必须有 Leader；
- 请求延时超过指定阈值的，打印包含来源 IP 的慢查询日志 (3.5 版本)。

server 收到 client 的 Range RPC 请求后，根据 ServiceName 和 RPC Method 将请求转发到对应的 handler 实现，handler 首先会将上面描述的一系列拦截器串联成一个执行，在拦截器逻辑中，通过调用 KVServer 模块的 Range 接口获取数据。

#### 串行读与线性读
进入 KVServer 模块后，我们就进入核心的读流程了，对应架构图中的流程三和四。我们知道 etcd 为了保证服务高可用，生产环境一般部署多个节点，那各个节点数据在任意时间点读出来都是一致的吗？什么情况下会读到旧数据呢？

这里为了帮助更好的理解读流程，先简单提下写流程。当 client 发起一个更新 hello 为 world 请求后，若 Leader 收到写请求，它会将此请求持久化到 WAL 日志，并广播给各个节点，若一半以上节点持久化成功，则该请求对应的日志条目被标识为已提交，etcdserver 模块异步从 Raft 模块获取已提交的日志条目，应用到状态机 (boltdb 等)。

此时若 client 发起一个读取 hello 的请求，假设此请求直接从状态机中读取， 如果连接到的是 C 节点，若 C 节点磁盘 I/O 出现波动，可能导致它应用已提交的日志条目很慢，则会出现更新 hello 为 world 的写命令，在 client 读 hello 的时候还未被提交到状态机，因此就可能读取到旧数据。

从以上介绍我们可以看出，在多节点 etcd 集群中，各个节点的状态机数据一致性存在差异。而我们不同业务场景的读请求对数据是否最新的容忍度是不一样的，有的场景它可以容忍数据落后几秒甚至几分钟，有的场景要求必须读到反映集群共识的最新数据。

- **我们首先来看一个对数据敏感度较低的场景。**

假如老板让你做一个旁路数据统计服务，希望你每分钟统计下 etcd 里的服务、配置信息等，这种场景其实对数据时效性要求并不高，读请求可直接从节点的状态机获取数据。即便数据落后一点，也不影响业务，毕竟这是一个定时统计的旁路服务而已。

**这种直接读状态机数据返回、无需通过 Raft 协议与集群进行交互的模式，在 etcd 里叫做串行 (Serializable) 读，它具有低延时、高吞吐量的特点，适合对数据一致性要求不高的场景。**

- **我们再看一个对数据敏感性高的场景。**

当你发布服务，更新服务的镜像的时候，提交的时候显示更新成功，结果你一刷新页面，发现显示的镜像的还是旧的，再刷新又是新的，这就会导致混乱。再比如说一个转账场景，Alice 给 Bob 转账成功，钱被正常扣出，一刷新页面发现钱又回来了，这也是令人不可接受的。

**以上的业务场景就对数据准确性要求极高了，在 etcd 里面，提供了一种线性读模式来解决对数据一致性要求高的场景。**

#### 什么是线性读呢?
你可以理解一旦一个值更新成功，随后任何通过线性读的 client 都能及时访问到。虽然集群中有多个节点，但 client 通过线性读就如访问一个节点一样。etcd 默认读模式是线性读，因为它需要经过 Raft 协议模块，反应的是集群共识，因此在延时和吞吐量上相比串行读略差一点，适用于对数据一致性要求高的场景。

如果你的 etcd 读请求显示指定了是串行读，就不会经过架构图流程中的流程三、四。默认是线性读，因此接下来我们看看读请求进入线性读模块，它是如何工作的。

#### 线性读之 ReadIndex
前面我们聊到串行读时提到，它之所以能读到旧数据，主要原因是 Follower 节点收到 Leader 节点同步的写请求后，应用日志条目到状态机是个异步过程，那么我们能否有一种机制在读取的时候，确保最新的数据已经应用到状态机中？
p3:
![p3](http://cdn.ipso.live/notes/etcd/etcd003.png)
其实这个机制就是叫 ReadIndex，它是在 etcd 3.1 中引入的。当收到一个线性读请求时，它首先会从 Leader 获取集群最新的已提交的日志索引 (committed index)，如上图中的流程二所示。

Leader 收到 ReadIndex 请求时，为防止脑裂等异常场景，会向 Follower 节点发送心跳确认，一半以上节点确认 Leader 身份后才能将已提交的索引 (committed index) 返回给节点 C(上图中的流程三)。

C 节点则会等待，直到状态机已应用索引 (applied index) 大于等于 Leader 的已提交索引时 (committed Index)(上图中的流程四)，然后去通知读请求，数据已赶上 Leader，你可以去状态机中访问数据了 (上图中的流程五)。

以上就是线性读通过 ReadIndex 机制保证数据一致性原理， 当然还有其它机制也能实现线性读，如在早期 etcd 3.0 中读请求通过走一遍 Raft 协议保证一致性， 这种 Raft log read 机制依赖磁盘 IO， 性能相比 ReadIndex 较差。

总体而言，KVServer 模块收到线性读请求后，通过架构图(p2)中流程三向 Raft 模块发起 ReadIndex 请求，Raft 模块将 Leader 最新的已提交日志索引封装在流程四的 ReadState 结构体，通过 channel 层层返回给线性读模块，线性读模块等待本节点状态机追赶上 Leader 进度，追赶完成后，就通知 KVServer 模块，进行架构图中流程五，与状态机中的 MVCC 模块进行进行交互了。

#### MVCC
流程五中的多版本并发控制 (Multiversion concurrency control) 模块是为了解决 etcd v2 不支持保存 key 的历史版本、不支持多 key 事务等问题而产生的。

它核心由内存树形索引模块 (treeIndex) 和嵌入式的 KV 持久化存储库 boltdb 组成。

首先我们需要简单了解下 boltdb，它是个基于 B+ tree 实现的 key-value 键值库，支持事务，提供 Get/Put 等简易 API 给 etcd 操作。

那么 etcd 如何基于 boltdb 保存一个 key 的多个历史版本呢?

比如我们现在有以下方案：方案 1 是一个 key 保存多个历史版本的值；方案 2 每次修改操作，生成一个新的版本号 (revision)，以版本号为 key， value 为用户 key-value 等信息组成的结构体。

很显然方案 1 会导致 value 较大，存在明显读写放大、并发冲突等问题，而方案 2 正是 etcd 所采用的。boltdb 的 key 是全局递增的版本号 (revision)，value 是用户 key、value 等字段组合成的结构体，然后通过 treeIndex 模块来保存用户 key 和版本号的映射关系。

treeIndex 与 boltdb 关系如下面的读事务流程图所示，从 treeIndex 中获取 key hello 的版本号，再以版本号作为 boltdb 的 key，从 boltdb 中获取其 value 信息。
p4:
![p4](http://cdn.ipso.live/notes/etcd/etcd004.png)

#### treeIndex
treeIndex 模块是基于 Google 开源的内存版 btree 库实现的，为什么 etcd 选择上图中的 B-tree 数据结构保存用户 key 与版本号之间的映射关系，而不是哈希表、二叉树呢？后面会再介绍。

treeIndex 模块只会保存用户的 key 和相关版本号信息，用户 key 的 value 数据存储在 boltdb 里面，相比 ZooKeeper 和 etcd v2 全内存存储，etcd v3 对内存要求更低。

简单介绍了 etcd 如何保存 key 的历史版本后，架构图(p2)中流程六也就非常容易理解了， 它需要从 treeIndex 模块中获取 hello 这个 key 对应的版本号信息。treeIndex 模块基于 B-tree 快速查找此 key，返回此 key 对应的索引项 keyIndex 即可。索引项中包含版本号等信息。

#### buffer
在获取到版本号信息后，就可从 boltdb 模块中获取用户的 key-value 数据了。不过有一点你要注意，并不是所有请求都一定要从 boltdb 获取数据。

etcd 出于数据一致性、性能等考虑，在访问 boltdb 前，首先会从一个内存读事务 buffer 中，二分查找你要访问 key 是否在 buffer 里面，若命中则直接返回。

#### boltdb
若 buffer 未命中，此时就真正需要向 boltdb 模块查询数据了，进入了流程七。

我们知道 MySQL 通过 table 实现不同数据逻辑隔离，那么在 boltdb 是如何隔离集群元数据与用户数据的呢？答案是 bucket。boltdb 里每个 bucket 类似对应 MySQL 一个表，用户的 key 数据存放的 bucket 名字的是 key，etcd MVCC 元数据存放的 bucket 是 meta。

因 boltdb 使用 B+ tree 来组织用户的 key-value 数据，获取 bucket key 对象后，通过 boltdb 的游标 Cursor 可快速在 B+ tree 找到 key hello 对应的 value 数据，返回给 client。

到这里，一个读请求之路执行完成。

#### 小结
小结一下，一个读请求从 client 通过 Round-robin 负载均衡算法，选择一个 etcd server 节点，发出 gRPC 请求，经过 etcd server 的 KVServer 模块、线性读模块、MVCC 的 treeIndex 和 boltdb 模块紧密协作，完成了一个读请求。

通过一个读请求，初步了解了 etcd 的基础架构以及各个模块之间是如何协作的。

在这过程中，特别总结下 client 的节点故障自动转移和线性读。

一方面， client 的通过负载均衡、错误处理等机制实现了 etcd 节点之间的故障的自动转移，它可助你的业务实现服务高可用，建议不使用低于 etcd 3.4 分支的 client 版本。

另一方面，详细解释了 etcd 提供的两种读机制 (串行读和线性读) 原理和应用场景。通过线性读，对业务而言，访问多个节点的 etcd 集群就如访问一个节点一样简单，能简洁、快速的获取到集群最新共识数据。

早期 etcd 线性读使用的 Raft log read，也就是说把读请求像写请求一样走一遍 Raft 的协议，基于 Raft 的日志的有序性，实现线性读。但此方案读涉及磁盘 IO 开销，性能较差，后来实现了 ReadIndex 读机制来提升读性能，满足了 Kubernetes 等业务的诉求。


---

## # 基础架构：etcd一个写请求是如何执行的？
在上一节里，通过分析 etcd 的一个读请求执行流程，介绍了 etcd 的基础架构，初步了解了在 etcd 的读请求流程中，各个模块是如何紧密协作，执行查询语句，返回数据给 client。

那么 etcd 一个写请求执行流程又是怎样的呢？在执行写请求过程中，如果进程 crash 了，如何保证数据不丢、命令不重复执行呢？

今天我就和你聊聊 etcd 写过程中是如何解决这些问题的。希望通过这小节让你了解一个 key-value 写入的原理，对 etcd 的基础架构中涉及写请求相关的模块有一定的理解，同时能触类旁通，当你在软件项目开发过程中遇到类似数据安全、幂等性等问题时，能设计出良好的方案解决它。

#### 整体架构
```sh
etcdctl put hello world --endpoints http://127.0.0.1:2379
OK
```

p5:
![p5](http://cdn.ipso.live/notes/etcd/etcd005.png)

为了能够更直观地理解 etcd 的写请求流程，在如上的架构图中，用序号标识了下面的一个 put hello 为 world 的写请求的简要执行流程，帮助你从整体上快速了解一个写请求的全貌。

首先 client 端通过负载均衡算法选择一个 etcd 节点，发起 gRPC 调用。然后 etcd 节点收到请求后经过 gRPC 拦截器、Quota 模块后，进入 KVServer 模块，KVServer 模块向 Raft 模块提交一个提案，提案内容为“大家好，请使用 put 方法执行一个 key 为 hello，value 为 world 的命令”。

随后此提案通过 RaftHTTP 网络模块转发、经过集群多数节点持久化后，状态会变成已提交，etcdserver 从 Raft 模块获取已提交的日志条目，传递给 Apply 模块，Apply 模块通过 MVCC 模块执行提案内容，更新状态机。

与读流程不一样的是写流程还涉及 Quota、WAL、Apply 三个模块。crash-safe 及幂等性也正是基于 WAL 和 Apply 流程的 consistent index 等实现的，因此会重点介绍这三个模块。

下面就让我们沿着写请求执行流程图，从 0 到 1 分析一个 key-value 是如何安全、幂等地持久化到磁盘的。

#### Quota 模块
首先是流程一 client 端发起 gRPC 调用到 etcd 节点，和读请求不一样的是，写请求需要经过流程二 db 配额（Quota）模块，它有什么功能呢？

我们先从此模块的一个常见错误说起，在使用 etcd 过程中是否遇到过"etcdserver: mvcc: database space exceeded"错误呢？

我相信只要你使用过 etcd 或者 Kubernetes，大概率见过这个错误。它是指当前 etcd db 文件大小超过了配额，当出现此错误后，你的整个集群将不可写入，只读，对业务的影响非常大。

哪些情况会触发这个错误呢？

一方面默认 db 配额仅为 2G，当你的业务数据、写入 QPS、Kubernetes 集群规模增大后，你的 etcd db 大小就可能会超过 2G。

另一方面我们知道 etcd v3 是个 MVCC 数据库，保存了 key 的历史版本，当你未配置压缩策略的时候，随着数据不断写入，db 大小会不断增大，导致超限。

最后你要特别注意的是，如果你使用的是 etcd 3.2.10 之前的旧版本，请注意备份可能会触发 boltdb 的一个 Bug，它会导致 db 大小不断上涨，最终达到配额限制。

了解完触发 Quota 限制的原因后，我们再详细了解下 Quota 模块它是如何工作的。

当 etcd server 收到 put/txn 等写请求的时候，会首先检查下当前 etcd db 大小加上你请求的 key-value 大小之和是否超过了配额（quota-backend-bytes）。

如果超过了配额，它会产生一个告警（Alarm）请求，告警类型是 NO SPACE，并通过 Raft 日志同步给其它节点，告知 db 无空间了，并将告警持久化存储到 db 中。

最终，无论是 API 层 gRPC 模块还是负责将 Raft 侧已提交的日志条目应用到状态机的 Apply 模块，都拒绝写入，集群只读。

那遇到这个错误时应该如何解决呢？

首先当然是调大配额。具体多大合适呢？etcd 社区建议不超过 8G。遇到过这个错误的你是否还记得，为什么当你把配额（quota-backend-bytes）调大后，集群依然拒绝写入呢?

原因就是我们前面提到的 NO SPACE 告警。Apply 模块在执行每个命令的时候，都会去检查当前是否存在 NO SPACE 告警，如果有则拒绝写入。所以还需要你额外发送一个取消告警（etcdctl alarm disarm）的命令，以消除所有告警。

其次你需要检查 etcd 的压缩（compact）配置是否开启、配置是否合理。etcd 保存了一个 key 所有变更历史版本，如果没有一个机制去回收旧的版本，那么内存和 db 大小就会一直膨胀，在 etcd 里面，压缩模块负责回收旧版本的工作。

压缩模块支持按多种方式回收旧版本，比如保留最近一段时间内的历史版本。不过你要注意，它仅仅是将旧版本占用的空间打个空闲（Free）标记，后续新的数据写入的时候可复用这块空间，而无需申请新的空间。

如果你需要回收空间，减少 db 大小，得使用碎片整理（defrag）， 它会遍历旧的 db 文件数据，写入到一个新的 db 文件。但是它对服务性能有较大影响，不建议你在生产集群频繁使用。

最后你需要注意配额（quota-backend-bytes）的行为，默认'0'就是使用 etcd 默认的 2GB 大小，你需要根据你的业务场景适当调优。如果你填的是个小于 0 的数，就会禁用配额功能，这可能会让你的 db 大小处于失控，导致性能下降，不建议你禁用配额。

#### KVServer 模块
通过流程二的配额检查后，请求就从 API 层转发到了流程三的 KVServer 模块的 put 方法，我们知道 etcd 是基于 Raft 算法实现节点间数据复制的，因此它需要将 put 写请求内容打包成一个提案消息，提交给 Raft 模块。不过 KVServer 模块在提交提案前，还有如下的一系列检查和限速。

**Preflight Check**

为了保证集群稳定性，避免雪崩，任何提交到 Raft 模块的请求，都会做一些简单的限速判断。首先，如果 Raft 模块已提交的日志索引（committed index）比已应用到状态机的日志索引（applied index）超过了 5000，那么它就返回一个"etcdserver: too many requests"错误给 client。

其次它会检查你写入的包大小是否超过默认的 1.5MB， 如果超过了会返回"etcdserver: request is too large"错误给给 client。

**Propose**

最后通过一系列检查之后，会生成一个唯一的 ID，将此请求关联到一个对应的消息通知 channel，然后向 Raft 模块发起（Propose）一个提案（Proposal），提案内容为“大家好，请使用 put 方法执行一个 key 为 hello，value 为 world 的命令”，也就是整体架构图里的流程四。

向 Raft 模块发起提案后，KVServer 模块会等待此 put 请求，等待写入结果通过消息通知 channel 返回或者超时。etcd 默认超时时间是 7 秒（5 秒磁盘 IO 延时 +2*1 秒竞选超时时间），如果一个请求超时未返回结果，则可能会出现你熟悉的 etcdserver: request timed out 错误。

#### WAL 模块
Raft 模块收到提案后，如果当前节点是 Follower，它会转发给 Leader，只有 Leader 才能处理写请求。Leader 收到提案后，通过 Raft 模块输出待转发给 Follower 节点的消息和待持久化的日志条目，日志条目则封装了我们上面所说的 put hello 提案内容。

etcdserver 从 Raft 模块获取到以上消息和日志条目后，作为 Leader，它会将 put 提案消息广播给集群各个节点，同时需要把集群 Leader 任期号、投票信息、已提交索引、提案内容持久化到一个 WAL（Write Ahead Log）日志文件中，用于保证集群的一致性、可恢复性，也就是我们图中的流程五模块。

WAL 日志结构是怎样的呢？p6:
![p6](http://cdn.ipso.live/notes/etcd/etcd006.png)

上图是 WAL 结构，它由多种类型的 WAL 记录顺序追加写入组成，每个记录由类型、数据、循环冗余校验码组成。不同类型的记录通过 Type 字段区分，Data 为对应记录内容，CRC 为循环校验码信息。

WAL 记录类型目前支持 5 种，分别是文件元数据记录、日志条目记录、状态信息记录、CRC 记录、快照记录：
- 文件元数据记录包含节点 ID、集群 ID 信息，它在 WAL 文件创建的时候写入；
- 日志条目记录包含 Raft 日志信息，如 put 提案内容；
- 状态信息记录，包含集群的任期号、节点投票信息等，一个日志文件中会有多条，以最后的记录为准；
- CRC 记录包含上一个 WAL 文件的最后的 CRC（循环冗余校验码）信息， 在创建、切割 WAL 文件时，作为第一条记录写入到新的 WAL 文件， 用于校验数据文件的完整性、准确性等；
- 快照记录包含快照的任期号、日志索引信息，用于检查快照文件的准确性。

WAL 模块又是如何持久化一个 put 提案的日志条目类型记录呢?

首先我们来看看 put 写请求如何封装在 Raft 日志条目里面。下面是 Raft 日志条目的数据结构信息，它由以下字段组成：
- Term 是 Leader 任期号，随着 Leader 选举增加；
- Index 是日志条目的索引，单调递增增加；
- Type 是日志类型，比如是普通的命令日志（EntryNormal）还是集群配置变更日志（EntryConfChange）；
- Data 保存我们上面描述的 put 提案内容。
```go
type Entry struct {
   Term             uint64    `protobuf:"varint，2，opt，name=Term" json:"Term"`
   Index            uint64    `protobuf:"varint，3，opt，name=Index" json:"Index"`
   Type             EntryType `protobuf:"varint，1，opt，name=Type，enum=Raftpb.EntryType" json:"Type"`
   Data             []byte    `protobuf:"bytes，4，opt，name=Data" json:"Data，omitempty"`
}
```

了解完 Raft 日志条目数据结构后，我们再看 WAL 模块如何持久化 Raft 日志条目。它首先先将 Raft 日志条目内容（含任期号、索引、提案内容）序列化后保存到 WAL 记录的 Data 字段， 然后计算 Data 的 CRC 值，设置 Type 为 Entry Type， 以上信息就组成了一个完整的 WAL 记录。

最后计算 WAL 记录的长度，顺序先写入 WAL 长度（Len Field），然后写入记录内容，调用 fsync 持久化到磁盘，完成将日志条目保存到持久化存储中。

当一半以上节点持久化此日志条目后， Raft 模块就会通过 channel 告知 etcdserver 模块，put 提案已经被集群多数节点确认，提案状态为已提交，你可以执行此提案内容了。

于是进入流程六，etcdserver 模块从 channel 取出提案内容，添加到先进先出（FIFO）调度队列，随后通过 Apply 模块按入队顺序，异步、依次执行提案内容。

#### Apply 模块

执行 put 提案内容对应我们架构图中的流程七，其细节图如下p7。那么 Apply 模块是如何执行 put 请求的呢？若 put 请求提案在执行流程七的时候 etcd 突然 crash 了， 重启恢复的时候，etcd 是如何找回异常提案，再次执行的呢？
![p7](http://cdn.ipso.live/notes/etcd/etcd007.png)

核心就是我们上面介绍的 WAL 日志，因为提交给 Apply 模块执行的提案已获得多数节点确认、持久化，etcd 重启时，会从 WAL 中解析出 Raft 日志条目内容，追加到 Raft 日志的存储中，并重放已提交的日志提案给 Apply 模块执行。

然而这又引发了另外一个问题，如何确保幂等性，防止提案重复执行导致数据混乱呢?

我们在上一节里讲到，etcd 是个 MVCC 数据库，每次更新都会生成新的版本号。如果没有幂等性保护，同样的命令，一部分节点执行一次，一部分节点遭遇异常故障后执行多次，则系统的各节点一致性状态无法得到保证，导致数据混乱，这是严重故障。

因此 etcd 必须要确保幂等性。怎么做呢？Apply 模块从 Raft 模块获得的日志条目信息里，是否有唯一的字段能标识这个提案？

答案就是我们上面介绍 Raft 日志条目中的索引（index）字段。日志条目索引是全局单调递增的，每个日志条目索引对应一个提案， 如果一个命令执行后，我们在 db 里面也记录下当前已经执行过的日志条目索引，是不是就可以解决幂等性问题呢？

是的。但是这还不够安全，如果执行命令的请求更新成功了，更新 index 的请求却失败了，是不是一样会导致异常？

因此我们在实现上，还需要将两个操作作为原子性事务提交，才能实现幂等。

正如我们上面的讨论的这样，etcd 通过引入一个 consistent index 的字段，来存储系统当前已经执行过的日志条目索引，实现幂等性。

Apply 模块在执行提案内容前，首先会判断当前提案是否已经执行过了，如果执行了则直接返回，若未执行同时无 db 配额满告警，则进入到 MVCC 模块，开始与持久化存储模块打交道。

#### MVCC
Apply 模块判断此提案未执行后，就会调用 MVCC 模块来执行提案内容。MVCC 主要由两部分组成，一个是内存索引模块 treeIndex，保存 key 的历史版本号信息，另一个是 boltdb 模块，用来持久化存储 key-value 数据。那么 MVCC 模块执行 put hello 为 world 命令时，它是如何构建内存索引和保存哪些数据到 db 呢？

**1. treeIndex**

首先我们来看 MVCC 的索引模块 treeIndex，当收到更新 key hello 为 world 的时候，此 key 的索引版本号信息是怎么生成的呢？需要维护、持久化存储一个全局版本号吗？

版本号（revision）在 etcd 里面发挥着重大作用，它是 etcd 的逻辑时钟。etcd 启动的时候默认版本号是 1，随着你对 key 的增、删、改操作而全局单调递增。

因为 boltdb 中的 key 就包含此信息，所以 etcd 并不需要再去持久化一个全局版本号。我们只需要在启动的时候，从最小值 1 开始枚举到最大值，未读到数据的时候则结束，最后读出来的版本号即是当前 etcd 的最大版本号 currentRevision。

MVCC 写事务在执行 put hello 为 world 的请求时，会基于 currentRevision 自增生成新的 revision 如{2,0}，然后从 treeIndex 模块中查询 key 的创建版本号、修改次数信息。这些信息将填充到 boltdb 的 value 中，同时将用户的 hello key 和 revision 等信息存储到 B-tree，也就是下面简易写事务图的流程一，整体架构图中的流程八。
![p8](http://cdn.ipso.live/notes/etcd/etcd008.png)

**2. boltdb**

MVCC 写事务自增全局版本号后生成的 revision{2,0}，它就是 boltdb 的 key，通过它就可以往 boltdb 写数据了，进入了整体架构图中的流程九。

boltdb 上一节我们提过它是一个基于 B+tree 实现的 key-value 嵌入式 db，它通过提供桶（bucket）机制实现类似 MySQL 表的逻辑隔离。

在 etcd 里面你通过 put/txn 等 KV API 操作的数据，全部保存在一个名为 key 的桶里面，这个 key 桶在启动 etcd 的时候会自动创建。

除了保存用户 KV 数据的 key 桶，etcd 本身及其它功能需要持久化存储的话，都会创建对应的桶。比如上面我们提到的 etcd 为了保证日志的幂等性，保存了一个名为 consistent index 的变量在 db 里面，它实际上就存储在元数据（meta）桶里面。

那么写入 boltdb 的 value 含有哪些信息呢？

写入 boltdb 的 value， 并不是简单的"world"，如果只存一个用户 value，索引又是保存在易失的内存上，那重启 etcd 后，我们就丢失了用户的 key 名，无法构建 treeIndex 模块了。

因此为了构建索引和支持 Lease 等特性，etcd 会持久化以下信息:
- key 名称；
- key 创建时的版本号（create_revision）、最后一次修改时的版本号（mod_revision）、key 自身修改的次数（version）；
- value 值；
- 租约信息（后面介绍）。

boltdb value 的值就是将含以上信息的结构体序列化成的二进制数据，然后通过 boltdb 提供的 put 接口，etcd 就快速完成了将你的数据写入 boltdb，对应上面简易写事务图的流程二。

但是 put 调用成功，就能够代表数据已经持久化到 db 文件了吗？

这里需要注意的是，在以上流程中，etcd 并未提交事务（commit），因此数据只更新在 boltdb 所管理的内存数据结构中。

事务提交的过程，包含 B+tree 的平衡、分裂，将 boltdb 的脏数据（dirty page）、元数据信息刷新到磁盘，因此事务提交的开销是昂贵的。如果我们每次更新都提交事务，etcd 写性能就会较差。

那么解决的办法是什么呢？etcd 的解决方案是合并再合并。

首先 boltdb key 是版本号，put/delete 操作时，都会基于当前版本号递增生成新的版本号，因此属于顺序写入，可以调整 boltdb 的 bucket.FillPercent 参数，使每个 page 填充更多数据，减少 page 的分裂次数并降低 db 空间。

其次 etcd 通过合并多个写事务请求，通常情况下，是异步机制定时（默认每隔 100ms）将批量事务一次性提交（pending 事务过多才会触发同步提交）， 从而大大提高吞吐量，对应上面简易写事务图的流程三。

但是这优化又引发了另外的一个问题， 因为事务未提交，读请求可能无法从 boltdb 获取到最新数据。

为了解决这个问题，etcd 引入了一个 bucket buffer 来保存暂未提交的事务数据。在更新 boltdb 的时候，etcd 也会同步数据到 bucket buffer。因此 etcd 处理读请求的时候会优先从 bucket buffer 里面读取，其次再从 boltdb 读，通过 bucket buffer 实现读写性能提升，同时保证数据一致性。

#### 思考
etcd 在执行读请求过程中涉及磁盘 IO 吗？如果涉及，是什么模块在什么场景下会触发呢？如果不涉及，又是什么原因呢？

大部分人会认为 buffer 没读到，从 boltdb 读时会产生磁盘 I/O，这是一个常见误区。

实际上，etcd 在启动的时候会通过 mmap 机制将 etcd db 文件映射到 etcd 进程地址空间，并设置了 mmap 的 MAP_POPULATE flag，它会告诉 Linux 内核预读文件，Linux 内核会将文件内容拷贝到物理内存中，此时会产生磁盘 I/O。节点内存足够的请求下，后续处理读请求过程中就不会产生磁盘 I/IO 了。

若 etcd 节点内存不足，可能会导致 db 文件对应的内存页被换出，当读请求命中的页未在内存中时，就会产生缺页异常，导致读过程中产生磁盘 IO，可以通过观察 etcd 进程的 majflt 字段来判断 etcd 是否产生了主缺页中断。


---

## # Raft协议：etcd如何实现高可用、数据强一致的？
在前面的 etcd 读写流程学习中，多次提到了 etcd 是基于 Raft 协议实现高可用、数据强一致性的。

那么 etcd 是如何基于 Raft 来实现高可用、数据强一致性的呢？

这小节就以上一节中的 hello 写请求为案例，深入分析 etcd 在遇到 Leader 节点 crash 等异常后，Follower 节点如何快速感知到异常，并高效选举出新的 Leader，对外提供高可用服务的。

同时，将通过一个日志复制整体流程图，介绍 etcd 如何保障各节点数据一致性，并介绍 Raft 算法为了确保数据一致性、完整性，对 Leader 选举和日志复制所增加的一系列安全规则。希望通过这小节，了解 etcd 在节点故障、网络分区等异常场景下是如何基于 Raft 算法实现高可用、数据强一致的。

#### 如何避免单点故障

在介绍 Raft 算法之前，我们首先了解下它的诞生背景，Raft 解决了分布式系统什么痛点呢？

首先我们回想下，早期我们使用的数据存储服务，它们往往是部署在单节点上的。但是单节点存在单点故障，一宕机就整个服务不可用，对业务影响非常大。

随后，为了解决单点问题，软件系统工程师引入了数据复制技术，实现多副本。通过数据复制方案，一方面我们可以提高服务可用性，避免单点故障。另一方面，多副本可以提升读吞吐量、甚至就近部署在业务所在的地理位置，降低访问延迟。

#### 多副本复制是如何实现的呢？

多副本常用的技术方案主要有主从复制和去中心化复制。主从复制，又分为全同步复制、异步复制、半同步复制，比如 MySQL/Redis 单机主备版就基于主从复制实现的。

**全同步复制**是指主收到一个写请求后，必须等待全部从节点确认返回后，才能返回给客户端成功。因此如果一个从节点故障，整个系统就会不可用。这种方案为了保证多副本的一致性，而牺牲了可用性，一般使用不多。

**异步复制**是指主收到一个写请求后，可及时返回给 client，异步将请求转发给各个副本，若还未将请求转发到副本前就故障了，则可能导致数据丢失，但是可用性是最高的。

**半同步复制**介于全同步复制、异步复制之间，它是指主收到一个写请求后，至少有一个副本接收数据后，就可以返回给客户端成功，在数据一致性、可用性上实现了平衡和取舍。

跟主从复制相反的就是**去中心化复制**，它是指在一个 n 副本节点集群中，任意节点都可接受写请求，但一个成功的写入需要 w 个节点确认，读取也必须查询至少 r 个节点。

你可以根据实际业务场景对数据一致性的敏感度，设置合适 w/r 参数。比如你希望每次写入后，任意 client 都能读取到新值，如果 n 是 3 个副本，你可以将 w 和 r 设置为 2，这样当你读两个节点时候，必有一个节点含有最近写入的新值，这种读我们称之为法定票数读（quorum read）。

AWS 的 Dynamo 系统就是基于去中心化的复制算法实现的。它的优点是节点角色都是平等的，降低运维复杂度，可用性更高。但是缺陷是去中心化复制，势必会导致各种写入冲突，业务需要关注冲突处理。

从以上分析中，为了解决单点故障，从而引入了多副本。但基于复制算法实现的数据库，为了保证服务可用性，大多数提供的是最终一致性，总而言之，不管是主从复制还是异步复制，都存在一定的缺陷。

#### 如何解决以上复制算法的困境呢？

答案就是共识算法，它最早是基于复制状态机背景下提出来的。 它由共识模块、日志模块、状态机组成。通过共识模块保证各个节点日志的一致性，然后各个节点基于同样的日志、顺序执行指令，最终各个复制状态机的结果实现一致。

共识算法的祖师爷是 Paxos， 但是由于它过于复杂，难于理解，工程实践上也较难落地，导致在工程界落地较慢。standford 大学的 Diego 提出的 Raft 算法正是为了可理解性、易实现而诞生的，它通过问题分解，将复杂的共识问题拆分成三个子问题，分别是：
- Leader 选举，Leader 故障后集群能快速选出新 Leader；
- 日志复制， 集群只有 Leader 能写入日志， Leader 负责复制日志到 Follower 节点，并强制 Follower 节点与自己保持相同；
- 安全性，一个任期内集群只能产生一个 Leader、已提交的日志条目在发生 Leader 选举时，一定会存在更高任期的新 Leader 日志中、各个节点的状态机应用的任意位置的日志条目内容应一样等。

下面我以实际场景为案例，分别深入讨论这三个子问题，看看 Raft 是如何解决这三个问题，以及在 etcd 中的应用实现。

#### **Leader 选举**
当 etcd server 收到 client 发起的 put hello 写请求后，KV 模块会向 Raft 模块提交一个 put 提案，我们知道只有集群 Leader 才能处理写提案，如果此时集群中无 Leader， 整个请求就会超时。

那么 Leader 是怎么诞生的呢？Leader crash 之后其他节点如何竞选呢？

首先在 Raft 协议中它定义了集群中的如下节点状态，任何时刻，每个节点肯定处于其中一个状态：
- Follower，跟随者， 同步从 Leader 收到的日志，etcd 启动的时候默认为此状态；
- Candidate，竞选者，可以发起 Leader 选举；
- Leader，集群领导者， 唯一性，拥有同步日志的特权，需定时广播心跳给 Follower 节点，以维持领导者身份。

当 Follower 节点接收 Leader 节点心跳消息超时后，它会转变成 Candidate 节点，并可发起竞选 Leader 投票，若获得集群多数节点的支持后，它就可转变成 Leader 节点。

下面以 Leader crash 场景为案例，给你详细介绍一下 etcd Leader 选举原理。

假设集群总共 3 个节点，A 节点为 Leader，B、C 节点为 Follower。

正常情况下，Leader 节点会按照心跳间隔时间，定时广播心跳消息（MsgHeartbeat 消息）给 Follower 节点，以维持 Leader 身份。 Follower 收到后回复心跳应答包消息（MsgHeartbeatResp 消息）给 Leader。

你可能注意到 Leader 节点有一个任期号（term）， 它具有什么样的作用呢？

这是因为 Raft 将时间划分成一个个任期，任期用连续的整数表示，每个任期从一次选举开始，赢得选举的节点在该任期内充当 Leader 的职责，随着时间的消逝，集群可能会发生新的选举，任期号也会单调递增。

通过任期号，可以比较各个节点的数据新旧、识别过期的 Leader 等，它在 Raft 算法中充当逻辑时钟，发挥着重要作用。

了解完正常情况下 Leader 维持身份的原理后，我们再看异常情况下，也就 Leader crash 后，etcd 是如何自愈的呢？

当 Leader 节点异常后，Follower 节点会接收 Leader 的心跳消息超时，当超时时间大于竞选超时时间后，它们会进入 Candidate 状态。

这里要提醒下你，etcd 默认心跳间隔时间（heartbeat-interval）是 100ms， 默认竞选超时时间（election timeout）是 1000ms， 你需要根据实际部署环境、业务场景适当调优，否则就很可能会频繁发生 Leader 选举切换，导致服务稳定性下降，后面实践部分会再详细介绍。

进入 Candidate 状态的节点，会立即发起选举流程，自增任期号，投票给自己，并向其他节点发送竞选 Leader 投票消息（MsgVote）。

C 节点收到 Follower B 节点竞选 Leader 消息后，这时候可能会出现如下两种情况：
- 第一种情况是 C 节点判断 B 节点的数据至少和自己一样新、B 节点任期号大于 C 当前任期号、并且 C 未投票给其他候选者，就可投票给 B。这时 B 节点获得了集群多数节点支持，于是成为了新的 Leader。

- 第二种情况是，恰好 C 也心跳超时超过竞选时间了，它也发起了选举，并投票给了自己，那么它将拒绝投票给 B，这时谁也无法获取集群多数派支持，只能等待竞选超时，开启新一轮选举。Raft 为了优化选票被瓜分导致选举失败的问题，引入了随机数，每个节点等待发起选举的时间点不一致，优雅的解决了潜在的竞选活锁，同时易于理解。

Leader 选出来后，它什么时候又会变成 Follower 状态呢？ 如果现有 Leader 发现了新的 Leader 任期号，那么它就需要转换到 Follower 节点。A 节点 crash 后，再次启动成为 Follower，假设因为网络问题无法连通 B、C 节点，这时候根据状态图，我们知道它将不停自增任期号，发起选举。等 A 节点网络异常恢复后，那么现有 Leader 收到了新的任期号，就会触发新一轮 Leader 选举，影响服务的可用性。

然而 A 节点的数据是远远落后 B、C 的，是无法获得集群 Leader 地位的，发起的选举无效且对集群稳定性有伤害。

那如何避免以上场景中的无效的选举呢？

在 etcd 3.4 中，etcd 引入了一个 PreVote 参数（默认 false），可以用来启用 PreCandidate 状态解决此问题，如下图所示。Follower 在转换成 Candidate 状态前，先进入 PreCandidate 状态，不自增任期号， 发起预投票。若获得集群多数节点认可，确定有概率成为 Leader 才能进入 Candidate 状态，发起选举流程。

因 A 节点数据落后较多，预投票请求无法获得多数节点认可，因此它就不会进入 Candidate 状态，导致集群重新选举。

这就是 Raft Leader 选举核心原理，使用心跳机制维持 Leader 身份、触发 Leader 选举，etcd 基于它实现了高可用，只要集群一半以上节点存活、可相互通信，Leader 宕机后，就能快速选举出新的 Leader，继续对外提供服务。

#### 日志复制
假设在上面的 Leader 选举流程中，B 成为了新的 Leader，它收到 put 提案后，它是如何将日志同步给 Follower 节点的呢？ 什么时候它可以确定一个日志条目为已提交，通知 etcdserver 模块应用日志条目指令到状态机呢？

这就涉及到 Raft 日志复制原理，为了帮助理解日志复制的原理，下面画了一幅 Leader 收到 put 请求后，向 Follower 节点复制日志的整体流程图p9，简称流程图，在图中我用序号给你标识了核心流程。结合流程图、后面的 Raft 的日志图简要分析 Leader B 收到 put hello 为 world 的请求后，是如何将此请求同步给其他 Follower 节点的。
![p9](http://cdn.ipso.live/notes/etcd/etcd009.png)

首先 Leader 收到 client 的请求后，etcdserver 的 KV 模块会向 Raft 模块提交一个 put hello 为 world 提案消息（流程图中的序号 2 流程）， 它的消息类型是 MsgProp。

Leader 的 Raft 模块获取到 MsgProp 提案消息后，为此提案生成一个日志条目，追加到未持久化、不稳定的 Raft 日志中，随后会遍历集群 Follower 列表和进度信息，为每个 Follower 生成追加（MsgApp）类型的 RPC 消息，此消息中包含待复制给 Follower 的日志条目。

这里就出现两个疑问了。第一，Leader 是如何知道从哪个索引位置发送日志条目给 Follower，以及 Follower 已复制的日志最大索引是多少呢？第二，日志条目什么时候才会追加到稳定的 Raft 日志中呢？Raft 模块负责持久化吗？

首先介绍下什么是 Raft 日志。下图是 Raft 日志复制过程中的日志细节图，简称日志图p10。

在日志图中，最上方的是日志条目序号 / 索引，日志由有序号标识的一个个条目组成，每个日志条目内容保存了 Leader 任期号和提案内容。最开始的时候，A 节点是 Leader，任期号为 1，A 节点 crash 后，B 节点通过选举成为新的 Leader， 任期号为 2。

日志图p10 描述的是 hello 日志条目未提交前的各节点 Raft 日志状态。

![p10](http://cdn.ipso.live/notes/etcd/etcd010.png)

我们现在就可以来回答第一个疑问了。Leader 会维护两个核心字段来追踪各个 Follower 的进度信息，一个字段是 NextIndex， 它表示 Leader 发送给 Follower 节点的下一个日志条目索引。一个字段是 MatchIndex， 它表示 Follower 节点已复制的最大日志条目的索引，比如上面的日志图10 中 C 节点的已复制最大日志条目索引为 5，A 节点为 4。

我们再看第二个疑问。etcd Raft 模块设计实现上抽象了网络、存储、日志等模块，它本身并不会进行网络、存储相关的操作，上层应用需结合自己业务场景选择内置的模块或自定义实现网络、存储、日志等模块。

上层应用通过 Raft 模块的输出接口（如 Ready 结构），获取到待持久化的日志条目和待发送给 Peer 节点的消息后（如上面的 MsgApp 日志消息），需持久化日志条目到自定义的 WAL 模块，通过自定义的网络模块将消息发送给 Peer 节点。

日志条目持久化到稳定存储中后，这时候你就可以将日志条目追加到稳定的 Raft 日志中。即便这个日志是内存存储，节点重启时也不会丢失任何日志条目，因为 WAL 模块已持久化此日志条目，可通过它重建 Raft 日志。

etcd Raft 模块提供了一个内置的内存存储（MemoryStorage）模块实现，etcd 使用的就是它，Raft 日志条目保存在内存中。网络模块并未提供内置的实现，etcd 基于 HTTP 协议实现了 peer 节点间的网络通信，并根据消息类型，支持选择 pipeline、stream 等模式发送，显著提高了网络吞吐量、降低了延时。

解答完以上两个疑问后，我们继续分析 etcd 是如何与 Raft 模块交互，获取待持久化的日志条目和发送给 peer 节点的消息。

正如刚刚讲到的，Raft 模块输入是 Msg 消息，输出是一个 Ready 结构，它包含待持久化的日志条目、发送给 peer 节点的消息、已提交的日志条目内容、线性查询结果等 Raft 输出核心信息。

etcdserver 模块通过 channel 从 Raft 模块获取到 Ready 结构后（流程图中的序号 3 流程），因 B 节点是 Leader，它首先会通过基于 HTTP 协议的网络模块将追加日志条目消息（MsgApp）广播给 Follower，并同时将待持久化的日志条目持久化到 WAL 文件中（流程图中的序号 4 流程），最后将日志条目追加到稳定的 Raft 日志存储中（流程图中的序号 5 流程）。

各个 Follower 收到追加日志条目（MsgApp）消息，并通过安全检查后，它会持久化消息到 WAL 日志中，并将消息追加到 Raft 日志存储，随后会向 Leader 回复一个应答追加日志条目（MsgAppResp）的消息，告知 Leader 当前已复制的日志最大索引（流程图中的序号 6 流程）。

Leader 收到应答追加日志条目（MsgAppResp）消息后，会将 Follower 回复的已复制日志最大索引更新到跟踪 Follower 进展的 Match Index 字段，如下面的日志图p11 中的 Follower C MatchIndex 为 6，Follower A 为 5，日志图p11 描述的是 hello 日志条目提交后的各节点 Raft 日志状态。

![p11](http://cdn.ipso.live/notes/etcd/etcd011.png)

最后 Leader 根据 Follower 的 MatchIndex 信息，计算出一个位置，如果这个位置已经被一半以上节点持久化，那么这个位置之前的日志条目都可以被标记为已提交。

**在我们这个案例中日志图11 里 6 号索引位置之前的日志条目已被多数节点复制，那么他们状态都可被设置为已提交。Leader 可通过在发送心跳消息（MsgHeartbeat）给 Follower 节点时，告知它已经提交的日志索引位置。**

最后各个节点的 etcdserver 模块，可通过 channel 从 Raft 模块获取到已提交的日志条目（流程图中的序号 7 流程），应用日志条目内容到存储状态机（流程图中的序号 8 流程），返回结果给 client。

通过以上流程，Leader 就完成了同步日志条目给 Follower 的任务，一个日志条目被确定为已提交的前提是，它需要被 Leader 同步到一半以上节点上。以上就是 etcd Raft 日志复制的核心原理。

#### 安全性
介绍完 Leader 选举和日志复制后，最后我们再来看看 Raft 是如何保证安全性的。

如果在上面的日志图p11 中，Leader B 在应用日志指令 put hello 为 world 到状态机，并返回给 client 成功后，突然 crash 了，那么 Follower A 和 C 是否都有资格选举成为 Leader 呢？

从日志图p11 中我们可以看到，如果 A 成为了 Leader 那么就会导致数据丢失，因为它并未含有刚刚 client 已经写入成功的 put hello 为 world 指令。

Raft 算法如何确保面对这类问题时不丢数据和各节点数据一致性呢？

这就是 Raft 的第三个子问题需要解决的。Raft 通过给选举和日志复制增加一系列规则，来实现 Raft 算法的安全性。

**1. 选举规则**

**当节点收到选举投票的时候，需检查候选者的最后一条日志中的任期号，若小于自己则拒绝投票。如果任期号相同，日志却比自己短，也拒绝为其投票。**

比如在日志图p11 中，Folllower A 和 C 任期号相同，但是 Follower C 的数据比 Follower A 要长，那么在选举的时候，Follower C 将拒绝投票给 A， 因为它的数据不是最新的。

同时，对于一个给定的任期号，最多只会有一个 leader 被选举出来，leader 的诞生需获得集群一半以上的节点支持。每个节点在同一个任期内只能为一个节点投票，节点需要将投票信息持久化，防止异常重启后再投票给其他节点。

通过以上规则就可防止日志图p11 中的 Follower A 节点成为 Leader。

**2. 日志复制规则**

在日志图p11 中，Leader B 返回给 client 成功后若突然 crash 了，此时可能还并未将 6 号日志条目已提交的消息通知到 Follower A 和 C，那么如何确保 6 号日志条目不被新 Leader 删除呢？ 同时在 etcd 集群运行过程中，Leader 节点若频繁发生 crash 后，可能会导致 Follower 节点与 Leader 节点日志条目冲突，如何保证各个节点的同 Raft 日志位置含有同样的日志条目？

以上各类异常场景的安全性是通过 Raft 算法中的 Leader 完全特性和只附加原则、日志匹配等安全机制来保证的。

- Leader 完全特性：是指如果某个日志条目在某个任期号中已经被提交，那么这个条目必然出现在更大任期号的所有 Leader 中。

Leader 只能追加日志条目，不能删除已持久化的日志条目（只附加原则），因此 Follower C 成为新 Leader 后，会将前任的 6 号日志条目复制到 A 节点。

**为了保证各个节点日志一致性，Raft 算法在追加日志的时候，引入了一致性检查。Leader 在发送追加日志 RPC 消息时，会把新的日志条目紧接着之前的条目的索引位置和任期号包含在里面。Follower 节点会检查相同索引位置的任期号是否与 Leader 一致，一致才能追加，这就是日志匹配特性。它本质上是一种归纳法，一开始日志空满足匹配特性，随后每增加一个日志条目时，都要求上一个日志条目信息与 Leader 一致，那么最终整个日志集肯定是一致的。**

通过以上的 Leader 选举限制、Leader 完全特性、只附加原则、日志匹配等安全特性，Raft 就实现了一个可严格通过数学反证法、归纳法证明的高可用、一致性算法，为 etcd 的安全性保驾护航。

#### 小结
从如何避免单点故障说起，介绍了分布式系统中实现多副本技术的一系列方案，从主从复制到去中心化复制、再到状态机、共识算法，了解了各个方案的优缺点，以及主流存储产品的选择。

Raft 虽然诞生晚，但它却是共识算法里面在工程界应用最广泛的。它将一个复杂问题拆分成三个子问题，分别是 Leader 选举、日志复制和安全性。

Raft 通过心跳机制、随机化等实现了 Leader 选举，只要集群半数以上节点存活可相互通信，etcd 就可对外提供高可用服务。

Raft 日志复制确保了 etcd 多节点间的数据一致性，通过一个 etcd 日志复制整体流程图详细介绍了 etcd 写请求从提交到 Raft 模块，到被应用到状态机执行的各个流程，剖析了日志复制的核心原理，即一个日志条目只有被 Leader 同步到一半以上节点上，此日志条目才能称之为成功复制、已提交。Raft 的安全性，通过对 Leader 选举和日志复制增加一系列规则，保证了整个集群的一致性、完整性。

#### 思考
expensive request 是否影响写请求性能？

要搞懂这个问题，得回顾下 etcd 读写性能优化历史。

在 etcd 3.0 中，线性读请求需要走一遍 Raft 协议持久化到 WAL 日志中，因此读性能非常差，写请求肯定也会被影响。

在 etcd 3.1 中，引入了 ReadIndex 机制提升读性能，读请求无需再持久化到 WAL 中。

在 etcd 3.2 中, 优化思路转移到了 MVCC/boltdb 模块，boltdb 的事务锁由粗粒度的互斥锁，优化成读写锁，实现“N reads or 1 write”的并行，同时引入了 buffer 来提升吞吐量。问题就出在这个 buffer，读事务会加读锁，写事务结束时要升级锁更新 buffer，但是 expensive request 导致读事务长时间持有锁，最终导致写请求超时。

在 etcd 3.4 中，实现了全并发读，创建读事务的时候会全量拷贝 buffer, 读写事务不再因为 buffer 阻塞，大大缓解了 expensive request 对 etcd 性能的影响。尤其是 Kubernetes List Pod 等资源场景来说，etcd 稳定性显著提升。

---

如果一个日志完整度相对较高的节点因为自己随机时间比其他节点的长，没能最先发起竞选，其他节点当上leader后同步自己的日志岂不是冲突了？

**所说的这个日志完整度相对较高的节点，投票时有竞选规则安全限制，如果它的节点比较新会拒绝投票，至于最终先发起选举的节点能否赢得选举，要看其他节点数据情况，如果多数节点的数据比它新，那么先发起选举的节点就无法获得多数选票，如果5个节点中，只有一个节点数据比较长，那的确会被覆盖，但是这是安全的，说明这个数据并未被集群节点多数确认**

---

首先在raft中并没有什么数据结构来保存提案状态，leader只维护了一个committed index, 它表示这个index之前的日志条目已被成功同步到了大多数follower节点。当原leader crash后，其他follower节点选举出新leader, 按照raft安全性原则，它是不能删除前任leader的任何日志条目，因leader crash前这条日志条目已经被持久化到了多数follower节点上，那么follower节点选举出新leader后，它含有这条日志条目，并且多数节点已经同步了，那么对新leader而言，它的状态就是已提交，可以直接提交给状态机模块执行。**对client而言，虽然写请求超时了，但最终它的提案是成功执行的，client需要自己确保幂等性，也就是写超时后，你的提交可能是成功的。**


---

## # 鉴权：如何保护你的数据安全？

不知道你有没有过这样的困惑，当你使用 etcd 存储业务敏感数据、多租户共享使用同 etcd 集群的时候，应该如何防止匿名用户访问你的 etcd 数据呢？多租户场景又如何最小化用户权限分配，防止越权访问的？

etcd 鉴权模块就是为了解决以上痛点而生。

那么 etcd 是如何实现多种鉴权机制和细粒度的权限控制的？在实现鉴权模块的过程中最核心的挑战是什么？又该如何确保鉴权的安全性以及提升鉴权性能呢？

本小节，将为你介绍 etcd 的鉴权模块，深入剖析 etcd 如何解决上面的这些痛点和挑战。帮助掌握 etcd 鉴权模块的设计、实现精要，了解各种鉴权方案的优缺点。能在实际应用中，根据自己的业务场景、安全诉求，选择合适的方案保护你的 etcd 数据安全。同时，你也可以参考其设计、实现思想应用到自己业务的鉴权系统上。

#### 整体架构
在详细介绍 etcd 的认证、鉴权实现细节之前，从整体上介绍下 etcd 鉴权体系。

etcd 鉴权体系架构由控制面和数据面组成。

![p12](http://cdn.ipso.live/notes/etcd/etcd012.png)

上图是是 etcd 鉴权体系控制面，你可以通过客户端工具 etcdctl 和鉴权 API 动态调整认证、鉴权规则，**AuthServer 收到请求后，为了确保各节点间鉴权元数据一致性，会通过 Raft 模块进行数据同步。**

当对应的 Raft 日志条目被集群半数以上节点确认后，Apply 模块通过鉴权存储 (AuthStore) 模块，执行日志条目的内容，将规则存储到 boltdb 的一系列“鉴权表”里面。

下图是数据面鉴权流程，由认证和授权流程组成。认证的目的是检查 client 的身份是否合法、防止匿名用户访问等。目前 etcd 实现了两种认证机制，分别是密码认证和证书认证。

![p13](http://cdn.ipso.live/notes/etcd/etcd013.png)

认证通过后，为了提高密码认证性能，会分配一个 Token（类似我们生活中的门票、通信证）给 client，client 后续其他请求携带此 Token，server 就可快速完成 client 的身份校验工作。

实现分配 Token 的服务也有多种，这是 TokenProvider 所负责的，目前支持 SimpleToken 和 JWT 两种。

通过认证后，在访问 MVCC 模块之前，还需要通过授权流程。授权的目的是检查 client 是否有权限操作你请求的数据路径，etcd 实现了 RBAC 机制，支持为每个用户分配一个角色，为每个角色授予最小化的权限。

好了，etcd 鉴权体系的整个流程讲完了，下面我们就以第三小节中提到的 put hello 命令为例，深入分析以上鉴权体系是如何进行身份认证来防止匿名访问的，又是如何实现细粒度的权限控制以防止越权访问的。

#### 认证
首先我们来看第一个问题，如何防止匿名用户访问你的 etcd 数据呢？

解决方案当然是认证用户身份。那 etcd 提供了哪些机制来验证 client 身份呢?

正如上面介绍的，etcd 目前实现了两种机制，分别是用户密码认证和证书认证，下面我分别给你介绍这两种机制在 etcd 中如何实现，以及这两种机制各自的优缺点。

**1. 密码认证**

首先我们来讲讲用户密码认证。etcd 支持为每个用户分配一个账号名称、密码。密码认证在我们生活中无处不在，从银行卡取款到微信、微博 app 登录，再到核武器发射，密码认证应用及其广泛，是最基础的鉴权的方式。

但密码认证存在两大难点，它们分别是如何保障密码安全性和提升密码认证性能。

**（1）如何保障密码安全性**

我们首先来看第一个难点：如何保障密码安全性。

也许你又会说，自己可以奇思妙想构建一个加密算法，然后将密码翻译下，比如将密码中的每个字符按照字母表序替换成字母后的第 XX 个字母。然而这种加密算法，它是可逆的，一旦被黑客识别到规律，还原出你的密码后，脱库后也将导致全部账号数据泄密。

那么是否我们用一种不可逆的加密算法就行了呢？比如常见的 MD5，SHA-1，这方案听起来似乎有点道理，然而还是不严谨，因为它们的计算速度非常快，黑客可以通过暴力枚举、字典、彩虹表等手段，快速将你的密码全部破解。

LinkedIn 在 2012 年的时候 650 万用户密码被泄露，黑客 3 天就暴力破解出 90% 用户的密码，原因就是 LinkedIn 仅仅使用了 SHA-1 加密算法。

**（2）那应该如何进一步增强不可逆 hash 算法的破解难度？**

一方面我们可以使用安全性更高的 hash 算法，比如 SHA-256，它输出位数更多、计算更加复杂且耗 CPU。

另一方面我们可以在每个用户密码 hash 值的计算过程中，引入一个随机、较长的加盐 (salt) 参数，它可以让相同的密码输出不同的结果，这让彩虹表破解直接失效。

彩虹表是黑客破解密码的一种方法之一，它预加载了常用密码使用 MD5/SHA-1 计算的 hash 值，可通过 hash 值匹配快速破解你的密码。

最后我们还可以增加密码 hash 值计算过程中的开销，比如循环迭代更多次，增加破解的时间成本。

**（3）etcd 的鉴权模块如何安全存储用户密码？**

etcd 的用户密码存储正是融合了以上讨论的高安全性 hash 函数（Blowfish encryption algorithm）、随机的加盐 salt、可自定义的 hash 值计算迭代次数 cost。

下面通过几个简单 etcd 鉴权 API，为你介绍密码认证的原理。

首先你可以通过如下的 auth enable 命令开启鉴权，注意 etcd 会先要求你创建一个 root 账号，它拥有集群的最高读写权限。
```sh
$ etcdctl user add root:root
User root created
$ etcdctl auth enable
Authentication Enabled
```

启用鉴权后，这时 client 发起如下 put hello 操作时， etcd server 会返回"user name is empty"错误给 client，就初步达到了防止匿名用户访问你的 etcd 数据目的。 那么 etcd server 是在哪里做的鉴权的呢?
```sh
$ etcdctl put hello world
Error: etcdserver: user name is empty
```

etcd server 收到 put hello 请求的时候，在提交到 Raft 模块前，它会从你请求的上下文中获取你的用户身份信息。如果你未通过认证，那么在状态机应用 put 命令的时候，检查身份权限的时候发现是空，就会返回此错误给 client。

下面通过鉴权模块的 user 命令，给 etcd 增加一个 alice 账号。我们一起来看看 etcd 鉴权模块是如何基于我上面介绍的技术方案，来安全存储 alice 账号信息。
```sh
$ etcdctl user add alice:alice --user root:root
User alice created
```

鉴权模块收到此命令后，它会使用 bcrpt 库的 blowfish 算法，基于明文密码、随机分配的 salt、自定义的 cost、迭代多次计算得到一个 hash 值，并将加密算法版本、salt 值、cost、hash 值组成一个字符串，作为加密后的密码。

最后，鉴权模块将用户名 alice 作为 key，用户名、加密后的密码作为 value，存储到 boltdb 的 authUsers bucket 里面，完成一个账号创建。

当你使用 alice 账号访问 etcd 的时候，你需要先调用鉴权模块的 Authenticate 接口，它会验证你的身份合法性。

那么 etcd 如何验证你密码正确性的呢？

鉴权模块首先会根据你请求的用户名 alice，从 boltdb 获取加密后的密码，因此 hash 值包含了算法版本、salt、cost 等信息，因此可以根据你请求中的明文密码，计算出最终的 hash 值，若计算结果与存储一致，那么身份校验通过。

**2. 如何提升密码认证性能**

通过以上的鉴权安全性的深入分析，我们知道身份验证这个过程开销极其昂贵，那么问题来了，如何避免频繁、昂贵的密码计算匹配，提升密码认证的性能呢？

这就是密码认证的第二个难点，如何保证性能。

想想我们办理港澳通行证的时候，流程特别复杂，需要各种身份证明、照片、指纹信息，办理成功后，下发通信证，每次过关你只需要刷下通信证即可，高效而便捷。

那么，在软件系统领域如果身份验证通过了后，我们是否也可以返回一个类似通信证的凭据给 client，后续请求携带通信证，只要通行证合法且在有效期内，就无需再次鉴权了呢？

是的，etcd 也有类似这样的凭据。当 etcd server 验证用户密码成功后，它就会返回一个 Token 字符串给 client，用于表示用户的身份。后续请求携带此 Token，就无需再次进行密码校验，实现了通信证的效果。

**etcd 目前支持两种 Token，分别为 Simple Token 和 JWT Token。**

#### Simple Token
Simple Token 实现正如名字所言，简单。

Simple Token 的核心原理是当一个用户身份验证通过后，生成一个随机的字符串值 Token 返回给 client，并在内存中使用 map 存储用户和 Token 映射关系。当收到用户的请求时， etcd 会从请求中获取 Token 值，转换成对应的用户名信息，返回给下层模块使用。

Token 是你身份的象征，若此 Token 泄露了，那你的数据就可能存在泄露的风险。etcd 是如何应对这种潜在的安全风险呢？

etcd 生成的每个 Token，都有一个过期时间 TTL 属性，Token 过期后 client 需再次验证身份，因此可显著缩小数据泄露的时间窗口，在性能上、安全性上实现平衡。

在 etcd v3.4.9 版本中，Token 默认有效期是 5 分钟，etcd server 会定时检查你的 Token 是否过期，若过期则从 map 数据结构中删除此 Token。

不过你要注意的是，Simple Token 字符串本身并未含任何有价值信息，因此 client 无法及时、准确获取到 Token 过期时间。所以 client 不容易提前去规避因 Token 失效导致的请求报错。

从以上介绍中，你觉得 Simple Token 有哪些不足之处？为什么 etcd 社区仅建议在开发、测试环境中使用 Simple Token 呢？

首先它是有状态的，etcd server 需要使用内存存储 Token 和用户名的映射关系。

其次，它的可描述性很弱，client 无法通过 Token 获取到过期时间、用户名、签发者等信息。

etcd 鉴权模块实现的另外一个 Token Provider 方案 JWT，正是为了解决这些不足之处而生。

#### JWT Token
JWT 是 Json Web Token 缩写， 它是一个基于 JSON 的开放标准（RFC 7519）定义的一种紧凑、独立的格式，可用于在身份提供者和服务提供者间，传递被认证的用户身份信息。它由 Header、Payload、Signature 三个对象组成， 每个对象都是一个 JSON 结构体。

第一个对象是 Header，它包含 alg 和 typ 两个字段，alg 表示签名的算法，etcd 支持 RSA、ESA、PS 系列，typ 表示类型就是 JWT。
```json
{
"alg": "RS256"，
"typ": "JWT"
}
```

第二对象是 Payload，它表示载荷，包含用户名、过期时间等信息，可以自定义添加字段。
```json
{
"username": username，
"revision": revision，
"exp":      time.Now().Add(t.ttl).Unix()
}
```

第三个对象是签名，首先它将 header、payload 使用 base64 url 编码，然后将编码后的字符串用"."连接在一起，最后用我们选择的签名算法比如 RSA 系列的私钥对其计算签名，输出结果即是 Signature。

```code
signature=RSA256(
base64UrlEncode(header) + "." +
base64UrlEncode(payload)，
key)

```

**JWT 就是由 base64UrlEncode(header).base64UrlEncode(payload).signature 组成。**\

为什么说 JWT 是独立、紧凑的格式呢？

**从以上原理介绍中我们知道，它是无状态的。JWT Token 自带用户名、版本号、过期时间等描述信息，etcd server 不需要保存它，client 可方便、高效的获取到 Token 的过期时间、用户名等信息。它解决了 Simple Token 的若干不足之处，安全性更高，etcd 社区建议大家在生产环境若使用了密码认证，应使用 JWT Token( --auth-token 'jwt')，而不是默认的 Simple Token。**

在给你介绍完密码认证实现过程中的两个核心挑战，密码存储安全和性能的解决方案之后，你是否对密码认证的安全性、性能还有所担忧呢？

接下来我给你介绍 etcd 的另外一种高性能、更安全的鉴权方案，x509 证书认证。

#### **证书认证**
密码认证一般使用在 client 和 server 基于 HTTP 协议通信的内网场景中。当对安全有更高要求的时候，你需要使用 HTTPS 协议加密通信数据，防止中间人攻击和数据被篡改等安全风险。

HTTPS 是利用非对称加密实现身份认证和密钥协商，因此使用 HTTPS 协议的时候，你需要使用 CA 证书给 client 生成证书才能访问。

那么一个 client 证书包含哪些信息呢？使用证书认证的时候，etcd server 如何知道你发送的请求对应的用户名称？

我们可以使用下面的 openssl 命令查看 client 证书的内容，下图是一个 x509 client 证书的内容，它含有证书版本、序列号、签名算法、签发者、有效期、主体名等信息，我们重点要关注的是主体名中的 CN 字段。

在 etcd 中，如果你使用了 HTTPS 协议并启用了 client 证书认证 (--client-cert-auth)，它会取 CN 字段作为用户名，在我们的案例中，alice 就是 client 发送请求的用户名。
```code
openssl x509 -noout -text -in client.pem
```

证书认证在稳定性、性能上都优于密码认证。

稳定性上，它不存在 Token 过期、使用更加方便、会让你少踩坑，避免了不少 Token 失效而触发的 Bug。性能上，证书认证无需像密码认证一样调用昂贵的密码认证操作 (Authenticate 请求)，此接口支持的性能极低，后面实践时会深入讨论。

**1. 授权**

当我们使用如上创建的 alice 账号执行 put hello 操作的时候，etcd 却会返回如下的"etcdserver: permission denied"无权限错误，这是为什么呢？
```sh
$ etcdctl put hello world --user alice:alice
Error: etcdserver: permission denied
```
这是因为开启鉴权后，put 请求命令在应用到状态机前，etcd 还会对发出此请求的用户进行权限检查， 判断其是否有权限操作请求的数据。常用的权限控制方法有 ACL(Access Control List)、ABAC(Attribute-based access control)、RBAC(Role-based access control)，etcd 实现的是 RBAC 机制。

**2. RBAC**

什么是基于角色权限的控制系统 (RBAC) 呢？

它由下图中的三部分组成，User、Role、Permission。User 表示用户，如 alice。Role 表示角色，它是权限的赋予对象。Permission 表示具体权限明细，比如赋予 Role 对 key 范围在[key，KeyEnd]数据拥有什么权限。目前支持三种权限，分别是 READ、WRITE、READWRITE。

![p14](http://cdn.ipso.live/notes/etcd/etcd014.png)

下面我们通过 etcd 的 RBAC 机制，给 alice 用户赋予一个可读写[hello,helly]数据范围的读写权限， 如何操作呢?

按照上面介绍的 RBAC 原理，首先你需要创建一个 role，这里我们命名为 admin，然后新增了一个可读写[hello,helly]数据范围的权限给 admin 角色，并将 admin 的角色的权限授予了用户 alice。详细如下：
```sh
$ #创建一个admin role 
etcdctl role add admin  --user root:root
Role admin created
# #分配一个可读写[hello，helly]范围数据的权限给admin role
$ etcdctl role grant-permission admin readwrite hello helly --user root:root
Role admin updated
# 将用户alice和admin role关联起来，赋予admin权限给user
$ etcdctl user grant-role alice admin --user root:root
Role admin is granted to user alice
```

然后当你再次使用 etcdctl 执行 put hello 命令时，鉴权模块会从 boltdb 查询 alice 用户对应的权限列表。

因为有可能一个用户拥有成百上千个权限列表，etcd 为了提升权限检查的性能，引入了区间树，检查用户操作的 key 是否在已授权的区间，时间复杂度仅为 O(logN)。

在我们的这个案例中，很明显 hello 在 admin 角色可读写的[hello，helly) 数据范围内，因此它有权限更新 key hello，执行成功。你也可以尝试更新 key hey，因为此 key 未在鉴权的数据区间内，因此 etcd server 会返回"etcdserver: permission denied"错误给 client，如下所示。
```sh
$ etcdctl put hello world --user alice:alice
OK
$ etcdctl put hey hey --user alice:alice
Error: etcdserver: permission denied
```

#### 思考

**1. 哪些场景会出现 Follower 日志与 Leader 冲突？**

leader 崩溃的情况下可能 (如老的 leader 可能还没有完全复制所有的日志条目)，如果 leader 和 follower 出现持续崩溃会加剧这个现象。follower 可能会丢失一些在新的 leader 中有的日志条目，他也可能拥有一些 leader 没有的日志条目，或者两者都发生。

**2.follower 如何删除无效日志？**

leader 处理不一致是通过强制 follower 直接复制自己的日志来解决。因此在 follower 中的冲突的日志条目会被 leader 的日志覆盖。leader 会记录 follower 的日志复制进度 nextIndex，如果 follower 在追加日志时一致性检查失败，就会拒绝请求，此时 leader 就会减小 nextIndex 值并进行重试，最终在某个位置让 follower 跟 leader 一致。

为什么 WAL 日志模块只通过追加，也能删除已持久化冲突的日志条目呢？ 其实这里 etcd 在实现上采用了一些比较有技巧的方法，在 WAL 日志中的确没删除废弃的日志条目，你可以在其中搜索到冲突的日志条目。只是 etcd 加载 WAL 日志时，发现一个 raft log index 位置上有多个日志条目的时候，会通过覆盖的方式，将最后写入的日志条目追加到 raft log 中，实现了删除冲突日志条目效果


---

## # 租约：如何检测你的客户端存活？

本小节的主题是租约（Lease）。etcd 的一个典型的应用场景是 Leader 选举，那么 etcd 为什么可以用来实现 Leader 选举？核心特性实现原理又是怎样的？

今天就聊一聊 Leader 选举背后技术点之一的 Lease， 解析它的核心原理、性能优化思路，希望通过本节让你对 Lease 如何关联 key、Lease 如何高效续期、淘汰、什么是 checkpoint 机制有深入的理解。同时希望你能基于 Lease 的 TTL 特性，解决实际业务中遇到分布式锁、节点故障自动剔除等各类问题，提高业务服务的可用性。

#### 什么是 Lease
在实际业务场景中，我们常常会遇到类似 Kubernetes 的调度器、控制器组件同一时刻只能存在一个副本对外提供服务的情况。然而单副本部署的组件，是无法保证其高可用性的。

那为了解决单副本的可用性问题，我们就需要多副本部署。同时，为了保证同一时刻只有一个能对外提供服务，我们需要引入 Leader 选举机制。那么 Leader 选举本质是要解决什么问题呢？

首先当然是要保证 Leader 的唯一性，确保集群不出现多个 Leader，才能保证业务逻辑准确性，也就是安全性（Safety）、互斥性。

其次是主节点故障后，备节点应可快速感知到其异常，也就是活性（liveness）检测。实现活性检测主要有两种方案。

方案一为被动型检测，你可以通过探测节点定时拨测 Leader 节点，看是否健康，比如 Redis Sentinel。

方案二为主动型上报，Leader 节点可定期向协调服务发送"特殊心跳"汇报健康状态，若其未正常发送心跳，并超过和协调服务约定的最大存活时间后，就会被协调服务移除 Leader 身份标识。同时其他节点可通过协调服务，快速感知到 Leader 故障了，进而发起新的选举。

我们今天的主题，Lease，正是基于主动型上报模式，提供的一种活性检测机制。Lease 顾名思义，client 和 etcd server 之间存在一个约定，内容是 etcd server 保证在约定的有效期内（TTL），不会删除你关联到此 Lease 上的 key-value。

若你未在有效期内续租，那么 etcd server 就会删除 Lease 和其关联的 key-value。

你可以基于 Lease 的 TTL 特性，解决类似 Leader 选举、Kubernetes Event 自动淘汰、服务发现场景中故障节点自动剔除等问题。为了帮助理解 Lease 的核心特性原理，以一个实际场景中的经常遇到的异常节点自动剔除为案例，围绕这个问题，深入介绍 Lease 特性的实现。

在这个案例中，我们期望的效果是，在节点异常时，表示节点健康的 key 能被从 etcd 集群中自动删除。

#### Lease 整体架构
在详细解读 Lease 特性如何解决上面的问题之前，先了解下 Lease 模块的整体架构，下图是 Lease 模块简要架构图。

![p15](http://cdn.ipso.live/notes/etcd/etcd015.png)

etcd 在启动的时候，创建 Lessor 模块的时候，它会启动两个常驻 goroutine，如上图所示，一个是 RevokeExpiredLease 任务，定时检查是否有过期 Lease，发起撤销过期的 Lease 操作。一个是 CheckpointScheduledLease，定时触发更新 Lease 的剩余到期时间的操作。

Lessor 模块提供了 Grant、Revoke、LeaseTimeToLive、LeaseKeepAlive API 给 client 使用，各接口作用如下:
- Grant 表示创建一个 TTL 为你指定秒数的 Lease，Lessor 会将 Lease 信息持久化存储在 boltdb 中；
- Revoke 表示撤销 Lease 并删除其关联的数据；
- LeaseTimeToLive 表示获取一个 Lease 的有效期、剩余时间；
- LeaseKeepAlive 表示为 Lease 续期。

#### key 如何关联 Lease
了解完整体架构后，我们再看如何基于 Lease 特性实现检测一个节点存活。

首先如何为节点健康指标创建一个租约、并与节点健康指标 key 关联呢?

如 KV 模块的一样，client 可通过 clientv3 库的 Lease API 发起 RPC 调用，可以使用如下的 etcdctl 命令为 node 的健康状态指标，创建一个 Lease，有效期为 600 秒。然后通过 timetolive 命令，查看 Lease 的有效期、剩余时间。
```sh
# 创建一个TTL为600秒的lease，etcd server返回LeaseID
$ etcdctl lease grant 600
lease 326975935f48f814 granted with TTL(600s)


# 查看lease的TTL、剩余时间
$ etcdctl lease timetolive 326975935f48f814
lease 326975935f48f814 granted with TTL(600s)， remaining(590s)
```
当 Lease server 收到 client 的创建一个有效期 600 秒的 Lease 请求后，会通过 Raft 模块完成日志同步，随后 Apply 模块通过 Lessor 模块的 Grant 接口执行日志条目内容。

首先 Lessor 的 Grant 接口会把 Lease 保存到内存的 ItemMap 数据结构中，然后它需要持久化 Lease，将 Lease 数据保存到 boltdb 的 Lease bucket 中，返回一个唯一的 LeaseID 给 client。

通过这样一个流程，就基本完成了 Lease 的创建。那么节点的健康指标数据如何关联到此 Lease 上呢？

很简单，KV 模块的 API 接口提供了一个"--lease"参数，你可以通过如下命令，将 key node 关联到对应的 LeaseID 上。然后你查询的时候增加 -w 参数输出格式为 json，就可查看到 key 关联的 LeaseID。
```sh
$ etcdctl put node healthy --lease 326975935f48f818
OK
$ etcdctl get node -w=json | python -m json.tool
{
    "kvs":[
        {
            "create_revision":24，
            "key":"bm9kZQ=="，
            "Lease":3632563850270275608，
            "mod_revision":24，
            "value":"aGVhbHRoeQ=="，
            "version":1
        }
    ]
}
```
以上流程原理如下图所示，它描述了用户的 key 是如何与指定 Lease 关联的。当你通过 put 等命令新增一个指定了"--lease"的 key 时，MVCC 模块它会通过 Lessor 模块的 Attach 方法，将 key 关联到 Lease 的 key 内存集合 ItemSet 中。

![p16](http://cdn.ipso.live/notes/etcd/etcd0016.png)

一个 Lease 关联的 key 集合是保存在内存中的，那么 etcd 重启时，是如何知道每个 Lease 上关联了哪些 key 呢?

答案是 etcd 的 MVCC 模块在持久化存储 key-value 的时候，保存到 boltdb 的 value 是个结构体（mvccpb.KeyValue）， 它不仅包含你的 key-value 数据，还包含了关联的 LeaseID 等信息。因此当 etcd 重启时，可根据此信息，重建关联各个 Lease 的 key 集合列表。

#### 如何优化 Lease 续期性能
通过以上流程，我们完成了 Lease 创建和数据关联操作。在正常情况下，你的节点存活时，需要定期发送 KeepAlive 请求给 etcd 续期健康状态的 Lease，否则你的 Lease 和关联的数据就会被删除。

那么 Lease 是如何续期的? 作为一个高频率的请求 API，etcd 如何优化 Lease 续期的性能呢？

Lease 续期其实很简单，核心是将 Lease 的过期时间更新为当前系统时间加其 TTL。关键问题在于续期的性能能否满足业务诉求。

然而影响续期性能因素又是源自多方面的。首先是 TTL，TTL 过长会导致节点异常后，无法及时从 etcd 中删除，影响服务可用性，而过短，则要求 client 频繁发送续期请求。其次是 Lease 数，如果 Lease 成千上万个，那么 etcd 可能无法支撑如此大规模的 Lease 数，导致高负载。

如何解决呢？

首先我们回顾下早期 etcd v2 版本是如何实现 TTL 特性的。在早期 v2 版本中，没有 Lease 概念，TTL 属性是在 key 上面，为了保证 key 不删除，即便你的 TTL 相同，client 也需要为每个 TTL、key 创建一个 HTTP/1.x 连接，定时发送续期请求给 etcd server。

很显然，v2 老版本这种设计，因不支持连接多路复用、相同 TTL 无法复用导致性能较差，无法支撑较大规模的 Lease 场景。

etcd v3 版本为了解决以上问题，提出了 Lease 特性，TTL 属性转移到了 Lease 上， 同时协议从 HTTP/1.x 优化成 gRPC 协议。

一方面不同 key 若 TTL 相同，可复用同一个 Lease， 显著减少了 Lease 数。另一方面，通过 gRPC HTTP/2 实现了多路复用，流式传输，同一连接可支持为多个 Lease 续期，大大减少了连接数。

通过以上两个优化，实现 Lease 性能大幅提升，满足了各个业务场景诉求。

#### 如何高效淘汰过期 Lease
在了解完节点正常情况下的 Lease 续期特性后，我们再看看节点异常时，未正常续期后，etcd 又是如何淘汰过期 Lease、删除节点健康指标 key 的。

淘汰过期 Lease 的工作由 Lessor 模块的一个异步 goroutine 负责。如下面架构图 RevokeExpiredLease 部分所示，它会定时从最小堆中取出已过期的 Lease，执行删除 Lease 和其关联的 key 列表数据的 RevokeExpiredLease 任务。

![p15](http://cdn.ipso.live/notes/etcd/etcd015.png)

从图中你可以看到，目前 etcd 是基于最小堆来管理 Lease，实现快速淘汰过期的 Lease。

etcd 早期的时候，淘汰 Lease 非常暴力。etcd 会直接遍历所有 Lease，逐个检查 Lease 是否过期，过期则从 Lease 关联的 key 集合中，取出 key 列表，删除它们，时间复杂度是 O(N)。

然而这种方案随着 Lease 数增大，毫无疑问它的性能会变得越来越差。我们能否按过期时间排序呢？这样每次只需轮询、检查排在前面的 Lease 过期时间，一旦轮询到未过期的 Lease， 则可结束本轮检查。

刚刚说的就是 etcd Lease 高效淘汰方案最小堆的实现方法。每次新增 Lease、续期的时候，它会插入、更新一个对象到最小堆中，对象含有 LeaseID 和其到期时间 unixnano，对象之间按到期时间升序排序。

etcd Lessor 主循环每隔 500ms 执行一次撤销 Lease 检查（RevokeExpiredLease），每次轮询堆顶的元素，若已过期则加入到待淘汰列表，直到堆顶的 Lease 过期时间大于当前，则结束本轮轮询。

相比早期 O(N) 的遍历时间复杂度，使用堆后，插入、更新、删除，它的时间复杂度是 O(Log N)，查询堆顶对象是否过期时间复杂度仅为 O(1)，性能大大提升，可支撑大规模场景下 Lease 的高效淘汰。

获取到待过期的 LeaseID 后，Leader 是如何通知其他 Follower 节点淘汰它们呢？

Lessor 模块会将已确认过期的 LeaseID，保存在一个名为 expiredC 的 channel 中，而 etcd server 的主循环会定期从 channel 中获取 LeaseID，发起 revoke 请求，通过 Raft Log 传递给 Follower 节点。

各个节点收到 revoke Lease 请求后，获取关联到此 Lease 上的 key 列表，从 boltdb 中删除 key，从 Lessor 的 Lease map 内存中删除此 Lease 对象，最后还需要从 boltdb 的 Lease bucket 中删除这个 Lease。

以上就是 Lease 的过期自动淘汰逻辑。Leader 节点按过期时间维护了一个最小堆，若你的节点异常未正常续期，那么随着时间消逝，对应的 Lease 则会过期，Lessor 主循环定时轮询过期的 Lease。获取到 ID 后，Leader 发起 revoke 操作，通知整个集群删除 Lease 和关联的数据。

#### 为什么需要 checkpoint 机制
了解完 Lease 的创建、续期、自动淘汰机制后，你可能已经发现，检查 Lease 是否过期、维护最小堆、针对过期的 Lease 发起 revoke 操作，都是 Leader 节点负责的，它类似于 Lease 的仲裁者，通过以上清晰的权责划分，降低了 Lease 特性的实现复杂度。

那么当 Leader 因重启、crash、磁盘 IO 等异常不可用时，Follower 节点就会发起 Leader 选举，新 Leader 要完成以上职责，必须重建 Lease 过期最小堆等管理数据结构，那么以上重建可能会触发什么问题呢？

当你的集群发生 Leader 切换后，新的 Leader 基于 Lease map 信息，按 Lease 过期时间构建一个最小堆时，etcd 早期版本为了优化性能，并未持久化存储 Lease 剩余 TTL 信息，因此重建的时候就会自动给所有 Lease 自动续期了。

然而若较频繁出现 Leader 切换，切换时间小于 Lease 的 TTL，这会导致 Lease 永远无法删除，大量 key 堆积，db 大小超过配额等异常。

为了解决这个问题，etcd 引入了检查点机制，也就是下面架构图中 CheckPointScheduledLeases 部分所示的任务。

![p15](http://cdn.ipso.live/notes/etcd/etcd015.png)

一方面，etcd 启动的时候，Leader 节点后台会运行此异步任务，定期批量地将 Lease 剩余的 TTL 基于 Raft Log 同步给 Follower 节点，Follower 节点收到 CheckPoint 请求后，更新内存数据结构 LeaseMap 的剩余 TTL 信息。

另一方面，当 Leader 节点收到 KeepAlive 请求的时候，它也会通过 checkpoint 机制把此 Lease 的剩余 TTL 重置，并同步给 Follower 节点，尽量确保续期后集群各个节点的 Lease 剩余 TTL 一致性。

最后你要注意的是，此特性对性能有一定影响，目前仍然是试验特性。你可以通过 experimental-enable-lease-checkpoint 参数开启。

#### 小结
通过一个实际案例解读了 Lease 创建、关联 key、续期、淘汰、checkpoint 机制。

Lease 的核心是 TTL，当 Lease 的 TTL 过期时，它会自动删除其关联的 key-value 数据。

首先是 Lease 创建及续期。当你创建 Lease 时，etcd 会保存 Lease 信息到 boltdb 的 Lease bucket 中。为了防止 Lease 被淘汰，需要定期发送 LeaseKeepAlive 请求给 etcd server 续期 Lease，本质是更新 Lease 的到期时间。

续期的核心挑战是性能，etcd 经历了从 TTL 属性在 key 上，到独立抽象出 Lease，支持多 key 复用相同 TTL，同时协议从 HTTP/1.x 优化成 gRPC 协议，支持多路连接复用，显著降低了 server 连接数等资源开销。

其次是 Lease 的淘汰机制，etcd 的 Lease 淘汰算法经历了从时间复杂度 O(N) 到 O(Log N) 的演进，核心是轮询最小堆的 Lease 是否过期，若过期生成 revoke 请求，它会清理 Lease 和其关联的数据。

最后是 Lease 的 checkpoint 机制，它是为了解决 Leader 异常情况下 TTL 自动被续期，可能导致 Lease 永不淘汰的问题而诞生。

淘汰过期lease最小堆中保存的时间是lease到期时间，比如lease TTL是600秒/10分钟，当前时间是00:00:00, 那么到期时间00:10:00。

checkpoint最小堆中保存的时间是定时触发lease剩余TTL的同步的间隔时间,  默认是每隔5分钟触发一次同步，如果leader在00:05:00 crash了，也没开启lease剩余TTL同步操作(还剩余5分钟），那么新的leader重建后的租约时长又是10分钟了，如果你开启checkpoint机制，那么同步的就是lease剩余TTL(5分钟)。

#### 问题思考

**1. 对于Lease操作，请求是否必须有Leader接收处理。这种写请求路由是通过client3客户端直接发到leader还是通过可以通过follower转发？**

从原理上我们知道lease是leader在内存中维护过期最小堆的，因此续期操作client是必须要直接发送给leader的，如果follower节点收到了keepalive请求，会转发给leader节点。续期操作不经过raft协议处理同步，而leaseGrant/Revoke请求会经过raft协议同步给各个节点，因此任意节点都可以处理它。 如果续期操作开启了 checkpoint 机制，续期操作最终是通过checkpoint同步给follower的，也是可以认为最终走了raft协议的。

**2. 多个kv公用一个lease？**

从实际使用场景上来，我认为是TTL几乎相同，为了降低etcd server的压力而把多个kv关联在一个lease上的，比如kubernetes场景中，有大量的event, 如果一个event,一个lease, lease数量是非常多的，lease过期会触发大量写请求，这对etcd server压力非常大，为了解决这个问题对etcd server性能的影响，lease过期淘汰会默认限速每秒1000个。因此kubernetes场景为了优化lease数，会将最近一分钟内产生的event key列表,复用在同一个lease,大大降低了lease数。

**3.lease 代表 过期的一个 ttl ，多个 key 复用一个 lease 的时候，lease 是不是没有办法保存每个 key 的具体过期时间点是多少？因为每个 key 的创建时间不一样，所以过期时间也不一样？**

一般情况下不要求每个key过期时间完全一致，比如kubernetes的event,就是误差1分钟内的event key,可以复用同一个lease.

**4. 还有就是当多个 key 复用同一个 lease 的时候， 某个客户端再发送  keepalive 请求的时候，是可以直接修改lease 的剩余 ttl吗？ 若能修改的话，不就把关联到该 lease 上所有 key 的 ttl 都修改了？**

keepalive请求就是更新lease在最小堆中的过期时间(now + ttl)，可简单理解为关联到此lease上的所有key ttl都延长了。

**5. 续期操作不通过raft协议同步到follower，那如果读带lease的key是不是得经过leader处理？因为只有leader的lease过期时间是最准确的（虽然会定时checkpoint同步ttl到follower，但是我理解这个不是非常准确到）？**

不需要经过leader处理，etcd对过期时间要求没那么严格，不需要精准到毫秒级。如果lease关联的key过期了，leader会立刻发送撤销租约请求给follower，正常etcd负载情况下，这个请求同步到follower延时大概是毫秒级的，高负载、磁盘IO异常等情况下，的确可能出现比较大的延迟。


---

## # MVCC：如何实现多版本并发控制？

在第一小节介绍 etcd v2 时，提到过它存在的若干局限，如仅保留最新版本 key-value 数据、丢弃历史版本。而 etcd 核心特性 watch 又依赖历史版本，因此 etcd v2 为了缓解这个问题，会在内存中维护一个较短的全局事件滑动窗口，保留最近的 1000 条变更事件。但是在集群写请求较多等场景下，它依然无法提供可靠的 Watch 机制。

那么不可靠的 etcd v2 事件机制，在 etcd v3 中是如何解决的呢？

MVCC（Multiversion concurrency control）机制，正是为解决这个问题而诞生的。

MVCC 机制的核心思想是保存一个 key-value 数据的多个历史版本，etcd 基于它不仅实现了可靠的 Watch 机制，避免了 client 频繁发起 List Pod 等 expensive request 操作，保障 etcd 集群稳定性。而且 MVCC 还能以较低的并发控制开销，实现各类隔离级别的事务，保障事务的安全性，是事务特性的基础。

希望通过本小节，帮助搞懂 MVCC 含义和 MVCC 机制下 key-value 数据的更新、查询、删除原理，了解 treeIndex 索引模块、boltdb 模块是如何相互协作，实现保存一个 key-value 数据多个历史版本。

#### 什么是 MVCC
首先聊聊什么是 MVCC，从名字上理解，它是一个基于多版本技术实现的一种并发控制机制。那常见的并发机制有哪些？MVCC 的优点在哪里呢？

提到并发控制机制你可能就没那么陌生了，比如数据库中的悲观锁，也就是通过锁机制确保同一时刻只能有一个事务对数据进行修改操作，常见的实现方案有读写锁、互斥锁、两阶段锁等。

悲观锁是一种事先预防机制，它悲观地认为多个并发事务可能会发生冲突，因此它要求事务必须先获得锁，才能进行修改数据操作。但是悲观锁粒度过大、高并发场景下大量事务会阻塞等，会导致服务性能较差。

MVCC 机制正是基于多版本技术实现的一种乐观锁机制，它乐观地认为数据不会发生冲突，但是当事务提交时，具备检测数据是否冲突的能力。

在 MVCC 数据库中，你更新一个 key-value 数据的时候，它并不会直接覆盖原数据，而是新增一个版本来存储新的数据，每个数据都有一个版本号。版本号它是一个逻辑时间，为了方便深入理解版本号意义，下面画了一个 etcd MVCC 版本号时间序列图。

![p17](http://cdn.ipso.live/notes/etcd/etcd017.png)

从图中你可以看到，随着时间增长，你每次修改操作，版本号都会递增。每修改一次，生成一条新的数据记录。当你指定版本号读取数据时，它实际上访问的是版本号生成那个时间点的快照数据。当你删除数据的时候，它实际也是新增一条带删除标识的数据记录。

#### MVCC 特性初体验
了解完什么是 MVCC 后，先通过几个简单命令，初步体验下 MVCC 特性，看看它是如何帮助你查询历史修改记录，以及找回不小心删除的 key 的。

启动一个空集群，更新两次 key hello 后，如何获取 key hello 的上一个版本值呢？ 删除 key hello 后，还能读到历史版本吗?

如下面的命令所示，第一次 key hello 更新完后，我们通过 get 命令获取下它的 key-value 详细信息。正如你所看到的，除了 key、value 信息，还有各类版本号，后面会详细介绍它们的含义。这里重点关注 mod_revision，它表示 key 最后一次修改时的 etcd 版本号。

当我们再次更新 key hello 为 world2 后，然后通过查询时指定 key 第一次更新后的版本号，你会发现我们查询到了第一次更新的值，甚至我们执行删除 key hello 后，依然可以获得到这个值。那么 etcd 是如何实现的呢?
```sh
# 更新key hello为world1
$ etcdctl put hello world1
OK
# 通过指定输出模式为json,查看key hello更新后的详细信息
$ etcdctl get hello -w=json
{
    "kvs":[
        {
            "key":"aGVsbG8=",
            "create_revision":2,
            "mod_revision":2,
            "version":1,
            "value":"d29ybGQx"
        }
    ],
    "count":1
}
# 再次修改key hello为world2
$ etcdctl put hello world2
OK
# 确认修改成功,最新值为wolrd2
$ etcdctl get hello
hello
world2
# 指定查询版本号,获得了hello上一次修改的值
$ etcdctl get hello --rev=2
hello
world1
# 删除key hello
$ etcdctl del  hello
1
# 删除后指定查询版本号3,获得了hello删除前的值
$ etcdctl get hello --rev=3
hello
world2
```

#### 整体架构
在详细介绍 etcd 如何实现 MVCC 特性前，先从整体上介绍下 MVCC 模块。下图是 MVCC 模块的一个整体架构图，整个 MVCC 特性由 treeIndex、Backend/boltdb 组成。

![p18](http://cdn.ipso.live/notes/etcd/etcd018.png)

当你执行 MVCC 特性初体验中的 put 命令后，请求经过 gRPC KV Server、Raft 模块流转，对应的日志条目被提交后，Apply 模块开始执行此日志内容。

Apply 模块通过 MVCC 模块来执行 put 请求，持久化 key-value 数据。MVCC 模块将请求划分成两个类别，分别是读事务（ReadTxn）和写事务（WriteTxn）。读事务负责处理 range 请求，写事务负责 put/delete 操作。读写事务基于 treeIndex、Backend/boltdb 提供的能力，实现对 key-value 的增删改查功能。

**treeIndex 模块基于内存版 B-tree 实现了 key 索引管理，它保存了用户 key 与版本号（revision）的映射关系等信息。**

Backend 模块负责 etcd 的 key-value 持久化存储，主要由 ReadTx、BatchTx、Buffer 组成，ReadTx 定义了抽象的读事务接口，BatchTx 在 ReadTx 之上定义了抽象的写事务接口，Buffer 是数据缓存区。

**etcd 设计上支持多种 Backend 实现，目前实现的 Backend 是 boltdb。boltdb 是一个基于 B+ tree 实现的、支持事务的 key-value 嵌入式数据库。**

**treeIndex 与 boltdb 关系你可参考下图。当你发起一个 get hello 命令时，从 treeIndex 中获取 key 的版本号，然后再通过这个版本号，从 boltdb 获取 value 信息。boltdb 的 value 是包含用户 key-value、各种版本号、lease 信息的结构体。**

![p19](http://cdn.ipso.live/notes/etcd/etcd019.png)

接下来重点聊聊 treeIndex 模块的原理与核心数据结构。

#### treeIndex 原理
为什么需要 treeIndex 模块呢?

对于 etcd v2 来说，当你通过 etcdctl 发起一个 put hello 操作时，etcd v2 直接更新内存树，这就导致历史版本直接被覆盖，无法支持保存 key 的历史版本。在 etcd v3 中引入 treeIndex 模块正是为了解决这个问题，**支持保存 key 的历史版本，提供稳定的 Watch 机制和事务隔离等能力**。

那 etcd v3 又是如何基于 treeIndex 模块，实现保存 key 的历史版本的呢?

在前面提到过 etcd 在每次修改 key 时会生成一个全局递增的版本号（revision），然后通过数据结构 B-tree 保存用户 key 与版本号之间的关系，再以版本号作为 boltdb key，以用户的 key-value 等信息作为 boltdb value，保存到 boltdb。

**下面就介绍下，etcd 保存用户 key 与版本号映射关系的数据结构 B-tree，为什么 etcd 使用它而不使用哈希表、平衡二叉树？**

- 从 etcd 的功能特性上分析， 因 etcd 支持范围查询，因此保存索引的数据结构也必须支持范围查询才行。所以哈希表不适合，而 B-tree 支持范围查询。
- 从性能上分析，平横二叉树每个节点只能容纳一个数据、导致树的高度较高，而 B-tree 每个节点可以容纳多个数据，树的高度更低，更扁平，涉及的查找次数更少，具有优越的增、删、改、查性能。

**树的基本概念**

- 结点 (Node)： 树中的每个元素称为结点。结点包含数据和指向子结点的指针。
- 边 (Edge)： 连接两个结点的线称为边。
- 根结点 (Root)： 树的顶端结点，没有父结点。
- 父结点 (Parent)： 若一个结点包含指向另一个结点的指针，则称前者为后者的父结点。
- 子结点 (Child)： 若一个结点被另一个结点指向，则称前者为后者的子结点。
- 兄弟结点 (Sibling)： 拥有同一个父结点的结点互称为兄弟结点。
- 叶结点 (Leaf)： 没有子结点的结点称为叶结点。
- 度 (Degree)： 一个结点的子结点个数称为该结点的度。
- 树的度 (Degree of a Tree)： 树中所有结点的度的最大值称为树的度。
- 路径 (Path)： 从一个结点到另一个结点所经过的结点序列称为路径。
- 路径长度 (Path Length)： 路径上边的数量。
- 层 (Level)： 根结点为第 1 层，其子结点为第 2 层，以此类推。
- 高度 (Height)： 从根结点到最远叶子结点的最长路径上的结点数（或边的数量加 1）。
- 深度 (Depth)： 从根结点到该结点的路径长度加 1。


Google 的开源项目 btree，使用 Go 语言实现了一个内存版的 B-tree，对外提供了简单易用的接口。etcd 正是基于 btree 库实现了一个名为 treeIndex 的索引模块，通过它来查询、保存用户 key 与版本号之间的关系。

下图是个最大度（degree > 1，简称 d）为 5 的 B-tree，度是 B-tree 中的一个核心参数，它决定了你每个节点上的数据量多少、节点的“胖”、“瘦”程度。

从图中你可以看到，节点越胖，意味着一个节点可以存储更多数据，树的高度越低。在一个度为 d 的 B-tree 中，节点保存的最大 key 数为 2d - 1，否则需要进行平衡、分裂操作。这里你要注意的是在 etcd treeIndex 模块中，创建的是最大度 32 的 B-tree，也就是一个叶子节点最多可以保存 63 个 key。

![p20](http://cdn.ipso.live/notes/etcd/etcd020.png)

从图中你可以看到，你通过 put/txn 命令写入的一系列 key，treeIndex 模块基于 B-tree 将其组织起来，节点之间基于用户 key 比较大小。当你查找一个 key k95 时，通过 B-tree 的特性，你仅需通过图中流程 1 和 2 两次快速比较，就可快速找到 k95 所在的节点。

在 treeIndex 中，每个节点的 key 是一个 keyIndex 结构，etcd 就是通过它保存了用户的 key 与版本号的映射关系。

那么 keyIndex 结构包含哪些信息呢？下面是字段说明，你可以参考一下。
```go
type keyIndex struct {
   key         []byte //用户的key名称，比如我们案例中的"hello"
   modified    revision //最后一次修改key时的etcd版本号,比如我们案例中的刚写入hello为world1时的，版本号为2
   generations []generation //generation保存了一个key若干代版本号信息，每代中包含对key的多次修改的版本号列表
}
```

keyIndex 中包含用户的 key、最后一次修改 key 时的 etcd 版本号、key 的若干代（generation）版本号信息，每代中包含对 key 的多次修改的版本号列表。那我们要如何理解 generations？为什么它是个数组呢?

generations 表示一个 key 从创建到删除的过程，每代对应 key 的一个生命周期的开始与结束。当你第一次创建一个 key 时，会生成第 0 代，后续的修改操作都是在往第 0 代中追加修改版本号。当你把 key 删除后，它就会生成新的第 1 代，一个 key 不断经历创建、删除的过程，它就会生成多个代。

generation 结构详细信息如下：
```go
type generation struct {
   ver     int64    //表示此key的修改次数
   created revision //表示generation结构创建时的版本号
   revs    []revision //每次修改key时的revision追加到此数组
}
```
generation 结构中包含此 key 的修改次数、generation 创建时的版本号、对此 key 的修改版本号记录列表。

你需要注意的是版本号（revision）并不是一个简单的整数，而是一个结构体。revision 结构及含义如下：
```go
type revision struct {
   main int64    // 一个全局递增的主版本号，随put/txn/delete事务递增，一个事务内的key main版本号是一致的
   sub int64    // 一个事务内的子版本号，从0开始随事务内put/delete操作递增
}
```
revision 包含 main 和 sub 两个字段，main 是全局递增的版本号，它是个 etcd 逻辑时钟，随着 put/txn/delete 等事务递增。sub 是一个事务内的子版本号，从 0 开始随事务内的 put/delete 操作递增。

比如启动一个空集群，全局版本号默认为 1，执行下面的 txn 事务，它包含两次 put、一次 get 操作，那么按照我们上面介绍的原理，全局版本号随读写事务自增，因此是 main 为 2，sub 随事务内的 put/delete 操作递增，因此 key hello 的 revison 为{2,0}，key world 的 revision 为{2,1}。
```shell
$ etcdctl txn -i
compares:


success requests (get，put，del):
put hello 1
get hello
put world 2
```
介绍完 treeIndex 基本原理、核心数据结构后，我们再看看在 MVCC 特性初体验中的更新、查询、删除 key 案例里，treeIndex 与 boltdb 是如何协作，完成以上 key-value 操作的?

#### MVCC 更新 key 原理
当你通过 etcdctl 发起一个 put hello 操作时，如下面的 put 事务流程图流程一所示，在 put 写事务中，首先它需要从 treeIndex 模块中查询 key 的 keyIndex 索引信息，keyIndex 中存储了 key 的创建版本号、修改的次数等信息，这些信息在事务中发挥着重要作用，因此会存储在 boltdb 的 value 中。

在我们的案例中，因为是第一次创建 hello key，此时 keyIndex 索引为空。

![p21](http://cdn.ipso.live/notes/etcd/etcd021.png)

其次 etcd 会根据当前的全局版本号（空集群启动时默认为 1）自增，生成 put hello 操作对应的版本号 revision{2,0}，这就是 boltdb 的 key。

boltdb 的 value 是 mvccpb.KeyValue 结构体，它是由用户 key、value、create_revision、mod_revision、version、lease 组成。它们的含义分别如下：
- create_revision 表示此 key 创建时的版本号。在我们的案例中，key hello 是第一次创建，那么值就是 2。当你再次修改 key hello 的时候，写事务会从 treeIndex 模块查询 hello 第一次创建的版本号，也就是 keyIndex.generations[i].created 字段，赋值给 create_revision 字段；
- mod_revision 表示 key 最后一次修改时的版本号，即 put 操作发生时的全局版本号加 1；
- version 表示此 key 的修改次数。每次修改的时候，写事务会从 treeIndex 模块查询 hello 已经历过的修改次数，也就是 keyIndex.generations[i].ver 字段，将 ver 字段值加 1 后，赋值给 version 字段。

填充好 boltdb 的 KeyValue 结构体后，这时就可以通过 Backend 的写事务 batchTx 接口将 key{2,0},value 为 mvccpb.KeyValue 保存到 boltdb 的缓存中，并同步更新 buffer，如上图p21中的流程二所示。

此时存储到 boltdb 中的 key、value 数据如下：

![p22](http://cdn.ipso.live/notes/etcd/etcd022.png)

然后 put 事务需将本次修改的版本号与用户 key 的映射关系保存到 treeIndex 模块中，也就是上图p21中的流程三。

因为 key hello 是首次创建，treeIndex 模块它会生成 key hello 对应的 keyIndex 对象，并填充相关数据结构。

keyIndex 填充后的结果如下所示：
```code
key hello的keyIndex:
key:     "hello"
modified: <2,0>
generations:
[{ver:1,created:<2,0>,revisions: [<2,0>]} ]
```
简易分析一下上面的结果。
- key 为 hello，modified 为最后一次修改版本号 <2,0>，key hello 是首次创建的，因此新增一个 generation 代跟踪它的生命周期、修改记录；
- generation 的 ver 表示修改次数，首次创建为 1，后续随着修改操作递增；
- generation.created 表示创建 generation 时的版本号为 <2,0>；
- revision 数组保存对此 key 修改的版本号列表，每次修改都会将将相应的版本号追加到 revisions 数组中。

通过以上流程，一个 put 操作终于完成。

**但是此时数据还并未持久化，为了提升 etcd 的写吞吐量、性能，一般情况下（默认堆积的写事务数大于 1 万才在写事务结束时同步持久化），数据持久化由 Backend 的异步 goroutine 完成，它通过事务批量提交，定时将 boltdb 页缓存中的脏数据提交到持久化存储磁盘中，也就是图p21中的流程四。**

#### MVCC 查询 key 原理
完成 put hello 为 world1 操作后，这时你通过 etcdctl 发起一个 get hello 操作，MVCC 模块首先会创建一个读事务对象（TxnRead），在 etcd 3.4 中 Backend 实现了 ConcurrentReadTx， 也就是并发读特性。

并发读特性的核心原理是创建读事务对象时，它会全量拷贝当前写事务未提交的 buffer 数据，并发的读写事务不再阻塞在一个 buffer 资源锁上，实现了全并发读。

在读事务中，它首先需要根据 key 从 treeIndex 模块获取版本号，因我们未带版本号读，默认是读取最新的数据。treeIndex 模块从 B-tree 中，根据 key 查找到 keyIndex 对象后，匹配有效的 generation，返回 generation 的 revisions 数组中最后一个版本号{2,0}给读事务对象。

读事务对象根据此版本号为 key，通过 Backend 的并发读事务（ConcurrentReadTx）接口，优先从 buffer 中查询，命中则直接返回，否则从 boltdb 中查询此 key 的 value 信息。

那指定版本号读取历史记录又是怎么实现的呢？

当再次发起一个 put hello 为 world2 修改操作时，key hello 对应的 keyIndex 的结果如下面所示，keyIndex.modified 字段更新为 <3,0>，generation 的 revision 数组追加最新的版本号 <3,0>，ver 修改为 2。
```code
key hello的keyIndex:
key:     "hello"
modified: <3,0>
generations:
[{ver:2,created:<2,0>,revisions: [<2,0>,<3,0>]}]
```
boltdb 插入一个新的 key revision{3,0}。

这时你再发起一个指定历史版本号为 2 的读请求时，实际是读版本号为 2 的时间点的快照数据。treeIndex 模块会遍历 generation 内的历史版本号，返回小于等于 2 的最大历史版本号，在我们这个案例中，也就是 revision{2,0}，以它作为 boltdb 的 key，从 boltdb 中查询出 value 即可。

#### MVCC 删除 key 原理
介绍完 MVCC 更新、查询 key 的原理后，接着往下看。当你执行 etcdctl del hello 命令时，etcd 会立刻从 treeIndex 和 boltdb 中删除此数据吗？还是增加一个标记实现延迟删除（lazy delete）呢？

答案为 etcd 实现的是延期删除模式，原理与 key 更新类似。

与更新 key 不一样之处在于，一方面，生成的 boltdb key 版本号{4,0,t}追加了删除标识（tombstone, 简写 t），boltdb value 变成只含用户 key 的 KeyValue 结构体。另一方面 treeIndex 模块也会给此 key hello 对应的 keyIndex 对象，追加一个空的 generation 对象，表示此索引对应的 key 被删除了。

当你再次查询 hello 的时候，treeIndex 模块根据 key hello 查找到 keyindex 对象后，若发现其存在空的 generation 对象，并且查询的版本号大于等于被删除时的版本号，则会返回空。

etcdctl hello 操作后的 keyIndex 的结果如下面所示：
```code 
key hello的keyIndex:
key:     "hello"
modified: <4,0>
generations:
[
{ver:3,created:<2,0>,revisions: [<2,0>,<3,0>,<4,0>(t)]}，             
{empty}
]
```
boltdb 此时会插入一个新的 key revision{4,0,t}。

那么 key 打上删除标记后有哪些用途呢？什么时候会真正删除它呢？

一方面删除 key 时会生成 events，Watch 模块根据 key 的删除标识，会生成对应的 Delete 事件。

另一方面，当你重启 etcd，遍历 boltdb 中的 key 构建 treeIndex 内存树时，你需要知道哪些 key 是已经被删除的，并为对应的 key 索引生成 tombstone 标识。而真正删除 treeIndex 中的索引对象、boltdb 中的 key 是通过压缩 (compactor) 组件异步完成。

正因为 etcd 的删除 key 操作是基于以上延期删除原理实现的，因此只要压缩组件未回收历史版本，我们就能从 etcd 中找回误删的数据。

#### 小结
通过 MVCC 特性初体验中的更新、查询、删除 key 案例，分析了 MVCC 整体架构、核心模块，它由 treeIndex、boltdb 组成。

treeIndex 模块基于 Google 开源的 btree 库实现，它的核心数据结构 keyIndex，保存了用户 key 与版本号关系。每次修改 key 都会生成新的版本号，生成新的 boltdb key-value。boltdb 的 key 为版本号，value 包含用户 key-value、各种版本号、lease 的 mvccpb.KeyValue 结构体。

当你未带版本号查询 key 时，etcd 返回的是 key 最新版本数据。当你指定版本号读取数据时，etcd 实际上返回的是版本号生成那个时间点的快照数据。

删除一个数据时，etcd 并未真正删除它，而是基于 lazy delete 实现的异步删除。删除原理本质上与更新操作类似，只不过 boltdb 的 key 会打上删除标记，keyIndex 索引中追加空的 generation。真正删除 key 是通过 etcd 的压缩组件去异步实现的，后面会继续深入介绍。

基于以上原理特性的实现，etcd 实现了保存 key 历史版本的功能，是高可靠 Watch 机制的基础。基于 key-value 中的各种版本号信息，etcd 可提供各种级别的简易事务隔离能力。基于 Backend/boltdb 提供的 MVCC 机制，etcd 可实现读写不冲突。

#### 思考
我理解etcd采用延迟删除，1是为了保证key对应的watcher能够获取到key的所有状态信息，留给watcher时间做相应的处理。2是实时从boltdb删除key，会可能触发树的不平衡，影响其他读写请求的性能。


---

## # Watch：如何高效获取数据变化通知？

在 Kubernetes 中，各种各样的控制器实现了 Deployment、StatefulSet、Job 等功能强大的 Workload。控制器的核心思想是监听、比较资源实际状态与期望状态是否一致，若不一致则进行协调工作，使其最终一致。

那么当你修改一个 Deployment 的镜像时，Deployment 控制器是如何高效的感知到期望状态发生了变化呢？

要回答这个问题，得从 etcd 的 Watch 特性说起，它是 Kubernetes 控制器的工作基础。本小节的主要内容就是 etcd 的核心特性 Watch 机制设计实现，通过分析 Watch 机制的四大核心问题，来了解一个变化数据是如何从 0 到 1 推送给 client，并给你介绍 Watch 特性从 etcd v2 到 etcd v3 演进、优化过程。

希望通过这小节，你能在实际业务中应用 Watch 特性，快速获取数据变更通知，而不是使用可能导致大量 expensive request 的轮询模式。更进一步，将帮助掌握 Watch 过程中，可能会出现的各种异常错误和原因，并知道在业务中如何优雅处理，让你的服务更稳地运行。

#### Watch 特性初体验
在详细介绍 Watch 特性实现原理之前，先通过几个简单命令，体验下 Watch 特性。

启动一个空集群，更新两次 key hello 后，使用 Watch 特性如何获取 key hello 的历史修改记录呢？

如下所示，你可以通过下面的 watch 命令，带版本号监听 key hello，集群版本号可通过 endpoint status 命令获取，空集群启动后的版本号为 1。

执行后输出如下代码所示，两个事件记录分别对应上面的两次的修改，事件中含有 key、value、各类版本号等信息，你还可以通过比较 create_revision 和 mod_revision 区分此事件是 add 还是 update 事件。

watch 命令执行后，你后续执行的增量 put hello 修改操作，它同样可持续输出最新的变更事件给你。

```sh
$ etcdctl put hello world1
$ etcdctl put hello world2
$ etcdctl watch hello -w=json --rev=1
{
    "Events":[
        {
            "kv":{
                "key":"aGVsbG8=",
                "create_revision":2,
                "mod_revision":2,
                "version":1,
                "value":"d29ybGQx"
            }
        },
        {
            "kv":{
                "key":"aGVsbG8=",
                "create_revision":2,
                "mod_revision":3,
                "version":2,
                "value":"d29ybGQy"
            }
        }
    ],
    "CompactRevision":0,
    "Canceled":false,
    "Created":false
}
```

从以上体验中，可以看到，基于 Watch 特性，可以快速获取到你感兴趣的数据变化事件，这也是 Kubernetes 控制器工作的核心基础。在这过程中，其实有以下四大核心问题：

- 第一，client 获取事件的机制，etcd 是使用轮询模式还是推送模式呢？两者各有什么优缺点？
- 第二，事件是如何存储的？ 会保留多久？watch 命令中的版本号具有什么作用？
- 第三，当 client 和 server 端出现短暂网络波动等异常因素后，导致事件堆积时，server 端会丢弃事件吗？若你监听的历史版本号 server 端不存在了，你的代码该如何处理？
- 第四，如果你创建了上万个 watcher 监听 key 变化，当 server 端收到一个写请求后，etcd 是如何根据变化的 key 快速找到监听它的 watcher 呢？

接下来就分别详细聊聊 etcd Watch 特性是如何解决这四大问题的。搞懂这四个问题，就明白 etcd 甚至各类分布式存储 Watch 特性的核心实现原理了。

#### 轮询 vs 流式推送

首先第一个问题是 client 获取事件机制，etcd 是使用轮询模式还是推送模式呢？两者各有什么优缺点？

答案是两种机制 etcd 都使用过。

在 etcd v2 Watch 机制实现中，使用的是 HTTP/1.x 协议，实现简单、兼容性好，每个 watcher 对应一个 TCP 连接。client 通过 HTTP/1.1 协议长连接定时轮询 server，获取最新的数据变化事件。

然而当你的 watcher 成千上万的时，即使集群空负载，大量轮询也会产生一定的 QPS，server 端会消耗大量的 socket、内存等资源，导致 etcd 的扩展性、稳定性无法满足 Kubernetes 等业务场景诉求。

etcd v3 的 Watch 机制的设计实现并非凭空出现，它正是吸取了 etcd v2 的经验、教训而重构诞生的。

在 etcd v3 中，为了解决 etcd v2 的以上缺陷，使用的是基于 HTTP/2 的 gRPC 协议，双向流的 Watch API 设计，实现了连接多路复用。

HTTP/2 协议为什么能实现多路复用呢？

**在 HTTP/2 协议中，HTTP 消息被分解独立的帧（Frame），交错发送，帧是最小的数据单位。每个帧会标识属于哪个流（Stream），流由多个数据帧组成，每个流拥有一个唯一的 ID，一个数据流对应一个请求或响应包。一个连接上有并行的多个数据流，HTTP/2 可基于帧的流 ID 将并行、交错发送的帧重新组装成完整的消息。**

通过以上机制，HTTP/2 就解决了 HTTP/1 的请求阻塞、连接无法复用的问题，实现了多路复用、乱序发送。

etcd 基于以上介绍的 HTTP/2 协议的多路复用等机制，实现了一个 client/TCP 连接支持多 gRPC Stream， 一个 gRPC Stream 又支持多个 watcher。同时事件通知模式也从 client 轮询优化成 server 流式推送，极大降低了 server 端 socket、内存等资源。

当然在 etcd v3 watch 性能优化的背后，也带来了 Watch API 复杂度上升, 不过你不用担心，etcd 的 clientv3 库已经帮助你搞定这些棘手的工作了。

在 clientv3 库中，Watch 特性被抽象成 Watch、Close、RequestProgress 三个简单 API 提供给开发者使用，屏蔽了 client 与 gRPC WatchServer 交互的复杂细节，实现了一个 client 支持多个 gRPC Stream，一个 gRPC Stream 支持多个 watcher，显著降低了你的开发复杂度。

同时当 watch 连接的节点故障，clientv3 库支持自动重连到健康节点，并使用之前已接收的最大版本号创建新的 watcher，避免旧事件回放等。

#### 滑动窗口 vs MVCC
介绍完 etcd v2 的轮询机制和 etcd v3 的流式推送机制后，再看第二个问题，事件是如何存储的？ 会保留多久呢？watch 命令中的版本号具有什么作用？

第二个问题的本质是历史版本存储，etcd 经历了从滑动窗口到 MVCC 机制的演变，滑动窗口是仅保存有限的最近历史版本到内存中，而 MVCC 机制则将历史版本保存在磁盘中，避免了历史版本的丢失，极大的提升了 Watch 机制的可靠性。

etcd v3 的 MVCC 机制，就是为解决 etcd v2 Watch 机制不可靠而诞生。相比 etcd v2 直接保存事件到内存的环形数组中，etcd v3 则是将一个 key 的历史修改版本保存在 boltdb 里面。boltdb 是一个基于磁盘文件的持久化存储，因此它重启后历史事件不像 etcd v2 一样会丢失，同时你可通过配置压缩策略，来控制保存的历史版本数，在压缩部分会详细讨论它。

最后 watch 命令中的版本号具有什么作用呢?

在上一小节中我们深入介绍了它的含义，版本号是 etcd 逻辑时钟，当 client 因网络等异常出现连接闪断后，通过版本号，它就可从 server 端的 boltdb 中获取错过的历史事件，而无需全量同步，它是 etcd Watch 机制数据增量同步的核心。

#### 可靠的事件推送机制
再看第三个问题，当 client 和 server 端出现短暂网络波动等异常因素后，导致事件堆积时，server 端会丢弃事件吗？若你监听的历史版本号 server 端不存在了，你的代码该如何处理？

第三个问题的本质是可靠事件推送机制，要搞懂它，我们就得弄懂 etcd Watch 特性的整体架构、核心流程，下图是 Watch 特性整体架构图。

![p23](http://cdn.ipso.live/notes/etcd/etcd023.png)

先通过上面的架构图，简要了解下一个 watch 请求流程，让你对全流程有个整体的认识。

当你通过 etcdctl 或 API 发起一个 watch key 请求的时候，etcd 的 gRPCWatchServer 收到 watch 请求后，会创建一个 serverWatchStream, 它负责接收 client 的 gRPC Stream 的 create/cancel watcher 请求 (recvLoop goroutine)，并将从 MVCC 模块接收的 Watch 事件转发给 client(sendLoop goroutine)。

当 serverWatchStream 收到 create watcher 请求后，serverWatchStream 会调用 MVCC 模块的 WatchStream 子模块分配一个 watcher id，并将 watcher 注册到 MVCC 的 WatchableKV 模块。

在 etcd 启动的时候，WatchableKV 模块会运行 syncWatchersLoop 和 syncVictimsLoop goroutine，分别负责不同场景下的事件推送，它们也是 Watch 特性可靠性的核心之一。

从架构图中你可以看到 Watch 特性的核心实现是 WatchableKV 模块，下面我就为你抽丝剥茧，看看"etcdctl watch hello -w=json --rev=1"命令在 WatchableKV 模块是如何处理的？面对各类异常，它如何实现可靠事件推送？

etcd 核心解决方案是复杂度管理，问题拆分。

etcd 根据不同场景，对问题进行了分解，将 watcher 按场景分类，实现了轻重分离、低耦合。首先介绍下 synced watcher、unsynced watcher 它们各自的含义。

**synced watcher**，顾名思义，表示此类 watcher 监听的数据都已经同步完毕，在等待新的变更。

如果你创建的 watcher 未指定版本号 (默认 0)、或指定的版本号大于 etcd sever 当前最新的版本号 (currentRev)，那么它就会保存到 synced watcherGroup 中。watcherGroup 负责管理多个 watcher，能够根据 key 快速找到监听该 key 的一个或多个 watcher。

**unsynced watcher**，表示此类 watcher 监听的数据还未同步完成，落后于当前最新数据变更，正在努力追赶。

如果你创建的 watcher 指定版本号小于 etcd server 当前最新版本号，那么它就会保存到 unsynced watcherGroup 中。比如我们的这个案例中 watch 带指定版本号 1 监听时，版本号 1 和 etcd server 当前版本之间的数据并未同步给你，因此它就属于此类。

从以上介绍中，我们可以将可靠的事件推送机制拆分成最新事件推送、异常场景重试、历史事件推送机制三个子问题来进行分析。

下面是第一个子问题，最新事件推送机制。

#### 最新事件推送机制

当 etcd 收到一个写请求，key-value 发生变化的时候，处于 syncedGroup 中的 watcher，是如何获取到最新变化事件并推送给 client 的呢？

![p24](http://cdn.ipso.live/notes/etcd/etcd024.png)

当你创建完成 watcher 后，此时你执行 put hello 修改操作时，如上图所示，请求经过 KVServer、Raft 模块后 Apply 到状态机时，在 MVCC 的 put 事务中，它会将本次修改的后的 mvccpb.KeyValue 保存到一个 changes 数组中。

在 put 事务结束时，如下面的精简代码所示，它会将 KeyValue 转换成 Event 事件，然后回调 watchableStore.notify 函数（流程 5）。notify 会匹配出监听过此 key 并处于 synced watcherGroup 中的 watcher，同时事件中的版本号要大于等于 watcher 监听的最小版本号，才能将事件发送到此 watcher 的事件 channel 中。

serverWatchStream 的 sendLoop goroutine 监听到 channel 消息后，读出消息立即推送给 client（流程 6 和 7），至此，完成一个最新修改事件推送。

```go
evs := make([]mvccpb.Event, len(changes))
for i, change := range changes {
   evs[i].Kv = &changes[i]
   if change.CreateRevision == 0 {
      evs[i].Type = mvccpb.DELETE
      evs[i].Kv.ModRevision = rev
   } else {
      evs[i].Type = mvccpb.PUT
   }
}
tw.s.notify(rev, evs)
```

注意接收 Watch 事件 channel 的 buffer 容量默认 1024(etcd v3.4.9)。若 client 与 server 端因网络波动、高负载等原因导致推送缓慢，buffer 满了，事件会丢失吗？

这就是第二个子问题，异常场景的重试机制。

#### 异常场景重试机制
若出现 channel buffer 满了，etcd 为了保证 Watch 事件的高可靠性，并不会丢弃它，而是将此 watcher 从 synced watcherGroup 中删除，然后将此 watcher 和事件列表保存到一个名为受害者 victim 的 watcherBatch 结构中，通过异步机制重试保证事件的可靠性。

还有一个点你需要注意的是，notify 操作它是在修改事务结束时同步调用的，必须是轻量级、高性能、无阻塞的，否则会严重影响集群写性能。

那么若因网络波动、CPU 高负载等异常导致 watcher 处于 victim 集合中后，etcd 是如何处理这种 slow watcher 呢？

在介绍 Watch 机制整体架构时，我们知道 WatchableKV 模块会启动两个异步 goroutine，其中一个是 syncVictimsLoop，正是它负责 slower watcher 的堆积的事件推送。

它的基本工作原理是，遍历 victim watcherBatch 数据结构，尝试将堆积的事件再次推送到 watcher 的接收 channel 中。若推送失败，则再次加入到 victim watcherBatch 数据结构中等待下次重试。

若推送成功，watcher 监听的最小版本号 (minRev) 小于等于 server 当前版本号 (currentRev)，说明可能还有历史事件未推送，需加入到 unsynced watcherGroup 中，由下面介绍的历史事件推送机制，推送 minRev 到 currentRev 之间的事件。

若 watcher 的最小版本号大于 server 当前版本号，则加入到 synced watcher 集合中，进入上面介绍的最新事件通知机制。

下面画了一幅图总结各类 watcher 状态转换关系，希望能帮助你快速理清之间关系。

![p25](http://cdn.ipso.live/notes/etcd/etcd025.png)

介绍完最新事件推送、异常场景重试机制后，那历史事件推送机制又是怎么工作的呢？

#### 历史事件推送机制

WatchableKV 模块的另一个 goroutine，syncWatchersLoop，正是负责 unsynced watcherGroup 中的 watcher 历史事件推送。

在历史事件推送机制中，如果你监听老的版本号已经被 etcd 压缩了，client 该如何处理？

要了解这个问题，我们就得搞清楚 syncWatchersLoop 如何工作，它的核心支撑是 boltdb 中存储了 key-value 的历史版本。

syncWatchersLoop，它会遍历处于 unsynced watcherGroup 中的每个 watcher，为了优化性能，它会选择一批 unsynced watcher 批量同步，找出这一批 unsynced watcher 中监听的最小版本号。

因 boltdb 的 key 是按版本号存储的，因此可通过指定查询的 key 范围的最小版本号作为开始区间，当前 server 最大版本号作为结束区间，遍历 boltdb 获得所有历史数据。

然后将 KeyValue 结构转换成事件，匹配出监听过事件中 key 的 watcher 后，将事件发送给对应的 watcher 事件接收 channel 即可。发送完成后，watcher 从 unsynced watcherGroup 中移除、添加到 synced watcherGroup 中，如下面的 watcher 状态转换图黑色虚线框所示。

![p26](http://cdn.ipso.live/notes/etcd/etcd026.png)

若 watcher 监听的版本号已经小于当前 etcd server 压缩的版本号，历史变更数据就可能已丢失，因此 etcd server 会返回 ErrCompacted 错误给 client。client 收到此错误后，需重新获取数据最新版本号后，再次 Watch。你在业务开发过程中，使用 Watch API 最常见的一个错误之一就是未处理此错误。

#### 高效的事件匹配
介绍完可靠的事件推送机制后，最后我们再看第四个问题，如果你创建了上万个 watcher 监听 key 变化，当 server 端收到一个写请求后，etcd 是如何根据变化的 key 快速找到监听它的 watcher 呢？一个个遍历 watcher 吗？

显然一个个遍历 watcher 是最简单的方法，但是它的时间复杂度是 O(N)，**在 watcher 数较多的场景下，会导致性能出现瓶颈。更何况 etcd 是在执行一个写事务结束时，同步触发事件通知流程的**，若匹配 watcher 开销较大，将严重影响 etcd 性能。

那使用什么数据结构来快速查找哪些 watcher 监听了一个事件中的 key 呢？

也许你会说使用 map 记录下哪些 watcher 监听了什么 key 不就可以了吗？ etcd 的确使用 map 记录了监听单个 key 的 watcher，但是你要注意的是 Watch 特性不仅仅可以监听单 key，它还可以指定监听 key 范围、key 前缀，因此 etcd 还使用了如下的区间树。

![p27](http://cdn.ipso.live/notes/etcd/etcd027.png)

当收到创建 watcher 请求的时候，它会把 watcher 监听的 key 范围插入到上面的区间树中，区间的值保存了监听同样 key 范围的 watcher 集合 /watcherSet。

区间树支持快速查找一个 key 是否在某个区间内，时间复杂度 O(LogN)，因此 etcd 基于 map 和区间树实现了 watcher 与事件快速匹配，具备良好的扩展性。

#### 小结
最后我们来总结一下这小节的内容，通过一个 Watch 特性初体验，提出了 Watch 特性设计实现的四个核心问题，分别是获取事件机制、事件历史版本存储、如何实现可靠的事件推送机制、如何高效的将事件与 watcher 进行匹配。

在获取事件机制、事件历史版本存储两个问题中，我给你介绍了 etcd v2 在使用 HTTP/1.x 轮询、滑动窗口时，存在大量的连接数、丢事件等问题，导致扩展性、稳定性较差。

而 etcd v3 Watch 特性优化思路是基于 HTTP/2 的流式传输、多路复用，实现了一个连接支持多个 watcher，减少了大量连接数，事件存储也从滑动窗口优化成稳定可靠的 MVCC 机制，历史版本保存在磁盘中，具备更好的扩展性、稳定性。

在实现可靠的事件推送机制问题中，通过一个整体架构图带你了解整个 Watch 机制的核心链路，数据推送流程。

Watch 特性的核心实现模块是 watchableStore，它通过将 watcher 划分为 synced/unsynced/victim 三类，将问题进行了分解，并通过多个后台异步循环 goroutine 负责不同场景下的事件推送，提供了各类异常等场景下的 Watch 事件重试机制，尽力确保变更事件不丢失、按逻辑时钟版本号顺序推送给 client。

最后一个事件匹配性能问题，etcd 基于 map 和区间树实现了 watcher 与事件快速匹配，保障了大规模场景下的 Watch 机制性能和读写稳定性。


---

## # 事务：如何安全地实现多key操作？
在软件开发过程中，我们经常会遇到需要批量执行多个 key 操作的业务场景，比如转账案例中，Alice 给 Bob 转账 100 元，Alice 账号减少 100，Bob 账号增加 100，这涉及到多个 key 的原子更新。

无论发生任何故障，我们应用层期望的结果是，要么两个操作一起成功，要么两个一起失败。我们无法容忍出现一个成功，一个失败的情况。那么 etcd 是如何解决多 key 原子更新问题呢？

这正是我今天要和你分享的主题——事务，它就是为了简化应用层的编程模型而诞生的。将通过转账案例为你剖析 etcd 事务实现，让你了解 etcd 如何实现事务 ACID 特性的，以及 MVCC 版本号在事务中的重要作用。希望通过本节课，帮助你在业务开发中正确使用事务，保证软件代码的正确性。

#### 事务特性初体验及 API
如何使用 etcd 实现 Alice 向 Bob 转账功能呢？

在 etcd v2 的时候， **etcd 提供了 CAS（Compare and swap），然而其只支持单 key，不支持多 key**，因此无法满足类似转账场景的需求。严格意义上说 CAS 称不上事务，无法实现事务的各个隔离级别。

etcd v3 为了解决多 key 的原子操作问题，提供了全新迷你事务 API，同时基于 MVCC 版本号，它可以实现各种隔离级别的事务。它的基本结构如下：
```go
client.Txn(ctx).If(cmp1, cmp2, ...).Then(op1, op2, ...,).Else(op1, op2, …)
```

从上面结构中可以看到，事务 API 由 If 语句、Then 语句、Else 语句组成，这与我们平时常见的 MySQL 事务完全不一样。

它的基本原理是，在 If 语句中，你可以添加一系列的条件表达式，若条件表达式全部通过检查，则执行 Then 语句的 get/put/delete 等操作，否则执行 Else 的 get/put/delete 等操作。

那么 If 语句支持哪些检查项呢？

首先是 key 的最近一次修改版本号 mod_revision，简称 mod。你可以通过它检查 key 最近一次被修改时的版本号是否符合你的预期。比如当你查询到 Alice 账号资金为 100 元时，它的 mod_revision 是 v1，当你发起转账操作时，你得确保 Alice 账号上的 100 元未被挪用，这就可以通过 mod(“Alice”) = “v1” 条件表达式来保障转账安全性。

其次是 key 的创建版本号 create_revision，简称 create。你可以通过它检查 key 是否已存在。比如在分布式锁场景里，只有分布式锁 key(lock) 不存在的时候，你才能发起 put 操作创建锁，这时你可以通过 create(“lock”) = "0"来判断，因为一个 key 不存在的话它的 create_revision 版本号就是 0。

接着是 key 的修改次数 version。你可以通过它检查 key 的修改次数是否符合预期。比如你期望 key 在修改次数小于 3 时，才能发起某些操作时，可以通过 version(“key”) < "3"来判断。

最后是 key 的 value 值。你可以通过检查 key 的 value 值是否符合预期，然后发起某些操作。比如期望 Alice 的账号资金为 200, value(“Alice”) = “200”。

If 语句通过以上 MVCC 版本号、value 值、各种比较运算符 (等于、大于、小于、不等于)，实现了灵活的比较的功能，满足你各类业务场景诉求。

下面给出了一个使用 etcdctl 的 txn 事务命令，基于以上介绍的特性，初步实现的一个 Alice 向 Bob 转账 100 元的事务。

Alice 和 Bob 初始账上资金分别都为 200 元，事务首先判断 Alice 账号资金是否为 200，若是则执行转账操作，不是则返回最新资金。etcd 是如何执行这个事务的呢？这个事务实现上有哪些问题呢？

```sh
$ etcdctl txn -i
compares: //对应If语句
value("Alice") = "200" //判断Alice账号资金是否为200


success requests (get, put, del): //对应Then语句
put Alice 100 //Alice账号初始资金200减100
put Bob 300 //Bob账号初始资金200加100


failure requests (get, put, del): //对应Else语句
get Alice  
get Bob


SUCCESS


OK

OK

```

#### 整体流程

![p28](http://cdn.ipso.live/notes/etcd/etcd028.png)

在和你介绍上面案例中的 etcd 事务原理和问题前，先介绍下事务的整体流程，为后面介绍 etcd 事务 ACID 特性的实现做准备。

上图是 etcd 事务的执行流程，当你通过 client 发起一个 txn 转账事务操作时，通过 gRPC KV Server、Raft 模块处理后，在 Apply 模块执行此事务的时候，它首先对你的事务的 If 语句进行检查，也就是 ApplyCompares 操作，如果通过此操作，则执行 ApplyTxn/Then 语句，否则执行 ApplyTxn/Else 语句。

在执行以上操作过程中，它会根据事务是否只读、可写，通过 MVCC 层的读写事务对象，执行事务中的 get/put/delete 各操作，也就是我们上一节课介绍的 MVCC 对 key 的读写原理。

#### 事务 ACID 特性
了解完事务的整体执行流程后，那么 etcd 应该如何正确实现上面案例中 Alice 向 Bob 转账的事务呢？别着急，我们先来了解一下事务的 ACID 特性。在你了解了 etcd 事务 ACID 特性实现后，这个转账事务案例的正确解决方案也就简单了。

ACID 是衡量事务的四个特性，由原子性（Atomicity）、一致性（Consistency）、隔离性（Isolation）、持久性（Durability）组成。接下来我就为你分析 ACID 特性在 etcd 中的实现。

#### 原子性与持久性
事务的原子性（Atomicity）是指在一个事务中，所有请求要么同时成功，要么同时失败。比如在我们的转账案例中，是绝对无法容忍 Alice 账号扣款成功，但是 Bob 账号资金到账失败的场景。

持久性（Durability）是指事务一旦提交，其所做的修改会永久保存在数据库。

软件系统在运行过程中会遇到各种各样的软硬件故障，如果 etcd 在执行上面事务过程中，刚执行完扣款命令（put Alice 100）就突然 crash 了，它是如何保证转账事务的原子性与持久性的呢？

![p29](http://cdn.ipso.live/notes/etcd/etcd029.png)

如上图转账事务流程图所示，etcd 在执行一个事务过程中，任何时间点都可能会出现节点 crash 等异常问题。我在图中给你标注了两个关键的异常时间点，它们分别是 T1 和 T2。接下来我分别为你分析一下 etcd 在这两个关键时间点异常后，是如何保证事务的原子性和持久性的。

**T1 时间点**

T1 时间点是在 Alice 账号扣款 100 元完成时，Bob 账号资金还未成功增加时突然发生了 crash。

从前面介绍的 etcd 写原理和上面流程图我们可知，此时 MVCC 写事务持有 boltdb 写锁，仅是将修改提交到了内存中，保证幂等性、防止日志条目重复执行的一致性索引 consistent index 也并未更新。同时，负责 boltdb 事务提交的 goroutine 因无法持有写锁，也并未将事务提交到持久化存储中。

因此，T1 时间点发生 crash 异常后，事务并未成功执行和持久化任意数据到磁盘上。在节点重启时，etcd server 会重放 WAL 中的已提交日志条目，再次执行以上转账事务。因此不会出现 Alice 扣款成功、Bob 到帐失败等严重 Bug，极大简化了业务的编程复杂度。

**T2 时间点**

T2 时间点是在 MVCC 写事务完成转账，server 返回给 client 转账成功后，boltdb 的事务提交 goroutine，批量将事务持久化到磁盘中时发生了 crash。这时 etcd 又是如何保证原子性和持久性的呢?

我们知道一致性索引 consistent index 字段值是和 key-value 数据在一个 boltdb 事务里同时持久化到磁盘中的。若在 boltdb 事务提交过程中发生 crash 了，简单情况是 consistent index 和 key-value 数据都更新失败。那么当节点重启，etcd server 重放 WAL 中已提交日志条目时，同样会再次应用转账事务到状态机中，因此事务的原子性和持久化依然能得到保证。

更复杂的情况是，当 boltdb 提交事务的时候，会不会部分数据提交成功，部分数据提交失败呢？这个问题，将在下一小节通过深入介绍 boltdb 为你解答。

了解完 etcd 事务的原子性和持久性后，那一致性又是怎么一回事呢？事务的一致性难道是指各个节点数据一致性吗？

#### 一致性
在软件系统中，到处可见一致性（Consistency）的表述，其实在不同场景下，它的含义是不一样的。

首先分布式系统中多副本数据一致性，它是指各个副本之间的数据是否一致，比如 Redis 的主备是异步复制的，那么它的一致性是最终一致性的。

其次是 CAP 原理中的一致性是指可线性化。核心原理是虽然整个系统是由多副本组成，但是通过线性化能力支持，对 client 而言就如一个副本，应用程序无需关心系统有多少个副本。

然后是一致性哈希，它是一种分布式系统中的数据分片算法，具备良好的分散性、平衡性。

最后是事务中的一致性，它是指事务变更前后，数据库必须满足若干恒等条件的状态约束，一致性往往是由数据库和业务程序两方面来保障的。

在 Alice 向 Bob 转账的案例中有哪些恒等状态呢？

很明显，转账系统内的各账号资金总额，在转账前后应该一致，同时各账号资产不能小于 0。

为了更好地理解前面转账事务实现的问题，下面画了幅两个并发转账事务的流程图。

图中有两个并发的转账事务，Mike 向 Bob 转账 100 元，Alice 也向 Bob 转账 100 元，按照我们上面的事务实现，从下图可知转账前系统总资金是 600 元，转账后却只有 500 元了，因此它无法保证转账前后账号系统内的资产一致性，导致了资产凭空消失，破坏了事务的一致性。

![p30](http://cdn.ipso.live/notes/etcd/etcd030.png)

事务一致性被破坏的根本原因是，事务中缺少对 Bob 账号资产是否发生变化的判断，这就导致账号资金被覆盖。

为了确保事务的一致性，一方面，业务程序在转账逻辑里面，需检查转账者资产大于等于转账金额。在事务提交时，通过账号资产的版本号，确保双方账号资产未被其他事务修改。若双方账号资产被其他事务修改，账号资产版本号会检查失败，这时业务可以通过获取最新的资产和版本号，发起新的转账事务流程解决。

另一方面，etcd 会通过 WAL 日志和 consistent index、boltdb 事务特性，去确保事务的原子性，因此不会有部分成功部分失败的操作，导致资金凭空消失、新增。

介绍完事务的原子性和持久化、一致性后，我们再看看 etcd 又是如何提供各种隔离级别的事务，在转账过程中，其他 client 能看到转账的中间状态吗 (如 Alice 扣款成功，Bob 还未增加时)？

#### 隔离性

ACID 中的 I 是指 Isolation，也就是事务的隔离性，它是指事务在执行过程中的可见性。常见的事务隔离级别有以下四种。

首先是未提交读（Read UnCommitted），也就是一个 client 能读取到未提交的事务。比如转账事务过程中，Alice 账号资金扣除后，Bob 账号上资金还未增加，这时如果其他 client 读取到这种中间状态，它会发现系统总金额钱减少了，破坏了事务一致性的约束。

其次是已提交读（Read Committed），指的是只能读取到已经提交的事务数据，但是存在不可重复读的问题。比如事务开始时，你读取了 Alice 和 Bob 资金，这时其他事务修改 Alice 和 Bob 账号上的资金，你在事务中再次读取时会读取到最新资金，导致两次读取结果不一样。

接着是可重复读（Repeated Read），它是指在一个事务中，同一个读操作 get Alice/Bob 在事务的任意时刻都能得到同样的结果，其他修改事务提交后也不会影响你本事务所看到的结果。

最后是串行化（Serializable），它是最高的事务隔离级别，读写相互阻塞，通过牺牲并发能力、串行化来解决事务并发更新过程中的隔离问题。对于串行化我要和你特别补充一点，很多人认为它都是通过读写锁，来实现事务一个个串行提交的，其实这只是在基于锁的并发控制数据库系统实现而已。为了优化性能，在基于 MVCC 机制实现的各个数据库系统中，提供了一个名为“可串行化的快照隔离”级别，相比悲观锁而言，它是一种乐观并发控制，通过快照技术实现的类似串行化的效果，事务提交时能检查是否冲突。

下面重点介绍下未提交读、已提交读、可重复读、串行化快照隔离。

**未提交读**

首先是最低的事务隔离级别，未提交读。我们通过如下一个转账事务时间序列图，来分析下一个 client 能否读取到未提交事务修改的数据，是否存在脏读。

![p31](http://cdn.ipso.live/notes/etcd/etcd031.png)

图中有两个事务，一个是用户查询 Alice 和 Bob 资产的事务，一个是我们执行 Alice 向 Bob 转账的事务。

如图中所示，若在 Alice 向 Bob 转账事务执行过程中，etcd server 收到了 client 查询 Alice 和 Bob 资产的读请求，显然此时我们无法接受 client 能读取到一个未提交的事务，因为这对应用程序而言会产生严重的 BUG。那么 etcd 是如何保证不出现这种场景呢？

我们知道 etcd 基于 boltdb 实现读写操作的，读请求由 boltdb 的读事务处理，你可以理解为快照读。写请求由 boltdb 写事务处理，etcd 定时将一批写操作提交到 boltdb 并清空 buffer。

由于 etcd 是批量提交写事务的，而读事务又是快照读，因此当 MVCC 写事务完成时，它需要更新 buffer，这样下一个读请求到达时，才能从 buffer 中获取到最新数据。

在我们的场景中，转账事务并未结束，执行 put Alice 为 100 的操作不会回写 buffer，因此避免了脏读的可能性。用户此刻从 boltdb 快照读事务中查询到的 Alice 和 Bob 资产都为 200。

从以上分析可知，etcd 并未使用悲观锁来解决脏读的问题，而是通过 MVCC 机制来实现读写不阻塞，并解决脏读的问题。

**已提交读、可重复读**

比未提交读隔离级别更高的是已提交读，它是指在事务中能读取到已提交数据，但是存在不可重复读的问题。已提交读，也就是说你每次读操作，若未增加任何版本号限制，默认都是当前读，etcd 会返回最新已提交的事务结果给你。

如何理解不可重复读呢?

在上面用户查询 Alice 和 Bob 事务的案例中，第一次查出来资产都是 200，第二次是 Alice 为 100，Bob 为 300，通过读已提交模式，你能及时获取到 etcd 最新已提交的事务结果，但是出现了不可重复读，两次读出来的 Alice 和 Bob 资产不一致。

那么如何实现可重复读呢？

你可以通过 MVCC 快照读，或者参考 etcd 的事务框架 STM 实现，它在事务中维护一个读缓存，优先从读缓存中查找，不存在则从 etcd 查询并更新到缓存中，这样事务中后续读请求都可从缓存中查找，确保了可重复读。

最后我们再来重点介绍下什么是串行化快照隔离。

**串行化快照隔离**

串行化快照隔离是最严格的事务隔离级别，它是指在在事务刚开始时，首先获取 etcd 当前的版本号 rev，事务中后续发出的读请求都带上这个版本号 rev，告诉 etcd 你需要获取那个时间点的快照数据，etcd 的 MVCC 机制就能确保事务中能读取到同一时刻的数据。

同时，它还要确保事务提交时，你读写的数据都是最新的，未被其他人修改，也就是要增加冲突检测机制。当事务提交出现冲突的时候依赖 client 重试解决，安全地实现多 key 原子更新。

那么我们应该如何为上面一致性案例中，两个并发转账的事务，增加冲突检测机制呢？

核心就是我们前面介绍 MVCC 的版本号，我通过下面的并发转账事务流程图为你解释它是如何工作的。

![p32](http://cdn.ipso.live/notes/etcd/etcd032.png)

如上图所示，事务 A，Alice 向 Bob 转账 100 元，事务 B，Mike 向 Bob 转账 100 元，两个事务同时发起转账操作。

一开始时，Mike 的版本号 (指 mod_revision) 是 4，Bob 版本号是 3，Alice 版本号是 2，资产各自 200。为了防止并发写事务冲突，etcd 在一个写事务开始时，会独占一个 MVCC 读写锁。

事务 A 会先去 etcd 查询当前 Alice 和 Bob 的资产版本号，用于在事务提交时做冲突检测。在事务 A 查询后，事务 B 获得 MVCC 写锁并完成转账事务，Mike 和 Bob 账号资产分别为 100，300，版本号都为 5。

事务 B 完成后，事务 A 获得写锁，开始执行事务。

为了解决并发事务冲突问题，事务 A 中增加了冲突检测，期望的 Alice 版本号应为 2，Bob 为 3。结果事务 B 的修改导致 Bob 版本号变成了 5，因此此事务会执行失败分支，再次查询 Alice 和 Bob 版本号和资产，发起新的转账事务，成功通过 MVCC 冲突检测规则 mod(“Alice”) = 2 和 mod(“Bob”) = 5 后，更新 Alice 账号资产为 100，Bob 资产为 400，完成转账操作。

通过上面介绍的快照读和 MVCC 冲突检测检测机制，etcd 就可实现串行化快照隔离能力。

#### 转账案例应用
介绍完 etcd 事务 ACID 特性实现后，你很容易发现事务特性初体验中的案例问题了，它缺少了完整事务的冲突检测机制。

首先你可通过一个事务获取 Alice 和 Bob 账号的上资金和版本号，用以判断 Alice 是否有足够的金额转账给 Bob 和事务提交时做冲突检测。 你可通过如下 etcdctl txn 命令，获取 Alice 和 Bob 账号的资产和最后一次修改时的版本号 (mod_revision):
```sh
$ etcdctl txn -i -w=json
compares:


success requests (get, put, del):
get Alice
get Bob


failure requests (get, put, del):


{
 "kvs":[
      {
          "key":"QWxpY2U=",
          "create_revision":2,
          "mod_revision":2,
          "version":1,
          "value":"MjAw"
      }
  ],
    ......
  "kvs":[
      {
          "key":"Qm9i",
          "create_revision":3,
          "mod_revision":3,
          "version":1,
          "value":"MzAw"
      }
  ],
}
```

其次发起资金转账操作，Alice 账号减去 100，Bob 账号增加 100。为了保证转账事务的准确性、一致性，提交事务的时候需检查 Alice 和 Bob 账号最新修改版本号与读取资金时的一致 (compares 操作中增加版本号检测)，以保证其他事务未修改两个账号的资金。

若 compares 操作通过检查，则执行转账操作，否则执行查询 Alice 和 Bob 账号资金操作，命令如下:
```sh
$ etcdctl txn -i
compares:
mod("Alice") = "2"
mod("Bob") = "3"


success requests (get, put, del):
put Alice 100
put Bob 300


failure requests (get, put, del):
get Alice
get Bob


SUCCESS


OK

OK
```

到这里我们就完成了一个安全的转账事务操作，从以上流程中你可以发现，自己从 0 到 1 实现一个完整的事务还是比较繁琐的，幸运的是，etcd 社区基于以上介绍的事务特性，提供了一个简单的事务框架[STM](https://github.com/etcd-io/etcd/blob/main/client/v3/concurrency/stm.go)，构建了各个事务隔离级别类，帮助你进一步简化应用编程复杂度。

#### 小结

最后我们来小结下今天的内容。首先我给你介绍了事务 API 的基本结构，它由 If、Then、Else 语句组成。

其中 If 支持多个比较规则，它是用于事务提交时的冲突检测，比较的对象支持 key 的 mod_revision、create_revision、version、value 值。随后介绍了整个事务执行的基本流程，Apply 模块首先执行 If 的比较规则，为真则执行 Then 语句，否则执行 Else 语句。

接着通过转账案例，四幅转账事务时间序列图，分析了事务的 ACID 特性，剖析了在 etcd 中事务的 ACID 特性的实现。

- 原子性是指一个事务要么全部成功要么全部失败，etcd 基于 WAL 日志、consistent index、boltdb 的事务能力提供支持。
- 一致性是指事务转账前后的，数据库和应用程序期望的恒等状态应该保持不变，这通过数据库和业务应用程序相互协作完成。
- 持久性是指事务提交后，数据不丢失，
- 隔离性是指事务提交过程中的可见性，etcd 不存在脏读，基于 MVCC 机制、boltdb 事务你可以实现可重复读、串行化快照隔离级别的事务，保障并发事务场景中你的数据安全性。


---

## # boltdb：如何持久化存储你的key-value数据？

在前面，已经多次提到过 etcd 数据存储在 boltdb。那么 boltdb 是如何组织 key-value 数据的呢？当读写一个 key 时，boltdb 是如何工作的？

现在将通过一个写请求在 boltdb 中执行的简要流程，分析其背后的 boltdb 的磁盘文件布局，帮助了解 page、node、bucket 等核心数据结构的原理与作用，搞懂 boltdb 基于 B+ tree、各类 page 实现查找、更新、事务提交的原理，让你明白 etcd 为什么适合读多写少的场景。

#### boltdb 磁盘布局
在介绍一个 put 写请求在 boltdb 中执行原理前，先从整体上介绍下平时所看到的 etcd db 文件的磁盘布局，来了解下 db 文件的物理存储结构。

boltdb 文件指的是你 etcd 数据目录下的 member/snap/db 的文件， etcd 的 key-value、lease、meta、member、cluster、auth 等所有数据存储在其中。etcd 启动的时候，会通过 mmap 机制将 db 文件映射到内存，后续可从内存中快速读取文件中的数据。写请求通过 fwrite 和 fdatasync 来写入、持久化数据到磁盘。

![p33](http://cdn.ipso.live/notes/etcd/etcd033.png)

上图是画的 db 文件磁盘布局，从图中的左边部分你可以看到，文件的内容由若干个 page 组成，一般情况下 page size 为 4KB。

page 按照功能可分为元数据页 (meta page)、B+ tree 索引节点页 (branch page)、B+ tree 叶子节点页 (leaf page)、空闲页管理页 (freelist page)、空闲页 (free page)。

文件最开头的两个 page 是固定的 db 元数据 meta page，空闲页管理页记录了 db 中哪些页是空闲、可使用的。索引节点页保存了 B+ tree 的内部节点，如图中的右边部分所示，它们记录了 key 值，叶子节点页记录了 B+ tree 中的 key-value 和 bucket 数据。

boltdb 逻辑上通过 B+ tree 来管理 branch/leaf page， 实现快速查找、写入 key-value 数据。

#### boltdb API

了解完 boltdb 的磁盘布局后，那么如果要在 etcd 中执行一个 put 请求，boltdb 中是如何执行的呢？ boltdb 作为一个库，提供了什么 API 给 client 访问写入数据？

boltdb 提供了非常简单的 API 给上层业务使用，当我们执行一个 put hello 为 world 命令时，boltdb 实际写入的 key 是版本号，value 为 mvccpb.KeyValue 结构体。

这里我们简化下，假设往 key bucket 写入一个 key 为 r94，value 为 world 的字符串，其核心代码如下：
```go
func main(){
  // 打开boltdb文件，获取db对象
  db,err := bolt.Open("db"， 0600， nil)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()
  // 参数true表示创建一个写事务，false读事务
  tx,err := db.Begin(true)
  if err != nil {
    return err
  }
  defer tx.Rollback()
  // 使用事务对象创建key bucket
  b,err := tx.CreatebucketIfNotExists([]byte("key"))
  if err != nil {
    return err
  }
  // 使用bucket对象更新一个key
  if err := b.Put([]byte("r94"),[]byte("world")); err != nil {
    return err
  }
  // 提交事务
  if err := tx.Commit(); err != nil {
    return err
  }
}
```
如上所示，通过 boltdb 的 Open API，我们获取到 boltdb 的核心对象 db 实例后，然后通过 db 的 Begin API 开启写事务，获得写事务对象 tx。

通过写事务对象 tx， 你可以创建 bucket。这里我们创建了一个名为 key 的 bucket（如果不存在），并使用 bucket API 往其中更新了一个 key 为 r94，value 为 world 的数据。最后我们使用写事务的 Commit 接口提交整个事务，完成 bucket 创建和 key-value 数据写入。

看起来是不是非常简单，神秘的 boltdb，并未有我们想象的那么难。然而其 API 简单的背后却是 boltdb 的一系列巧妙的设计和实现。

一个 key-value 数据如何知道该存储在 db 在哪个 page？如何快速找到你的 key-value 数据？事务提交的原理又是怎样的呢？

接下来就浅析 boltdb 背后的奥秘。

#### 核心数据结构介绍
上面介绍 boltdb 的磁盘布局时提到，**boltdb 整个文件由一个个 page 组成。最开头的两个 page 描述 db 元数据信息**，而它正是在 client 调用 boltdb Open API 时被填充的。那么描述磁盘页面的 page 数据结构是怎样的呢？元数据页又含有哪些核心数据结构？

boltdb 本身自带了一个工具 bbolt，它可以按页打印出 db 文件的十六进制的内容，下面我们就使用此工具来揭开 db 文件的神秘面纱。

下图左边的十六进制是执行如下bbolt dump命令，所打印的 boltdb 第 0 页的数据，图的右边是对应的 page 磁盘页结构和 meta page 的数据结构。
```sh
$ ./bbolt dump ./infra1.etcd/member/snap/db 0
```

![p34](http://cdn.ipso.live/notes/etcd/etcd034.png)

一看上图中的十六进制数据，你可能很懵，没关系，在你了解 page 磁盘页结构、meta page 数据结构后，你就能读懂其含义了。

#### page 磁盘页结构
我们先了解下 page 磁盘页结构，如上图所示，它由页 ID(id)、页类型 (flags)、数量 (count)、溢出页数量 (overflow)、页面数据起始位置 (ptr) 字段组成。

页类型目前有如下四种：0x01 表示 branch page，0x02 表示 leaf page，0x04 表示 meta page，0x10 表示 freelist page。

数量字段仅在页类型为 leaf 和 branch 时生效，溢出页数量是指当前页面数据存放不下，需要向后再申请 overflow 个连续页面使用，页面数据起始位置指向 page 的载体数据，比如 meta page、branch/leaf 等 page 的内容。

#### meta page 数据结构
meta page 数据结构第 0、1 页我们知道它是固定存储 db 元数据的页 (meta page)，那么 meta page 它为了管理整个 boltdb 含有哪些信息呢？

如上图中的 meta page 数据结构所示，你可以看到它由 boltdb 的文件标识 (magic)、版本号 (version)、页大小 (pagesize)、boltdb 的根 bucket 信息 (root bucket)、freelist 页面 ID(freelist)、总的页面数量 (pgid)、上一次写事务 ID(txid)、校验码 (checksum) 组成。

#### bucket 数据结构
如下命令所示，你可以使用 bbolt buckets 命令，输出一个 db 文件的 bucket 列表。执行完此命令后，我们可以看到之前介绍过的 auth/lease/meta 等熟悉的 bucket，它们都是 etcd 默认创建的。那么 boltdb 是如何存储、管理 bucket 的呢？
```sh
$ ./bbolt buckets  ./infra1.etcd/member/snap/db
alarm
auth
authRoles
authUsers
cluster
key
lease
members
members_removed
meta

```
在上面我们提到过 meta page 中的，有一个名为 root、类型 bucket 的重要数据结构，如下所示，bucket 由 root 和 sequence 两个字段组成，root 表示该 bucket 根节点的 page id。注意 meta page 中的 bucket.root 字段，存储的是 db 的 root bucket 页面信息，你所看到的 key/lease/auth 等 bucket 都是 root bucket 的子 bucket。
```go
type bucket struct {
   root     pgid   // page id of the bucket's root-level page
   sequence uint64 // monotonically incrementing, used by NextSequence()
}
```

我们可以通过如下 bbolt pages 命令看看各个 page 类型和元素数量，从下图结果可知，4 号页面为 leaf page。
```sh
$ ./bbolt pages  ./infra1.etcd/member/snap/db
ID       TYPE       ITEMS  OVRFLW
======== ========== ====== ======
0        meta       0
1        meta       0
2        free
3        freelist   2
4        leaf       10
5        free
```
通过上面的分析可知，当 bucket 比较少时，我们子 bucket 数据可直接从 meta page 里指向的 leaf page 中找到。

#### leaf page
leaf page 的磁盘布局，前半部分是 leafPageElement 数组，后半部分是 key-value 数组。

leafPageElement 包含 leaf page 的类型 flags， 通过它可以区分存储的是 bucket 名称还是 key-value 数据。

当 flag 为 bucketLeafFlag(0x01) 时，表示存储的是 bucket 数据，否则存储的是 key-value 数据，leafPageElement 它还含有 key-value 的读取偏移量，key-value 大小，根据偏移量和 key-value 大小，我们就可以方便地从 leaf page 中解析出所有 key-value 对。

当存储的是 bucket 数据的时候，key 是 bucket 名称，value 则是 bucket 结构信息。bucket 结构信息含有 root page 信息，通过 root page（基于 B+ tree 查找算法），你可以快速找到你存储在这个 bucket 下面的 key-value 数据所在页面。

从上面分析你可以看到，每个子 bucket 至少需要一个 page 来存储其下面的 key-value 数据，如果子 bucket 数据量很少，就会造成磁盘空间的浪费。实际上 boltdb 实现了 inline bucket，在满足一些条件限制的情况下，可以将小的子 bucket 内嵌在它的父亲叶子节点上，友好的支持了大量小 bucket。

为了方便快速理解核心原理，这里讨论的 bucket 是假设都是非 inline bucket。

那么 boltdb 是如何管理大量 bucket、key-value 的呢？

#### branch page
boltdb 使用了 B+ tree 来高效管理所有子 bucket 和 key-value 数据，因此它可以支持大量的 bucket 和 key-value，只不过 B+ tree 的根节点不再直接指向 leaf page，而是 branch page 索引节点页。branch page flags 为 0x01。它的磁盘布局，前半部分是 branchPageElement 数组，后半部分是 key 数组。

branchPageElement 包含 key 的读取偏移量、key 大小、子节点的 page id。根据偏移量和 key 大小，我们就可以方便地从 branch page 中解析出所有 key，然后二分搜索匹配 key，获取其子节点 page id，递归搜索，直至从 bucketLeafFlag 类型的 leaf page 中找到目的 bucket name。

注意，boltdb 在内存中使用了一个名为 node 的数据结构，来保存 page 反序列化的结果。下面我给出了一个 boltdb 读取 page 到 node 的代码片段，你可以直观感受下。
```go
func (n *node) read(p *page) {
   n.pgid = p.id
   n.isLeaf = ((p.flags & leafPageFlag) != 0)
   n.inodes = make(inodes, int(p.count))


   for i := 0; i < int(p.count); i++ {
      inode := &n.inodes[i]
      if n.isLeaf {
         elem := p.leafPageElement(uint16(i))
         inode.flags = elem.flags
         inode.key = elem.key()
         inode.value = elem.value()
      } else {
         elem := p.branchPageElement(uint16(i))
         inode.pgid = elem.pgid
         inode.key = elem.key()
      }
   }
}
```
从上面分析过程中你会发现，boltdb 存储 bucket 和 key-value 原理是类似的，将 page 划分成 branch page、leaf page，通过 B+ tree 来管理实现。boltdb 为了区分 leaf page 存储的数据类型是 bucket 还是 key-value，增加了标识字段（leafPageElement.flags），因此 key-value 的数据存储过程就不再重复分析了。

#### freelist
介绍完 bucket、key-value 存储原理后，我们再看 meta page 中的另外一个核心字段 freelist，它的作用是什么呢？

我们知道 boltdb 将 db 划分成若干个 page，那么它是如何知道哪些 page 在使用中，哪些 page 未使用呢？

答案是 boltdb 通过 meta page 中的 freelist 来管理页面的分配，freelist page 中记录了哪些页是空闲的。当你在 boltdb 中删除大量数据的时候，其对应的 page 就会被释放，页 ID 存储到 freelist 所指向的空闲页中。当你写入数据的时候，就可直接从空闲页中申请页面使用。

可以通过 bbolt page 命令查看 page 内容，如下所示，它记录了 2 和 5 为空闲页
```sh
$ ./bbolt page  ./infra1.etcd/member/snap/db 3
page ID:    3
page Type:  freelist
Total Size: 4096 bytes
Item Count: 2
Overflow: 0

2
5
```

freelist page 存储结构，pageflags 为 0x10，表示 freelist 类型的页，ptr 指向空闲页 id 数组。注意在 boltdb 中支持通过多种数据结构（数组和 hashmap）来管理 free page。

#### Open 原理
了解完核心数据结构后，我们就很容易搞懂 boltdb Open API 的原理了。

首先它会打开 db 文件并对其增加文件锁，目的是防止其他进程也以读写模式打开它后，操作 meta 和 free page，导致 db 文件损坏。

其次 boltdb 通过 mmap 机制将 db 文件映射到内存中，并读取两个 meta page 到 db 对象实例中，然后校验 meta page 的 magic、version、checksum 是否有效，若两个 meta page 都无效，那么 db 文件就出现了严重损坏，导致异常退出。

#### Put 原理
那么成功获取 db 对象实例后，通过 bucket API 创建一个 bucket、发起一个 Put 请求更新数据时，boltdb 是如何工作的呢？

根据我们上面介绍的 bucket 的核心原理，它首先是根据 meta page 中记录 root bucket 的 root page，按照 B+ tree 的查找算法，从 root page 递归搜索到对应的叶子节点 page 面，返回 key 名称、leaf 类型。

如果 leaf 类型为 bucketLeafFlag，且 key 相等，那么说明已经创建过，不允许 bucket 重复创建，结束请求。否则往 B+ tree 中添加一个 flag 为 bucketLeafFlag 的 key，key 名称为 bucket name，value 为 bucket 的结构。

创建完 bucket 后，你就可以通过 bucket 的 Put API 发起一个 Put 请求更新数据。它的核心原理跟 bucket 类似，根据子 bucket 的 root page，从 root page 递归搜索此 key 到 leaf page，如果没有找到，则在返回的位置处插入新 key 和 value。

在核心数据结构介绍中，提到 boltdb 在内存中通过 node 数据结构来存储 page 磁盘页内容，它记录了 key-value 数据、page id、parent 及 children 的 node、B+ tree 是否需要进行重平衡和分裂操作等信息。

因此，当我们执行完一个 put 请求时，它只是将值更新到 boltdb 的内存 node 数据结构里，并未持久化到磁盘中。

#### 事务提交原理
那么 boltdb 何时将数据持久化到 db 文件中呢？

当你的代码执行 tx.Commit API 时，它才会将我们上面保存到 node 内存数据结构中的数据，持久化到 boltdb 中。接下来我就分别和你简要分析下各个核心步骤。

首先从上面 put 案例中我们可以看到，插入了一个新的元素在 B+ tree 的叶子节点，它可能已不满足 B+ tree 的特性，因此事务提交时，第一步首先要调整 B+ tree，进行重平衡、分裂操作，使其满足 B+ tree 树的特性。上面案例里插入一个 key r94 后，经过重平衡、分裂操作后的 B+ tree 如下图所示。

在重平衡、分裂过程中可能会申请、释放 free page，freelist 所管理的 free page 也发生了变化。因此事务提交的第二步，就是持久化 freelist。

注意，在 etcd v3.4.9 中，为了优化写性能等，freelist 持久化功能是关闭的。etcd 启动获取 boltdb db 对象的时候，boltdb 会遍历所有 page，构建空闲页列表。

事务提交的第三步就是将 client 更新操作产生的 dirty page 通过 fdatasync 系统调用，持久化存储到磁盘中。

最后，在执行写事务过程中，meta page 的 txid、freelist 等字段会发生变化，因此事务的最后一步就是持久化 meta page。

通过以上四大步骤，我们就完成了事务提交的工作，成功将数据持久化到了磁盘文件中，安全地完成了一个 put 操作。

#### 小结

最后来小结下今天的内容。首先通过一幅 boltdb 磁盘布局图和 bbolt 工具，为你解密了 db 文件的本质。db 文件由 meta page、freelist page、branch page、leaf page、free page 组成。随后结合 bbolt 工具，深入介绍了 meta page、branch page、leaf page、freelist page 的数据结构，帮助了解 key、value 数据是如何存储到文件中的。

然后通过分析一个 put 请求在 boltdb 中如何执行的。从 Open API 获取 db 对象说起，介绍了其通过 mmap 将 db 文件映射到内存，构建 meta page，校验 meta page 的有效性，再到创建 bucket，通过 bucket API 往 boltdb 添加 key-value 数据。

添加 bucket 和 key-value 操作本质，是从 B+ tree 管理的 page 中找到插入的页和位置，并将数据更新到 page 的内存 node 数据结构中。

真正持久化数据到磁盘是通过事务提交执行的。它首先需要通过一系列重平衡、分裂操作，确保 boltdb 维护的 B+ tree 满足相关特性，其次需要持久化 freelist page，并将用户更新操作产生的 dirty page 数据持久化到磁盘中，最后则是持久化 meta page。


---

## # Kubernetes基础应用：创建一个Pod背后etcd发生了什么？

#### Kubernetes 基础架构

**1. 控制面 Master 节点主要包含以下组件：**

- kube-apiserver，负责对外提供集群各类资源的增删改查及 Watch 接口，它是 Kubernetes 集群中各组件数据交互和通信的枢纽。kube-apiserver 在设计上可水平扩展，高可用 Kubernetes 集群中一般多副本部署。当收到一个创建 Pod 写请求时，它的基本流程是对请求进行认证、限速、授权、准入机制等检查后，写入到 etcd 即可。

- kube-scheduler 是调度器组件，负责集群 Pod 的调度。基本原理是通过监听 kube-apiserver 获取待调度的 Pod，然后基于一系列筛选和评优算法，为 Pod 分配最佳的 Node 节点。

- kube-controller-manager 包含一系列的控制器组件，比如 Deployment、StatefulSet 等控制器。控制器的核心思想是监听、比较资源实际状态与期望状态是否一致，若不一致则进行协调工作使其最终一致。

- etcd 组件，Kubernetes 的元数据存储。

**2. Node 节点主要包含以下组件：**

- kubelet，部署在每个节点上的 Agent 的组件，负责 Pod 的创建运行。基本原理是通过监听 APIServer 获取分配到其节点上的 Pod，然后根据 Pod 的规格详情，调用运行时组件创建 pause 和业务容器等。

- kube-proxy，部署在每个节点上的网络代理组件。基本原理是通过监听 APIServer 获取 Service、Endpoint 等资源，基于 Iptables、IPVS 等技术实现数据包转发等功能。

kube-apiserver 是唯一直接与 etcd 打交道的组件，各组件都通过 kube-apiserver 实现数据交互，它们极度依赖 kube-apiserver 提供的资源变化监听机制。而 kube-apiserver 对外提供的监听机制，也正是由 etcd Watch 特性提供的底层支持。


那么在 kubectl create 命令发出后，kube-apiserver 是如何与 etcd 打交道的呢？ 它是通过什么接口安全写入资源到 etcd 的？

同时，使用 kubectl 带标签查询 Pod 背后，kube-apiserver 是直接从缓存读取还是向 etcd 发出一个线性读或串行读请求呢？ 若同 namespace 下存在大量的 Pod，此操作性能又是怎样的呢?

接下来就了解下 kube-apiserver 收到创建和查询请求后，是如何与 etcd 交互的。

#### kube-apiserver 请求执行链路
kube-apiserver 作为 Kubernetes 集群交互的枢纽、对外提供 API 供用户访问的组件，因此保障集群安全、保障本身及后端 etcd 的稳定性的等重任也是非它莫属。比如校验创建请求发起者是否合法、是否有权限操作相关资源、是否出现 Bug 产生大量写和读请求等。

当收到一个请求后，它主要经过以下处理链路来完成以上若干职责后，才能与 etcd 交互。

- 认证模块，校验发起的请求的用户身份是否合法。支持多种方式，比如 x509 客户端证书认证、静态 token 认证、webhook 认证等。

- 限速模块，对请求进行简单的限速，默认读 400/s 写 200/s，不支持根据请求类型进行分类、按优先级限速，存在较多问题。Kubernetes 1.19 后已新增 Priority and Fairness 特性取代它，它支持将请求重要程度分类进行限速，支持多租户，可有效保障 Leader 选举之类的高优先级请求得到及时响应，能防止一个异常 client 导致整个集群被限速。

- 审计模块，可记录用户对资源的详细操作行为。

- 授权模块，检查用户是否有权限对其访问的资源进行相关操作。支持多种方式，RBAC(Role-based access control)、ABAC(Attribute-based access control)、Webhhook 等。Kubernetes 1.12 版本后，默认授权机制使用的 RBAC。

- 准入控制模块，提供在访问资源前拦截请求的静态和动态扩展能力，比如要求镜像的拉取策略始终为 AlwaysPullImages。

经过上面一系列的模块检查后，这时 kube-apiserver 就开始与 etcd 打交道了。在了解 kube-apiserver 如何将我们创建的 Deployment 资源写入到 etcd 前，先介绍下 Kubernetes 的资源是如何组织、存储在 etcd 中。


#### Kubernetes 资源存储格式

我们知道 etcd 仅仅是个 key-value 存储，但是在 Kubernetes 中存在各种各样的资源，并提供了以下几种灵活的资源查询方式：
- 按具体资源名称查询，比如 PodName、kubectl get po/PodName。
- 按 namespace 查询，获取一个 namespace 下的所有 Pod，比如 kubectl get po -n kube-system。
- 按标签名，标签是极度灵活的一种方式，你可以为你的 Kubernetes 资源打上各种各样的标签，比如 kubectl get po -l app=nginx。

你知道以上这几种查询方式它们的性能优劣吗？假设你是 Kubernetes 开发者，你会如何设计存储格式来满足以上功能点？

首先是按具体资源名称查询。它本质就是个 key-value 查询，只需要写入 etcd 的 key 名称与资源 key 一致即可。

其次是按 namespace 查询。这种查询也并不难。因为我们知道 etcd 支持范围查询，若 key 名称前缀包含 namespace、资源类型，查询的时候指定 namespace 和资源类型的组合的最小开始区间、最大结束区间即可。

最后是标签名查询。这种查询方式非常灵活，业务可随时添加、删除标签，各种标签可相互组合。实现标签查询的办法主要有以下两种：
- 方案一，在 etcd 中存储标签数据，实现通过标签可快速定位（时间复杂度 O(1)）到具体资源名称。然而一个标签可能容易实现，但是在 Kubernetes 集群中，它支持按各个标签组合查询，各个标签组合后的数量相当庞大。在 etcd 中维护各种标签组合对应的资源列表，会显著增加 kube-apiserver 的实现复杂度，导致更频繁的 etcd 写入。
- 方案二，在 etcd 中不存储标签数据，而是由 kube-apiserver 通过范围遍历 etcd 获取原始数据，然后基于用户指定标签，来筛选符合条件的资源返回给 client。此方案优点是实现简单，但是大量标签查询可能会导致 etcd 大流量等异常情况发生。

那么 Kubernetes 集群选择的是哪种实现方式呢?

下面是一个 Kubernetes 集群中的 coredns 一系列资源在 etcd 中的存储格式：
```txt
/registry/clusterrolebindings/system:coredns
/registry/clusterroles/system:coredns
/registry/configmaps/kube-system/coredns
/registry/deployments/kube-system/coredns
/registry/events/kube-system/coredns-7fcc6d65dc-6njlg.1662c287aabf742b
/registry/events/kube-system/coredns-7fcc6d65dc-6njlg.1662c288232143ae
/registry/pods/kube-system/coredns-7fcc6d65dc-jvj26
/registry/pods/kube-system/coredns-7fcc6d65dc-mgvtb
/registry/pods/kube-system/coredns-7fcc6d65dc-whzq9
/registry/replicasets/kube-system/coredns-7fcc6d65dc
/registry/secrets/kube-system/coredns-token-hpqbt
/registry/serviceaccounts/kube-system/coredns
```
从中可以看到，一方面 Kubernetes 资源在 etcd 中的存储格式由 prefix + "/" + 资源类型 + "/" + namespace + "/" + 具体资源名组成，基于 etcd 提供的范围查询能力，非常简单地支持了按具体资源名称查询和 namespace 查询。

kube-apiserver 提供了如下参数给你配置 etcd prefix，并支持将资源存储在多个 etcd 集群。
```txt
--etcd-prefix string     Default: "/registry"
The prefix to prepend to all resource paths in etcd.
--etcd-servers stringSlice
List of etcd servers to connect with (scheme://ip:port), comma separated.
--etcd-servers-overrides stringSlice
Per-resource etcd servers overrides, comma separated. The individual override format: group/resource#servers, where servers are URLs, 
semicolon separated.
```

另一方面，我们未看到任何标签相关的 key。Kubernetes 实现标签查询的方式显然是方案二，即由 kube-apiserver 通过范围遍历 etcd 获取原始数据，然后基于用户指定标签，来筛选符合条件的资源返回给 client（资源 key 的 value 中记录了资源 YAML 文件内容等，如标签）。

也就是当你执行"kubectl get po -l app=nginx"命令，按标签查询 Pod 时，它会向 etcd 发起一个范围遍历整个 default namespace 下的 Pod 操作。
```sh
$ kubectl get po -l app=nginx -v 8
I0301 23:45:25.597465   32411 loader.go:359] Config loaded from file /root/.kube/config
I0301 23:45:25.603182   32411 round_trippers.go:416] GET https://ip:port/api/v1/namespaces/default/pods?
labelSelector=app%3Dnginx&limit=500
```
etcd 收到的请求日志如下，由此可见当一个 namespace 存在大量 Pod 等资源时，若频繁通过 kubectl，使用标签查询 Pod 等资源，后端 etcd 将出现较大的压力。
```json
{
    "level":"debug",
    "ts":"2021-03-01T23:45:25.609+0800",
    "caller":"v3rpc/interceptor.go:181",
    "msg":"request stats",
    "start time":"2021-03-01T23:45:25.608+0800",
    "time spent":"1.414135ms",
    "remote":"127.0.0.1:44664",
    "response type":"/etcdserverpb.KV/Range",
    "request count":0,
    "request size":61,
    "response count":11,
    "response size":81478,
    "request content":"key:"/registry/pods/default/" range_end:"/registry/pods/default0" limit:500 "
}
```
了解完 Kubernetes 资源的存储格式后，再看看 nginx Deployment 资源是如何由 kube-apiserver 写入 etcd 的。

#### 通用存储模块
kube-apiserver 启动的时候，会将每个资源的 APIGroup、Version、Resource Handler 注册到路由上。当请求经过认证、限速、授权、准入控制模块检查后，请求就会被转发到对应的资源逻辑进行处理。

同时，kube-apiserver 实现了类似数据库 ORM 机制的通用资源存储机制，提供了对一个资源创建、更新、删除前后的 hook 能力，将其封装成策略接口。当你新增一个资源时，你只需要编写相应的创建、更新、删除等策略即可，不需要写任何 etcd 的 API。

创建一个资源主要由 BeforeCreate、Storage.Create 以及 AfterCreate 三大步骤组成。

当收到创建 nginx Deployment 请求后，通用存储模块首先会回调各个资源自定义实现的 BeforeCreate 策略，为资源写入 etcd 做一些初始化工作。

下面是 Deployment 资源的创建策略实现，它会进行将 deployment.Generation 设置为 1 等操作。
```go
// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (deploymentStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
   deployment := obj.(*apps.Deployment)
   deployment.Status = apps.DeploymentStatus{}
   deployment.Generation = 1

   pod.DropDisabledTemplateFields(&deployment.Spec.Template, nil)
}
```
执行完 BeforeCreate 策略后，它就会执行 Storage.Create 接口，也就是由它真正开始调用底层存储模块 etcd3，将 nginx Deployment 资源对象写入 etcd。

那么 Kubernetes 是使用 etcd Put 接口写入资源 key-value 的吗？如果是，那要如何防止同名资源并发创建被覆盖的问题？

#### 资源安全创建及更新

我们知道 etcd 提供了 Put 和 Txn 接口给业务添加 key-value 数据，但是 Put 接口在并发场景下若收到 key 相同的资源创建，就会导致被覆盖。

因此 Kubernetes 很显然无法直接通过 etcd Put 接口来写入数据。

而 etcd 事务接口 Txn，它正是为了多 key 原子更新、并发操作安全性等而诞生的，它提供了丰富的冲突检查机制。

Kubernetes 集群使用的正是事务 Txn 接口来防止并发创建、更新被覆盖等问题。当执行完 BeforeCreate 策略后，这时 kube-apiserver 就会调用 Storage 的模块的 Create 接口写入资源。1.6 版本后的 Kubernete 集群默认使用的存储是 etcd3，它的创建接口简要实现如下：
```go
// Create implements storage.Interface.Create.
func (s *store) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
   ......
   key = path.Join(s.pathPrefix, key)

   opts, err := s.ttlOpts(ctx, int64(ttl))
   if err != nil {
      return err
   }

   newData, err := s.transformer.TransformToStorage(data, authenticatedDataString(key))
   if err != nil {
      return storage.NewInternalError(err.Error())
   }

   startTime := time.Now()
   txnResp, err := s.client.KV.Txn(ctx).If(
      notFound(key),
   ).Then(
      clientv3.OpPut(key, string(newData), opts...),
   ).Commit
}
```
从上面的代码片段中，我们可以得出首先它会按照我们介绍的 Kubernetes 资源存储格式拼接 key。

然后若 TTL 非 0，它会根据 TTL 从 leaseManager 获取可复用的 Lease ID。Kubernetes 集群默认若不同 key（如 Kubernetes 的 Event 资源对象）的 TTL 差异在 1 分钟内，可复用同一个 Lease ID，避免大量 Lease 影响 etcd 性能和稳定性。

其次若开启了数据加密，在写入 etcd 前数据还将按加密算法进行转换工作。

最后就是使用 etcd 的 Txn 接口，向 etcd 发起一个创建 deployment 资源的 Txn 请求。

下面是 etcd 收到创建 nginx deployment 资源的请求日志：
```json
{
    "level":"debug",
    "ts":"2021-02-11T09:55:45.914+0800",
    "caller":"v3rpc/interceptor.go:181",
    "msg":"request stats",
    "start time":"2021-02-11T09:55:45.911+0800",
    "time spent":"2.697925ms",
    "remote":"127.0.0.1:44822",
    "response type":"/etcdserverpb.KV/Txn",
    "request count":1,
    "request size":479,
    "response count":0,
    "response size":44,
    "request content":"compare:<target:MOD key:"/registry/deployments/default/nginx-deployment" mod_revision:0 > success:<request_put:<key:"/registry/deployments/default/nginx-deployment" value_size:421 >> failure:<>"
}
```
从这个请求日志中，你可以得到以下信息：
- 请求的模块和接口，KV/Txn；
- key 路径，/registry/deployments/default/nginx-deployment，由 prefix + "/" + 资源类型 + "/" + namespace + "/" + 具体资源名组成；
- 安全的并发创建检查机制，mod_revision 为 0 时，也就是此 key 不存在时，才允许执行 put 更新操作。

通过 Txn 接口成功将数据写入到 etcd 后，kubectl create -f nginx.yml 命令就执行完毕，返回给 client 了。在以上介绍中可以看到，kube-apiserver 并没有任何逻辑去真正创建 Pod，但是为什么我们可以马上通过 kubectl get 命令查询到新建并成功运行的 Pod 呢？

这就涉及到了k8s中的控制器、调度器、Kubelet 等组件。下面就浅析它们是如何基于 etcd 提供的 Watch 机制工作，最终实现创建 Pod、调度 Pod、运行 Pod 的。

#### Watch 机制在 Kubernetes 中应用
kube-controller-manager 组件中包含一系列 WorkLoad 的控制器。Deployment 资源就由其中的 Deployment 控制器来负责的，那么它又是如何感知到新建 Deployment 资源，最终驱动 ReplicaSet 控制器创建出 Pod 的呢？

获取数据变化的方案，主要有轮询和推送两种方案组成。轮询会产生大量 expensive request，并且存在高延时。而 etcd Watch 机制提供的流式推送能力，赋予了 kube-apiserver 对外提供数据监听能力。

我们知道在 etcd 中版本号是个逻辑时钟，随着 client 对 etcd 的增、删、改操作而全局递增，它被广泛应用在 MVCC、事务、Watch 特性中。

尤其是在 Watch 特性中，版本号是数据增量同步的核心。当 client 因网络等异常出现连接闪断后，它就可以通过版本号从 etcd server 中快速获取异常后的事件，无需全量同步。

那么在 Kubernetes 集群中，它提供了什么概念来实现增量监听逻辑呢？

答案是 Resource Version。

#### Resource Version 与 etcd 版本号

Resource Version 是 Kubernetes API 中非常重要的一个概念，顾名思义，它是一个 Kubernetes 资源的内部版本字符串，client 可通过它来判断资源是否发生了变化。同时，你可以在 Get、List、Watch 接口中，通过指定 Resource Version 值来满足你对数据一致性、高性能等诉求。

那么 Resource Version 有哪些值呢？跟 etcd 版本号是什么关系？

下面分别以 Get 和 Watch 接口中的 Resource Version 参数值为例，剖析它与 etcd 的关系。

在 Get 请求查询案例中，ResourceVersion 主要有以下这三种取值：
- 第一种是未指定 ResourceVersion，默认空字符串。kube-apiserver 收到一个此类型的读请求后，它会向 etcd 发出共识读 / 线性读请求获取 etcd 集群最新的数据。
- 第二种是设置 ResourceVersion="0"，赋值字符串 0。kube-apiserver 收到此类请求时，它可能会返回任意资源版本号的数据，但是优先返回较新版本。一般情况下它直接从 kube-apiserver 缓存中获取数据返回给 client，有可能读到过期的数据，适用于对数据一致性要求不高的场景。
- 第三种是设置 ResourceVersion 为一个非 0 的字符串。kube-apiserver 收到此类请求时，它会保证 Cache 中的最新 ResourceVersion 大于等于你传入的 ResourceVersion，然后从 Cache 中查找你请求的资源对象 key，返回数据给 client。基本原理是 kube-apiserver 为各个核心资源（如 Pod）维护了一个 Cache，通过 etcd 的 Watch 机制来实时更新 Cache。当你的 Get 请求中携带了非 0 的 ResourceVersion，它会等待缓存中最新 ResourceVersion 大于等于你 Get 请求中的 ResoureVersion，若满足条件则从 Cache 中查询数据，返回给 client。若不满足条件，它最多等待 3 秒，若超过 3 秒，Cache 中的最新 ResourceVersion 还小于 Get 请求中的 ResourceVersion，就会返回 ResourceVersionTooLarge 错误给 client。

你要注意的是，若你使用的 Get 接口，那么 kube-apiserver 会取资源 key 的 ModRevision 字段填充 Kubernetes 资源的 ResourceVersion 字段（v1.meta/ObjectMeta.ResourceVersion）。若你使用的是 List 接口，kube-apiserver 会在查询时，使用 etcd 当前版本号填充 ListMeta.ResourceVersion 字段（v1.meta/ListMeta.ResourceVersion）。

**这里要提醒下你，在规模较大的集群中，尽量不要使用 kubectl 频繁查询资源。正如我们上面所分析的，它会直接查询 etcd 数据，可能会产生大量的 expensive request 请求，导致集群不稳定。**

watch ResourceVersion 它同样含有查询案例中的三种取值，官方定义的含义分别如下：

- 未指定 ResourceVersion，默认空字符串。一方面为了帮助 client 建立初始状态，它会将当前已存在的资源通过 Add 事件返回给 client。另一方面，它会从 etcd 当前版本号开始监听，后续新增写请求导致数据变化时可及时推送给 client。
- 设置 ResourceVersion="0"，赋值字符串 0。它同样会帮助 client 建立初始状态，但是它会从任意版本号开始监听（当前 kube-apiserver 的实现指定 ResourceVersion=0 和不指定行为一致，在获取初始状态后，都会从 cache 最新的 ResourceVersion 开始监听），这种场景可能会导致集群返回陈旧的数据。
- 设置 ResourceVersion 为一个非 0 的字符串。从精确的版本号开始监听数据，它只会返回大于等于精确版本号的变更事件。

Kubernetes 的控制器组件就基于以上的 Watch 特性，在快速感知到新建 Deployment 资源后，进入一致性协调逻辑，创建 ReplicaSet 控制器，整体交互流程如下所示。（watch Deployment 创建 ReplicaSet；watch ReplicaSet 创建 pod）

![p38](http://cdn.ipso.live/notes/etcd/etcd038.png)

Deployment 控制器创建 ReplicaSet 资源对象的日志如下所示。
```json
{
    "level":"debug",
    "ts":"2021-02-11T09:55:45.923+0800",
    "caller":"v3rpc/interceptor.go:181",
    "msg":"request stats",
    "start time":"2021-02-11T09:55:45.917+0800",
    "time spent":"5.922089ms",
    "remote":"127.0.0.1:44828",
    "response type":"/etcdserverpb.KV/Txn",
    "request count":1,
    "request size":766,
    "response count":0,
    "response size":44,
    "request content":"compare:<target:MOD key:"/registry/replicasets/default/nginx-deployment-756d9fd5f9" mod_revision:0 > success:<request_put:<key:"/registry/replicasets/default/nginx-deployment-756d9fd5f9" value_size:697 >> failure:<>"
}
```
真正创建 Pod 则是由 ReplicaSet 控制器负责，它同样基于 Watch 机制感知到新的 RS 资源创建后，发起请求创建 Pod，确保实际运行 Pod 数与期望一致。
```json
{
    "level":"debug",
    "ts":"2021-02-11T09:55:46.023+0800",
    "caller":"v3rpc/interceptor.go:181",
    "msg":"request stats",
    "start time":"2021-02-11T09:55:46.019+0800",
    "time spent":"3.519326ms",
    "remote":"127.0.0.1:44664",
    "response type":"/etcdserverpb.KV/Txn",
    "request count":1,
    "request size":822,
    "response count":0,
    "response size":44,
    "request content":"compare:<target:MOD key:"/registry/pods/default/nginx-deployment-756d9fd5f9-x6r6q" mod_revision:0 > success:<request_put:<key:"/registry/pods/default/nginx-deployment-756d9fd5f9-x6r6q" value_size:754 >> failure:<>"
}
```
在这过程中也产生了若干 Event，下面是 etcd 收到新增 Events 资源的请求，你可以看到 Event 事件 key 关联了 Lease，这个 Lease 正是由上面所介绍的 leaseManager 所负责创建。
```json
{
    "level":"debug",
    "ts":"2021-02-11T09:55:45.930+0800",
    "caller":"v3rpc/interceptor.go:181",
    "msg":"request stats",
    "start time":"2021-02-11T09:55:45.926+0800",
    "time spent":"3.259966ms",
    "remote":"127.0.0.1:44632",
    "response type":"/etcdserverpb.KV/Txn",
    "request count":1,
    "request size":449,
    "response count":0,
    "response size":44,
    "request content":"compare:<target:MOD key:"/registry/events/default/nginx-deployment.16628eb9f79e0ab0" mod_revision:0 > success:<request_put:<key:"/registry/events/default/nginx-deployment.16628eb9f79e0ab0" value_size:369 lease:5772338802590698925 >> failure:<>"
}
```
Pod 创建出来后，这时 kube-scheduler 监听到待调度的 Pod，于是为其分配 Node，通过 kube-apiserver 的 Bind 接口，将调度后的节点 IP 绑定到 Pod 资源上。kubelet 通过同样的 Watch 机制感知到新建的 Pod 后，发起 Pod 创建流程即可。

以上就是当我们在 Kubernetes 集群中创建一个 Pod 后，Kubernetes 和 etcd 之间交互的简要分析。


---

## # Kubernetes高级应用：如何优化业务场景使etcd能支撑上万节点集群？

你知道吗？ 虽然 Kubernetes 社区官网文档目前声称支持最大集群节点数为 5000，但是云厂商已经号称支持 15000 节点的 Kubernetes 集群了，那么为什么一个小小的 etcd 能支撑 15000 节点 Kubernetes 集群呢？

了支撑 15000 节点，Kubernetes 和 etcd 做的一系列优化。重点分析 Kubernetes 针对 etcd 的瓶颈是如何从应用层采取一系列优化措施，去解决大规模集群场景中各个痛点。

当遇到 etcd 性能瓶颈时，希望这里介绍的大规模 Kubernetes 集群的最佳实践经验和优化技术，能让你获得启发，帮助解决类似问题。

#### 大集群核心问题分析

在大规模 Kubernetes 集群中会遇到哪些问题呢？

大规模 Kubernetes 集群的外在表现是节点数成千上万，资源对象数量高达几十万。本质是更频繁地查询、写入更大的资源对象。

首先是查询相关问题。在大集群中最重要的就是如何最大程度地减少 expensive request。因为对几十万级别的对象数量来说，按标签、namespace 查询 Pod，获取所有 Node 等场景时，很容易造成 etcd 和 kube-apiserver OOM 和丢包，乃至雪崩等问题发生。

其次是写入相关问题。Kubernetes 为了维持上万节点的心跳，会产生大量写请求。而按照 etcd MVCC、boltdb、线性读等原理，etcd 适用场景是读多写少，大量写请求可能会导致 db size 持续增长、写性能达到瓶颈被限速、影响读性能。

最后是大资源对象相关问题。etcd 适合存储较小的 key-value 数据，etcd 本身也做了一系列硬限制，比如 key 的 value 大小默认不能超过 1.5MB。

现在重点分析下 Kubernetes 是如何优化以上问题，以实现支撑上万节点的。以及我会简单和你讲下 etcd 针对 Kubernetes 场景做了哪些优化。

#### 如何减少 expensive request
首先是第一个问题，Kubernetes 如何减少 expensive request？

在这个问题中，将 Kubernetes 解决此问题的方案拆分成几个核心点分析：

**1. 分页**

首先 List 资源操作是个基本功能点。各个组件在启动的时候，都不可避免会产生 List 操作，从 etcd 获取集群资源数据，构建初始状态。因此优化的第一步就是要避免一次性读取数十万的资源操作。

解决方案是 Kubernetes List 接口支持分页特性。分页特性依赖底层存储支持，早期的 etcd v2 并未支持分页被饱受诟病，非常容易出现 kube-apiserver 大流量、高负载等问题。在 etcd v3 中，实现了指定返回 Limit 数量的范围查询，因此也赋能 kube-apiserver 对外提供了分页能力。

如下所示，在 List 接口的 ListOption 结构体中，Limit 和 Continue 参数就是为了实现分页特性而增加的。

```go
// ListOptions is the query options to a standard REST list call.
type ListOptions struct {
   ...
   Limit int64 `json:"limit,omitempty" 
protobuf:"varint,7,opt,name=limit"`
   Continue string `json:"continue,omitempty" 
protobuf:"bytes,8,opt,name=continue"`
}
```
了解完 kube-apiserver 的分页特性后，接着往下看 Continue 字段具体含义，以及它是如何影响 etcd 查询结果的。

Limit 表示一次 List 请求最多查询的对象数量，一般为 500。如果实际对象数量大于 Limit，kube-apiserver 则会更新 ListMeta 的 Continue 字段，client 发起的下一个 List 请求带上这个字段就可获取下一批对象数量。直到 kube-apiserver 返回空的 Continue 值，就获取完成了整个对象结果集。

我们知道 etcd 分页是通过范围查询和 Limit 实现，ListOption 中的 Limit 对应 etcd 查询接口中的 Limit 参数。可以大胆猜测下，Continue 字段是不是跟查询的范围起始 key 相关呢？

Continue 字段的确包含查询范围的起始 key，它本质上是个结构体，还包含 APIVersion 和 ResourceVersion。之所以看到的是一个奇怪字符串，那是因为 kube-apiserver 使用 base64 库对其进行了 URL 编码，下面是它的原始结构体。
```go
type continueToken struct {
   APIVersion      string `json:"v"`
   ResourceVersion int64  `json:"rv"`
   StartKey        string `json:"start"`
}
```
当 kube-apiserver 收到带 Continue 的分页查询时，解析 Continue，获取 StartKey、ResourceVersion，etcd 查询 Range 接口指定 startKey，增加 clienv3.WithRange、clientv3.WithLimit、clientv3.WithRev 即可。

当通过分页多次查询 Kubernetes 资源对象，得到的最终结果集合与不带 Limit 查询结果是一致的吗？kube-apiserver 是如何保证分页查询的一致性呢？（自行查找答案）

#### 资源按 namespace 拆分
通过分页特性提供机制避免一次拉取大量资源对象后，接下来就是业务最佳实践上要避免同 namespace 存储大量资源，尽量将资源对象拆分到不同 namespace 下。

为什么拆分到不同 namespace 下有助于提升性能呢?

Kubernetes 资源对象存储在 etcd 中的 key 前缀包含 namespace，因此它相当于是个高效的索引字段。etcd treeIndex 模块从 B-tree 中匹配前缀时，可快速过滤出符合条件的 key-value 数据。

Kubernetes 社区承诺SLO达标的前提是，你在使用 Kubernetes 集群过程中必须合理配置集群和使用扩展特性，并遵循[一系列条件限制](https://github.com/kubernetes/community/blob/master/sig-scalability/configs-and-limits/thresholds.md)（比如同 namespace 下的 Service 数量不超过 5000 个）。

#### Informer 机制
各组件启动发起一轮 List 操作加载完初始状态数据后，就进入了控制器的一致性协调逻辑。在一致性协调逻辑中， Kubernetes 使用的是 Watch 特性来获取数据变化通知，而不是 List 定时轮询，这也是减少 List 操作一大核心策略。

Kubernetes 社区在 client-go 项目中提供了一个通用的 Informer 组件来负责 client 与 kube-apiserver 进行资源和事件同步，显著降低了开发者使用 Kubernetes API、开发高性能 Kubernetes 扩展组件的复杂度。

Informer 机制的 Reflector 封装了 Watch、List 操作，结合本地 Cache、Indexer，实现了控制器加载完初始状态数据后，接下来的其他操作都只需要从本地缓存读取，极大降低了 kube-apiserver 和 etcd 的压力。

**Informer 机制的基本工作流程如下：**
- client 启动或与 kube-apiserver 出现连接中断再次 Watch 时，报"too old resource version"等错误后，通过 Reflector 组件的 List 操作，从 kube-apiserver 获取初始状态数据，随后通过 Watch 机制实时监听数据变化。
- 收到事件后添加到 Delta FIFO 队列，由 Informer 组件进行处理。
- Informer 将 delta FIFO 队列中的事件转发给 Indexer 组件，Indexer 组件将事件持久化存储在本地的缓存中。
- 控制器开发者可通过 Informer 组件注册 Add、Update、Delete 事件的回调函数。Informer 组件收到事件后会回调业务函数，比如典型的控制器使用场景，一般是将各个事件添加到 WorkQueue 中，控制器的各个协调 goroutine 从队列取出消息，解析 key，通过 key 从 Informer 机制维护的本地 Cache 中读取数据。

通过以上流程分析，可以发现除了启动、连接中断等场景才会触发 List 操作，其他时候都是从本地 Cache 读取。

那连接中断等场景为什么触发 client List 操作呢？

#### Watch bookmark 机制
要搞懂这个问题，你得了解 kube-apiserver Watch 特性的原理。

我们知道 Kubernetes 通过全局递增的 Resource Version 来实现增量数据同步逻辑，尽量避免连接中断等异常场景下 client 发起全量 List 同步操作。

那么在什么场景下会触发全量 List 同步操作呢？这就取决于 client 请求的 Resource Version 以及 kube-apiserver 中是否还保存了相关的历史版本数据。

在 Watch 特性中，提到实现历史版本数据存储两大核心机制，滑动窗口和 MVCC。与 etcd v3 使用 MVCC 机制不一样的是，Kubernetes 采用的是滑动窗口机制。

kube-apiserver 的滑动窗口机制是如何实现的呢?

它通过为每个类型资源（Pod,Node 等）维护一个 cyclic buffer，来存储最近的一系列变更事件实现。

下面 Kubernetes 核心的 watchCache 结构体中的 cache 数组、startIndex、endIndex 就是用来实现 cyclic buffer 的。滑动窗口中的第一个元素就是 cache[startIndex%capacity]，最后一个元素则是 cache[endIndex%capacity]。
```go
// watchCache is a "sliding window" (with a limited capacity) of objects
// observed from a watch.
type watchCache struct {
   sync.RWMutex

   // Condition on which lists are waiting for the fresh enough
   // resource version.
   cond *sync.Cond

   // Maximum size of history window.
   capacity int

   // upper bound of capacity since event cache has a dynamic size.
   upperBoundCapacity int

   // lower bound of capacity since event cache has a dynamic size.
   lowerBoundCapacity int

   // cache is used a cyclic buffer - its first element (with the smallest
   // resourceVersion) is defined by startIndex, its last element is defined
   // by endIndex (if cache is full it will be startIndex + capacity).
   // Both startIndex and endIndex can be greater than buffer capacity -
   // you should always apply modulo capacity to get an index in cache array.
   cache      []*watchCacheEvent
   startIndex int
   endIndex   int

   // store will effectively support LIST operation from the "end of cache
   // history" i.e. from the moment just after the newest cached watched event.
   // It is necessary to effectively allow clients to start watching at now.
   // NOTE: We assume that <store> is thread-safe.
   store cache.Indexer

   // ResourceVersion up to which the watchCache is propagated.
   resourceVersion uint64
}
```

Kubernetes 的 Watch bookmark 机制核心思想。即使队列中无 client 关注的更新事件，Informer 机制的 Reflector 组件中 Resource Version 也需要更新。

Watch bookmark 机制通过新增一个 bookmark 类型的事件来实现的。kube-apiserver 会通过定时器将各类型资源最新的 Resource Version 推送给 kubelet 等 client，在 client 与 kube-apiserver 网络异常重连等场景，大大降低了 client 重建 Watch 的开销，减少了 relist expensive request。

#### 更高效的 Watch 恢复机制
虽然 Kubernetes 社区通过 Watch bookmark 机制缓解了 client 与 kube-apiserver 重连等场景下可能导致的 relist expensive request 操作，然而在 kube-apiserver 重启、滚动更新时，它依然还是有可能导致大量的 relist 操作，这是为什么呢？ 如何进一步减少 kube-apiserver 重启场景下的 List 操作呢？

在 kube-apiserver 重启后，kubelet 等 client 会立刻带上 Resource Version 发起重建 Watch 的请求。问题就在 kube-apiserver 重启后，watchCache 中的 cyclic buffer 是空的，此时 watchCache 中的最小 Resource Version(listResourceVersion) 是 etcd 的最新全局版本号。

在不少场景下，client 请求重建 Watch 的 Resource Version 是可能小于 listResourceVersion 的。

比如，集群内 Pod 稳定运行未发生变化，kubelet 假设收到了最新的事件。然而这个集群其他资源如 ConfigMap，被管理员不断的修改，它就会导致导致 etcd 版本号新增，ConfigMap 滑动窗口也会不断存储变更事件。

因此 kube-apiserver 重启后，client 请求重建 Pod Watch 的 Resource Version 假如是 RV100，而 Pod watchCache 中的滑动窗口最小 Resource Version 是 RV200。很显然，RV100 不在 Pod watchCache 所维护的滑动窗口中，kube-apiserver 就会返回"too old resource version"错误给 client，client 只能发起 relist expensive request 操作同步最新数据。

为了进一步降低 kube-apiserver 重启对 client Watch 中断的影响，Kubernetes 在 1.20 版本中又进一步实现了更高效的 Watch 恢复机制。它通过 etcd Watch 机制的 Notify 特性，实现了将 etcd 最新的版本号定时推送给 kube-apiserver。kube-apiserver 在将其转换成 ResourceVersion 后，再通过 bookmark 机制推送给 client，避免了 kube-apiserver 重启后 client 可能发起的 List 操作。

#### 如何控制 db size
分析完 Kubernetes 如何减少 expensive request，再看看 Kubernetes 是如何控制 db size 的。

首先，我们知道 Kubernetes 的 kubelet 组件会每隔 10 秒上报一次心跳给 kube-apiserver。

其次，Node 资源对象因为包含若干个镜像、数据卷等信息，导致 Node 资源对象会较大，一次心跳消息可能高达 15KB 以上。

最后，**etcd 是基于 COW(Copy-on-write) 机制实现的 MVCC 数据库，每次修改都会产生新的 key-value，若大量写入会导致 db size 持续增长。**

早期 Kubernetes 集群由于以上原因，当节点数成千上万时，kubelet 产生的大量写请求就较容易造成 db 大小达到配额，无法写入。

那么如何解决呢？

本质上还是 Node 资源对象大的问题。实际上我们需要更新的仅仅是 Node 资源对象的心跳状态，而在 etcd 中我们存储的是整个 Node 资源对象，并未将心跳状态拆分出来。

因此 Kuberentes 的解决方案就是将 Node 资源进行拆分，把心跳状态信息从 Node 对象中剥离出来，通过下面的 Lease 对象来描述它。
```go
// Lease defines a lease concept.
type Lease struct {
   metav1.TypeMeta `json:",inline"`
   metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
   Spec LeaseSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// LeaseSpec is a specification of a Lease.
type LeaseSpec struct {
   HolderIdentity *string `json:"holderIdentity,omitempty" protobuf:"bytes,1,opt,name=holderIdentity"`
   LeaseDurationSeconds *int32 `json:"leaseDurationSeconds,omitempty" protobuf:"varint,2,opt,name=leaseDurationSeconds"`
   AcquireTime *metav1.MicroTime `json:"acquireTime,omitempty" protobuf:"bytes,3,opt,name=acquireTime"`
   RenewTime *metav1.MicroTime `json:"renewTime,omitempty" protobuf:"bytes,4,opt,name=renewTime"`
   LeaseTransitions *int32 `json:"leaseTransitions,omitempty" protobuf:"varint,5,opt,name=leaseTransitions"`
}
```
因为 Lease 对象非常小，更新的代价远小于 Node 对象，所以这样显著降低了 kube-apiserver 的 CPU 开销、etcd db size，Kubernetes 1.14 版本后已经默认启用 Node 心跳切换到 Lease API。

#### 如何优化 key-value 大小
最后，我们再看看 Kubernetes 是如何解决 etcd key-value 大小限制的。

在成千上万个节点的集群中，一个服务可能背后有上万个 Pod。而服务对应的 Endpoints 资源含有大量的独立的 endpoints 信息，这会导致 Endpoints 资源大小达到 etcd 的 value 大小限制，etcd 拒绝更新。

另外，kube-proxy 等组件会实时监听 Endpoints 资源，一个 endpoint 变化就会产生较大的流量，导致 kube-apiserver 等组件流量超大、出现一系列性能瓶颈。

如何解决以上 Endpoints 资源过大的问题呢？

答案依然是拆分、化大为小。Kubernetes 社区设计了 EndpointSlice 概念，每个 EndpointSlice 最大支持保存 100 个 endpoints，成功解决了 key-value 过大、变更同步导致流量超大等一系列瓶颈。

#### etcd 优化

Kubernetes 社区在解决大集群的挑战的同时，etcd 社区也在不断优化、新增特性，提升 etcd 在 Kubernetes 场景下的稳定性和性能。简单列举两个，一个是 etcd 并发读特性，一个是 Watch 特性的 Notify 机制。

#### 并发读特性
通过以上介绍的各种机制、策略，虽然 Kubernetes 能大大缓解 expensive read request 问题，但是它并不是从本质上来解决问题的。

为什么 etcd 无法支持大量的 read expensive request 呢？

除了我们一直强调的容易导致 OOM、大流量导致丢包外，etcd 根本性瓶颈是在 etcd 3.4 版本之前，expensive read request 会长时间持有 MVCC 模块的 buffer 读锁 RLock。而写请求执行完后，需升级锁至 Lock，expensive request 导致写事务阻塞在升级锁过程中，最终导致写请求超时。

为了解决此问题，etcd 3.4 版本实现了并发读特性。核心解决方案是去掉了读写锁，每个读事务拥有一个 buffer。在收到读请求创建读事务对象时，全量拷贝写事务维护的 buffer 到读事务 buffer 中。

通过并发读特性，显著降低了 List Pod 和 CRD 等 expensive read request 对写性能的影响，延时不再突增、抖动。

#### 改善 Watch Notify 机制

为了配合 Kubernetes 社区实现更高效的 Watch 恢复机制，etcd 改善了 Watch Notify 机制，早期 Notify 消息发送间隔是固定的 10 分钟。

在 etcd 3.4.11 版本中，新增了 --experimental-watch-progress-notify-interval 参数使 Notify 间隔时间可配置，最小支持为 100ms，满足了 Kubernetes 业务场景的诉求。

最后，要注意的是，默认通过 clientv3 Watch API 创建的 watcher 是不会开启此特性的。你需要创建 Watcher 的时候，设置 clientv3.WithProgressNotify 选项，这样 etcd server 就会定时发送提醒消息给 client，消息中就会携带 etcd 当前最新的全局版本号。


---

## # 压缩：如何回收旧版本数据？
这小节是基础知识里的最后一节

我们知道 etcd 中的每一次更新、删除 key 操作，treeIndex 的 keyIndex 索引中都会追加一个版本号，在 boltdb 中会生成一个新版本 boltdb key 和 value。也就是随着你不停更新、删除，你的 etcd 进程内存占用和 db 文件就会越来越大。很显然，这会导致 etcd OOM 和 db 大小增长到最大 db 配额，最终不可写。

那么 etcd 是通过什么机制来回收历史版本数据，控制索引内存占用和 db 大小的呢？

这就是将要分享的 etcd 压缩机制。希望通过这小节，能帮助理解 etcd 压缩原理，在使用 etcd 过程中能根据自己的业务场景，选择适合的压缩策略，避免 db 大小增长失控而不可写入，帮助你构建稳定的 etcd 服务。

#### 整体架构

![p35](http://cdn.ipso.live/notes/etcd/etcd035.png)

在了解 etcd 压缩模块实现细节前，上面是一幅压缩模块的整体架构图。从图中可知，可以通过 client API 发起人工的压缩 (Compact) 操作，也可以配置自动压缩策略。在自动压缩策略中，你可以根据你的业务场景选择合适的压缩模式。目前 etcd 支持两种压缩模式，**分别是时间周期性压缩和版本号压缩**。

当你通过 API 发起一个 Compact 请求后，KV Server 收到 Compact 请求提交到 Raft 模块处理，在 Raft 模块中提交后，Apply 模块就会通过 MVCC 模块的 Compact 接口执行此压缩任务。

Compact 接口首先会更新当前 server 已压缩的版本号，并将耗时昂贵的压缩任务保存到 FIFO 队列中异步执行。压缩任务执行时，它首先会压缩 treeIndex 模块中的 keyIndex 索引，其次会遍历 boltdb 中的 key，删除已废弃的 key。

以上就是压缩模块的一个工作流程。接下来首先介绍如何人工发起一个 Compact 操作，然后详细介绍周期性压缩模式、版本号压缩模式的工作原理，最后再介绍 Compact 操作核心的原理。

#### 压缩特性初体验
在使用 etcd 过程中，当你遇到"etcdserver: mvcc: database space exceeded"错误时，若是你未开启压缩策略导致 db 大小达到配额，这时你可以使用 etcdctl compact 命令，主动触发压缩操作，回收历史版本。

如下所示，你可以先通过 endpoint status 命令获取 etcd 当前版本号，然后再通过 etcdctl compact 命令发起压缩操作即可。
```sh
# 获取etcd当前版本号
$ rev=$(etcdctl endpoint status --write-out="json" | egrep -o '"revision":[0-9]*' | egrep -o '[0-9].*')
$ echo $rev
9
# 执行压缩操作，指定压缩的版本号为当前版本号
$ etcdctl compact $rev
Compacted revision 9
# 压缩一个已经压缩的版本号
$ etcdctl compact $rev
Error: etcdserver: mvcc: required revision has been compacted
# 压缩一个比当前最大版号大的版本号
$ etcdctl compact 12
Error: etcdserver: mvcc: required revision is a future revision
```
请注意，如果你压缩命令传递的版本号小于等于当前 etcd server 记录的压缩版本号，etcd server 会返回已压缩错误 ("mvcc: required revision has been compacted") 给 client。如果版本号大于当前 etcd server 最新的版本号，etcd server 则返回一个未来的版本号错误给 client("mvcc: required revision is a future revision")。

执行压缩命令的时候，不少初学者有一个常见的误区，就是担心压缩会不会把我最新版本数据给删除？

压缩的本质是回收历史版本，目标对象仅是历史版本，不包括一个 key-value 数据的最新版本，因此你可以放心执行压缩命令，不会删除你的最新版本数据。不过在介绍 Watch 机制时提到，Watch 特性中的历史版本数据同步，依赖于 MVCC 中是否还保存了相关数据，因此建议你不要每次简单粗暴地回收所有历史版本。

在生产环境中，建议精细化的控制历史版本数，那如何实现精细化控制呢？

主要有两种方案，一种是使用 etcd server 的自带的自动压缩机制，根据你的业务场景，配置合适的压缩策略即可。

另外一种方案是如果你觉得 etcd server 的自带压缩机制无法满足你的诉求，想更精细化的控制 etcd 保留的历史版本记录，你就可以基于 etcd 的 Compact API，在业务逻辑代码中、或定时任务中主动触发压缩操作。你需要确保发起 Compact 操作的程序高可用，压缩的频率、保留的历史版本在合理范围内，并最终能使 etcd 的 db 大小保持平稳，否则会导致 db 大小不断增长，直至 db 配额满，无法写入。

在一般情况下，建议使用 etcd 自带的压缩机制。它支持两种模式，分别是按时间周期性压缩和保留版本号的压缩，配置相应策略后，etcd 节点会自动化的发起 Compact 操作。

接下来就详细介绍下 etcd 的周期性和保留版本号压缩模式。

#### 周期性压缩
首先是周期性压缩模式，它适用于什么场景呢？

当你希望 etcd 只保留最近一段时间写入的历史版本时，你就可以选择配置 etcd 的压缩模式为 periodic，保留时间为你自定义的 1h 等。

如何给 etcd server 配置压缩模式和保留时间呢?

如下所示，etcd server 提供了配置压缩模式和保留时间的参数：
```code
--auto-compaction-retention '0'
Auto compaction retention length. 0 means disable auto Compaction.
--auto-compaction-mode 'periodic'
Interpret 'auto-Compaction-retention' one of: periodic|revision.
```

auto-compaction-mode 为 periodic 时，它表示启用时间周期性压缩，auto-compaction-retention 为保留的时间的周期，比如 1h。

auto-compaction-mode 为 revision 时，它表示启用版本号压缩模式，auto-compaction-retention 为保留的历史版本号数，比如 10000。

注意，etcd server 的 auto-compaction-retention 为'0'时，将关闭自动压缩策略，

那么周期性压缩模式的原理是怎样的呢？ etcd 是如何知道你配置的 1h 前的 etcd server 版本号呢？

其实非常简单，etcd server 启动后，根据你的配置的模式 periodic，会创建 periodic Compactor，它会异步的获取、记录过去一段时间的版本号。periodic Compactor 组件获取你设置的压缩间隔参数 1h， 并将其划分成 10 个区间，也就是每个区间 6 分钟。每隔 6 分钟，它会通过 etcd MVCC 模块的接口获取当前的 server 版本号，追加到 rev 数组中。

因为你只需要保留过去 1 个小时的历史版本，periodic Compactor 组件会通过当前时间减去上一次成功执行 Compact 操作的时间，如果间隔大于一个小时，它会取出 rev 数组的首元素，通过 etcd server 的 Compact 接口，发起压缩操作。

需要注意的一点是，在 etcd v3.3.3 版本之前，不同的 etcd 版本对周期性压缩的行为是有一定差异的，具体的区别可以参考官方文档。

#### 版本号压缩
了解完周期性压缩模式，我们再看看版本号压缩模式，它又适用于什么场景呢？

当你写请求比较多，可能产生比较多的历史版本导致 db 增长时，或者不确定配置 periodic 周期为多少才是最佳的时候，你可以通过设置压缩模式为 revision，指定保留的历史版本号数。比如你希望 etcd 尽量只保存 1 万个历史版本，那么你可以指定 compaction-mode 为 revision，auto-compaction-retention 为 10000。

它的实现原理又是怎样的呢?

也很简单，etcd 启动后会根据你的压缩模式 revision，创建 revision Compactor。revision Compactor 会根据你设置的保留版本号数，每隔 5 分钟定时获取当前 server 的最大版本号，减去你想保留的历史版本数，然后通过 etcd server 的 Compact 接口发起如下的压缩操作即可。
```sh
# 获取当前版本号，减去保留的版本号数
rev := rc.rg.Rev() - rc.retention
# 调用server的Compact接口压缩
_，err := rc.c.Compact(rc.ctx，&pb.CompactionRequest{Revision: rev})
```

#### 压缩原理
介绍完两种自动化的压缩模式原理后，接下来我们就深入分析下压缩的本质。当 etcd server 收到 Compact 请求后，它是如何执行的呢？ 核心原理是什么？

如前面的整体架构图所述，Compact 请求经过 Raft 日志同步给多数节点后，etcd 会从 Raft 日志取出 Compact 请求，应用此请求到状态机执行。

执行流程如下图所示，MVCC 模块的 Compact 接口首先会检查 Compact 请求的版本号 rev 是否已被压缩过，若是则返回 ErrCompacted 错误给 client。其次会检查 rev 是否大于当前 etcd server 的最大版本号，若是则返回 ErrFutureRev 给 client，这就是我们上面执行 etcdctl compact 命令所看到的那两个错误原理。

通过检查后，Compact 接口会通过 boltdb 的 API 在 meta bucket 中更新当前已调度的压缩版本号 (scheduledCompactedRev) 号，然后将压缩任务追加到 FIFO Scheduled 中，异步调度执行。

![p36](http://cdn.ipso.live/notes/etcd/etcd036.png)

为什么 Compact 接口需要持久化存储当前已调度的压缩版本号到 boltdb 中呢？

试想下如果不保存这个版本号，etcd 在异步执行的 Compact 任务过程中 crash 了，那么异常节点重启后，各个节点数据就会不一致。

因此 etcd 通过持久化存储 scheduledCompactedRev，节点 crash 重启后，会重新向 FIFO Scheduled 中添加压缩任务，已保证各个节点间的数据一致性。

异步的执行压缩任务会做哪些工作呢？

首先我们回顾下 treeIndex 索引模块，它是 etcd 支持保存历史版本的核心模块，每个 key 在 treeIndex 模块中都有一个 keyIndex 数据结构，记录其历史版本号信息。

![p37](http://cdn.ipso.live/notes/etcd/etcd037.png)

如上图所示，因此异步压缩任务的第一项工作，就是压缩 treeIndex 模块中的各 key 的历史版本、已删除的版本。为了避免压缩工作影响读写性能，首先会克隆一个 B-tree（treeIndex使用B-tree 保存用户 key 与版本号之间的映射，不存储value，占用的空间很小。另外，etcd单库也有大小的限制，使得treeIndex不会太大。），然后通过克隆后的 B-tree 遍历每一个 keyIndex 对象，压缩历史版本号、清理已删除的版本。

假设当前压缩的版本号是 CompactedRev， 它会保留 keyIndex 中最大的版本号，移除小于等于 CompactedRev 的版本号，并通过一个 map 记录 treeIndex 中有效的版本号返回给 boltdb 模块使用。

为什么要保留最大版本号呢?

因为最大版本号是这个 key 的最新版本，移除了会导致 key 丢失。而 Compact 的目的是回收旧版本。当然如果 keyIndex 中的最大版本号被打了删除标记 (tombstone)， 就会从 treeIndex 中删除这个 keyIndex，否则会出现内存泄露。

Compact 任务执行完索引压缩后，它通过遍历 B-tree、keyIndex 中的所有 generation 获得当前内存索引模块中有效的版本号，这些信息将帮助 etcd 清理 boltdb 中的废弃历史版本。

![p37](http://cdn.ipso.live/notes/etcd/etcd037.png)

压缩任务的第二项工作就是删除 boltdb 中废弃的历史版本数据。如上图所示（没有黑线框的部分），它通过 etcd 一个名为 scheduleCompaction 任务来完成。

scheduleCompaction 任务会根据 key 区间，从 0 到 CompactedRev 遍历 boltdb 中的所有 key，通过 treeIndex 模块返回的有效索引信息，判断这个 key 是否有效，无效则调用 boltdb 的 delete 接口将 key-value 数据删除。

在这过程中，scheduleCompaction 任务还会更新当前 etcd 已经完成的压缩版本号 (finishedCompactRev)，将其保存到 boltdb 的 meta bucket 中。

scheduleCompaction 任务遍历、删除 key 的过程可能会对 boltdb 造成压力，为了不影响正常读写请求，它在执行过程中会通过参数控制每次遍历、删除的 key 数（默认为 100，每批间隔 10ms），分批完成 boltdb key 的删除操作。

#### 为什么压缩后 db 大小不减少呢?
当你执行完压缩任务后，db 大小减少了吗？ 事实是并没有减少。那为什么我们都通过 boltdb API 删除了 key，db 大小还不减少呢？

前面介绍 boltdb 实现时，提到过 boltdb 将 db 文件划分成若干个 page 页，page 页又有四种类型，分别是 meta page、branch page、leaf page 以及 freelist page。branch page 保存 B+ tree 的非叶子节点 key 数据，leaf page 保存 bucket 和 key-value 数据，freelist 会记录哪些页是空闲的。

当我们通过 boltdb 删除大量的 key，在事务提交后 B+ tree 经过分裂、平衡，会释放出若干 branch/leaf page 页面，然而 boltdb 并不会将其释放给磁盘，调整 db 大小操作是昂贵的，会对性能有较大的损害。

boltdb 是通过 freelist page 记录这些空闲页的分布位置，当收到新的写请求时，优先从空闲页数组中申请若干连续页使用，实现高性能的读写（而不是直接扩大 db 大小）。当连续空闲页申请无法得到满足的时候， boltdb 才会通过增大 db 大小来补充空闲页。

一般情况下，压缩操作释放的空闲页就能满足后续新增写请求的空闲页需求，db 大小会趋于整体稳定。

#### 小结

**etcd 压缩操作可通过 API 人工触发，也可以配置压缩模式由 etcd server 自动触发。压缩模式支持按周期和版本两种。在周期模式中你可以实现保留最近一段时间的历史版本数，在版本模式中你可以实现保留期望的历史版本数。**

压缩的核心工作原理分为两大任务，第一个任务是压缩 treeIndex 中的各 key 历史索引，清理已删除 key，并将有效的版本号保存到 map 数据结构中。

第二个任务是删除 boltdb 中的无效 key。基本原理是根据版本号遍历 boltdb 已压缩区间范围的 key，通过 treeIndex 返回的有效索引 map 数据结构判断 key 是否有效，无效则通过 boltdb API 删除它。

最后在执行压缩的操作中，虽然我们删除了 boltdb db 的 key-value 数据，但是 db 大小并不会减少。db 大小不变的原因是存放 key-value 数据的 branch 和 leaf 页，它们释放后变成了空闲页，并不会将空间释放给磁盘。

boltdb 通过 freelist page 来管理一系列空闲页，后续新增的写请求优先从 freelist 中申请空闲页使用，以提高性能。在写请求速率稳定、新增 key-value 较少的情况下，压缩操作释放的空闲页就可以基本满足后续写请求对空闲页的需求，db 大小就会处于一个基本稳定、健康的状态。


