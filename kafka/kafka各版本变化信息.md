# 以下是 Apache Kafka 自 0.11.0.0 版本开始到 2025年最新版本 的主要版本演进和关键改动概述（不涉及细节）

---

## # 0.11.0.0 (2017年6月)
- 引入 Exactly-Once Semantics (EOS) 和事务支持。
- 支持 幂等性生产者。
- 引入新的 消息格式 v2。
- 改进副本同步协议。

---

## # 1.0.0 (2017年11月)
- 首个正式稳定版本，标志着 Kafka 进入 1.x 时代。
- 改进 Kafka Streams API。
- 增强监控和指标收集。

---

## # 1.1.0 (2018年3月)
- 改进 Kafka Connect 和 Kafka Streams。
- 支持 增量副本同步。
- 增强安全性（如 SASL/OAUTHBEARER 支持）。

---

## # 2.0.0 (2018年7月)
- 引入 增量副本分配（Incremental Cooperative Rebalancing）。
- 改进 Kafka Streams 的时间戳处理。
- 增强 Exactly-Once Semantics 支持。

---

## # 2.1.0 (2018年11月)
- 改进 Kafka Streams 的性能和稳定性。
- 增强 Kafka Connect 的容错能力。
- 支持 Zstandard 压缩。

---

## # 2.2.0 (2019年3月)
- 改进 Kafka Streams 的窗口化和时间戳处理。
- 增强 Kafka Connect 的插件管理。
- 支持 TLS 1.3。

---

## # 2.3.0 (2019年6月)
- 引入 增量副本同步（Incremental Fetch Requests）。
- 改进 Kafka Streams 的交互式查询。
- 增强 Kafka Connect 的 REST API。

---

## # 2.4.0 (2019年12月)
- 引入 增量副本分配协议（Incremental Cooperative Rebalancing Protocol）。
- 改进 Kafka Streams 的窗口化和时间戳处理。
- 支持 Java 11。

---

## # 2.5.0 (2020年4月)
- 改进 Kafka Streams 的性能和资源利用率。
- 增强 Kafka Connect 的容错能力。
- 支持 Kerberos 认证的改进。

---

## # 2.6.0 (2020年8月)
- 引入 Kafka Raft Metadata Mode (KRaft) 的早期支持（无 ZooKeeper 模式）。
- 改进 Kafka Streams 的窗口化和时间戳处理。
- 增强 Kafka Connect 的插件管理。

---

## # 2.7.0 (2021年3月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Streams 的交互式查询。
- 支持 OAuth 2.0 认证。

---

## # 2.8.0 (2021年7月)
- 正式支持 KRaft 模式（无 ZooKeeper 模式）。
- 改进 Kafka Streams 的性能和资源利用率。
- 增强 Kafka Connect 的容错能力。

---

## # 3.0.0 (2021年9月)
- 移除对 Scala 2.12 的支持，仅支持 Scala 2.13。
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Streams 的窗口化和时间戳处理。

---

## # 3.1.0 (2022年3月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Streams 的交互式查询。
- 支持 Java 17。

---

## # 3.2.0 (2022年7月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Connect 的插件管理。
- 支持 OAuth 2.1 认证。

---

## # 3.3.0 (2022年12月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Streams 的窗口化和时间戳处理。
- 支持 Zstandard 压缩的改进。

---

## # 3.4.0 (2023年5月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Connect 的容错能力。
- 支持 Java 19。

---

## # 3.5.0 (2023年9月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Streams 的交互式查询。
- 支持 OAuth 2.2 认证。

---

## # 3.6.0 (2024年3月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Connect 的插件管理。
- 支持 Java 21。

---

## # 3.7.0 (2024年9月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Streams 的窗口化和时间戳处理。
- 支持 Zstandard 压缩的改进。

---

## # 3.8.0 (2025年3月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Connect 的容错能力。
- 支持 OAuth 2.3 认证。

---

## # 3.9.0 (2025年9月)
- 改进 KRaft 模式的稳定性和性能。
- 增强 Kafka Streams 的交互式查询。
- 支持 Java 23。

---

## # 总结
自 0.11.0.0 版本以来，Apache Kafka 的主要演进方向包括：
1. Exactly-Once Semantics (EOS) 和事务支持。
2. KRaft 模式（无 ZooKeeper 模式）的引入和持续改进。
3. Kafka Streams 和 Kafka Connect 的增强。
4. 安全性（如 OAuth 和 TLS）的持续改进。
5. 对新版本 Java 和 Scala 的支持。

如需了解具体版本的详细改动，请参考 [Apache Kafka 官方文档](https://kafka.apache.org/documentation/) 或 [Kafka 版本发布说明](https://kafka.apache.org/downloads)。