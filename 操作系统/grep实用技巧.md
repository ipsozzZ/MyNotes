# 一、grep 是什么

`grep` = **文本搜索工具**

👉 作用：

* 在文件中查找字符串
* 支持正则表达式
* 支持递归、过滤、统计等

---

# 二、基础用法

```bash
grep "关键词" 文件名
```

例：

```bash
grep "error" app.log
```

👉 输出包含 `error` 的行

---

# 三、常用参数（核心必会）

## 1. 忽略大小写 `-i`

```bash
grep -i "error" app.log
```

👉 匹配 ERROR / Error / error

---

## 2. 显示行号 `-n`

```bash
grep -n "error" app.log
```

👉 输出：

```
12:error occurred
```

---

## 3. 只显示文件名 `-l`

```bash
grep -l "error" *.log
```

👉 哪些文件包含关键词

---

## 4. 反向匹配 `-v`

```bash
grep -v "info" app.log
```

👉 排除包含 info 的行

---

## 5. 统计数量 `-c`

```bash
grep -c "error" app.log
```

👉 返回匹配行数

---

## 6. 递归搜索 `-r`

```bash
grep -r "error" /var/log/
```

👉 搜索目录下所有文件

---

## 7. 精确匹配单词 `-w`

```bash
grep -w "cat" file.txt
```

👉 不匹配 `category`

---

## 8. 只输出匹配内容 `-o`

```bash
grep -o "error" file.txt
```

👉 每个匹配单独一行

---

# 四、正则表达式（重点🔥）

## 1. 基础正则

```bash
grep "a.*b" file
```

👉 匹配：

```
a...b
```

---

## 2. 扩展正则 `-E`

```bash
grep -E "error|warn" log.txt
```

👉 等价：

```
error 或 warn
```

---

## 3. 匹配开头 / 结尾

```bash
grep "^ERROR" log.txt   # 开头
grep "ERROR$" log.txt   # 结尾
```

---

## 4. 匹配数字

```bash
grep "[0-9]" file.txt
```

---

## 5. 匹配 IP（常见面试）

```bash
grep -E "([0-9]{1,3}\.){3}[0-9]{1,3}" log.txt
```

---

# 五、上下文查看（超实用🔥）

## 1. 上下文行

```bash
grep -C 3 "error" log.txt
```

👉 显示前后 3 行

---

## 2. 只看前几行

```bash
grep -A 5 "error" log.txt
```

👉 After（后5行）

---

## 3. 只看前几行

```bash
grep -B 5 "error" log.txt
```

👉 Before（前5行）

---

# 六、管道组合（高手必会）

## 1. 配合 `ps` 查进程

```bash
ps aux | grep nginx
```

👉 常见问题：
会匹配到 grep 本身

✅ 解决：

```bash
ps aux | grep nginx | grep -v grep
```

---

## 2. 查端口

```bash
netstat -anp | grep 8080
```

---

## 3. 查日志错误

```bash
cat app.log | grep "ERROR"
```

👉 更好写法：

```bash
grep "ERROR" app.log
```

---

## 4. 多条件过滤

```bash
grep "error" log.txt | grep "timeout"
```

---

# 七、文件过滤技巧

## 1. 指定文件类型

```bash
grep "error" *.log
```

---

## 2. 排除文件

```bash
grep -r "error" . --exclude="*.txt"
```

---

## 3. 排除目录

```bash
grep -r "error" . --exclude-dir=node_modules
```

---

# 八、高级技巧（实战🔥）

## 1. 多关键词匹配

```bash
grep -E "error|fail|warn" log.txt
```

---

## 2. 从文件读取关键词

```bash
grep -f keywords.txt log.txt
```

---

## 3. 精确匹配整行

```bash
grep -x "hello world" file.txt
```

---

## 4. 静默模式（脚本常用）

```bash
grep -q "error" file.txt
```

👉 用于判断：

```bash
if grep -q "error" file.txt; then
  echo "found"
fi
```

---

## 5. 显示匹配文件 + 行

```bash
grep -rn "error" .
```

---

# 九、性能 & 注意事项

## 🚀 性能建议

* 大文件优先：

```bash
grep "xxx" file
```

不要：

```bash
cat file | grep xxx
```

---

## ⚠️ 注意点

### 1. 中文乱码

```bash
LANG=C grep ...
```

---

### 2. 二进制文件

```bash
grep -a "text" file
```

---

### 3. 正则转义

```bash
grep "\." file
```

---

# 十、常见问题

## ❓ grep 和 egrep 区别

* `egrep` = `grep -E`

---

## ❓ 如何查不包含某字符串的行

```bash
grep -v "xxx"
```

---

## ❓ 如何统计关键词出现次数（不是行数）

```bash
grep -o "error" file | wc -l
```

---

# 十一、实战场景总结

## 🔥 查日志错误

```bash
grep -i "error" app.log
```

---

## 🔥 定位异常上下文

```bash
grep -C 5 "Exception" app.log
```

---

## 🔥 查接口慢日志

```bash
grep "timeout" access.log
```

---

## 🔥 查代码关键字

```bash
grep -rn "TODO" .
```

---

# 十二、一句话总结

👉 `grep` 本质：

> **一个基于正则表达式的高效文本过滤器**


