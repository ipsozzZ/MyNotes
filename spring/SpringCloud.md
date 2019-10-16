# SpringCloud

分布式微服务集群，主要使用demo演示

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
            <artifactId>spring-cloud-starter-netflix-eureka-server</artifactId>
        </dependency>
    </dependencies>

</project>

```

## # Ribbon负载均衡

springcloud请求处理过程：系统接收到一个请求 --> 负载均衡器 --> 微服务集群中算法计算到的微服务

上面注册了服务这里就可以调用服务了
