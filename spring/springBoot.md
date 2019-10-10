# SpringBoot知识整理

## # 关于springBoot

>SpringBoot是Spring项目中的一个子工程与我们所熟知的Spring-framework 同属于spring的产品设计目的是用来简化新Spring应用的初始搭建以及开发过程最主要作用就是帮我们快速的构建庞大的spring项目，并且尽可能的减少一切xml配置做到开箱即用，迅速上手，让我们关注与业务而非配置以jar包方式独立运行（jar -jar xxx.jar）内嵌Servlet容器（tomcat, jetty）,无需以war包形式部署到独立的servlet容器中提供starter简化maven依赖包配置自动装配bean(大多数场景)提倡使用java配置和注解配置结合而无需xml配置

## # 搭建入门程序

这里搭建一个入门的springBoot web项目

### pom配置

```xml

<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>live.ipso</groupId>
    <artifactId>01-SpringBootPro</artifactId>
    <version>1.0.0</version>

    <!-- parent pom中的dependencyManager标签用来定义工程中所有可能用到的依赖，如果没用到则不会将其中定义的依赖打包 -->
    <!-- 父pom中的依赖在子pom中依然生效，并且打包的时候也会一同打包 -->

    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>2.1.3.RELEASE</version>
    </parent>

    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
            <!-- parent中已经定义过版本号了 -->
        </dependency>

        <!-- 添加热部署依赖 -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-devtools</artifactId>
        </dependency>

        <dependency>
            <groupId>com.alibaba</groupId>
            <artifactId>druid</artifactId>
            <version>1.1.6</version>
        </dependency>
    </dependencies>


</project>

```

### 启动类

```java

/* 启动类 */
@SpringBootApplication
public class Application {
   public static void main(String[] args) {
      SpringApplication.run(Application.class);
   }

}

// 按住ctrl键鼠标单击SpringBootApplication进入SpringBootApplication类，里面有注解：
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Inherited
@SpringBootConfiguration  // 标记springBoot配置文件，使用@SpringBootApplication注解的类使用@Bean时会将该类交由spring管理，一般一个项目中只在一个文件中用一次
@EnableAutoConfiguration  // 根据加入的依赖判断项目的类型，比如当看到依赖spring-boot-starter-web，就会给你加载springMVC对应的配置文件。springBoot内部定义了大量的第三方jar包配置文件，当发现你调用了某个第三方依赖，并且内部也有该第三方默认的配置文件的话就会自动加载该配置文件，比如springmvc等
@ComponentScan(excludeFilters = {
   @Filter(type = FilterType.CUSTOM, classes = TypeExcludeFilter.class),
   @Filter(type = FilterType.CUSTOM, classes = AutoConfigurationExcludeFilter.class) }) // 相当于xml文件中的包扫描<context:compenent-scan class="" />，可以设置basePackageClasses=""或basePackages="",来指定扫描特定包，@Filter指定哪些类型不适合进行组件扫描(这个例子就是这种类型)。如果什么都没写，就会扫描@ComponentScan注解所在包及其子包，在开发中常用的是什么都不写的形式，所以一般将这个注解放在有类的基础包里面

```

### Controller类

```java

@RestController
public class MyController {

   @RequestMapping("/hello")
   public String hello(){
      return "Hello111 SpringBoot11ss";
   }
}

```

## # 编写配置

**配置注解**：

1. @Configuration：声明一个类作为配置类，代替xml文件
2. @Bean：声明在方法上，将方法的返回值加入Bean容器，代替bean标签
3. @value：属性注入
4. @PropertySource：指定外部属性文件

**java类配置JDBC方式一**:

```properties

##------------- ipso -------------##

# 部署环境
jdbc.Driver=com.mysql.jdbc.Driver
jdbc.url=jdbc:mysql://www.ipso.live:3306/mybatis?characterEncoding=utf-8
jdbc.username=root
jdbc.password=gqm1975386453

# 测试环境
jdbc.DriverTest=com.mysql.jdbc.Driver
jdbc.urlTest=jdbc:mysql://localhost:3306/mybatis?characterEncoding=utf-8
jdbc.usernameTest=root
jdbc.passwordTest=gqm1975386453

```

