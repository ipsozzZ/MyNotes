# Thymeleaf学习

SpringBoot在开发B/S项目的时候不推荐使用jsp开发界面，但是使用Thymeleaf、Velocity、FreeMarker等模板引擎，它们完全可以替代jsp

## # Thymeleaf模板引擎的特点

1. 动静结合，没有网络照样可以使用thymeleaf（有数据的情况下，数据还是需要网络获取的），可以查看页面的静态效果，也可以让程序员在服务器查看带数据的动态页面效果，这是因为它支持html原型，然后在html标签里增加额外的属性来达到模板+数据的展示方式，浏览器解释htlm时会忽略掉未定义的标签属性，所以thymeleaf模板可以静态运行。当有数据返回到页面时，thymeleaf会动态的替换掉静态内容，使页面动态显示

2. 开箱即用，它提供标准和spring标准两种方言，可以套用模板实现JSTL、OGNL表达式效果(就是和jsp写法大致相同)，避免了每天套模板、改JSTL、改标签的困扰。同时开发人员也可以扩展和创建自定义方言。

3. 多方言支持，thymeleaf提供spring标准方言和springMVC完美结合集成的可选模板，可以快速的实现表单绑定、属性编辑器、国际化等功能。

4. 与SpringBoot完美结合，这也是官方推荐模板引擎，springBoot提供了thymeleaf的默认配置，并且为thymeleaf设置了视图解析器，我们可以像以前SpringMVC中操作jsp一样操作thymeleaf。

## # SpringBoot整合Thymeleaf

1. 加载依赖

```xml

<!-- thymeleaf模板引擎 -->
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-thymeleaf</artifactId>
</dependency>

```

2. 创建模板文件夹：

SpringBoot会自动为Thymeleaf注册一个视图解析器ThymeleafViewResolver，还配置了模板文件(html)的位置，与jsp类似的前缀+视图名+后缀风格。与解析jsp的InternalViewResolver类似，Thymeleaf也会根据前后缀来确定模板文件的位置。在配置文件中配置缓存、编码。在resource下创建templates文件夹，里面是项目视图文件(默认配置的位置，可以改)

3. Thymeleaf中默认是将html文件的地址设置为"classpath:/templates/*.html"，也就是resource/templates/下的html文件

```html

<!DOCTYPE html>
<html lang="en" xmlns:th="http://www.thymeleaf.org"> <!-- 这里是比常规html多出来的地方 -->
<head>
    <meta charset="UTF-8">
    <title>Title</title>
</head>
<body>
<!-- 当域里有值的时候就显示域里的name的值，没有(或者没网)就显示标签中的值，这就是动静结合 -->
<h1 th:text="${name}">Hello ipso</h1>
</body>
</html>

```

4. 数据传递方式和jsp类似都是使用Model对象传递，这里不再举例，可以复习之前的笔记巩固

## # 最后

以上就是基本的Thymeleaf与SpringBoot的整合，更多用法可以自行查找文档，多练多用自然能记住
