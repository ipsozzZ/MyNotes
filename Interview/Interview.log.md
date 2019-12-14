
## # 19-12-13 学海教育

笔试题概括：

1. java基础：线程、集合等
2. 框架：spring、mybatis、SpringCloud等
3. sql基础：多表查询、左连接、引擎、索引等
4. 算法基础：排序算法
5. **问：一下代码输出内容**：

```java

package test3;

public class ClassA {

   public ClassA(){
      System.out.println("Class A");
   }

   {
      System.out.println("this is A");
   }

   static {
      System.out.println("Static A");
   }
}

package test3;

public class ClassB extends ClassA {

   public ClassB(){
      System.out.println("Class B");
   }

   {
      System.out.println("this is B");
   }

   static {
      System.out.println("Static B");
   }
}


import test3.ClassB;

public class test4 {

   public static void main(String[] args) {
      new ClassB();
   }
}

```

**输出内容**：

```statck

Static A
Static B
this is A
Class A
this is B
Class B

```

面试概括：

1. 项目提问：介绍项目以及我负责的模块
2. java基础：gc、集合等
3. 框架：进行一次业务程序步骤，框架优劣势等
