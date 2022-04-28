# SpringCloud

分布式微服务集群，主要使用demo演示，dome项目springCloud

## # 基础框架及Eureka服务管理

1. 创建父工程
2. 在父工程下创建个springboot微服务

父工程中删除src目录，然后编辑pom.xml

```xml

<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>live.ipso.springcloud</groupId>
    <artifactId>microservice-cloud-01</artifactId>
    <version>1.0-SNAPSHOT</version>

    <!-- 子模块,这里创建子工程后自动添加 -->
    <modules>
        <module>../microservice-cloud-02-api</module>
        <module>../microservice-cloud-03-provider-product-8081</module>
        <module>../microservice-cloud-04-consumer-product-8082</module>
        <module>../microservice-cloud-05-eureka-8083</module>
        <module>../microservice-cloud-05-eureka-8084</module>
    </modules>
    <packaging>pom</packaging>  <!-- 类型为作为pom，因为这是父工程 -->

    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>2.0.7.RELEASE</version>
        <relativePath/>
    </parent>

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>1.8</maven.compiler.source>
        <maven.compiler.target>1.8</maven.compiler.target>
        <junit.version>4.12</junit.version>
        <!-- spring cloud 采用 Finchley.SR2 版本 -->
        <spring-cloud.version>Finchley.SR2</spring-cloud.version>
    </properties>

    <!-- 依赖声明,并不是真实的导入 -->
    <dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>org.springframework.cloud</groupId>
                <artifactId>spring-cloud-dependencies</artifactId>
                <version>${spring-cloud.version}</version>
                <type>pom</type>
                <!--maven不支持多继承，使用import来依赖管理配置-->
                <scope>import</scope>
            </dependency>
            <!--导入 mybatis 启动器-->
            <dependency>
                <groupId>org.mybatis.spring.boot</groupId>
                <artifactId>mybatis-spring-boot-starter</artifactId>
                <version>1.3.2</version>
            </dependency>
            <!--druid数据源-->
            <dependency>
                <groupId>com.alibaba</groupId>
                <artifactId>druid</artifactId>
                <version>1.1.12</version>
            </dependency>
            <dependency>
                <groupId>mysql</groupId>
                <artifactId>mysql-connector-java</artifactId>
                <version>8.0.13</version>
            </dependency>
            <dependency>
                <groupId>junit</groupId>
                <artifactId>junit</artifactId>
                <version>${junit.version}</version>
                <scope>test</scope>
            </dependency>
        </dependencies>
    </dependencyManagement>

</project>

```

3. 创建实体类模块项目(就是数据库映射对象)

创建并实现实体类即可，相对简单

4. 创建一个微服务

一个微服务就是一个springboot项目

**增加的地方**:

```xml

<!-- pom.xml -->
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <parent>
        <artifactId>microservice-cloud-01</artifactId>
        <groupId>live.ipso.springcloud</groupId>
        <version>1.0-SNAPSHOT</version>
        <!--<relativePath>../microservice-cloud-01/pom.xml</relativePath>-->
    </parent>
    <modelVersion>4.0.0</modelVersion>

    <artifactId>microservice-cloud-03-provider-product-8081</artifactId>

    <dependencies>
        <dependency>
            <groupId>live.ipso.springcloud</groupId>
            <artifactId>microservice-cloud-02-api</artifactId>
            <version>1.0-SNAPSHOT</version>
        </dependency>

        <!--springboot web启动器-->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>

        <!-- mybatis 启动器-->
        <dependency>
            <groupId>org.mybatis.spring.boot</groupId>
            <artifactId>mybatis-spring-boot-starter</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-test</artifactId>
        </dependency>

        <!-- 导入Eureka客户端的依赖，将 微服务提供者 注册进 Eureka -->
        <dependency>
            <groupId>org.springframework.cloud</groupId>
            <artifactId>spring-cloud-starter-netflix-eureka-client</artifactId>
        </dependency>

        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
        </dependency>
        <dependency>
            <groupId>mysql</groupId>
            <artifactId>mysql-connector-java</artifactId>
        </dependency>
        <dependency>
            <groupId>com.alibaba</groupId>
            <artifactId>druid</artifactId>
        </dependency>
    </dependencies>

</project>

```

