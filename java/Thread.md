# JAVA多线程编程

JAVA多线程

java程序入口就是由JVM虚拟机启动main线程

* main线程又可以启动其它线程
* 当所有线程都运行结束时，JVM退出，进程结束

## # 线程基础

### 线程创建、启动

首先创建一个线程对象

```java

/**
 * 线程对象
 */
public class MyThread extends Thread{
  
  // 覆写run方法
  public void run(){
    System.out.println("Thread")
  }
}

```

主线程中通过```Thread thread = new MyThread()```方法创建，```thread.start()```启动，自动调用线程中的run方法，不可通过```thread.run()```的方式调用，这种方式只是调用了普通对象的run方法而已

### 线程状态、中断

一个线程对象只能调用一次start();线程执行代码是run()方法；线程调度由操作系统决定，程序本身无法决定。

**线程状态**：

1. New（新创建）
2. Runnable（运行中）
3. Blocked（被阻塞）
4. waiting（等待）
5. Timed Waiting（计时等待）
6. Terminated（已终止）

**线程中止的原因**：

1. run()方法执行结束（正常终止）
2. 因为未捕获的异常导致线程终止（意外终止）
3. 对某个线程的Thread实例调用stop()方法强制终止（不推荐使用）

**join()方法**：

1. 一个线程可以等待另一个线程直到其运行结束。即使用```thread.join();```方法会让当前线程执行结束才执行下一个线程
2. 可以对join()指定等待时间，超过等待时间线程仍然没有结束就不再等待
3. 对已经结束的线程调用join()方法会立刻返回

**中断线程**：

* 中断线程需要通过检测isInterrupted()标志获取当前线程是否已中断
* 其它线程可以通过调用interrupt()方法中断该线程
* 如果线程处于等待状态，该线程会捕获InterruptedException
* isInterrupted()为true或者捕获了InterruptedException都应该立刻结束
* 通过标志位判断需要正确使用volatile关键字
* volatile关键字解决了共享变量在线程间的可见性问题(就是两个线程对内存中的同一个数据进行操作时可能读写问题)，所以共享的变量一定要使用volatile关键字标记，确保读取到最新更新的值。
* 设置标志位在run()方法中判断标准位终止线程

**volatile关键字**：

volatile关键字解决了共享变量在线程间的可见性问题(就是两个线程对内存中的同一个数据进行操作时可能读写问题)，所以共享的变量一定要使用volatile关键字标记，确保读取到最新更新的值

java内存模型是：虚拟机中有一块主内存(包括了我们定义的各种变量)，当线程创建后调用一个主内存中的值时会在线程的工作内存中保存一个从主内存获取过来的变量副本，对工作内存中的变量进行操作，再修改后的某一时刻将操作后的变量返回给主内存中的原变量，此期间如果再有其它线程对该变量进行操作时就有可能产生变量不同步问题。所以线程共享变量一定要使用volatile关键字修饰。

* 每次访问变量时，总是获取主内存的最新值
* 每次修改变量值后立即回写到主内存

### 线程守护

java程序入口就是由JVM虚拟机启动main线程

* main线程又可以启动其它线程
* 当所有线程都运行结束时，JVM退出，进程结束
  
如果某个线程不结束，JVM进程就无法结束所以就有了守护线程

**守护线程**：

一般做定时任务会用

* 守护线程是为其它线程服务的线程
* 所有非守护线程都执行完毕后虚拟机才退出
* 守护线程不能持有资源
* 在创建线程后使用setDaemon(true)将线程定义为守护线程，假设现在有创建好的线程对象MyThread对象，在主线程中```Thread thread = new MyThread(); thread.setDaemon(true)```将该线程设置为守护线程。

如下例，不使用守护线程将会死循环

```java

import java.util.Date;

public class MyThread extends Thread {

   @Override
   public void run() {
      super.run();
      Date date = new Date();
      while (true){
         System.out.println(date.getTime());
         try {
            Thread.sleep(1000);
         } catch (InterruptedException e) {
            e.printStackTrace();
         }
      }
   }
}

```

