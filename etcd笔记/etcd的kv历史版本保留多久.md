## etcd 中 Key 的历史版本保留机制

---

### 1. 默认行为
在 etcd 中，一个 Key 的所有历史版本默认会被保留。每次对 Key 的更新（`put` 操作）都会生成一个新的版本（`Revision`），旧版本不会被自动删除。

---

### 2. 历史版本的存储方式
etcd 使用 多版本并发控制（MVCC） 机制存储 Key 的历史版本。每个 Key 的修改都会记录以下信息：
- Revision：全局递增的版本号，表示修改的序号。
- CreateRevision：Key 创建时的 Revision。
- ModRevision：Key 最后一次修改的 Revision。
- Value：Key 的值。

---

### 3. 历史版本的保留策略
etcd 提供了两种机制来控制历史版本的保留：
#### 3.1 手动压缩（Compact）
通过 `etcdctl compact` 命令手动删除指定 Revision 之前的所有历史版本：
```bash
# 压缩到指定 Revision
etcdctl compact 1000

# 压缩到当前 Revision
CURRENT_REV=$(etcdctl endpoint status --write-out=json | jq -r '.header.revision')
etcdctl compact $CURRENT_REV
```
- 作用：删除指定 Revision 之前的所有历史版本，释放存储空间。
- 注意事项：
  - 压缩后，被删除的历史版本无法再访问。
  - 压缩操作不可逆。

#### 3.2 自动压缩（Auto-compaction）
在启动 etcd 时，可以通过 `--auto-compaction` 参数启用自动压缩：
```bash
# 每 1 小时压缩一次，保留最近 1000 个 Revision
etcd --auto-compaction-retention=1h --auto-compaction-mode=revision
```
- 模式：
  - `revision`：基于 Revision 压缩。
  - `periodic`：基于时间间隔压缩。
- 参数：
  - `--auto-compaction-retention`：保留时间或 Revision 数量。

---

### 4. 访问历史版本
即使启用了压缩，仍然可以通过指定 Revision 访问未被压缩的历史版本：
```bash
# 获取指定 Revision 的 Key 值
etcdctl get --rev=500 /config/app1/key1
```

---

### 5. 注意事项
- 性能影响：保留大量历史版本会增加存储开销和查询延迟。
- 压缩策略：根据业务需求合理设置压缩策略，避免过早删除有用数据。
- 监控：定期监控 etcd 的存储使用情况，及时调整压缩参数。

---

### 6. 总结
- etcd 默认会保留 Key 的所有历史版本。
- 通过 手动压缩 或 自动压缩 可以控制历史版本的保留策略。
- 合理使用压缩功能可以优化存储性能和资源利用率。

---

参考文档：
- [etcd 官方文档 - 压缩](https://etcd.io/docs/v3.5/op-guide/maintenance/#history-compaction)
- [etcd MVCC 设计](https://etcd.io/docs/v3.5/learning/api/#mvcc)