```yml
# application.yml
server:
  port: 8081

mybatis:
  config-location: classpath:mybatis/mybatis.cfg.xml        # mybatis配置文件所在路径
  type-aliases-package: live.ipso.springcloud.entities  # 所有Entity别名类所在包
  mapper-locations: classpath:mybatis/mapper/**/*.xml       # mapper映射文件

spring:
  application:
    name: microservice-product #这个很重要，这在以后的服务与服务之间相互调用一般都是根据这个name
  datasource:
    type: com.alibaba.druid.pool.DruidDataSource            # 当前数据源操作类型
    driver-class-name: com.mysql.cj.jdbc.Driver             # mysql驱动包
    url: jdbc:mysql://127.0.0.1:3306/springcloud01?serverTimezone=GMT%2B8  # 数据库名称
    username: root
    password: gqm1975386453
    dbcp2:
      min-idle: 5                                # 数据库连接池的最小维持连接数
      initial-size: 5                            # 初始化连接数
      max-total: 5                               # 最大连接数
      max-wait-millis: 150                       # 等待连接获取的最大超时时间

# Eureka服务配置
eureka:
  client:
    registerWithEureka: true # 服务注册开关，true表示将自己注册到eureka服务中
    fetchRegistry: true      # 服务发现，true表示从eureka中获取注册信息
    serviceUrl:              # eureka客户端与eureka服务端的交互地址，集群版配置对方的地址，单机版配置自己（如果不配置默认本机的8761端口）
      # defaultZone: http://www.ipso.me:8083/eureka  # 就是向其它微服务暴漏自己的地址，方便其它微服务调用
      defaultZone: http://www.ipso.me:8083/eureka,http://localhost:8084/eureka  # 注册到多个服务器使用逗号隔开两个地址
  instance:
    instanceId: ${spring.application.name}:${server.port} # status 显示的内容
    prefer-ip-address: true  # 前缀显示ip地址

```

5. 创建一个服务注册服务

主要实现一个启动类

```java

package live.ipso.springcloud;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.eureka.server.EnableEurekaServer;

@EnableEurekaServer  // 标识一个Eureka服务注册中心
@SpringBootApplication
public class EurekaServer_8083 {

   public static void main(String[] args) {
      SpringApplication.run(EurekaServer_8083.class, args);
   }
}

配置application.yml

```yml

server:
  port: 8083

eureka:
  instance:
    hostname: www.ipso.me # eureka服务端的实例名称
  client:
    register-with-eureka: false
    fetch-registry: false
    service-url:
      # defaultZone: http://${eureka.instance.hostname}:${server.port}/eureka/ # 单机版
      defaultZone: http://localhost:8084/eureka/  # 集群版配对方的地址,如果有多台服务注册服务器就用逗号将两个地址隔开
  server:
    enable-self-preservation: false # 禁用自我保护机制， 注意禁用后对于挂掉的服务，会在90后清除，部署后应该启用

    # registerWithEureka: false #注册服务，false表示不将自己注册到eureka服务中
    # fetch-egistry: false # 服务发现，false表示不从eureka中获取注册信息

    # serviceUrl:  # eureka客户端与eureka服务端的交互地址，集群版配置对方的地址，单机版配置自己（如果不配置默认本机的8761端口）
    # defaultZone: 就是向其它微服务暴漏自己的地址，方便其它微服务调用

```

配置pom.xml

```xml

<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <parent>
        <artifactId>microservice-cloud-01</artifactId>
        <groupId>live.ipso.springcloud</groupId>
        <version>1.0-SNAPSHOT</version>
        <!--<relativePath>../microservice-cloud-01/pom.xml</relativePath>-->
    </parent>
    <modelVersion>4.0.0</modelVersion>

    <artifactId>microservice-cloud-05-eureka-8083</artifactId>

    <dependencies>
        <!-- 导入 eureka-server 服务端依赖 -->
        <dependency>
            <groupId>org.springframework.cloud</groupId>
            <artifactId>spring-cloud-starter-netflix-eureka-client</artifactId>
        </dependency>
    </dependencies>

