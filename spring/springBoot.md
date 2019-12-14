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

## SpringBoot + Thymeleaf

在SpringMVC中我们通常使用JSP渲染后台数据，展示在浏览器中。而在SpringBoot我们通常使用Thymeeaf渲染数据在浏览器中展示。

**介绍Thymeleaf**:

* SpringBoot并不推荐使用jsp
* Thymeleaf 是一个跟 Velocity、FreeMarker 类似的模板引擎，它可以完全替代 JSP

**特点**：

* 动静结合
  1. Thymeleaf 在有网络和无网络的环境下皆可运行
  2. 它可以让美工在浏览器查看页面的静态效果，也可以让程序员在服务器查看带数据的动态页面效果
  3. 这是由于它支持 html 原型，然后在 html 标签里增加额外的属性来达到模板+数据的展示方式
  4. 浏览器解释 html 时会忽略未定义的标签属性，所以 thymeleaf 的模板可以静态地运行；
  5. 当有数据返回到页面时，Thymeleaf 标签会动态地替换掉静态内容，使页面动态显示。

* 开箱即用
  它提供标准和spring标准两种方言，可以直接套用模板实现JSTL、 OGNL表达式效果，避免每天套模板、该jstl、改标签的困扰。同时开发人员也可扩展和创建自定义的方言。

* 多方言支持
  Thymeleaf 提供spring标准方言和一个与 SpringMVC 完美集成的可选模块，可以快速的实现表单绑定、属性编辑器、国际化等功能。

* 与SpringBoot完美整合
  与SpringBoot完美整合，SpringBoot提供了Thymeleaf的默认配置，并且为Thymeleaf设置了视图解析器，我们可以像以前操作jsp一样来操作Thymeleaf。

**添加启动器**：

```xml

  <dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-thymeleaf</artifactId>
  </dependency>

```

**创建模板文件夹**:

SpringBoot会自动为Thymeleaf注册一个视图解析器ThymeleafViewResolver还配置了模板文件（html）的位置，与jsp类似的前缀+ 视图名 + 后缀风格，与解析JSP的InternalViewResolver类似，Thymeleaf也会根据前缀和后缀来确定模板文件的位置,在配置文件中 配置缓存,编码。pom.xml文件如下：

```xml

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
        <!--<dependency>
            <groupId>org.mybatis.spring.boot</groupId>
            <artifactId>mybatis-spring-boot-starter</artifactId>
            <version>1.3.2</version>
        </dependency>-->

        <!-- 通用mapper启动器(它包含mybatis启动器，所以上面mybatis启动器可以注释) -->
        <dependency>
            <groupId>tk.mybatis</groupId>
            <artifactId>mapper-spring-boot-starter</artifactId>
            <version>2.1.5</version>
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

        <!-- thymeleaf模板引擎 -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-thymeleaf</artifactId>
        </dependency>

    </dependencies>

</project>

```

### 基本使用

引入名称空间
	<html lang="en" xmlns:th="http://www.thymeleaf.org">
表达式
	${}:变量表达式
	*{} ：选择变量表达式
	#{...} : Message 表达式

**URL**：

* 绝对网址：
  绝对URL用于创建到其他服务器的链接,它们需要指定一个协议名称("http"或"https")开头如：```<a th:href="@{https://www.itlike.com/}">```
* 上下文相关URL
  与Web应用程序根相关联URL如：```<a th:href="@{/hello}">跳转</a>```
* 与服务器相关URL
  服务器相关的URL与上下文相关的URL非常相似, 如：```<a th:href="@{~/hello}">跳转</a>```
* 携带参数
  如：```<a th:href="@{/hero/detail(id=3,action='show_all')}">aa</a>```

**字面值**:

  有的时候，我们需要在指令中填写基本类型如：字符串、数值、布尔等，并不希望被Thymeleaf解析为变量，这个时候称为字面值。

* 字符串字面值

 如：```<h3 th:text="我是String字串"></h3>```

* 数字字面值

 如：```<h3 th:text="2 + 1"></h3>```

* 布尔字面值
   布尔类型的字面值是true或false

**拼接**：

* 普通字符串与表达式拼接的情况
如：```<h1 th:text="'我是' + ${name}">Hello ipso</h1>```

