# 整合spring、springMVC与Mybatis

## # 前言

**把自己曾经入门时学习SSM整合的笔记整理出来希望能帮到别人，同时也方便自己复习知识点，小小入门java开发程序员，错误在所难免，希望路过大神指正**。

## # 整合SSM web层

包含相当于MVC中的C和V(jsp)

### 创建web工程

### 引入jar包与配置文件

```md

# spring项目基础jar包

com.springsource.org.aopalliance-1.0.0.jar
com.springsource.org.apache.commons.logging-1.1.1.jar
com.springsource.org.apache.log4j-1.2.15.jar
com.springsource.org.aspectj.weaver-1.6.8.RELEASE.jar
spring-aop-5.0.7.RELEASE.jar
spring-aspects-5.0.7.RELEASE.jar
spring-beans-5.0.7.RELEASE.jar
spring-context-5.0.7.RELEASE.jar
spring-core-5.0.7.RELEASE.jar
spring-expression-5.0.7.RELEASE.jar
spring-jdbc-5.0.7.RELEASE.jar
spring-orm-5.0.7.RELEASE.jar
spring-test-5.0.7.RELEASE.jar
spring-tx-5.0.7.RELEASE.jar
spring-web-5.0.7.RELEASE.jar

```

```xml

<!-- applicationContext.xml -->

<?xml version="1.0" encoding="UTF-8"?>
<beans xmlns="http://www.springframework.org/schema/beans"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       xmlns:context="http://www.springframework.org/schema/context"
       xmlns:aop="http://www.springframework.org/schema/aop"
       xmlns:tx="http://www.springframework.org/schema/tx"
       xsi:schemaLocation="http://www.springframework.org/schema/beans
        http://www.springframework.org/schema/beans/spring-beans.xsd
        http://www.springframework.org/schema/context
        http://www.springframework.org/schema/context/spring-context.xsd
        http://www.springframework.org/schema/aop
        http://www.springframework.org/schema/aop/spring-aop.xsd
        http://www.springframework.org/schema/tx
        http://www.springframework.org/schema/tx/spring-tx.xsd">

    <!--注解扫描-->
    <context:component-scan base-package="live.ipso"/>

</beans>

```

### 在web.xml当中配置spring监听器

```xml

<?xml version="1.0" encoding="UTF-8"?>
<web-app xmlns="http://xmlns.jcp.org/xml/ns/javaee"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://xmlns.jcp.org/xml/ns/javaee http://xmlns.jcp.org/xml/ns/javaee/web-app_4_0.xsd"
         version="4.0">

    <!-- spring的核心监听器 -->
    <listener>
        <listener-class>org.springframework.web.context.ContextCleanupListener</listener-class>
    </listener>

    <!-- 加载spring的配置文件，默认加载的是/WEB_INFO/applicationContext.xml -->
    <context-param>
        <param-name>contextConfigLocation</param-name>
        <param-value>classpath:applicationContext.xml</param-value>
    </context-param>
</web-app>

```

### 添加springMVC相关jar包与配置文件

```md

spring-webmvc-5.0.7.RELEASE.jar

```

**配置文件springmvc.xml**：

```xml

<?xml version="1.0" encoding="UTF-8"?>
<beans xmlns="http://www.springframework.org/schema/beans"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       xmlns:context="http://www.springframework.org/schema/context"
       xmlns:mvc="http://www.springframework.org/schema/mvc"
       xmlns:aop="http://www.springframework.org/schema/aop"
       xmlns:tx="http://www.springframework.org/schema/tx"
       xsi:schemaLocation="http://www.springframework.org/schema/beans
       http://www.springframework.org/schema/beans/spring-beans.xsd
       http://www.springframework.org/schema/context
       http://www.springframework.org/schema/context/spring-context.xsd
       http://www.springframework.org/schema/mvc
       http://www.springframework.org/schema/mvc/spring-mvc.xsd
       http://www.springframework.org/schema/aop
       http://www.springframework.org/schema/aop/spring-aop.xsd
       http://www.springframework.org/schema/tx
       http://www.springframework.org/schema/tx/spring-tx.xsd">

    <!-- 扫描注解 -->
    <context:component-scan base-package="live.ipso"/>

    <!-- 视图解析器的前后缀配置 -->
    <bean class="org.springframework.web.servlet.view.InternalResourceViewResolver">
        <property name="prefix" value="/page/" />
        <property name="suffix" value=".jsp" />
    </bean>

    <!-- 不经过Controller,由jsp直接跳转jsp -->
    <mvc:view-controller path="toRequest" view-name="request" />

    <!-- 开放静态资源的访问，判断访问是否是静态资源，是就放行，不是就去@RequestMapping中匹配 -->
    <!-- 方法一 -->
    <!-- <mvc:default-servlet-handler /> -->

    <!-- 方法二 -->
    <!--<mvc:resources location="/images/" mapping="/images/**" />
    <mvc:resources location="/js/" mapping="/js/**" />-->
    <!--<mvc:resources mapping="/css/" location="/css/**" />-->

    <!-- 重新注册3个Bean（一般都会加上） -->
    <mvc:annotation-driven />

</beans>

```