</project>

```

## # Ribbon客户端负载均衡

springcloud请求处理过程：系统接收到一个请求 --> 负载均衡器 --> 微服务集群中算法计算到的微服务

**服务端负载均衡**：

负载均衡器维护一份服务列表，根据负载均衡算法将请求转发到相应的微服务上，负载均衡算法有：轮训(默认)，随机，加权轮训，加权随机，地址哈希等，所以负载均衡可以为微服务集群分担请求，降低系统压力。

**客户端负载均衡**：

客户端负载均衡与服务端负载均衡的区别在于：客户端负载均衡要维护一份服务列表。两者的最大区别在于服务清单存储的位置。在客户端负载均衡中，每个客户端服务都有一份自己要访问的服务清单，这些清单都是从Eureka服务注册中心获取的，而在服务器负载均衡中，只要负载均衡器维护一份服务列表。Ribbon实现的就是客户端负载均衡，它不用在创建新的微服务，只需要在已有的服务中加入Ribbon即可(一个业务微服务就是一个客户端)，如：现有多个同一产品生产者服务，负载均衡就是负责合理分配每一次该产品生产请求到合理的产品生产微服务。

我们在消费者服务中引入eureka客户端启动器，application.yml配置文件如下：

```yml

server:
  port: 80

eureka:
  client:
    register-with-eureka: false # false表示不将当前服务注册到eureka, 因为当前服务是返回给用户，其它服务并不需要调用，所以不用注册到eureka，相当于客户端
    fetch-registry: true # 服务发现，因为要生成服务清单，所以需要发现注册到eureka中的服务
    service-url: http://www.ipso.me:8084/eureka/,http://localhost:8083/eureka/

```

自定义配置类如下：

```java

package live.ipso.springcloud.config;

import org.springframework.cloud.client.loadbalancer.LoadBalanced;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestTemplate;

@Configuration // 标识这是一个配置类
public class ConfigBean {

   @LoadBalanced  // 实现负载均衡，并且调用地址可以变成服务名，之前是使用IP地址，一般来说IP会因部署改变，但是服务名一般不会变，所以使用服务名更好
   @Bean
   public RestTemplate getRestTemplate(){
      return new RestTemplate();
   }
}

```

## # Feign客户端接口调用

功能类似RestTemplate + Ribbon，区别在于，Feign是面向接口编程的风格，实现的功能是一样的，都是调用eureka注册的微服务。

使用Feign可以很方便的、简单的实现HTTP客户端(也就是上例中的消费者，负责返回数据给用户的)。使用Feign只需要定义一个接口，然后在接口上添加注解即可。spring cloud对Feign进行了封装，Feign默认集成了Ribbon实现客户端负载均衡

**Feign注意事项**：

SpringCloud对Feign进行了增强兼容了SpringMVC的注解，我们使用SpringMVC的注解时需要注意：

1. @Feign接口方法有基本类型参数时,在参数前必须加@PathVariable("xxx")或@RequestParam("xxx")
2. @Feign接口方法返回值为复杂对象时，返回的对象必须有无参的构造方法。

## # Hystrix熔断器

在分布式微服务中，服务之间调用的链路上，由于网络原因、资源繁忙或者自身问题，服务并不能100%可用，如果单个服务出现问题，调用这个服务就会出现线程堵塞，导致响应时间过长或不可用，此时如有大量的请求，容器的线程资源会被完全消耗完毕，会导致服务瘫痪，服务与服务之间的依赖性，导致故障会传播，会对整个微服务框架造成灾难性后果。这就是服务故障中的"雪崩"效应。为了解决这个问题，业界提出熔断器模型

**解决"雪崩"效应的方法**：

Hystrix熔断器：

1. 服务的熔断
2. 服务监控

**注意**：

我们既可以在生产者服务中做熔断器，也可以在消费者服务中做熔断器

**在生产者中做熔断器示例**：

生产者中主要的关于Hystrix熔断器的代码示例

* pom.xml中添加依赖

```xml

