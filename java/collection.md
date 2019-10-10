# java 集合

## # 介绍

Java集合，java.util提供了集合类，包括：

* Collection：根接口
* List：有序列表
* Set：无重复元素集合
* Map：通过Key查找Value的映射表

Java集合支持范型，通过迭代器（Iterator）访问集合。

## # List

List是一种有序列表，通过索引访问元素。

常用方法：

* void add(E e) 在末尾添加一个元素
* void add(int index, E e) 在指定索引添加一个元素
* int remove(int index) 删除指定索引的元素
* int remove(Object e) 删除某个元素
* E get(int index) 获取指定索引的元素
* int size() 获取链表大小（包含元素的个数）

List有ArrayList(数组，线程不安全)和LinkedList(指针列表，线程安全)两种实现。遍历List使用Iterator或者foreach循环。List和Array可以相互转换。

**list特点**：

* 按索引顺序访问的长度可变的列表
* 优先使用ArrayList而不是LinkList
* 可是使用for遍历，或者foreach遍历
* 可以和Array相互转换

**list和array相互转换**：

* 将Array变成List

```java

Integer[] array = {1, 2, 3};
List<Integer> list = Arrays.asList(array); // 注意这样转换出来的list只是可读的list，不能进行其它操作
List<Integer> arraylist = new ArrayList<>();
arraylist.addAll(list) // 将上面只读的list转换成ArrayList

// 遍历list输出
for (Iterator<Integer> it = arraylist.iterator(); it.hasNext(); ) {
  System.out.println(it.next());
}

// 整理代码
Integer[] array = {1, 2, 3};
List<Integer> arraylist = new ArrayList<>(ArrayList.asList(array));

```

* List转Array

如果提供的数组长度小于list长度的话，java会丢掉提供的数组，重新new一个长度符合的数组，再转换

如果数组长度大于list长度的话，java会保留多出来的部分，并使用null值代替

```java

List<String> list = new LinkedList<>();
list.add("apple");
list.add("pear");
list.add("orange");

// 这里在支持的格式转换范围内，可允许使用其它类型接收，比如list存的是Integer, 就可以使用Number来接收，但是这种情况不能使用String接收
String[] ss = list.toArray(new String[list.size()]); 
for (String s : ss){
    System.out.println(s);
}

```

**在List中使用contains()和indexOf()方法**：

boolean contains(Object o) 是否包含某个元素
int indexOf(Object o) 查找某个元素的索引，不存在返回-1

```java

/* 在List中使用contains()和indexOf()方法 */
List<String> list2 = new ArrayList<>();
list2.add("A");
list2.add("B");
list2.add("C");

list2.contains("B"); // true, 存在返回true
list2.contains("X"); // false, 不存在返回false
list.indexOf("B"); // 1（从0开始的）, 返回元素的索引
list2.indexOf("X"); // -1，不存在时返回-1

```

**重写equals**:

在上面的例子中我们给出contains()和indexOf()方法的例子并说明其作用，但是上面例子List存的是普通java类型, 假设现在有一个对象Person,并将不同的Person实例存入List中，再调用contains()和indexOf()方法