```java

public class Test {

   public static void main(String[] args) throws Exception {
      System.out.println("main start");

      Thread thread = new MyThread();
      thread.setDaemon(true); // 设置为守护进程
      thread.start();
      Thread.sleep(5000);

      System.out.println("main end");
   }
}

```

## # 线程同步

### 线程同步

1. 多线程同时修改变量，会造成逻辑错误
2. 需要通过synchronized同步
3. 同步的本质就是给指定对象加锁
4. 注意加锁对象必须是同一个实例
5. 对JVM定义的单个原子操作不需要同步

**原子性**：

* 对共享变量进行写入时，必须保证原子操作
* 原子操作是指不能被中断的一个或一系列操作

为了保证一系列操作作为原子性，必须保证一系列操作执行中没有被其它线程执行这些操作。

java使用synchronized对一个对象(一系列操作)加锁，如:

```java

synchronized(lock){
  n = n+1;
  m = m-1;
  p = n-m;
}

```

假设现在有两个线程T1、T2，它们都需要执行上面的共享语句块，假设当前是T1获取到了共享语句块的琐，此时T1就可以执行语句块中的程序，而T2因为没有共享语句块的琐，所以只能处于等待状态，等待T1执行完共享语句块并且释放了共享语句块的琐后才能执行。这就是synchronized(lock)的作用之一。

synchronized的缺点：它在保证代码块只能给一个线程获取的同时降低了并发能力，并且在加锁和解锁时也会消耗一定的时间

**如何使用synchronized**：

找出修改共享变量的线程代码块
选择一个实例作为琐
使用synchronized(lockObject){ ... }

**JVM规范定义几种原子操作**：

对于原子操作是不需要同步的，其次对于局部变量也不需要同步

* 基本类型(long和double除外)赋值如：```int n = 100;```
* 应用类型赋值如：```List<String> list = anotherList;```
* 将简单的非原子操作变为原子操作：如对两个int类型赋值可以将其变成数组(引用类型)赋值，就可以实现简单的非引用类型赋值变引用类型赋值了。

**总结**：

多线程同时修改变量，会造成逻辑错误：

1. 需要通过synchronized同步
2. 同步的实质就是给指定对象加锁
3. 注意加锁对象必须是同一个实例

对JVM定义的单个原子操作、局部变量不需要同步

**synchronized使用实例**：

```java

/**
 * 线程AddThread对象
 * 对Test(main线程中的共享变量count进行加1操作)
 */
public class AddThread extends Thread {

   @Override
   public void run() {
      for (int i=0; i<Test.LOOP; i++){

        // Test.count += 1; 直接使用

        synchronized (Test.Lock){
          Test.count += 1; // 涉及到修改线程共享变量的程序块
        }
      }
   }
}


/**
 * 线程DecThread对象
 * 对Test(main线程中的共享变量count进行减1操作)
 */
public class DecThread extends Thread {

   @Override
   public void run() {
      for (int i=0; i<Test.LOOP; i++){
         synchronized(Test.Lock){
            Test.count -= 1; // 涉及到修改线程共享变量的程序块
         }
      }
   }
}


/**
 * 测试主线程(main线程)
 */
public class Test {

   final static int LOOP = 10000; // 线程共享变量
   public static int count = 0; // 线程共享变量
   public static final Object Lock = new Object(); // 可以使用Object类型的Lock变量作为琐

   public static void main(String[] args) throws Exception {
      System.out.println("main start");

      /* 测试synchronized加锁解锁 */
      Thread t1 = new AddThread();
      Thread t2 = new DecThread();
      t1.start();
      t2.start();
      t1.join();
      t2.join(); // 如果没有使用synchronized,就算使用join()也不能避免线程共享变量操作错误，因为线程结束并不代表线程工作内存中的变量已经同步到主内存中
      System.out.println(count);

      System.out.println("main end");
   }
}


// 其中的随机打印的4次结果：

我们想要的结果应该是：
main start
0
main end

Process finished with exit code 0

/* ** 没有使用synchronized加锁 ** */

// 第一次
main start
-1579
main end

Process finished with exit code 0

// 第二次
main start
1253
main end

Process finished with exit code 0

// 第三次
main start
3110
main end

Process finished with exit code 0

// 第四次
main start
2474
main end

Process finished with exit code 0



/* ** 使用synchronized加锁 ** */

// 第一次
main start
0
main end

Process finished with exit code 0

// 第二次
main start
0
main end

Process finished with exit code 0

// 第三次
main start
0
main end

Process finished with exit code 0

// 第四次
main start
0
main end

Process finished with exit code 0

```