<!-- pom.xml中添加Hystrix依赖 -->
<!-- hystrix熔断器 -->
<dependency>
    <groupId>org.springframework.cloud</groupId>
    <artifactId>spring-cloud-starter-netflix-hystrix</artifactId>
</dependency>

```

* 在启动类中添加注解"@EnableHystrix"开启Hystrix 熔断服务
* 控制器代码示例

```java

package live.ipso.springcloud.Controller;

import com.netflix.hystrix.contrib.javanica.annotation.HystrixCommand;
import live.ipso.springcloud.entities.Product;
import live.ipso.springcloud.service.ProductService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController  // 响应的都是json字符串
public class ProductController {

   @Autowired
   ProductService productService;

   /**
    * 根据id获取一条数据
    * @param id 查询的id
    * @return Product json字符串
    */
   // 当前方法出现异常时转到熔断器方法"getFallBack()",两个方法返回值、参数要一致
   @HystrixCommand(fallbackMethod = "getFallBack")
   @RequestMapping(value = "/product/get/{id}", method = RequestMethod.GET)
   public Product get(@PathVariable("id") Integer id){
      Product product = productService.getOne(id);
      // product为空则模拟一个异常
      if (product == null){
         throw new RuntimeException("ID= " + id + "无效");
      }
      return product;
   }

   /**
    * get方法熔断器，当get方法异常时调用此方法
    * @param id 查询的id
    * @return Product json字符串
    */
   public Product getFallBack(@PathVariable("id") Integer id){
      return new Product(id, "ID=" + id + "id无效", "无","无法找到对应数据库");
   }

}

```

**消费者中的熔断器示例**：

* 在application.yml中

```yml

server:
  port: 8082

eureka:
  client:
    register-with-eureka: false # false表示不将当前服务注册到eureka, 因为当前服务是返回给用户，其它服务并不需要调用，所以不用注册到eureka，相当于客户端
    fetch-registry: true # 服务发现，因为要生成服务清单，所以需要发现注册到eureka中的服务
    service-url:
      defaultZone: http://www.ipso.me:8084/eureka/,http://localhost:8083/eureka/

feign:
  hystrix:
    enabled: true  # 开启服务熔断器

```

* service包中接收生产者服务的接口

```java

package live.ipso.springcloud.service;

import live.ipso.springcloud.entities.Product;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;

import java.util.List;

// value指定调用微服务的名称，不区分大小写
// fallback 作用，指定熔断处理类，如果被调用的方法出现异常，就会调用熔断处理类中的方法进行处理
@FeignClient(value = "microservice-product", fallback = ProductClientServiceFallBack.class)
public interface ProductClientService {

   /**
    * 通过RestTemplate向生产者服务中添加产品
    * @return bool
    */
   @RequestMapping(value = "/product/add") // 特别注意这里的value是调用的微服务中的RequestMapping
   Boolean add(Product product);

   /**
    * 通过RestTemplate从生产者服务中获取产品
    * @param id 产品id
    * @return Product
    */
   @RequestMapping(value = "/product/get/{id}", method = RequestMethod.GET)
   Product get(@PathVariable("id") Integer id);

   /**
    * 从生产者服务中获取产品列表
    * @return List Product
    */
   @RequestMapping(value = "/product/list", method = RequestMethod.GET)
   List<Product> list();
}

```

* 在service包中添加熔断处理类

```java

package live.ipso.springcloud.service;

import live.ipso.springcloud.entities.Product;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * 熔断处理器类
 * 如果在调用ProductClientService中的方法出现异常时，就会调用熔断器类中对应的方法
 * 这里以get方法为例，其它方法做返回null处理
 */
