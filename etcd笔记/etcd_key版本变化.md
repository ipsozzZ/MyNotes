# etcd 版本变化
在 etcd 中，键（key）的版本号由 全局 Revision 和 键自身版本信息 共同管理，具体机制如下：

---

### 1. 全局 Revision
- 定义：  
  Revision 是 etcd 集群中所有修改操作（如 `Put`、`Delete`、`Txn`）的 全局唯一、严格递增的版本号，由 Raft 日志索引直接决定。
- 生成规则：  
  - 每个成功提交的 Raft 日志条目对应一个全局 Revision。  
  - 每次修改操作（无论是否同一个 key）均递增 Revision。例如：  
    - `Put key1 value1` → Revision=1  
    - `Put key2 value2` → Revision=2  
    - `Put key1 value3` → Revision=3  

---

### 2. 键的版本信息
每个键的元数据中包含两个关键版本字段：  
| 字段           | 说明                                                                 |
|--------------------|--------------------------------------------------------------------------|
| `create_revision` | 该键 首次创建时的全局 Revision。若键被删除后重建，此值会更新为新的 Revision。 |
| `mod_revision`    | 该键 最后一次修改时的全局 Revision。每次更新操作均会更新此值。                 |

---

### 3. 创建 Key 时的版本号
- 首次创建：  
  执行 `Put key1 value1` 时：  
  - 生成新的全局 Revision（假设为 `N`）。  
  - 设置 `create_revision = N`，`mod_revision = N`。  

- 示例：  
  ```bash
  etcdctl put key1 value1
  # 输出：OK（假设 Revision=1001）

  etcdctl get key1 -w json
  # 输出：
  # {
  #   "kvs": [{
  #     "key": "key1",
  #     "create_revision": 1001,
  #     "mod_revision": 1001,
  #     "value": "value1"
  #   }]
  # }
  ```

---

### 4. 更新 Key 时的版本号变化
- 更新操作：  
  执行 `Put key1 value2` 时：  
  - 生成新的全局 Revision（假设为 `N+1`）。  
  - 更新 `mod_revision = N+1`，`create_revision` 保持不变。  

- 示例：  
  ```bash
  etcdctl put key1 value2
  # 输出：OK（假设 Revision=1002）

  etcdctl get key1 -w json
  # 输出：
  # {
  #   "kvs": [{
  #     "key": "key1",
  #     "create_revision": 1001,  # 仍为首次创建的 Revision
  #     "mod_revision": 1002,     # 更新为最新 Revision
  #     "value": "value2"
  #   }]
  # }
  ```

---

### 5. 删除 Key 后重建的版本号
- 删除操作：  
  执行 `Delete key1` 后：  
  - 该键的元数据被标记为删除，但仍保留历史版本（可通过指定 Revision 读取）。  
  - 再次执行 `Put key1 value3` 时：  
    - 视为 重新创建，生成新的全局 Revision（假设为 `N+2`）。  
    - 重置 `create_revision = N+2`，`mod_revision = N+2`。

- 示例：  
  ```bash
  etcdctl del key1
  etcdctl put key1 value3
  # 输出：OK（假设 Revision=1003）

  etcdctl get key1 -w json
  # 输出：
  # {
  #   "kvs": [{
  #     "key": "key1",
  #     "create_revision": 1003,  # 重置为新的 Revision
  #     "mod_revision": 1003,
  #     "value": "value3"
  #   }]
  # }
  ```

---

### 6. 版本号结构（Main Revision + Sub Revision）
etcd 的 Revision 是一个 64 位整数，分为两部分：  
- 高位 32 位：`main_revision`，对应 Raft 日志索引，全局递增。  
- 低位 32 位：`sub_revision`，用于同一事务内的多次操作（如一个事务中多次修改同一个 key）。  

- 示例：  
  - Revision `0x0000000300000001` 表示：  
    - `main_revision = 3`  
    - `sub_revision = 1`  

---

### 总结
| 操作               | 全局 Revision | create_revision | mod_revision     |
|------------------------|-------------------|---------------------|----------------------|
| 首次创建（`Put key1`） | 递增              | 等于当前 Revision   | 等于当前 Revision    |
| 更新（`Put key1`）     | 递增              | 不变                | 更新为当前 Revision  |
| 删除后重建（`Put key1`）| 递增              | 重置为当前 Revision | 等于当前 Revision    |

---

### 应用场景
1. Watch 机制：  
   - 客户端可以指定从某个 Revision 开始监听键的变化（如 `watch key1 --rev=1001`）。  
2. 事务操作：  
   - 使用 `mod_revision` 实现条件更新（如 `Compare-And-Swap`）。  
3. 数据回滚：  
   - 通过指定 Revision 读取历史版本数据。  

通过理解版本号机制，可以更好地设计基于 etcd 的分布式系统，确保数据一致性和正确性。