```java

/**
 * 数据库配置类
 1. @Configuration：声明一个类作为配置类，代替xml文件
 2. @Bean：声明在方法上，将方法的返回值加入Bean容器，代替bean标签
 3. @value：属性注入
 4. @PropertySource：指定外部属性文件
 */

@PropertySource("classpath:db.properties")
@Configuration
public class JdbcConfig {

   @Value("${jdbc.urlTest}")
   private String url;

   @Value("${jdbc.usernameTest}")
   private String username;

   @Value("${jdbc.passwordTest}")
   private String password;

   @Value("${jdbc.DriverTest}")
   private String driven;

   @Bean
   public DataSource dataSource(){
      DruidDataSource dataSource = new DruidDataSource();
      dataSource.setUrl(url);
      dataSource.setDriverClassName(driven);
      dataSource.setUsername(username);
      dataSource.setPassword(password);
      return dataSource;
   }
}

```

**java类配置JDBC方式二常用**:

```properties

# 该属性名必须为application.properties

##------------- ipso -------------##

# 部署环境
#jdbc.Driver=com.mysql.jdbc.Driver
#jdbc.url=jdbc:mysql://www.ipso.live:3306/mybatis?characterEncoding=utf-8
#jdbc.username=root
#jdbc.password=gqm1975386453

# 测试环境
jdbc.driverTest=com.mysql.jdbc.Driver
jdbc.urlTest=jdbc:mysql://localhost:3306/mybatis?characterEncoding=utf-8
jdbc.usernameTest=root
jdbc.passwordTest=gqm1975386453

```

```java

/**
 * JdbcProperties.java
 * 获取属性对象，通过该对象即可获取属性文件内容
 * 调用方式： @Autowired
 *          private JdbcProperties jdbcProperties;
 * 或者public DataSource dataSource(JdbcProperties jdbcProperties){}
 */
@Component
@ConfigurationProperties(prefix = "jdbc")
@Data
public class JdbcProperties {

   private String urlTest;

   private String usernameTest;

   private String passwordTest;

   private String driverTest;

}

```

```java

/**
 * 数据库配置类
 1. @Configuration：声明一个类作为配置类，代替xml文件
 2. @Bean：声明在方法上，将方法的返回值加入Bean容器，代替bean标签
 3. @value：属性注入
 4. @PropertySource：指定外部属性文件
 */

@Configuration
@EnableConfigurationProperties(JdbcProperties.class)
public class JdbcConfig {

   /* ----------- @Value方式配置属性普通properties文件方式 ---------- */
   /* ---- 这种方式用注解@PropertySource将属性文件中的值加载过来，再通过@value获取，如下 ---- */
   /* -- 这种方式会在每个调用的地方都必须有一下代码，所以基本不使用 -- */


   /*@Value("${jdbc.urlTest}")
   private String url;

   @Value("${jdbc.usernameTest}")
   private String username;

   @Value("${jdbc.passwordTest}")
   private String password;

   @Value("${jdbc.DriverTest}")
   private String driven;*/

   /* ------------ 使用application.properties属性文件方式 ------------ */
   /* ----- 使用EnableConfigurationProperties注解将JdbcConfig对象获取到的属性加载进来 ------ */
   /* -- 通过一个对象将属性值获取，每次只要使用@EnableConfigurationProperties注解， -- */
   /* -- 然后创建加载属性的对象就可以获取到属性,非常方便 -- */

   /* 因为@EnableConfigurationProperties注解，所以也可以直接写在方法的参数上： */
   /* public DataSource dataSource(JdbcProperties jdbcProperties){} */

   @Autowired
   private JdbcProperties jdbcProperties;


   @Bean
   public DataSource dataSource(){
      DruidDataSource dataSource = new DruidDataSource();

      // 创建获取对象属性

      dataSource.setUrl(jdbcProperties.getUrlTest());
      dataSource.setDriverClassName(jdbcProperties.getDriverTest());
      dataSource.setUsername(jdbcProperties.getUsernameTest());
      dataSource.setPassword(jdbcProperties.getPasswordTest());
      return dataSource;
   }
}

```

**两种方式在控制器中的调用**：

