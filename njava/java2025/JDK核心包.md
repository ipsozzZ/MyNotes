## # SE
Java 的核心代码库即 JDK 自带的 **Java 标准类库（Java Class Library）**，是 Java 语言最基础的运行时支持。下面按其功能模块进行整体梳理，并说明主要用途。

---

### 1. 语言基础与系统支持（java.lang）
- **提供 Java 语言最核心的类，自动导入**
- 主要类/接口：`Object`, `String`, `Integer/Long/Double 等包装类`, `Math`, `System`, `Thread`, `Runnable`, `Throwable/Exception/Error`, `ClassLoader`, `Process`, `StackTraceElement` 等
- 用途：对象根类、基本类型封装、字符串处理、数学运算、系统交互、多线程基础、异常体系、反射入口（`Class`）、动态语言支持（`java.lang.invoke`）

### 2. 集合框架与数据结构（java.util）
- **提供常用数据结构和工具类**
- 接口：`Collection`, `List`, `Set`, `Queue`, `Deque`, `Map`, `Iterator`, `Comparator`
- 实现：`ArrayList`, `LinkedList`, `HashSet`, `TreeSet`, `HashMap`, `TreeMap`, `PriorityQueue`, `ArrayDeque` 等
- 工具类：`Collections`, `Arrays`, `Objects`, `Random`, `Scanner`, `StringTokenizer`, `Timer`, `UUID`, `Optional`（实际在 `java.util`）
- 遗留类：`Vector`, `Stack`, `Hashtable`, `Properties`, `Date`, `Calendar` 等
- 用途：各类容器、排序查找、随机数、配置属性、日期时间（旧API）、定时任务、文本扫描

### 3. 并发编程（java.util.concurrent）
- **高并发与多线程工具集**
- 子包/核心类：  
  - 原子类：`AtomicInteger`, `AtomicLong`, `AtomicReference` 等（`java.util.concurrent.atomic`）  
  - 锁：`Lock`, `ReentrantLock`, `ReadWriteLock`, `StampedLock`（`java.util.concurrent.locks`）  
  - 线程池：`Executor`, `ExecutorService`, `ThreadPoolExecutor`, `ScheduledExecutorService`  
  - 同步器：`Semaphore`, `CountDownLatch`, `CyclicBarrier`, `Phaser`, `Exchanger`  
  - 并发集合：`ConcurrentHashMap`, `CopyOnWriteArrayList`, `BlockingQueue`（`ArrayBlockingQueue`, `LinkedBlockingQueue` 等）  
  - 工具：`ForkJoinPool`, `CompletableFuture`, `CompletionService`
- 用途：构建高性能、线程安全的并发应用，任务调度与异步编排

### 4. 函数式编程与流（java.util.function, java.util.stream）
- **函数式接口库**：`Function<T,R>`, `Predicate<T>`, `Consumer<T>`, `Supplier<T>`, `BiFunction`, `UnaryOperator` 等
- **流式操作**：`Stream<T>`, `IntStream`, `LongStream`, `DoubleStream`, `Collectors`
- 用途：支持 Lambda 表达式和方法引用，提供 map/filter/reduce 风格的集合流水线操作

### 5. 输入输出与文件系统（java.io, java.nio）
- **java.io**：传统阻塞 I/O  
  - 字节流：`InputStream`, `OutputStream`, `FileInputStream`, `FileOutputStream`, `BufferedInputStream`  
  - 字符流：`Reader`, `Writer`, `FileReader`, `FileWriter`, `BufferedReader`, `PrintWriter`  
  - 文件/随机访问：`File`, `RandomAccessFile`  
  - 序列化：`Serializable`, `ObjectInputStream`, `ObjectOutputStream`
- **java.nio**：新 I/O（非阻塞与缓冲）  
  - 缓冲区：`ByteBuffer`, `CharBuffer`, `MappedByteBuffer`  
  - 通道：`FileChannel`, `SocketChannel`, `ServerSocketChannel`, `DatagramChannel`  
  - 多路复用：`Selector`, `SelectionKey`  
  - 文件操作：`Files`, `Path`, `Paths`, `FileSystem`（`java.nio.file`）  
  - 文件监听：`WatchService`
- 用途：文件读写、网络 I/O 底层操作、高效大文件处理、非阻塞通信模型

### 6. 网络编程（java.net）
- 类：`URL`, `URLConnection`, `HttpURLConnection`, `URI`
- TCP/UDP：`Socket`, `ServerSocket`, `DatagramSocket`, `InetAddress`
- 用途：实现 HTTP 客户端（基础）、Socket 通信、网络地址解析  
  （注：Java 11 引入的 `java.net.http` 提供现代 HTTP 客户端，支持 HTTP/2、WebSocket）

