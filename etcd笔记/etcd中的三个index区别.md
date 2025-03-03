# committed index, applied index 和 consistent index 的区别

在 etcd 中，`committed index`、`applied index` 和 `consistent index` 是与 Raft 协议和存储状态密切相关的三个关键概念，它们的区别和关系如下：

---

### 1. Committed Index（提交索引）
- 定义：  
  表示 Raft 日志中已被 集群多数节点持久化 的最高日志条目索引（即 Raft 协议中的 `Commit Index`）。
- 作用：  
  - 确定哪些日志条目可以被安全地应用到状态机（即数据存储引擎）。  
  - Leader 节点根据多数派节点的确认更新 `committed index`。
- 持久化：  
  - 不直接持久化，仅在内存中维护。节点重启后通过 Raft 日志重新计算。
- 示例：  
  日志条目索引为 `1~100`，若多数节点确认到索引 `95`，则 `committed index = 95`。

---

### 2. Applied Index（已应用索引）
- 定义：  
  表示已经 应用到状态机（如 etcd 的键值存储引擎） 的最高日志条目索引。
- 作用：  
  - 跟踪状态机的处理进度，确保所有提交的日志最终被应用到数据存储。  
  - 客户端可见的数据更新必须等待日志应用到状态机后才能生效。
- 持久化：  
  - 通过 `consistent index` 间接持久化（见下文）。
- 示例：  
  若 `committed index = 95`，但状态机只处理到索引 `90`，则 `applied index = 90`。

---

### 3. Consistent Index（持久化一致索引）
- 定义：  
  存储在 etcd 后端存储（如 BoltDB） 中的一个元数据，记录已应用到状态机的最高日志索引。
- 作用：  
  - 在 etcd 重启时，用于恢复 `applied index`，确保数据一致性。  
  - 防止因节点崩溃导致已应用的日志被重复应用或丢失。
- 持久化：  
  - 直接持久化到磁盘，与键值数据一起存储。
- 示例：  
  - 节点崩溃前 `applied index = 90`，重启后从 `consistent index = 90` 恢复，继续应用后续日志。

---

### 三者的关系与协同
1. 写入流程：  
   - 日志条目被复制到多数节点 → `committed index` 更新。  
   - 应用层（状态机）处理日志 → `applied index` 更新。  
   - 每次更新 `applied index` 后，同步更新 `consistent index` 到存储引擎。

2. 重启恢复：  
   - 节点重启后，从存储引擎读取 `consistent index`，恢复 `applied index`。  
   - 重放 Raft 日志中 `applied index + 1` 到 `committed index` 之间的日志。

---

### 关键区别总结
| 索引类型       | 维护者          | 持久化         | 作用场景                     |
|--------------------|---------------------|--------------------|----------------------------------|
| `committed index`  | Raft 模块（内存）   | 不持久化           | 控制日志提交进度                 |
| `applied index`    | 应用层（内存）      | 通过 `consistent index` 持久化 | 跟踪状态机处理进度               |
| `consistent index` | 存储引擎（磁盘）    | 直接持久化         | 崩溃恢复后保证数据一致性         |

---

### 示例场景
#### 场景：节点崩溃后重启
1. 持久化数据：  
   - 存储引擎中记录的 `consistent index = 100`。  
2. 恢复过程：  
   - 从 `consistent index` 恢复 `applied index = 100`。  
   - 重放 Raft 日志中索引 `101` 到当前 `committed index` 的日志。  
3. 结果：  
   - 数据状态与崩溃前一致，避免重复应用日志。

---

### 监控与运维
- 查看索引值：  
  ```bash
  etcdctl endpoint status --write-out=json | jq '.[] | {committedIndex: .Status.raftAppliedIndex, appliedIndex: .Status.dbSize}'
  ```
- 健康检查：  
  - 若 `applied index` 长期落后于 `committed index`，可能表示状态机处理阻塞。  
  - `consistent index` 与 `applied index` 不一致时，可能发生数据损坏。

---

### 总结
- `committed index` 是 Raft 层的提交进度，决定日志的可见性。  
- `applied index` 是应用层的处理进度，决定数据的最终状态。  
- `consistent index` 是崩溃恢复的基石，保证应用进度的持久化。  

理解这三者的区别，是诊断 etcd 性能问题（如数据延迟）和保障集群一致性的关键。

