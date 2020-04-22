# Spring MVC基础

## # Spring MVC 找Controller流程

1. 扫描整个项目，定义一个Map集合（Spring已经做了）
2. 拿到所有加了@Controller注解的类
3. 遍历类中的所有方法
4. 判断方法是否加了@RequestMapping注解
5. 把@RequestMapping注解的value值作为Map的key，把Method