### 7. 数据库编程（java.sql, javax.sql）
- **java.sql**：`DriverManager`, `Connection`, `Statement`, `PreparedStatement`, `ResultSet`, `SQLException`, `Date/Time/Timestamp`
- **javax.sql**：`DataSource`, `ConnectionPoolDataSource`, `RowSet`
- 用途：JDBC 接口标准，连接各种关系型数据库，执行 SQL，处理结果集

### 8. 数学与高精度计算（java.math）
- `BigInteger`：任意精度的整数
- `BigDecimal`：任意精度的定点数，常用于金融计算
- `MathContext`, `RoundingMode`
- 用途：超越基本类型的精确数值运算

### 9. 现代日期时间 API（java.time）
- 核心类：`LocalDate`, `LocalTime`, `LocalDateTime`, `ZonedDateTime`, `Instant`, `Duration`, `Period`
- 格式化：`DateTimeFormatter`
- 时区：`ZoneId`, `ZoneOffset`
- 用途：不可变、线程安全的日期时间处理，替代旧 `Date/Calendar`

### 10. 文本格式化与国际化（java.text）
- `NumberFormat`, `DecimalFormat`：数字格式化
- `DateFormat`, `SimpleDateFormat`：日期格式化（旧式，推荐 `java.time.format`）
- `MessageFormat`, `ChoiceFormat`：消息模板
- `Collator`：字符串排序
- 用途：本地化显示、多语言支持

### 11. 正则表达式（java.util.regex）
- `Pattern`, `Matcher`
- 用途：字符串匹配、查找、替换与分割

### 12. 日志（java.util.logging）
- `Logger`, `LogRecord`, `Handler`, `Formatter`, `Level`
- 用途：JDK 内置日志系统（轻量级，生产环境多使用 SLF4J/Logback 等替代）

### 13. 反射、注解与元编程
- **java.lang.reflect**：`Class`, `Field`, `Method`, `Constructor`, `Array`, `Modifier`
- **java.lang.annotation**：`@Retention`, `@Target`, `@Inherited` 等元注解，`Annotation` 接口
- **java.lang.invoke**：`MethodHandle`, `MethodHandles`, `CallSite` 等动态语言支持
- 用途：运行时类型探查与操作、动态代理、注解处理器基础、方法句柄高性能调用

### 14. 安全与加密（java.security, javax.crypto）
- **java.security**：`MessageDigest`, `Signature`, `KeyPairGenerator`, `KeyStore`, `SecureRandom`, 权限控制 (`Permission`, `AccessController`)
- **javax.crypto**：`Cipher`, `KeyGenerator`, `SecretKey`, `Mac`
- **javax.net.ssl**：`SSLSocket`, `SSLContext`, `TrustManager`
- 用途：哈希、签名、加解密、密钥管理、安全通信

### 15. 压缩与归档（java.util.zip, java.util.jar）
- `ZipInputStream` / `ZipOutputStream`, `GZIPInputStream`, `JarInputStream` / `JarOutputStream`
- `ZipEntry`, `JarEntry`, `Manifest`
- 用途：处理 ZIP/GZIP 压缩包和 JAR 文件

### 16. 图形用户界面（java.awt, javax.swing）
- **java.awt**：`Frame`, `Panel`, `Button`, `Graphics`, `Image`, `Font`, 布局管理器，AWT 事件模型
- **javax.swing**：`JFrame`, `JPanel`, `JButton`, `JTable`, `JTree`, `JList`, 可插拔观感
- 用途：桌面 GUI 应用开发（现已不是主流，但仍属核心库）

### 17. XML 处理（javax.xml.parsers, org.w3c.dom, org.xml.sax, javax.xml.transform）
- DOM 解析器、SAX 解析器、StAX 流式解析、XSLT 转换、XPath 查询
- 用途：读取、操作和转换 XML 文档

### 18. RMI 与远程调用（java.rmi）
- 接口/类：`Remote`, `UnicastRemoteObject`, `Naming`, `Registry`
- 用途：构建分布式对象应用（如今多被 REST/gRPC 取代）

### 19. 管理与监控（java.lang.management, javax.management）
- **java.lang.management**：`ManagementFactory`, `MemoryMXBean`, `ThreadMXBean` 等，访问 JVM 运行时信息
- **javax.management（JMX）**：`MBeanServer`, `ObjectName`, 暴露管理接口
- 用途：监控堆内存、线程、类加载，构建可管理的系统