### synchronized

synchronized方法
用synchronized修饰方法可以把整个方法变为同步代码块；

synchronized方法加锁对象是this；

通过合理的设计和数据封装可以让一个类变为“线程安全”；

一个类没有特殊说明，默认不是thread-safe；

多线程能否访问某个非线程安全的实例，需要具体问题具体分析。

上面我们的加锁对象也就是synchronized(LockObject)中的LockObject怎么琐？锁住哪个对象？

让线程自己决定锁住哪个对象不是一个好的方法，我们应该自己定义锁住对象：

```java

/**
 * 针对刚才的自增自减封装一个count类，把同步的逻辑都些到这个类中
 * 使用synchronized(this)能保证每个实例调用add、或者dec进行操作时都是线程安全的
 */
public class Counter {
   int count = 0;

   /*public void add(int n){
      synchronized (this){
         count += n;
      }
   }*/

   /**
    * 和上面注释掉的写法的区别是这种写法整个方法的操作都会加锁，而上面的只是synchronized锁起来的部分有效
    * @param n
    */
   public synchronized void add(int n){
      count += n;
   }

   /*public void dec(int n){
      synchronized (this){
         count -= n;
      }
   }*/

   public synchronized  void dec(int n){
      count -= n;
   }

   public int get(){
      return count;
   }
}

// 在线程对象AddThread中

/**
 * 线程AddThread对象
 * 对Test(main线程中的共享变量count进行加1操作)
 */
public class AddThread extends Thread {

   Counter counter;

   public AddThread(Counter counter){
      this.counter = counter;
   }

   @Override
   public void run() {
      for (int i=0; i<Test.LOOP; i++){
         /*// Test.count += 1;
         synchronized (Test.Lock){
            Test.count += 1;
         }*/

         counter.add(1); // 调用Counter中的线程安全方法执行加1
      }
   }
}


// 在DecThread中

/**
 * 线程DecThread对象
 * 对Test(main线程中的共享变量count进行减1操作)
 */
public class DecThread extends Thread {

   Counter counter;

   public DecThread(Counter counter){
      this.counter = counter;
   }

   @Override
   public void run() {
      for (int i=0; i<Test.LOOP; i++){
         /*// Test.count -= 1;
         synchronized(Test.Lock){
            Test.count -= 1;
         }*/

         counter.dec(1); // 调用Counter对象中的线程安全方法执行count减1
      }
   }
}

// 在主线程测试类中

/**
 * 测试主线程(main线程)
 */
public class Test {

   final static int LOOP = 10000;
   public static int count = 0;
   // public static final Object Lock = new Object(); // 可以使用Object类型的Lock变量作为琐

   public static void main(String[] args) throws Exception {
      System.out.println("main start");

      /* 使用synchronized定义线程安全方法来实现同步(自定义锁定对象) */
      Counter counter = new Counter();
      Thread t1 = new AddThread(counter);
      Thread t2 = new DecThread(counter);
      t1.start();
      t2.start();
      t1.join();
      t2.join();
      count = counter.get();
      System.out.println(count);

      System.out.println("main end");
   }
}


// 测试结果

main start
0
main end

Process finished with exit code 0

```

**注意**上面使用synchronized修饰方法时，如果方法是static修饰也就是静态方法时如：

```java

public class A{
  static count;
  static synchromized void add(int n){
    count += n;
  }
}

// 以上类就相当于：

public class A{
  static count;
  static void add(int n){
    synchromized(A.class){
      count += n;
    }
  }
}

```

