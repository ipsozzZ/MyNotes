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

