```java

@RestController
public class MyController {

   /* 获取数据源 */
   @Autowired
   private DataSource dataSource;

   @RequestMapping("/hello")
   public String  hello(){

      JSONObject jsonObject = new JSONObject();

      jsonObject.put("ipso", dataSource.toString());
      System.out.println(dataSource);
      return  "result";
   }
}

```

## # yaml/yml文件

YAML是"YAML Ain't a Markup Language"（YAML不是一种标记语言）的递归缩写，是一个可读性高，用来表达数据序列化的格式，来个例子比较明了

```yaml

# 配置JDBC连接信息
jdbc:
  driverClassName: com.mysql.jdbc.Driver
  url: jdbc:mysql://localhost:3306/mybatis?characterEncoding=utf-8
  username: root
  password: gqm1975386453
  arrayPros: 1,2,3,4  # 数组的写法
  listPros:  # list 类型
    - value1
    - value2
    - value3
  mapPros:  # map 类型
    key1: value1
    key2: value2
    key3: value3

  listMapPros: # List<Map> 类型
    - key1-1: value1  # list的第一个元素Map1
      key1-2: value2
    - key2-1: value2-1 # list的第二个元素Map2
      key2-2: value2-2
      key2-3: value2-3

# 打印结果
# MyProperties(driverClassName=com.mysql.jdbc.Driver,
# url=jdbc:mysql://localhost:3306/mybatis?characterEncoding=utf-8,
# username=root, password=gqm1975386453, arrayPros=[1, 2, 3, 4],
# listPros=[value1, value2, value3], mapPros={key1=value1, key2=value2, key3=value3},
# listMapPros=[{key1-1=value1, key1-2=value2}, {key2-1=value2-1, key2-2=value2-2, key2-3=value2-3}]
# )

# 属性的属性也可以是一个对象，以linux中的用户关系为例：
root:
  username: ipso
  password: password
  users:
    user1:
      name: name1
      pass: password1
    user2:
      name: name2
      pass: password2

# 就是一个字符串，注意字符串可以不加双引号或单引号，但是特殊情况需要加双引号，比如：如果123是字符串的形式，则就需要使用"123"
github: https://github.com/ipsozzZ

```

**SpringBoot单元测试**：

```xml

<!-- pos.xml的dependencies中添加依赖 -->
<!-- spring Test测试 -->
<dependency>
  <groupId>org.springframework.boot</groupId>
  <artifactId>spring-boot-starter-test</artifactId>
</dependency>

```

```java

import live.ipso.Application;
import live.ipso.config.MyProperties;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)  // 固定格式
@SpringBootTest(classes = Application.class) // 指定启动类
public class MyPropertiesTest {

   // 上面已给出过MyProperties
   @Autowired
   private MyProperties myProperties;

   @Test
   public void test(){
      System.out.println(myProperties);
   }
}

```

## # 日志级别设置

日志级别有：info，debug、trace、error、warn

* 默认开启的是info: 通常会将info的信息写到日志文件中，运行日志
* debug: 调试日志
* trace: 栈追踪日志
* error: 报错日志
* warn:  警告日志

```yml

# application.yml
# 设置服务器端口号， 仅限本地测试使用
server:
  port: 8080

# 设置控制器加载时间，-1，创建完毕请求时加载前端控制器；1(大于0),在tomcat启动时加载前端控制器；默认是-1
spring:
  jersey:
    servlet:
      load-on-startup: 1

# 设置日志级别，开启debug
logging:
  level:
    live.ipso: debug
  path: "E:/test/test.log" # 配置日志文件路径

```

```java

package live.ipso.web;

import net.minidev.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.Map;

@RestController
@Slf4j // 使用注解也可以获取当前类的日志操作对象Logger(常用)，需要lombok依赖
public class MyController {

   /* 获取当前类的日志操作对象Logger，如果使用注解@Slf4j，就不用写下面这句代码了 */
   // private static final Logger log = LoggerFactory.getLogger(MyController.class);

   @RequestMapping("/hello")
   public String hello(){

      /* 日志输出(开发时输出一些想要查看的数据) */
      log.info("日志输出----info级别-----------"); // 默认开启

      // 需要开发者手动开启, 开启后info和debug都会打印，部署的时候将日志级别改为info，debug级别的日志就不会输出
      // 就不用到处去删除sout代码
      log.debug("日志输出----debug级别-----------");

      // 将数据存储到map中传递
      Map<String, Object> maptt = new HashMap<>();
      maptt.put("key", "ipso");
      maptt.put("name", "高启明");
      maptt.put("age", 33);

      // 返回json字符串类型数据
      return JSONObject.toJSONString(maptt);
   }
}

```