* 字符串字面值需要用''，拼接起来非常麻烦，Thymeleaf对此进行了简化，使用一对|即可
如：```<h1 th:text="|欢迎您:${name}|">Hello ipso</h1>```

**运算符**:

* 算术操作符 + - * / %
* 比较运算 >, <, >= and <=, 但是>, <不能直接使用，因为xml会解析为标签
所以：

1. gt表示大于
2. lt表示小于
3. ge表示大于等于
4. le表示小于等于

* 三元运算
  conditon ? then : else

**内联写法**：

在标签内部写如：```<p>内联写法：[[ ${name} ]]</p>```

**局部变量**：

如：

```html

<div th:with="first=${heros[0]}">
    <p>
      <span th:text="${first.username}"></span>
    </p>
</div>

```

**判断**:

  th:if

  th:unless

  th:switch

**迭代**:

```html

<table>
    <tr th:each="item:${hero}">
        <td th:text="${item.username}"></td>
    </tr>
</table>

<!-- stat对象 -->
<table>
    <tr th:each="item,stat:${hero}">
        <td th:text="${item.username}"></td>
        <td th:text="${item.index}"></td>
    </tr>
</table>

```

* stat对象包含以下属性
   1. index，从0开始的角标
   2. count，元素的个数，从1开始
   3. size，总元素个数
   4. current，当前遍历到的元素
   5. even/odd，返回是否为奇偶，boolean值
   6. first/last，返回是否为第一或最后，boolean值

**内置对象**：