### 20. 其他常用工具库
- **java.beans**：JavaBean 内省、属性编辑器、持久化
- **java.util.prefs**：`Preferences` 轻量级跨平台配置存储
- **java.util.spi**：服务提供者接口，用于扩展机制（如 `ResourceBundleProvider`）
- **javax.script**：`ScriptEngineManager`, `ScriptEngine` 脚本引擎支持（如 JavaScript）
- **java.net.http**（Java 11+）：`HttpClient`, `HttpRequest`, `HttpResponse`，现代 HTTP 客户端

---

以上覆盖了 Java SE 的核心代码库及其典型用途。这些库构成了 Java 开发的基石，也是成为资深工程师必须深度掌握的内容。

---
---
---
---
--- Jet135975

## # Jakarta EE
Java EE（现 Jakarta EE）的核心是**一套用于构建分布式、事务性、可扩展企业应用的规范集合**。它本身不是代码库，而是定义了一系列接口和合约，由应用服务器（如 WildFly、TomEE）提供实现。

下面先整体列出 Java EE 的核心知识模块及其用途，然后从中重点拆解 **3 种最常用的技术**：**Servlet**、**JPA**、**JAX-RS**。

---

### 一、Java EE / Jakarta EE 核心知识全景图

| 核心规范 | 主要用途 |
|----------|----------|
| **Servlet** | Web 应用基石，处理 HTTP 请求/响应，所有 Java Web 框架的底层 |
| **JPA** (Java Persistence API) | ORM 标准，将 Java 对象映射到关系数据库，代替复杂 JDBC |
| **JAX-RS** (Java API for RESTful Web Services) | 构建 RESTful 接口的标准方式，轻量级替代 SOAP |
| **CDI** (Contexts and Dependency Injection) | 统一的依赖注入和上下文管理（类似 Spring IoC，但是 Java EE 原生标准） |
| **EJB** (Enterprise Java Beans) | 封装业务逻辑的组件模型，提供声明式事务、安全、远程调用 |
| **JTA** (Java Transaction API) | 管理跨多个数据库、消息队列的分布式事务 |
| **JMS** (Java Message Service) | 异步消息传递，解耦系统间的通信 |
| **Bean Validation** | 数据校验统一标准 (Hibernate Validator 是参考实现) |
| **JSF** (JavaServer Faces) | 基于组件的 MVC 框架（现在主要用于遗留企业内部系统） |
| **WebSocket** | 双向实时通信 (以注解驱动，简化开发) |
| **Security API** | 统一的认证与授权接口 |
| **Batch** (JSR 352) | 处理大批量任务的批处理框架 |

上面这些规范构成了传统意义的企业应用架构，一个完整的企业应用往往会同时用到多个规范：比如用 **Servlet + JAX-RS** 做前后端分离，用 **JPA** 持久化，用 **CDI** 管理 Bean，用 **JTA** 控制事务，用 **JMS** 解耦服务。

下面重点介绍**现代 Java 开发中依然高频使用、任何资深工程师都无法绕开的三个核心知识**。

---

### 二、重点 1：Servlet — Java Web 的“操作系统”

#### 用途
所有 Java Web 框架（Spring MVC、Struts、JSF）的底层都运行在 **Servlet 容器**（如 Tomcat、Jetty）之上。Servlet 定义了如何处理客户端请求、生成响应，并提供了会话管理、过滤器链、监听器等基础能力。

#### 核心概念
- **`Servlet` 接口**：`init()`, `service()`, `destroy()` 生命周期方法。
- **`HttpServlet`**：最常用的抽象类，覆写 `doGet()`, `doPost()` 等处理不同 HTTP 方法。
- **`Filter`**：在请求到达 Servlet 前或响应返回客户端前进行拦截（如编码、鉴权、日志）。
- **`HttpSession`**：服务器端的会话管理。
- **`ServletContext`**：整个 Web 应用的全局上下文。
- **部署描述符 `web.xml`** 或 **注解 `@WebServlet`**：将 URL 映射到 Servlet。

#### 为什么重要
即使今天直接写原生 Servlet 的项目变少，但 Spring Boot 内嵌的 Tomcat 本质上依然是 Servlet 容器。理解 **请求→过滤器链→Servlet→响应** 的完整过程，是调优性能、排查乱码、设计安全拦截的必备功底。

```java
// 最简原生 Servlet 示例
@WebServlet("/hello")
public class HelloServlet extends HttpServlet {
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) 
            throws IOException {
        resp.setContentType("text/plain");
        resp.getWriter().write("Hello, Servlet!");
    }
}
```

---

