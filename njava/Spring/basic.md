# spring基础知识

>所有对象在spring中都叫bean.

## # java中的反射

1. 概述

JAVA反射机制是在运行状态中，对于任意一个类，都能够知道这个类的所有属性和方法；对于任意一个对象，都能够调用它的任意一个方法和属性；这种动态获取的信息以及动态调用对象的方法的功能称为java语言的反射机制。
要想解剖一个类,必须先要获取到该类的字节码文件对象。而解剖使用的就是Class类中的方法.所以先要获取到每一个字节码文件对应的Class类型的对象.

2. 获取class对象的三种方法

* Object --> getClass() (因为所有类的继承Object类所以可以用该方法获取)
* 任何数据类型(包括基本数据类型)都有一个“静态”的class属性
* 通过Class类的静态方法：forName(String className) （常用）

## # XML

>可扩展标记语言,XML 是基于文本的标记语言。主要用于数据的存储、数据传递、配置文件。而HTML(超文本标记语言)主要用于展示数据。xml没有像html一样有预定义标签，所以xml只要遵循xml规定的语法格式，标签名完全可以根据实际需要命名，规则就是：不能使用xml或xml(XML)开头、数字开头、特殊字符开头的标签名即可。

### 5大预定义实体应用

1. "(\&lt;)小于符号"
2. "(\&gt;)小于符号"
3. "(\&amp;)&符号"
4. "(\&apos;)单引号"
5. "(\&quot;)双引号"

### 基本知识

* 命名空间。避免元素命名冲突。命名空间语法：
命名空间声明的语法如下。xmlns:前缀="URI"。使用在某个根标签的开始标签中。'前缀名才是命名空间名'。用法是：（<前缀:开始标签><\/结束标签>）

URI:统一资源标识符,是一串可以标识因特网资源的字符.最常用的 URI 是用来标识因特网域名地址的统一资源定位器（URL）。另一个不那么常用的 URI 是统一资源命名（URN）。

命名空间中的URI我们一般使用URL。它是不被解析器用于查找信息的。其目的是赋予命名空间一个惟一的名称。不过，很多公司常常会作为指针来使用命名空间指向实际存在的网页，这个网页包含关于命名空间的信息。

* 元素：
从开始标签(包括结束标签)到结束标签的所有内容(包括其它的元素、文本、属性、以及以上的混合)

* 属性：提供有关元素的额外信息。（固定一直不会变得信息可以作为属性，否则建议使用元素不建议使用属性）

"形式良好"的 XML 文档拥有正确的语法规则：
XML 文档必须有一个根元素；
XML元素都必须有一个关闭标签；
XML 标签对大小写敏感；
XML 元素必须被正确的嵌套；
XML 属性值必须加引号；

* XML样式表语言：XSLT

* XML CDATA
XML 文档中的所有文本均会被解析器解析。只有 CDATA 区段中的文本会被解析器忽略。
格式：
```
<![CDATA[
	// 这里的代码不会被xml解析器解析
]]>
```

## # ioc ##

### 什么是ioc

>Ioc相当于一个容器，统一管理spring中所有的bean的权限

### IoC容器实例化Bean(程序演示项目：spring项目的inBean模块)

1. 创建一个xml配置文件
```
<!-- xml配置文件的同一头部代码 -->
<beans xmlns="http://www.springframework.org/schema/beans"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       xsi:schemaLocation="http://www.springframework.org/schema/beans http://www.springframework.org/schema/beans/spring-beans.xsd">

    <!--  IoC容器实例化Bean(类)方法  -->

    <!-- 通过构造方法实例化bean -->
    <bean id="bean1" class="getbean.Bean1"></bean>

    <!-- 工厂模式静态使用IoC容器获取实例方法 -->
    <bean id="bean2" class="getbean.Bean2Factory" factory-method="getBean2"></bean>

    <!-- 工厂模式非静态使用IoC容器获取实例方法 -->
    <bean id="bean3Factory" class="getbean.Bean3Factory"></bean>
    <bean id="bean3" class="getbean.Bean3" factory-bean="bean3Factory" factory-method="getBean3"></bean>

    <!-- 设置别名：一：可以通过name属性来设置,可以设置多个; 二可以通过标签<alias>来设置,一个标签只能设置一个如下 -->
    <!-- <bean id="bean1" class="getbean.Bean1" name="bean1_2, bean1_3"></bean> -->
    <!-- <alias name="bean2" alias="bean2_1"></alias> -->
</beans>

// 创建spring上下文，这里省略类的其余部分

// 使用ClassPathXmlApplicationContext读取xml配置文件
ApplicationContext context = new ClassPathXmlApplicationContext("spring.xml");

// 获取Bean
Bean bean = context.getBean("(这里是唯一标识)", Bean.class)  （Bean为具体的类名；Bean.class指定具体类的类型）
```

### IoC容器注入Bean方式(程序演示项目：spring项目的inBean模块)

1. 通过构造方法注入Bean
2. 通过set方法注入Bean
3. 集合类型Bean的注入

* List
* Set
* Map
* Properties

1. null值注入

2. 注入时创建内部Bean