```java

// Person类
public class Person {

   private String name;
   private int age;

   public Person(String name, int age){
      this.name = name;
      this.age  = age;
   }

   @Override
   public String toString() {
      return "Person{" + "name='" + name + '\'' + ", age=" + age + '}';
   }
}

// Test测试类

import java.util.ArrayList;
import java.util.List;

public class Test {

   public static void main(String[] args) {

      /*-------- 测试普通类型下的contains()和indexOf()方法 --------*/
      System.out.println("案例一：\n");
      List<String> strList = new ArrayList<>();
      strList.add(new String("ipso"));
      strList.add(new String("ipso1"));
      strList.add(new String("ipso2"));
      strList.add(new String("ipso3"));
      Boolean flag1 = strList.contains(new String("ipso"));
      System.out.println("普通类型contains结果:" + flag1);

      int index1 = strList.indexOf("ipso2");
      System.out.println("普通类型下标为：" + index1);

      System.out.println("\n/* ---------分-----------界------------线----------- */\n");

      /*-------- 测试自定义类型下的contains()和indexOf()方法 --------*/
      System.out.println("案例二：\n");
      List<Person> list = new ArrayList<>();

      list.add(new Person("ipso", 11));
      list.add(new Person("ipso1", 21));
      list.add(new Person("ipso2", 31));
      list.add(new Person("ipso3", 41));

      Boolean flag = list.contains(new Person("ipso", 11));
      System.out.println("contains结果:" + flag);

      int index = list.indexOf(new Person("ipso2", 31));
      System.out.println("下标为：" + index);

      /* 打印结果：
      普通类型contains结果:true
      普通类型下标为：2

      *//* ---------分-----------界------------线----------- *//*

      contains结果:false
      下标为：-1
      */

   }
}

```

**?!?!?**为什么呢？上面普通类型是我们想要的结果，可是一样的操作下面为什么就不行了？

**揭秘**：

根据我对普通类型案例(案例一)的调试追踪我发现，在ArrayList类的contains()方法中调用了自身类的indexOf()方法，而在indexOf()中有一个关键方法, String类中的equals()方法, 该方法在最基本的Object类中被创建，并且其是将另一个对象与当前对象进行'=='比较并返回比较结果。但是String类将该方法进行了重写：

```java

// 这是Object.java中的equals()
public boolean equals(Object obj) {
  return (this == obj);  // 这里不同实例是不会相等的
}



// 这是String.java中的equals()
public boolean equals(Object anObject) {
  if (this == anObject) {
    return true;
  }
  // 这里判断anObject是否是String的一个实例，是(包括String的子类的实例)就返回true,取结果true执行if
  if (anObject instanceof String) {
    String anotherString = (String)anObject; // 做强类型转换
    int n = value.length;   // 获取传入字符串的长度并赋值给n(字符个数)
    if (n == anotherString.value.length) { // 再将n与做强类型转换后的字符数组长度做'=='运算，取结果true执行if
      char v1[] = value; // 将传入对象的数据存入v1字符数组
      char v2[] = anotherString.value; // 将强类型转换后的对象值存入v2字符数组
      int i = 0;
     // 再判断两个字符数组的值是否相等，不等返回false,等返回true
      while (n-- != 0) {
        if (v1[i] != v2[i])
          return false;
        i++;
      }
      return true;
    }
  }
  return false;
}

```

可以看到调试追踪的代码中contains()方法就是调用indexOf()方法实现的，而indexOf()方法又调用了(```List<myObject>```)myObject中的equals()方法，没有重写该方法的话会调用Object对象中创建的equals()方法，如果不对equals()进行重写的话，如果我们不再同一实例中操作contains()或indexOf()方法的话就会得不到想要的结果，所以现在我们来重新编写我们的Person类

```java

/**
  * 重新equals
  * @param anObject 传入对象实例
  * @return boolean
  */
  public boolean equals(Object anObject){
     if (this == anObject) {
        return true;
     }
     if (anObject instanceof Person){
        Person anotherPerson = (Person)anObject;
        if (Objects.equals(anotherPerson.name, this.name) && anotherPerson.age==this.age){
           return true;
        }
        return false;
     }
     return false;
  }

// 用到的Objects类中的equals()方法
public static boolean equals(Object a, Object b) {
  return (a == b) || (a != null && a.equals(b));
}

/* 添加自定义equals()方法后的打印结果 */
/*
案例一：

普通类型contains结果:true
普通类型下标为：2

*//* ---------分-----------界------------线----------- *//*

案例二：

contains结果:true
下标为：2

Process finished with exit code 0
*/


```

嗯！很棒没错现在结果一致了，是我们要的结果了