### 三、重点 2：JPA — 对象关系映射的标准

#### 用途
JPA 让开发者可以用面向对象的方式操作数据库，无需写大量 JDBC 样板代码和 SQL。它管理实体对象的生命周期、表之间的关联、继承映射、缓存策略等，是 Java 持久层的事实标准。

#### 核心概念
- **Entity**：轻量级持久化域对象，通过 `@Entity` 标注，对应数据库表。
- **EntityManager**：核心接口，负责实体 CRUD、查询、事务持久化上下文。
- **JPQL**：类似 SQL 但面向对象的查询语言（也可以直接使用原生 SQL）。
- **Metamodel / Criteria API**：类型安全的动态查询。
- **关联映射**：`@OneToMany`, `@ManyToOne`, `@OneToOne`, `@ManyToMany`。
- **继承策略**：单表、具体表、类层次结构。
- **持久化上下文与事务**：一级缓存、脏检查、乐观锁 (`@Version`)。
- **标准实现**：Hibernate 是最常见的 JPA 提供者。

#### 为什么重要
无论是传统的 Java EE 项目还是 Spring Data JPA，底层都是 JPA 规范。掌握 JPA 意味着你可以：
- 用极少的代码实现复杂查询与关联操作
- 理解 N+1 查询问题并优化 Fetch 策略
- 设计合理的主键和并发控制方案

```java
@Entity
public class Order {
    @Id @GeneratedValue
    private Long id;
    
    @ManyToOne
    private Customer customer;
    
    @OneToMany(mappedBy = "order", cascade = ALL)
    private List<OrderItem> items;
    // ...
}

// 典型使用
EntityManager em = ...;
Order order = em.find(Order.class, 1L);
TypedQuery<Order> query = em.createQuery(
    "SELECT o FROM Order o JOIN FETCH o.items WHERE o.customer = :cust", Order.class);
```

---

### 四、重点 3：JAX-RS — RESTful 服务的标准写法

#### 用途
JAX-RS 使用声明式注解将 Java 方法直接映射到 HTTP 资源和操作，是构建 REST API 最优雅的方式之一。它屏蔽了底层 Servlet API，让开发者专注于资源和表示。

#### 核心概念
- **资源类**：用 `@Path` 标注的普通 Java 类。
- **HTTP 方法绑定**：`@GET`, `@POST`, `@PUT`, `@DELETE`, `@PATCH`。
- **参数注入**：`@PathParam`, `@QueryParam`, `@HeaderParam`, `@FormParam`, `@CookieParam`。
- **内容协商**：`@Produces`, `@Consumes` 指定媒体类型（JSON/XML）。
- **响应构建**：`Response` 类灵活构建状态码、头信息。
- **客户端 API**：JAX-RS 也定义了标准客户端，如 `Client`、`WebTarget`，可发起 HTTP 请求。
- **实现**：Jersey、RESTEasy（Spring MVC 虽不是 JAX-RS 实现，但设计受其影响）。

#### 为什么重要
微服务和前后端分离架构下，REST API 是系统间交互的主流方式。JAX-RS 是 Java 世界里设计 REST 接口的官方标准，其**资源导向**的设计思想至今指导着 API 设计。即使使用 Spring MVC，JAX-RS 的许多概念（路径映射、参数注入）也高度一致，深入理解有助于写出更符合 HTTP 语义的服务。

```java
@Path("/customers")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class CustomerResource {
    
    @GET
    @Path("/{id}")
    public Customer getCustomer(@PathParam("id") Long id) {
        return customerService.findById(id);
    }
    
    @POST
    public Response create(Customer customer, @Context UriInfo uriInfo) {
        Customer saved = customerService.save(customer);
        return Response.created(
            uriInfo.getAbsolutePathBuilder().path(saved.getId().toString()).build()
        ).build();
    }
}
```

---

### 五、总结与补充

- **Servlet** 是 Java Web 的根基，所有 HTTP 处理都建立在它之上。  
- **JPA** 是持久化的标准语言，解决了面向对象与关系数据库的“阻抗失配”。  
- **JAX-RS** 是构建 RESTful 服务的官方哲学，让 HTTP 接口开发变得自然且规范。

这三者，加上 **CDI** 管理 Bean 的生命周期与依赖（`@Inject`），构成了大多数 Java 企业应用的基础骨架。即使你主要使用 Spring Boot，底层也大量依赖 Servlet 和 JPA，而 JAX-RS 的思想则深刻影响了 Spring MVC 的设计。作为资深工程师，深入掌握这些规范，不仅能更好地理解框架底层，还能在技术选型和架构设计时做出更准确的判断。