实例
```
具体过程阅读inBean项目实例体会
```

### Bean作用域(程序演示项目：spring项目的scope模块，改变scope属性查看测试结果)

1. Singleton作用域
2. Prototype作用域
3. web环境作用域（包括：request作用域、session作用域、application作用域、websocket作用域）
4. 自定义作用域(spring 默认的自定义作用域是：SimpleThreadScope作用域)

#### Singleton作用域(默认模式)

>所谓Singleton(也叫单例模式)作用域，就是每次向spring上下文去请求这个Bean的实例的时候spring都会返回同一个实例(或者说在spring上下文的整个生命周期中，只存在一个实例)，简单来说就是在同一个new Context实例范围获取的是同一个实例。

#### Prototype作用域

>所谓Prototype(也叫多例模式)作用域，就是每次向spring上下文去请求这个Bean的实例的时候spring都会返回不同的实例。

#### Singleton作用域与Prototype作用域有依赖关系时同样各自遵循自己的规定，仔细检查，这种情况容易混乱，同理多个单例模式依赖或者多个多例模式依赖也需仔细

#### 方法注入

情景：bean1是scope=simpleton, bean2是scope=prototype, bean2依赖bean1。在bean1 scope不变的情况下希望能够实例化到不同的bean2，这个时候就需要用到方法注入了。

#### web作用域

#### simpleThreadScope作用域(spring内置作用域)

>每个线程给一个新的实例

### bean懒加载

>单个bean懒加载：在bean标签中加lazy-init="true"属性
>全部bean懒加载：在beans标签中加default-lazy-init属性即可实现当前beans下的所有bean懒加载。

### 添加bean的初始化和销毁逻辑

>有的bean对象实例化或者销毁后要进行一些初始化或销毁后的逻辑处理

#### 初始化逻辑处理

在配置文件的bean标签中使用init-method="（bean中的某个逻辑处理方法）"；其次还可以让bean实现InitializingBean接口来实现相识的效果。也可以在beans标签中加defualt-init/destory-method来实现多个bean的默认逻辑处理

#### 销毁逻辑处理

在配置文件的bean标签中使用destory-method="（bean中的某个逻辑处理方法）"；其次还可以让bean实现DisposableBean接口来实现相识的效果。也可以在beans标签中加defualt-init/destory-method来实现多个bean的默认逻辑处理

### bean属性的继承

>一个类可能包含很多的子类和一个父类，所以在bean依赖配置的时候如果不设置属性继承的话就会有很多重复代码。此外没有继承关系的类也可能有很多相同的属性。

有继承关系的：
>将父类bean的id给子类bean的parent属性，而且父类的bean要加abstract="true"属性来告诉beans这bean这个bean不需要实例化。

无继承关系：
>将公共的属性提取出来放在一个新的bean标签中，该新标签不需要class属性，其它就和有继承关系的类似了。

## # 通过注解来管理bean

## # 注解

>它不能实现任何逻辑功能，本质就是一个标记，java底层通过反射原理获取注解，每个注解后都有一个类支持，当程序调用注解时，Java底层就会实例化这个类。三个基本的注解：@Deprecated（过时标记）、@Override（重写、覆盖）、@SuppressWarnings（压缩警告）

### 自定义注解

1. 编写注解类(支持注解功能的背后类，一个特殊的类)，使用@interface关键字
2. 注解类中再使用的注解称元注解，一般的注解都会使用@Retention和@Target注解
3. 可以给注解类属性，特别地，注解添加属性的格式为："类型 + 属性名();"

#### @Retention元注解

>（重要）当我们自定义注解类时使用的@Retention注解的作用是设定我们的注解类在被调用类中存在的时长（即生命周期），通常来说我们的Java程序会的生命周期是这样的：源代码阶段(.java)->(经过javac编译)->字节码阶段(.class)->进入jvm虚拟机内存阶段，三大阶段。@Retention的三个参数：@Retention(RetentionPolicy.SOURCE)注解在源代码执行之后被去掉，@Retention(RetentionPolicy.CLASS)注解在字节码阶段后被去除，@Retention(RetentionPolicy.RUNTIME)注解会一直保存到跟随程序进入jvm内存中，基本都是设置为@Retention(RetentionPolicy.CLASS)注解在字节码阶段后被去除，@Retention(RetentionPolicy.RUNTIME)

#### @Target元注解

>设置当前注解类的作用位置，可以作用在类、包、方法等。如：@Target(ElementType.METHOD)作用在方法。

#### 给注解类添加属性

语法：类型 + 属性名();
给定义属性并设置默认值：类型(任何常用类型) + 属性名() default "（属性值）";
不设置默认值时必须给属性传值
属性的注解也可以是一个注解或枚举(不常用)。数组常用。

#### 注解的获取(反射方式)

> 例：现Test类和注解类MyAnnotation; 要判断Test类是否使用了@MyAnnotation: Test.class.isAnnotationPresent(MyAnnotation.class)，返回true或false。
> 获取注解类实例：MyAnnotation annotation = (MyAnnotation)Test.class.isAnnotationPresent(MyAnnotation.class)