@Component  // 一定要添加，将当前类纳入容器中,否者直接报错
public class ProductClientServiceFallBack implements ProductClientService {
   @Override
   public Boolean add(Product product) {
      return null;
   }

   /**
    * 原方法处理异常时，触发
    * @param id 产品id
    * @return Product
    */
   @Override
   public Product get(Integer id) {
      return new Product(id, "id= " + id + "无效，当前为熔断处理器的处理结果", "无", "无数据源");
   }

   @Override
   public List<Product> list() {
      return null;
   }
}

```

* 注意生产者中的熔断器与消费者中的熔断器的区别

## # Hystrix Dashboard监控平台搭建

**什么是服务监控**：

* 除了隔离依赖服务的调用外，Hystrix还提供了准实时的调用监控(Hystrix Dashboard), Hystrix会持续的地记录所有通过Hystrix发起的请求执行信息，并以统计报表和图形的形式展示给用户，包括每秒执行多少请求多少成功，多少失败等。
* Netfix通过hystrix-metrics-event-stream项目实现了对以上指标的监控，SpringCloud也提供了Hystrix Dashboard的整合，对监控内容转化成了可视化界面。

**创建一个微服务用来做"服务监控"**:

这个微服务只需要做：

* application.yml中

```yml

server:
  port: 9001  # 配置端口号

```

* 启动类

```java

package live.ipso.springcloud;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.hystrix.dashboard.EnableHystrixDashboard;

/**
 * "服务监控"微服务启动类
 */
@EnableHystrixDashboard // 开启服务监控
@SpringBootApplication
public class HystrixDashboard_9001 {
   public static void main(String[] args) {
      SpringApplication.run(HystrixDashboard_9001.class, args);
   }
}

```

* 访问

这里是本地测试所以访问的是：```http://localhost:9001/hystrix``` 即可看到可视化的服务监控界面

* 在其它微服务中配置服务监控(配置被监控的微服务)

  1. 在需要被监控的服务的pom中的dependencies节点添加spring-boot-starter-actuator监控依赖；
  2. 开启依赖相关的断点,也就是被监控的地址
  3. 确保已经引入熔断器的依赖spring-cloud-starter-netfix-hystrix

**常见错误**：

Unable to connect to Command Metric Stream

这个错误就是因为上面的3个配置步骤出现问题导致的。

## # Zuul路由网关

Zuul路由就是针对微服务做一次路由转发，让用户不知到真正的访问路由，从而保护微服务安全

**什么是Zuul**:

Spring Cloud Zuul是整合Netfix公司的Zuul开源项目。Zuul包含了对请求路由和校验过滤两个主要功能：

* 其中路由功能负责将外部请求转发到具体的微服务实例上，是实现外部访问的统一入口的基础
  
   1. 客户端请求网关/api/product，通过路由转发到product服务
   2. 客户端请求网关/api/order，通过路由转发到order服务

* 而过滤功能则负责对请求的处理过程进行干预，是实现请求校验等功能的基础

Zuul和Eureka进行整合，将Zuul自身注册为Eureka服务治理中的服务，同时从Eureka中获得其他微服务的消息，也就是说以后的访问微服务都是通过Zuul跳转后获得。**注意: Zuul服务最终还是会注册进Eureka**

**实现路由功能实例**：

```xml

<!-- pom.xml -->

<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <parent>
        <artifactId>microservice-cloud-01</artifactId>
        <groupId>live.ipso.springcloud</groupId>
        <version>1.0-SNAPSHOT</version>
        <!--<relativePath>../microservice-cloud-01/pom.xml</relativePath>-->
    </parent>
    <modelVersion>4.0.0</modelVersion>

    <artifactId>microservice-cloud-10-zuul-gateway-7001</artifactId>

    <dependencies>
        <dependency>
            <groupId>org.springframework.cloud</groupId>
            <artifactId>spring-cloud-starter-netflix-eureka-client</artifactId>
        </dependency>

        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>

        <!-- zuul路由网关依赖 -->
        <dependency>
            <groupId>org.springframework.cloud</groupId>
            <artifactId>spring-cloud-starter-netflix-zuul</artifactId>
        </dependency>
    </dependencies>

</project>

```

