# Java运行管理

## # OS管理
- 进程级别的管理(黑盒)
- CPU/内存/IO等具体功能监控

1. linux平台管理：
- top命令
- vmstat命令：查看整体的cpu/内存/IO/Swap等信息，格式：vmstat [采样间隔时间] [采样几次]。可以用来查看系统在某段时间的性能变化
- iostat命令：用来统计系统详细的IO信息的，格式：iostat -d [采样间隔时间] [采样几次]
- pidstat命令，多功能诊断器，可以定位到线程

2. windows平台：
- 任务管理器
- perfmon工具

## # JVM管理
- 线程/程序级管理(白盒)
- 查看虚拟机运行时各项信息
- 跟踪程序的执行过程，查看程序运行时信息
- 限制程序对资源的使用
- 将内存导出为文件进行具体分析


1. JDK工具集：位于jdk bin目录下
- 编译运行工具：    javac/java
- 打包工具：jar
- 文档工具：javadoc
- 国际化工具：native2ascii
- 混合编程工具：javah
- 反编译工具: javap
- 程序运行管理工具：jps/jstat/jinfo/jstack/jstatd/jcmd

2. jps（这是第一个要用的工具，因为需要用它获取Java进程号）
- 查看当前系统的Java进程信息
- 显示main函数所在类的名字
- -m可以显示进程的参数
- -l显示程序的全路径
- -v显示传递给java的main函数参数

3. jstat
- 查看堆信息 ：jstat -gc [进程号]

4. jinfo
- 用来查看虚拟机的参数是什么：jinfo -flags [进程号
- 可以通过jinfo来修改某些参数，注意并不是所有参数

5. jstack
- 查看线程堆栈信息
- jstack -l [进程号]
- 查看线程拥有的锁，分析线程死锁的原因

6. jstatd
- 客户机工具可以查看远程的Java进程，但是jstatd需要在远程的java进程所在服务器启动
- 本质上是一个RMI服务器，默认驻守在1099端口
- 远程机启动 jstatd -J-Djava.security.policy=E:/...;
- 启动需要权限支持，需要配置一个安全策略文件
- 启动客户机jps，检查远程机器的Java进程如：jps www.ipso.live:1099

```java

// 示例jdk的位置根据自己服务器jdk位置变化，策略文件的内容
grant codebase "file:E://java/jdk1.8.0_45/lib/tools.jar"{
    permission java.security.AllPermission; // tools.jar拥有所有的安全权限。
};
```

7. jcmd
- 从JDK7+新增，综合性工具
- 查看Java进程，导出进程信息，执行GC等操作
- jcmd 直接查看进程
- jcmd [进程号] help 展示命令的参数列表

### 可视化Java运行管理工具

1. JConsole （综合性管理可视化工具）
- 可以监管本地Java进程
- 监管远程Java进程（需要远程进程开启JMX服务）

2. Visual VM（综合性更强的可视化管理工具）
- 自JDK7发布，一个综合性的工具
- 可以查看、统计，也可以支持插件扩展
- 使用jvisualvm启动(位于java bin目录下)
- 可以强制将堆内存中的无用垃圾回收，这是JConsole无法完成的
- 可以将堆dump为一个文件，从而更好的分析堆内存

3. Mission Control

