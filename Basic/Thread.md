# Thread线程学习

## # 进程、线程基本概念

### 进程 ###

>进程是一段执行的程序，持有资源(共享内存、共享文件)和线程

### 线程 ###

>线程是系统中最小的执行单元；同一进程中有多个线程；线程共享进程的资源；线程交互(同步和互斥)

## # java线程

### java线程支持

* class Thread 和 interface Runnable（都属java.lang包，即默认包，所以使用时不必做引入申明）

### 常用方法

#### 线程创建方法

* Thread()
* Thread(String name)
* Thread(Runnable target)
* Thread(Runnable target, String name)

#### 常用方法

* void start() 启动线程
* static void sleep(long millis)、static void sleep(long millis, int nanos) 线程休眠
* void join()、void join(long millis)、void join(long millis, int nanos) 使其它线程等待当前线程终止
* static void yield() 当前运行线程释放处理器资源
* static Thread currentThread() 返回当前运行的线程引用