## # SpringBoot配置拦截器

```java

// 实现拦截器类
package live.ipso.interceptor;

import lombok.extern.slf4j.Slf4j;
import org.springframework.web.servlet.HandlerInterceptor;
import org.springframework.web.servlet.ModelAndView;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

/**
 * 拦截器器类
 */
@Slf4j
public class MyInterceptor implements HandlerInterceptor {

   /**
    * 地址映射之前(即请求到达之前)，拦截
    * @param request request
    * @param response response
    * @param handler handler
    * @return return
    * @throws Exception exception
    */
   @Override
   public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
      log.debug("MyInterceptor-------preHandle");
      return true;
   }

   /**
    * 控制器类执行结束后拦截
    * @param request request
    * @param response response
    * @param handler handler
    * @param modelAndView modelAndView
    * @throws Exception exception
    */
   @Override
   public void postHandle(HttpServletRequest request, HttpServletResponse response, Object handler, ModelAndView modelAndView) throws Exception {
      log.debug("MyInterceptor-------postHandle");
   }

   /**
    * 程序执行结束后触发改拦截器
    * @param request request
    * @param response response
    * @param handler handler
    * @param ex ex
    * @throws Exception exception
    */
   @Override
   public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) throws Exception {
      log.debug("MyInterceptor-------afterCompletion");
   }
}

```

```java

package live.ipso.config;

import live.ipso.interceptor.MyInterceptor;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration // 表明当前类为配置类，spring扫描的时候才会读取其中的配置信息
public class WebMvcConfig implements WebMvcConfigurer {

   /**
    * 用来添加拦截器的方法
    * @param registry registry
    */
   @Override
   public void addInterceptors(InterceptorRegistry registry) {

      // addInterceptor(),设置添加的拦截器类
      // addPathPatterns()，设置拦截的路径，'/**': 拦截所有请求，即所有请求都会触发该拦截器，也可以设置其它路径
      registry.addInterceptor(new MyInterceptor()).addPathPatterns("/**");
   }
}

```

## # 整合mybatis

springboot没有给出mybatis的启动器org.mybatis.spring.boot（看不是org.springframework.spring.boot吧），但是mybatis自己提供了启动器。

```xml

<!-- pom.xml -->

<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>live.ipso</groupId>
    <artifactId>03SpringBootSSMPro</artifactId>
    <version>1.0.0</version>

    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>2.1.3.RELEASE</version>
    </parent>

    <dependencies>

        <!-- springMVC -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>

        <!--jdbc-->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-jdbc</artifactId>
        </dependency>

        <!-- mysql -->
        <dependency>
            <groupId>mysql</groupId>
            <artifactId>mysql-connector-java</artifactId>
        </dependency>

        <!-- mybatis启动器 -->
        <dependency>
            <groupId>org.mybatis.spring.boot</groupId>
            <artifactId>mybatis-spring-boot-starter</artifactId>
            <version>1.3.2</version>
        </dependency>

        <!-- spring Test测试 -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-test</artifactId>
        </dependency>

        <!-- spring热部署 -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-devtools</artifactId>
        </dependency>

        <!-- 数据库工具 -->
        <dependency>
            <groupId>com.alibaba</groupId>
            <artifactId>druid</artifactId>
            <version>1.1.6</version>
        </dependency>

        <!-- json格式支持 -->
        <!--<dependency>
            <groupId>com.alibaba</groupId>
            <artifactId>fastjson</artifactId>
        </dependency>-->

        <!-- 编译时自动为属性生成构造器、getter/setter、equals、hashcode、toString方法 -->
        <dependency>
            <groupId>org.projectlombok</groupId>
            <artifactId>lombok</artifactId>
        </dependency>

        <!-- 异常 -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-configuration-processor</artifactId>
            <optional>true</optional>
        </dependency>

    </dependencies>

</project>

```