```java

// 入口类

package live.ipso.springcloud;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.zuul.EnableZuulProxy;

@EnableZuulProxy  // 开启Zuul功能
@SpringBootApplication
public class Zuulserver_7001 {
   public static void main(String[] args) {
      SpringApplication.run(Zuulserver_7001.class, args);
   }
}

```

```yml

# application.yml

server:
  port: 7001

spring:
  application:
    name: microservice-zuul-gateway  # 微服务名称

# 配置服务注册到Eureka服务注册中心
eureka:
  client:
    register-with-eureka: true
    fetch-registry: true
    service-url:
      defaultZone: http://localhost:8084/eureka/, http://www.ipso.me:8083/eureka/
  instance:
    instance-id: ${spring.application.name}:${server.port}  # 指定实例ID，显示为服务名:端口
    prefer-ip-address: true  # 访问路径可以显示IP地址

# zuul路由服务配置
zuul:
  routes: # 路由配置，可配置多组路由

    # 这是第一组路由信息
    provider-product: # 路由名称，名称任意，路由名称唯一
      path: /product/**  # 访问路径
      serviceId: microservice-product # 指定服务ID(也就是服务名)会自动从Eureka中找到此服务的ip和端口
      stripPrefix: true # true值代理转发请求时去掉前缀，false值不做处理。
      # 如：/product/get/2  为true时： /get/2   为false时： /product/get/2

      # 这是第二组路由信息
      # provider-order:
        # path: /order/**
        # serviceId: microservice-order
        # stripPrefix: false

```

**实现路由服务中过滤器功能实例**：

**自定义过滤器**：

需要继承ZuulFilter,ZuulFilter是一个抽象类，需要实现它的4个方法，如下：

* filterType: 返回字符串代表过滤器的类型，返回值有：
  1. pre: 在请求之前执行
  2. route: 在请求路由时调用
  3. post: 在请求路由之后调用，也就是在route、error过滤器之后调用
  4. error: 处理请求发生错误时调用

* filterOrder: 此方法返回整数值，通过此值来定义过滤器的执行顺序，数字越小优先级越高
* shouldFilter: 返回Boolen值判断该过滤器是否执行，返回true表示要执行此过滤器，false不执行
* run: 要执行的过滤器业务逻辑

**实例**：

```java

// 登录验证过滤器类

package live.ipso.springcloud.filter;

import com.netflix.zuul.ZuulFilter;
import com.netflix.zuul.context.RequestContext;
import com.netflix.zuul.exception.ZuulException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import java.io.IOException;

@Component  // 将过滤器注册的spring容器中，一定不能少了
public class LoginFilter extends ZuulFilter {

   private Logger logger = LoggerFactory.getLogger(getClass()); // 获取日志对象

   /**
    * 返回值有：
    * 1. pre: 在请求之前执行
    * 2. route: 在请求路由时调用
    * 3. post: 在请求路由之后调用，也就是在route、error过滤器之后调用
    * 4. error: 处理请求发生错误时调用
    * @return
    */
   @Override
   public String filterType() {
      return "pre";
   }

   /**
    * 此方法返回整数值，通过此值来定义过滤器的执行顺序，数字越小优先级越高
    * @return int
    */
   @Override
   public int filterOrder() {
      return 1;
   }

   /**
    * 返回Boolen值判断该过滤器是否执行，返回true表示要执行此过滤器，false不执行
    * @return
    */
   @Override
   public boolean shouldFilter() {
      return true; // 设置为true表示当前过滤器需要执行
   }

   /**
    * 要执行的过滤器业务逻辑
    * @return
    * @throws ZuulException
    */
   @Override
   public Object run() throws ZuulException {

      // 1. 获取请求上下文
      RequestContext context = RequestContext.getCurrentContext();
      HttpServletRequest request = context.getRequest(); // 获取Request
      String token = request.getParameter("token"); // 获取token参数

      // 判断是否有token, 有token表示已经登录过，可以放行
      if (token == null){
         // 没有登录，不进行路由转发，或者将路由转发到登录服务
         logger.warn("此操作需要先登录系统"); // 打印日志
         context.setSendZuulResponse(false); // 拒绝访问
         context.setResponseStatusCode(200); // 响应状态码
         try {
            // 设置响应信息,它将输出在浏览器上
            context.getResponse().getWriter().write("token is empty.... please to login");
         } catch (IOException e) {
            e.printStackTrace();
         }
         return null;
      }
      // token不为空，进行路由转发
      return null;
   }
}

```