* 环境相关对象
   1. ${#ctx} 上下文对象，可用于获取其它内置对象。
   2. ${#vars}:    上下文变量。
   3. ${#locale}：上下文区域设置。
   4. ${#request}: HttpServletRequest对象。
   5. ${#response}: HttpServletResponse对象。
   6. ${#session}: HttpSession对象。
   7. ${#servletContext}:  ServletContext对象。

* 全局对象功能

   1. #strings：字符串工具类
   2. #lists：List 工具类
   3. #arrays：数组工具类
   4. #sets：Set 工具类
   5. #maps：常用Map方法。
   6. #objects：一般对象类，通常用来判断非空
   7. #bools：常用的布尔方法。
   8. #execInfo：获取页面模板的处理信息。
   9. #messages：在变量表达式中获取外部消息的方法，与使用＃{...}语法获取的方法相同。
   10. #uris：转义部分URL / URI的方法。
   11. #conversions：用于执行已配置的转换服务的方法。
   12. #dates：时间操作和时间格式化等。
   13. #calendars：用于更复杂时间的格式化。
   14. #numbers：格式化数字对象的方法。
   15. #aggregates：在数组或集合上创建聚合的方法。
   16. #ids：处理可能重复的id属性的方法。

* 示例

```${#strings.abbreviate(str,10)}```  str截取0-10位，后面的全部用…这个点代替，注意，最小是3位
```${#strings.toUpperCase(name)}```

* 示例1

1. 判断是不是为空:null:

```html

<span th:if="${name} != null">不为空</span>

<span th:if="${name1} == null">为空</span>

```

1. 判断是不是为空字符串: “”

```html

<span th:if="${#strings.isEmpty(name1)}">空的</span>

```

1. 判断是否相同：

```html

<span th:if="${name} eq 'jack'">相同于jack,</span> 

<span th:if="${name} eq 'ywj'">相同于ywj,</span> 

<span th:if="${name} ne 'jack'">不相同于jack,</span>

```

1. 不存在设置默认值：

```html

<span th:text="${name2} ?: '默认值'"></span> 

```

1. 是否包含(分大小写): 

```html

<span th:if="${#strings.contains(name,'ez')}">包ez</span>

<span th:if="${#strings.contains(name,'y')}">包j</span>

```

1. 是否包含（不分大小写）:

```html

<span th:if="${#strings.containsIgnoreCase(name,'y')}">包</span>

```

1. 其它

```html

${#strings.startsWith(name,'o')}

${#strings.endsWith(name, 'o')}

${#strings.indexOf(name,frag)}// 下标

${#strings.substring(name,3,5)}// 截取

${#strings.substringAfter(name,prefix)}// 从 prefix之后的一位开始截取到最后,比如 (ywj,y) = wj, 如果是(abccdefg,c) = cdefg//里面有2个c,取的是第一个c

${#strings.substringBefore(name,suffix)}// 同上，不过是往前截取

${#strings.replace(name,'las','ler')}// 替换

${#strings.prepend(str,prefix)}// 拼字字符串在str前面

${#strings.append(str,suffix)}// 和上面相反，接在后面

${#strings.toUpperCase(name)}

${#strings.toLowerCase(name)}

${#strings.trim(str)}

${#strings.length(str)}

${#strings.abbreviate(str,10)}//  str截取0-10位，后面的全部用…这个点代替，注意，最小是3位

```

**布局**：

* 方式1

```html

<nav th:fragment="header">
  <h1>导航头部</h1>
</nav>

<div th:include="common/base::header"></div>

```

* 方式2

```html

<footer id="footer">
  <h1>尾部</h1>
</footer>

<div th:insert="`{common/base::#footer}"></div>

```

* 引入方式
   1. th:insert
   将公共的标签及内容插入到指定标签当中
   2. th:replace
   将公共的标签替换指定的标签
   3. th:include
   将公共标签的内容包含到指定标签当中

* 传值

```html

<nav th:fragment="header">
  <h1>导航头部</h1>
</nav>

<div th:include="common/base::header(active='header')"></div>

<div th:include="common/base::#footer(active='footer')"></div>

```

```html

<nav th:fragment="header">
  <h1 th:class="${active=='header'}?'active':''">导航头部</h1>
</nav>

```

**js模板**:

* 模板引擎不仅可以渲染html，也可以对JS中的进行预处理。而且为了在纯静态环境下可以运行
* 在script标签中通过th:inline="javascript"来声明这是要特殊处理的js脚本

## 更改数据库连接池

Mybatis默认的数据库连接池是HikariCP，相对于HikariCP，Druid既可以作为数据库连接池，还可以做数据监控

**引入依赖**：

```xml

<!-- 数据库连接池,druid不仅做数据库连接池，还做一些数据的监控工作等 -->
<!-- JDBC默认用的是HikariCP -->
<dependency>
    <groupId>com.alibaba</groupId>
    <artifactId>druid</artifactId>
    <version>1.1.6</version>
</dependency>

<dependency>
    <groupId>log4j</groupId>
    <artifactId>log4j</artifactId>
    <version>1.2.17</version>
</dependency>

```

**appication.yml**:

```yml

server:
  port: 8080

spring:
  mvc:
    servlet:
      load-on-startup: 1
  datasource:
    driver-class-name: com.mysql.jdbc.Driver
    url: jdbc:mysql://localhost:3306/spring?useUnicode=true&characterEncoding=utf8&serverTimezone=GMT%2B8
    username: root
    password: gqm1975386453
    type: com.alibaba.druid.pool.DruidDataSource  # 使用Druid
    initialSize: 5
    minIdle: 5
    maxActive: 20
    maxWait: 60000
    timeBetweenEvictionRunsMillis: 60000
    minEvictableIdleTimeMillis: 300000
    validationQuery: SELECT 1 FROM DUAL
    testWhileIdle: true
    testOnBorrow: false
    testOnReturn: false
    poolPreparedStatements: true
    #   配置监控统计拦截的filters，去掉后监控界面sql无法统计，'wall'用于防火墙
    filters: stat,wall,log4j
    maxPoolPreparedStatementPerConnectionSize: 20
    useGlobalDataSourceStat: true
    connectionProperties: druid.stat.mergeSql=true;druid.stat.slowSqlMillis=500

  # thymeleaf 配置
  thymeleaf:
    cache: false   # 关掉thymeleaf缓存，否则有时候刷新可能会失败
    mode: HTML5
    encoding: UTF-8

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

**添加Druid配置类**：对数据库性能要求不太高，使用，对数据的监控很适用

```java

package live.ipso.config;

import com.alibaba.druid.pool.DruidDataSource;
import com.alibaba.druid.support.http.StatViewServlet;
import com.alibaba.druid.support.http.WebStatFilter;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.boot.web.servlet.ServletRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import javax.sql.DataSource;
import java.io.SerializablePermission;
import java.util.Arrays;
import java.util.HashMap;

/**
 * 配置使用druid数据库连接池，不是用默认的HikariCP
 */
@Configuration
public class DruidConfig {

   /**
    * 数据库连接池
    * @return
    */
   @ConfigurationProperties(prefix = "spring.datasource") // 配置前缀
   @Bean
   public DataSource druid(){
      return new DruidDataSource();
   }

   /**
    * 配置servlet
    * 当输入‘/druid/*’地址就会进入druid后台输入密码后可进行视图形式的数据监控
    * @return
    */
   @Bean
   public ServletRegistrationBean statViewServlet(){
      ServletRegistrationBean bean = new ServletRegistrationBean(new StatViewServlet(), "/druid/*");
      HashMap<Object, Object> hashMap = new HashMap<>();
      hashMap.put("loginUsername","admin");
      hashMap.put("loginPassword", "123456");
      hashMap.put("allow", ""); // 允许访问所有
      bean.setInitParameters(hashMap);
      return bean;
   }

   /**
    * 过滤器，设置静态资源及/druid/*路径不用拦截
    * @return
    */
   @Bean
   public FilterRegistrationBean webStatFilter(){
      FilterRegistrationBean bean = new FilterRegistrationBean(new WebStatFilter());
      HashMap<Object, Object> hashMap = new HashMap<>();
      hashMap.put("exclusions", "*.js,*.css,/druid/*");
      bean.setInitParameters(hashMap); // 设置不用过滤*.js,*.css,/druid/*
      bean.setUrlPatterns(Arrays.asList("/*")); // 所有请求都做过滤，除了上面的*.js,*.css,/druid/*
      return bean;
   }

}


```

## # 集成Swagger2

**Swagger2简介**:
  1.随项目自动生成强大RESTful API文档，减少工作量
  2.API文档与代码整合在一起，便于同步更新API说明
  3.页面测试功能来调试每个RESTful API

**添加依赖**：

```xml

<dependency>
    <groupId>io.springfox</groupId>
    <artifactId>springfox-swagger2</artifactId>
    <version>2.2.2</version>
</dependency>
<dependency>
    <groupId>io.springfox</groupId>
    <artifactId>springfox-swagger-ui</artifactId>
    <version>2.2.2</version>
</dependency>

```

**创建swagger2配置类**：

```java

@Configuration

@EnableSwagger2

public class SwaggerConfig {

        @Bean

        public Docket createRestApi() {

            return new Docket(DocumentationType.SWAGGER_2)

                    .apiInfo(apiInfo())

                    .select()

                    .apis(RequestHandlerSelectors.basePackage("com.itlike"))// 指定扫描包下面的注解

                    .paths(PathSelectors.any())

                    .build();

        }

        // 创建api的基本信息

        private ApiInfo apiInfo() {

            return new ApiInfoBuilder()

                    .title("集成Swagger2构建RESTful APIs")

                    .description("集成Swagger2构建RESTful APIs")

                    .termsOfServiceUrl("https://www.baidu.com")

                    .contact("itlike")

                    .version("1.0.0")

                    .build();

        }

}

```

**在控制器方法上添加对应api信息**：

```java

package live.ipso.web;

import live.ipso.domain.Hero;
import live.ipso.service.HeroService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.RequestMapping;

import java.util.List;

@Controller
@Slf4j
@RequestMapping("hero")
public class MyController {

   @Autowired
   private HeroService heroService;

   @ApiIgnore // 不让当前方法展示在文档(生产环境时有的方法不便暴露，可以用这个注解隐藏)
   @ApiOperation(value="获取英雄信息", notes="根据id来获取英雄详细信息")
   @ApiImplicitParam(name="id", value="用户id", requied=true, dataType="String")
   @RequestMapping("{id}")
   @ResponseBody
   public Hero hello(Model model){

      /* 查询所有 */
      List<Hero> allHero = heroService.getAllHero();
      System.out.println("getAll-----------------------");
      System.out.println(allHero);

      /* 根据id查询 */
      Hero hero = heroService.getOneHero(10);
      System.out.println("getOne-----------------------");
      System.out.println(hero);

      // 返回json字符串类型数据 JSONObject.toJSONString(maptt)
      return hero;
   }
}

```

