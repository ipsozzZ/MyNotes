# Java代理

## # 代理模式
23个经典模式的一种，又称委托模式，为目标对象提供一个代理，这个代理可以控制对目标对象的访问，外界不用直接访问目标对象，而是访问代理对象，由代理对象再访问目标对象，代理对象可以添加监控和审查处理

## # Java中的代理

### 静态代理
1. 代理对象持有目标对象的句柄(就是拥有目标对象的一个实例作为成员变量)
2. 所有调用目标对象的方法，都调用代理对象的方法
3. 对每个方法，需要静态编码（理解简单，但代码繁琐，即目标对象中有多少可供调用的方法，代理对象中就需要有多少个方法且代理类中的每个方法需做前置处理和后置处理）

### 动态代理
1. 对目标对象的方法每次被调用，进行动态拦截，交给代理处理器处理
2. 代理处理器：
- 持有目标对象的句柄
- 实现InvocationHandler接口
-- 实现invoke方法
-- 所有的代理对象方法调用，都会转发到invoke方法来
-- invoke的形参method，就是指代理对象方法的调用
-- 在invoke内部，可以根据method，使用目标对象不同的方法来响应请求

### 代理对象
1. 根据给定的接口，由Proxy类自动生成的对象
2. 类型com.sun.proxy.$Proxy0，继承自java.lang.reflect.Proxy
3. 通常和目标对象实现同样的接口(可另实现其它接口)
3. 实现多个接口
- -接口的排序非常重要
- -当多个接口里面有方法同名，则默认以第一个接口的方法调用
```java
// 动态生成代理对象实例
Subject proxySubject =
                (Subject) Proxy.newProxyInstance(
                        SubjectImpl.class.getClassLoader(), SubjectImpl.class.getInterfaces(), handler
                );
```

### 动态代理实例

1. 目标类

```java

package ipso.proxy;

/**
 * 目标类继承的接口
 */
public interface Subject {
    public void request();
}

/**
 * 目标类
 */
package ipso.proxy;

public class SubjectImpl implements Subject {
    public void request() {
        System.out.println("I am dealing the request.");
    }
}

```

2. 代理处理器
```java
package ipso.proxy;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;

public class ProxyHandler implements InvocationHandler {
    private Subject subject;

    public ProxyHandler(Subject subject){
        this.subject = subject;
    }

    /**
     * 此函数在代理对象调用任何一个方法时都会被调用
     * @param proxy  代理对象
     * @param method 调用的方法
     * @param args   具体形参
     * @return
     * @throws Throwable
     */
    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {

        System.out.println(proxy.getClass().getName()); // 输出一下代理类的名字，不重要

        // 定义预处理的工作，当然也可以根据method的不同进行不同的预处理工作，这里只做输出
        System.out.println("************ before ************");
        Object result = method.invoke(subject, args); // 调用真实的目标对象来工作
        System.out.println("************ after ************"); // 后置处理，这里只做输出
        return result;
    }
}
```

3. 测试入口（动态生成代理对象，模拟请求）

```java
package ipso.proxy;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Proxy;

public class DynamicProxyTest {
    public static void main(String[] args) {
        // 创建目标对象
        SubjectImpl realSubject = new SubjectImpl();

        // 创建调用处理器对象
        ProxyHandler handler = new ProxyHandler(realSubject);

        // 当代理对象有多个接口继承时
        /*Class<?> proxyClass = Proxy.getProxyClass(SubjectImpl.class.getClassLoader(), new Class<?>[]{Subject.class, ...});
        Object proxy = proxyClass.getConstructor(new Class[]{InvocationHandler.class}).newInstance(new Object[]{handler});*/

        // 使用Proxy.newProxyInstance方法来动态生成代理对象
        Subject proxySubject =
                (Subject) Proxy.newProxyInstance(
                        SubjectImpl.class.getClassLoader(), SubjectImpl.class.getInterfaces(), handler
                );

        // 客户端通过代理对象调用方法
        // 本次调用将自动被代理处理器的invoke方法接收
        proxySubject.request();

        System.out.println(proxySubject.getClass().getName());
    }
}
```

## # AOP编程
AOP编程也叫面向切面编程，在面向对象编程中，将需求功能划分为不同的、独立，封装良好的类，并让它们通过继承和多态实现相同和不同行为。而面向切面编程中通过需求功能从众多类中分离出来，使得很多类共享一个行为，一旦发生变化，不必修改很多类，而只需要修改这个行为即可。

1. 分离代码的耦合(高内聚，低耦合)
2. 业务逻辑变化不需要修改源代码/不用重启
3. 加快编程和测试速度

可以基于代理模式实现面向切面的编程案例

## # 最后
动态代理类相较于静态代理类来说，静态代理中逻辑易于理解，但是代码实现相对繁琐，目标类中有多少个方法就需要代理类中对应多少个方法，当目标类方法较多时不易于实现和代码维护。为何需要代理降低调用方与被调用方的耦合；再者就是保护目标对象的安全性，防止重要信息泄露，另外还可以过滤调用方的不合法请求等

**扩展知识点：WatchServive是java NIO引入的类，可用于文件系统的文件变化监控可处理。**