### 在web.xml中配置springMVC前端控制器和编码

```xml

<?xml version="1.0" encoding="UTF-8"?>
<web-app xmlns="http://xmlns.jcp.org/xml/ns/javaee"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://xmlns.jcp.org/xml/ns/javaee http://xmlns.jcp.org/xml/ns/javaee/web-app_4_0.xsd"
         version="4.0">

    <!-- 配置SpringMVC前端控制器 -->
    <servlet>
        <servlet-name>MySpringMVC</servlet-name>
        <servlet-class>org.springframework.web.servlet.DispatcherServlet</servlet-class>

        <init-param>
            <param-name>contextConfigLocation</param-name>
            <param-value>classpath:springmvc.xml</param-value>
        </init-param>

        <load-on-startup>1</load-on-startup>
    </servlet>
    <servlet-mapping>
        <servlet-name>MySpringMVC</servlet-name>
        <url-pattern>/</url-pattern>
    </servlet-mapping>

    <!-- spring的核心监听器 -->
    <listener>
        <listener-class>org.springframework.web.context.ContextCleanupListener</listener-class>
    </listener>

    <!-- 加载spring的配置文件，默认加载的是/WEB_INFO/applicationContext.xml -->
    <context-param>
        <param-name>contextConfigLocation</param-name>
        <param-value>classpath:applicationContext.xml</param-value>
    </context-param>
</web-app>

```

### 测试Spring MVC

**index.jsp**:

```jsp

<%@ page contentType="text/html;charset=UTF-8" language="java" %>
<html>
  <head>
    <title>TEST-SpringMVC</title>
  </head>
  <body>
  <form action="${pageContext.request.contextPath}/testSpringMvc" method="post">
    用户名：<input type="text" name="name"><br>
    年龄：<input type="text" name="age"> <br>
    <input type="submit" value="提交">
  </form>
  </body>
</html>

```

## # sevice层整合

业务层介于MVC的M层与C层之间，做数据逻辑处理

### 创建sevice包

### 在sevice包中创建Sevice接口

```java

// CustomerService接口
package live.ipso.service;

import live.ipso.domain.Customer;

public interface CustomerService {
   /* 数据存储业务处理 */
   public void save(Customer customer);
}

// service接口实现类
package live.ipso.service;

import live.ipso.domain.Customer;
import org.springframework.stereotype.Service;

@Service("customerService") // 业务层，并将该业务层取名为CustomerService,调用时使用@Autowired可以不用取名
public class CustomerServiceImpl implements CustomerService {
   @Override
   public void save(Customer customer) {
      System.out.println("来到业务层， 保存Customer:" + customer);
   }
}

// 在控制器中调用service对象
@Controller
public class CustomerController {

   /* 注入业务层 */
   // @Autowired是用在JavaBean中的注解，通过byType形式，用来给指定的属性或方法注入所需的外部资源
   // 通过byType形式，如果容器中存在一个与指定属性类型相同的bean,那么指定属性将于该类型自动匹配，如果有多个类型相同的bean则抛出异常，表示不能使用byType形式；
   // 通过byName形式, 根据属性名自动装配，根据属性名查找与属性名一致的bean
   @Autowired
   private CustomerService customerService;

   @RequestMapping("addCustomer")
   public String addUser(Customer customer){

      /* 接收数据 */
      System.out.println(customer);

      /* 调用业务层将数据保存到数据库中 */
      customerService.save(customer);

      return "result";
   }

}

```

## # 整合SSM的Mybatis框架 dao层

相当于MVC中的M层，对数据库进行CURD操作

### 引入jar包

```md

ant-1.9.6.jar
ant-launcher-1.9.6.jar
asm-5.2.jar
cglib-3.2.5.jar
commons-logging-1.2.jar
druid-1.0.15.jar
javassist-3.22.0-GA.jar
lombok.jar
mybatis-3.4.6.jar
mybatis-spring-1.3.2.jar
mysql-connector-java-5.1.7-bin.jar
ognl-3.1.16.jar
slf4j-api-1.7.25.jar
slf4j-log4j12-1.7.25.jar

mybatis-spring-1.3.2.jar 为整合SSM时的特有包，在独立使用Mybatis开发时不需要此jar包

```

### 添加配置文件SqlMappingConfig.xml数据库全局配置文件

```xml

<!-- 注意environments数据库连接信息已经不再写在这个地方 -->

<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE configuration
        PUBLIC "-//mybatis.org//DTD config 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-config.dtd">
<configuration>

    <!-- 配置sql打印日志信息 -->
    <settings>
        <setting name="logImpl" value="STDOUT_LOGGING"/>
    </settings>

    <!-- 别名配置，在xml文件中使用live.ipso.mybatis.domain.Customer时可以简化成Customer -->
    <typeAliases>

        <!-- 在包live.ipso.mybatis.domain下的所有对象都可以直接使用其类名调用(即自动使用类名作为别名) -->
        <package name="live.ipso.domain"/>
    </typeAliases>

</configuration>

```

### 数据库配置信息属性文件db.properties