**常见线程安全类**：

* 不变类：String、Integer、LocalDate等
* 没有成员变量的类：Math等
* 正确使用synchronized的类：StringBuffer等

**常见非线程安全类**：

* 不能在多线程中共享实例并修改：如ArrayList
* 可以在多线程中以只读的方式共享

### 死锁

死锁产生的条件：

多线程各自持有不同的锁，并互相试图获取对方已持有的锁，双方无限等待下去：导致死锁

如何避免死锁：

多线程获取锁的顺序要一致

**java的线程琐是可以重入的**：

什么是可重入？

```java

public void add(int m){
  synchronized(lock){
    this.value += m;
    addAnother(m);
  }
}

public void addAnother(int m){
  synchronized(lock){
    this.another += m;
  }
}

```

这就叫可重入

**java的线程可以获取多个不同对象的琐**：

```java

public void add(int m){
  synchronized(lockA){ // 获取lockA琐
    this.value += m;
    synchronized(lockB){ // 获取lockB琐
      this.another += m;
    } // 释放lockB琐
  } // 释放lockA琐
}

```

当有多个线程时,每个线程都需要各种各样的琐，当一个线程拥有一个或多个琐后可能还需要其它的琐来配合已有的琐来完成一个业务，所以就去请求所需要的琐，但是所需要的琐又被其它进程占用着，而且这个进程也在等待所缺的琐，如果两个线程所缺的琐都被对方占有时，就会导致死锁，双方都在无限等待对方释放琐。

**简单来说死锁的形成条件**：

1. 两个线程各自持有不同的琐
2. 两个线程各自试图获取对方已经拥有的琐
3. 双方无限等待下去->死锁

**死锁形成后**：

* 没有任何机制能解除死锁
* 只能强制结束进程(这里是JVM进程)

**死锁避免**：

* 多线程获取琐的顺序要完全一致

### wait和notify/notifyAll

wait()方法内部非常复杂，是由在JVM中的c代码实现的

上面synchronized为我们解决线程资源竞争问题，但没有解决多线程协调问题，所以需要wait和notify来协调线程执行

* wait / notify用于多线程协调运行：
* 在synchronized内部可以调用wait()使线程进入等待状态
* 必须在已获得的锁对象上调用wait()方法
* 在synchronized内部可以调用notify()/notifyAll()唤醒其他等待线程
* 必须在已获得的锁对象上调用notify()/notifyAll()方法

例如我们有一个队列对象类其中有一个方法持有不断的往这个队列中push数据并且该方法为线程安全(就是对这个队列对象加锁)，又有另一个方法不断的从这个队列中pop数据并且该方法也为线程安全，假设有业务是如果队列为空就导致死循环，在某一时刻不断pop的方法发现栈顶为空，进入死循环，并且一直锁着这个队列对象，导致push获取不到这个队列的琐也不能操作push了,这个时候在死循环内部如果有this.wait()，线程就会进入等待状态，并释放线程获得的琐，直到其它线程使用notify唤醒。代码实现

```java

import java.util.LinkedList;
import java.util.Queue;

/**
 * 定义一个线程安全类
 * 这个类在getTask中添加图片或者文件下载方式，就可以从数据库中获取文件地址并顺序将文件下载
 */
public class TaskQueue {

   Queue<String> queue = new LinkedList<>();

   /**
    * 当有线程调用该方法时会锁住queue对象实例
    * @param s 存入队列的数据
    */
   public synchronized void addTask(String s){
      this.queue.add(s);
      this.notifyAll(); // 向队列中添加了数据后需要将所有正在等待的线程唤醒并释放this琐，建议使用
      // this.notify(); // 向队列中添加了数据后需要将正在等待的线程唤醒
   }

   /**
    * 当有线程调用该方法是，线程会锁住queue对象实例
    * @return String
    * @throws InterruptedException Exception
    */
   public synchronized String getTask() throws InterruptedException{
      while (queue.isEmpty()){
         this.wait(); // 如果队列中没有元素，就将线程设为等待状态，这样会释放该线程占有的琐，供其它线程使用
      }
      return queue.remove();
   }
}

```