```yml

# application.yml
# 设置服务器端口号， 仅限本地测试使用
server:
  port: 8080

# 设置控制器加载时间，-1，创建完毕请求时加载前端控制器；1(大于0),在tomcat启动时加载前端控制器；默认是-1
spring:
  mvc:
    servlet:
      load-on-startup: 1
  datasource:
    driver-class-name: com.mysql.jdbc.Driver
    url: jdbc:mysql://localhost:3306/spring?useUnicode=true&characterEncoding=utf8&serverTimezone=GMT%2B8
    username: root
    password: gqm1975386453

logging:
  level:
   com.itlike: debug
  path: "D:/test/test.log"

mybatis:
  mapper-locations: classpath:mapper/*Mapper.xml
  type-aliases-package: com.itlike.pojo
  configuration:
    map-underscore-to-camel-case: true

```

```java

/*-------- 启动类 -----------*/

package live.ipso;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
@MapperScan("live.ipso.mapper") // 配置扫描mapper接口路径
public class Application {

   public static void main(String[] args) {
      SpringApplication.run(Application.class, args);
   }
}

```

```java

/*-------- Controller -----------*/

package live.ipso.web;

import live.ipso.domain.Hero;
import live.ipso.service.HeroService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@Slf4j
public class MyController {

   @Autowired
   private HeroService heroService;

   @RequestMapping("/hello")
   public String hello(){

      /* 日志输出(开发时输出一些想要查看的数据) */
      // log.info("日志输出----info级别-----------"); // 默认开启

      // 需要开发者手动开启, 开启后info和debug都会打印，部署的时候将日志级别改为info，debug级别的日志就不会输出
      // 就不用到处去删除sout代码
      // log.debug("日志输出----debug级别-----------");

      // map对象存储数据
      /*Map<String, Object> maptt = new HashMap<>();
      maptt.put("key", "ipso");
      maptt.put("name", "高启明");
      maptt.put("age", 33);*/

      log.debug("hello info log-debug");
      log.debug("hello info log-debug");
      log.debug("hello info log-debug");

      List<Hero> allHero = heroService.getAllHero();
      System.out.println("getAll-----------------------");
      System.out.println(allHero);

      Hero hero = heroService.getOneHero(10);
      System.out.println("getOne-----------------------");
      System.out.println(hero);
      // 返回json字符串类型数据 JSONObject.toJSONString(maptt)
      return "hello";
   }
}


/*-------- service接口 -----------*/
package live.ipso.service;

import live.ipso.domain.Hero;

import java.util.List;

public interface HeroService {

   public List<Hero> getAllHero();
   public Hero getOneHero(Integer id); // getById
}


/*-------- serviceImpl接口实现类 -----------*/

package live.ipso.service;

import live.ipso.domain.Hero;
import live.ipso.mapper.HeroMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class HeroServiceImpl implements HeroService {

   @Autowired
   private HeroMapper heroMapper;

   @Override
   public List<Hero> getAllHero() {
      return heroMapper.getAllHero();
   }

   @Override
   public Hero getOneHero(Integer id) {
      return heroMapper.getById(id);
   }
}


/*-------- mapper接口类或者dao接口类 -----------*/

package live.ipso.mapper;

import live.ipso.domain.Hero;
import org.apache.ibatis.annotations.Param;
import org.junit.runners.Parameterized;

import java.util.List;

public interface HeroMapper {

   public List<Hero> getAllHero();
   public Hero getById(@Param("id") Integer id);
}


/*-------- 自定义拦截器类 -----------*/
package live.ipso.interceptor;

import lombok.extern.slf4j.Slf4j;
import org.springframework.lang.Nullable;
import org.springframework.web.servlet.HandlerInterceptor;
import org.springframework.web.servlet.ModelAndView;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

/**
 * 拦截器器类
 */
@Slf4j
public class MyInterceptor implements HandlerInterceptor {

   /**
    * 地址映射之前(即请求到达之前)，拦截
    * @param request request
    * @param response response
    * @param handler handler
    * @return return
    * @throws Exception exception
    */
   @Override
   public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
      log.debug("MyInterceptor-------preHandle");
      return true;
   }

   /**
    * 控制器类执行结束后拦截
    * @param request request
    * @param response response
    * @param handler handler
    * @param modelAndView modelAndView
    * @throws Exception exception
    */
   @Override
   public void postHandle(HttpServletRequest request, HttpServletResponse response, Object handler, @Nullable ModelAndView modelAndView) throws Exception {
      log.debug("MyInterceptor-------postHandle");
   }

   /**
    * 程序执行结束后触发改拦截器
    * @param request request
    * @param response response
    * @param handler handler
    * @param ex ex
    * @throws Exception exception
    */
   @Override
   public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) throws Exception {
      log.debug("MyInterceptor-------afterCompletion");
   }
}


/*-------- 自定义拦截器类调用 -----------*/

package live.ipso.config;

import live.ipso.interceptor.MyInterceptor;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration // 表明当前类为配置类，spring扫描的时候才会读取其中的配置信息
public class WebMvcConfig implements WebMvcConfigurer {

   /**
    * 用来添加拦截器的方法
    * @param registry registry
    */
   @Override
   public void addInterceptors(InterceptorRegistry registry) {

      // addInterceptor(),设置添加的拦截器类
      // addPathPatterns()，设置拦截的路径，'/**': 拦截所有请求，即所有请求都会触发该拦截器，也可以设置其它路径
      registry.addInterceptor(new MyInterceptor()).addPathPatterns("/**");
   }
}



/*-------- 数据库映射对象或POJO -----------*/

package live.ipso.domain;

import lombok.Data;

import java.util.Date;

@Data
public class Hero {
   private Integer id;
   private String username;
   private String profession;
   private String phone;
   private String email;
   private Date onlinetime;
}

```