```properties

# 数据库配置信息
##------------- ipso -------------##

# 部署环境
jdbc.Driver=com.mysql.jdbc.Driver
jdbc.url=jdbc:mysql://myHost:3306/mybatis?characterEncoding=utf-8
jdbc.username=root
jdbc.password=mypassword

# 测试环境
jdbc.DriverTest=com.mysql.jdbc.Driver
jdbc.urlTest=jdbc:mysql://localhost:3306/mybatis?characterEncoding=utf-8
jdbc.usernameTest=root
jdbc.passwordTest=mypassword

```

### 创建mapper包

**在mapper包中创建对应映射对象mapper接口及映射xml文件

```java

package live.ipso.mapper;

import live.ipso.domain.Customer;

public interface CustomerMapper {
   public void insertOne(Customer customer);
}


```

```xml

<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<!-- 数据库映射文件 -->
<!-- 使用传统的实现dao层, namespace可以随意赋值 -->
<!-- 使用Mybatis动态代理实现dao层，namespace的值必须是和Mapper接口类路径一致 -->
<mapper namespace="live.ipso.mapper.CustomerMapper">

  <insert id="insertOne">
    insert into `customer`(cust_name, cust_profession, cust_phone, email)
    values (
    #{cust_name}, #{cust_profession}, #{cust_phone}, #{email}
    )
  </insert>
</mapper>

```

### 在applicationContext.xml中整合Mybatis

```xml

<!-- 整合后的applicationContext.xml文件 -->

<?xml version="1.0" encoding="UTF-8"?>
<beans xmlns="http://www.springframework.org/schema/beans"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       xmlns:context="http://www.springframework.org/schema/context"
       xmlns:aop="http://www.springframework.org/schema/aop"
       xmlns:tx="http://www.springframework.org/schema/tx"
       xsi:schemaLocation="http://www.springframework.org/schema/beans
        http://www.springframework.org/schema/beans/spring-beans.xsd
        http://www.springframework.org/schema/context
        http://www.springframework.org/schema/context/spring-context.xsd
    http://www.springframework.org/schema/aop
    http://www.springframework.org/schema/aop/spring-aop.xsd
     http://www.springframework.org/schema/tx
        http://www.springframework.org/schema/tx/spring-tx.xsd">
    <!--注解扫描-->
    <context:component-scan base-package="live.ipso"/>

    <!-- Spring与Mybatis的整合 -->

    <!-- 加载属性文件 -->
    <context:property-placeholder location="classpath:db.properties" />

    <!-- 数据库连接池 -->
    <bean id="dataSource" class="com.alibaba.druid.pool.DruidDataSource">
        <!-- 注意属性文件中的名称不能和name一样 -->
        <property name="driverClassName" value="${jdbc.DriverTest}"/>
        <property name="url" value="${jdbc.urlTest}" />
        <property name="username" value="${jdbc.usernameTest}" />
        <property name="password" value="${jdbc.passwordTest}" />
    </bean>

    <!-- 配置事务管理器 -->
    <bean id="transactionManager"
          class="org.springframework.jdbc.datasource.DataSourceTransactionManager">

        <!-- 配置数据源 -->
        <property name="dataSource" ref="dataSource" />
    </bean>

    <!-- 开启注解管理事务 -->
    <tx:annotation-driven transaction-manager="transactionManager" />

    <!--******     开始整合Mybatis   ******-->

    <!-- Mybatis工厂 -->
    <bean id="SqlSessionFactoryBean" class="org.mybatis.spring.SqlSessionFactoryBean">
        <property name="dataSource" ref="dataSource" />
        <!-- 核心配置文件的位置 -->
        <property name="configLocation" value="classpath:SqlMappingConfig.xml"/>
        <!-- 配置映射文件的路径 -->
        <property name="mapperLocations" value="classpath:live/ipso/mapper/*.xml" />
    </bean>

    <!-- 配置Mapper扫描 -->
    <bean class="org.mybatis.spring.mapper.MapperScannerConfigurer">
        <!-- 配置Mapper扫描包 -->
        <property name="basePackage" value="live.ipso.mapper" />
    </bean>

    <!--******     结束整合Mybatis   ******-->

</beans>

```

### 在业务层service层中调用dao层

```java

package live.ipso.service;

import live.ipso.domain.Customer;
import live.ipso.mapper.CustomerMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service("CustomerService") // 业务层，并将该业务层取名为CustomerService
@Transactional // 自动提交事务
public class CustomerServiceImpl implements CustomerService {

   @Autowired
   // 不需要再向独立使用Mybatis时一样通过openSession来获取SqlSession会话，
   // 再来获取SqlSession.getMapper()了
   private CustomerMapper customerMapper;

   @Override
   public void save(Customer customer) {
      System.out.println("来到业务层， 保存Customer:" + customer);

      /* 调用dao层 */
      customerMapper.insertOne(customer);
   }
}

```

## # 整合完毕

至此一个简单的SSM架构项目就运转起来了，项目已经可以运行起来并将数据写入到数据库，其它更复杂的CURD可以自己慢慢在此基础上扩展