## # 高级concurrent包（java并发包）

我们知道java提供synchronized/wait/notify/notifyAll来解决多线程同步问题，但是我们实现多线程同步依然很困难，所以从JDK1.5开始java提供了java.util.cuncurrent包来实现更高级得同步功能，并且简化多线程程序得编写。

### ReentrantLock

ReentrantLock可以替代synchronized，ReentrantLock获取锁更安全，必须使用try … finally保证正确获取和释放锁，
tryLock()可指定超时

### ReadWriteLock

使用ReadWriteLock可以提高读取效率：

* ReadWriteLock只允许一个线程写入
* ReadWriteLock允许多个线程同时读取
* ReadWriteLock适合读多写少的场景

### Condition

* Condition可以替代wait / notify
* Condition对象必须从ReentrantLock对象获取
* ReentrantLock＋Condition可以替代synchronized + wait / notify

### Concurrent集合

使用java.util.concurrent提供的Blocking集合可以简化多线程编程：

* CopyOnWriteArrayList
* ConcurrentHashMap
* CopyOnWriteArraySet
* ArrayBlockingQueue
* LinkedBlockingQueue
* LinkedBlockingDeque
* 多线程同时访问Blocking集合是安全的

尽量使用JDK提供的concurrent集合，避免自己编写同步代码

### Atomic

* 使用java.util.atomic提供的原子操作可以简化多线程编程：
* AtomicInteger／AtomicLong／AtomicIntegerArray等
* 原子操作实现了无锁的线程安全
* 适用于计数器，累加器等

### ExecutorService

JDK提供了ExecutorService实现了线程池功能，线程池内部维护一组线程，可以高效执行大量小任务，Executors提供了静态方法创建不同类型的ExecutorService。

常用ExecutorService：

* FixedThreadPool：线程数固定
* CachedThreadPool：线程数根据任务动态调整
* SingleThreadExecutor：仅单线程执行
* 必须调用shutdown()关闭ExecutorService
* ScheduledThreadPool可以定期调度多个任务（可取代Timer）

### Future

* Future表示一个未来可能会返回的结果
* 提交Callable任务，可以获得一个Future对象
* 可以用Future在将来某个时刻获取结果

### CompletableFuture

CompletableFuture的优点：

* 异步任务结束时，会自动回调某个对象的方法
* 异步任务出错时，会自动回调某个对象的方法
* 主线程设置好回调后，不再关心异步任务的执行

CompletableFuture对象可以指定异步处理流程：

* thenAccept()处理正常结果
* exceptional()处理异常结果
* thenApplyAsync() 用于串行化另一个CompletableFuture
* anyOf / allOf 用于并行化两个CompletableFuture

### Fork/Join

* Fork/Join是一种基于“分治”的算法：分解任务＋合并结果
* ForkJoinPool线程池可以把一个大任务分拆成小任务并行执行
* 任务类必须继承自RecursiveTask／RecursiveAction
* 使用Fork/Join模式可以进行并行计算提高效率

## # 线程工具类ThreadLocal

调用Thread.currentThread()获取当前线程。

JDK提供了ThreadLocal，在一个线程中传递同一个对象。

ThreadLocal表示线程的“局部变量”，它确保每个线程的ThreadLocal变量都是各自独立的。

ThreadLocal适合在一个线程的处理流程中保持上下文（避免了同一参数在所有方法中传递）

使用ThreadLocal要用try … finally结构，并且中finally中调用remove()方法将ThreadLocal清除，否则可能将ThreadLocal放到线程池中，就会将上个线程得状态带入到下一次线程中

多线程是java实现多任务得基础：

* Thread

调用Thread.currentThread()获取当前线程

* ExecuterService
* ScheduledThreadPool
* Fork/Join

可以将ThreadLocal当作一个全局得Map<Thread, Object>;
每个线程获取ThreadLocal变量时，使用自身自身作为key