```xml

<!-- 映射文件*mapper.xml -->

<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd" >

<mapper namespace="live.ipso.mapper.HeroMapper">
    <select id="getAllHero" resultType="live.ipso.domain.Hero">
        select * from `hero`;
    </select>
    <select id="getById" parameterType="INTEGER" resultType="live.ipso.domain.Hero">
        select * from `hero` where id=#{id};
    </select>
</mapper>

```

### 通用mapper映射文件

实现通用的数据库操作(curd), 注意通用mapper只支持单表操作

在以上demo的基础上做一下改变

1. 引入依赖

```xml
<!-- 通用mapper启动器(它包含mybatis启动器，所以上面mybatis启动器可以注释) -->
<dependency>
  <groupId>tk.mybatis</groupId>
  <artifactId>mapper-spring-boot-starter</artifactId>
  <version>2.1.5</version>
</dependency>

```

2. 修改入口类mapper包扫描注解

```java

// application.java入口类

package live.ipso;

//import org.mybatis.spring.annotation.MapperScan; // 没有引入通用mapper依赖时使用
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import tk.mybatis.spring.annotation.MapperScan;

@SpringBootApplication
@MapperScan("live.ipso.mapper") // 配置扫描mapper接口路径
public class Application {
   public static void main(String[] args) {
      SpringApplication.run(Application.class,args);
   }
}

```

3. mapper接口继承tk.mybatis中的Mapper接口并指明数据表映射类

```java
package live.ipso.mapper;

import live.ipso.domain.Hero;
import org.apache.ibatis.annotations.Param;
import org.junit.runners.Parameterized;
import tk.mybatis.mapper.common.Mapper;

import java.util.List;

/**
 * 实现通用mapper接口
 */
public interface HeroMapper extends Mapper<Hero> {  // 继承tk.mybatis中的Mapper接口并指明数据表映射类


   public List<Hero> getAllHero();
   public Hero getById(@Param("id") Integer id);
}

```

4. 在实体类(或者叫数据库映射类，也就是domain中的类)上加上注解@Table(name="表名")

```java

package live.ipso.domain;

import lombok.Data;
import tk.mybatis.mapper.annotation.KeySql;

import javax.persistence.Id;
import javax.persistence.Table;
import java.util.Date;

@Data
@Table(name = "hero")
public class Hero {

   @Id  // 指明主键
   @KeySql(useGeneratedKeys = true) // 添加时生成主键

   private Integer id;

   private String username;

   private String profession;

   @Transient  // 查询的时候过滤该字段,有需要时使用
   private String phone;

   private String email;

   private Date onlinetime;
}

```