## # SpringCloud Config 分布式配置中心

在分布式微服务架构中，由于微服务数量众多，使得有很多配置文件，在更新配置文件时很麻烦。我们每个微服务中都有自己的application.yml，上百个配置文件管理起来就会很麻烦，所以一套集中式的，动态的配置管理功能是必不可少的，在SpringCloud中，有分布式配置中心组件SpringCloud Config来解决这个问题。

### SpringCloud Config分客户端和服务端

**服务端 config server**：也就是配置服务中心，是一个集中管理配置文件的微服务

**客户端 config client**：是将配置文件交给Config配置服务中心管理的每一个微服务

**作用**：

* 集中式管理配置文件
* 不同环境不同配置，动态化的配置更新，根据不同环境部署，如dev/test/prod
* 运行期间动态调整配置，不需要在每个每个服务部署的机器上编写配置，服务会向配置中心统一拉取自己的配置信息
* 当配置发生变动时，服务不需要重启即可感知到配置的变化并使用修改后的配置信息
* 将配置信息以REST接口的形式暴露

**与github整合配置信息**：

由于SpringCloud Config官方推荐使用Git来管理配置文件(也支持其它方式，如SVN和本地文件),而且使用https/http访问的形式

**整合案例**：

```java

// 入口类

package live.ipso.springcloud;


import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.config.server.EnableConfigServer;

@EnableConfigServer  // 配置为服务端
@SpringBootApplication
public class ConfigServer_5001 {

   public static void main(String[] args) {
      SpringApplication.run(ConfigServer_5001.class, args);
   }
}

```

```yml

# application.yml

server:
  port: 5001
spring:
  application:
    name: microservice-config
  cloud:
    config:
      server:
        git:  # 远程库的git地址
          uri: https://github.com/ipsozzZ/Microservice-SpringCoud-Config.git

```

```yml

# github仓库中的yml

# 此yml文件是给客户端使用，而resource目录下的application.yml是给自己这个微服务使用
spring:
  profiles:
    active: dev  # 激活开发环境配置

server:
  port: 4001

spring:
  profiles: dev  # 开发环境
  application:
    name: microservice-config-dev  # 微服务名称

server:
  port: 4002

spring:
  profiles: prod  # 开发环境
  application:
  name: microservice-config-prod  # 微服务名称

```

**调用**：

* 方式1

格式：/{appication}-{profile}.yml   读取的配置文件名-环境配置项.yml  (默认分支为master分支)
例： http://localhost:5001/Microservice-config-application-dev.yml  (注意这里{application}是：Microservice-config-application。{profile}是：dev)

* 方式2

其它方式(master分支或非master分支)：

格式：/{appication}/{profile}/{label}   读取的配置文件名/环境配置项/分支名
例： http://localhost:5001/Microservice-config-application/dev/master

* 方式3

格式：/label/{appication}-{profile}.yml   /分支名/读取的配置文件名-环境配置项.yml
例： http://localhost:5001/master/Microservice-config-application-dev.yml

### 配置bootstrap.yml

* application.yml 是用户级别的配置项文件
* bootstrap.yml   是系统级别的配置项文件

SpringCloud 会创建一个Bootstrap Context（bootstrap上下文）,Bootstrap Context会负责从外部资源加载配置属性并解析配置；Bootstrap属性有高优先级，默认情况下，它们不会被本地配置覆盖

## SpringCloud Bus使用机制
