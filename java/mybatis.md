# Mybatis学习

## # 入门程序(连接数据库等)

```xml

<!--  全局配置文件  -->

<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE configuration
        PUBLIC "-//mybatis.org//DTD Config 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-config.dtd">
<configuration>

    <!-- 配置sql打印 -->
    <settings>
        <setting name="logImpl" value="STDOUT_LOGGING"/>
    </settings>

    <!-- spring整合后 environments配置将废除 使用spring中的连接池 -->
    <environments default="development">
        <environment id="development">
            <!-- 使用jdbc事务管理 -->
            <transactionManager type="JDBC" />
            <!-- 数据库连接池 -->
            <dataSource type="POOLED">
                <property name="driver" value="com.mysql.jdbc.Driver" />
                <property name="url"
                          value="jdbc:mysql://localhost:3306/mybatis?characterEncoding=utf-8" />
                <property name="username" value="root" />
                <property name="password" value="password" />
            </dataSource>

        </environment>
    </environments>

    <!-- 加载映射文件 -->
    <mappers>
        <mapper resource="live/ipso/mybatis/domain/Customer.xml"/>
    </mappers>
</configuration>

<!--  映射配置文件示例  -->

<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="myTest">
    <!--根据cust_id查询客户-->
    <select id="queryCustomerById" parameterType="Int" resultType="live.ipso.mybatis.domain.Customer">
        SELECT * FROM `customer` WHERE cust_id  = #{cust_id}
    </select>
</mapper>

```

```java

// 操作示例

// 抽取工具类（注意当配合其它框架(如Springmvc等)使用时就没必要抽取工具类，因为其它框架会有响应的配置）

/**
 * 数据库工具类
 */
public class MybatisUilts {

   public static final SqlSessionFactory sessionFactory;

   /*  只执行一次  */
   static {
      // 1. sqlSessionFactoryBuilder 加载配置文件
      SqlSessionFactoryBuilder sqlSessionFactoryBuilder = new SqlSessionFactoryBuilder();

      // 2. 读取配置文件
      InputStream resourceAsStream = null;
      try {
         resourceAsStream = Resources.getResourceAsStream("SqlMappingConfig.xml");
      } catch (IOException e) {
         e.printStackTrace();
      }

      // 3. 获取session工厂
      sessionFactory = sqlSessionFactoryBuilder.build(resourceAsStream);
   }

   public static SqlSession openSession(){
      return sessionFactory.openSession();
   }
}


// 测试类

public class MyTest {

  private Customer customer;

  @Test
  public void test() throws IOException {
     getOne(12);
  }

  /**
   * 根据id查询一条数据
   * 查询时可以不手动提交事务，当涉及到数据库增删改时一定要手动提交事务，最好是都加上以免混淆
   */
  public void getOne(int id){
    // 获取会话 （连接数据库，相当于JDBC中的连接数据库）
    SqlSession sqlSession = MybatisUilts.openSession();

    // 执行Sql
    Customer customer = sqlSession.selectOne("queryCustomerById", id);
    System.out.println(customer);

    // 关闭session(会话)
    sqlSession.close();
  }
}

```

## # 增删改查

### Mybatis配置文件sql语句中#{}与${}的区别

**#{}**：

1. 表示一个占位符；
2. 通过#{}可以实现perparedStatement向占位符(解析sql后会加上单引号，包含int类型也会加)中设置值，自动进行java类型和JDBC类型转换；
3. #{}可以有效防止sql注入;
4. #{}可以接收简单类型值或pojo属性值；
5. 如果parammeterType传输单个简单类型值,#{}中可以是value或其它名称；

**$()**:

1. 表示拼接sql串
2. 通过${}可以将perparedStatement传入的内容拼接在sql且不进行JDBC类型转换
3. ${}可以接收接收简单类型值或pojo属性值
4. 如果perparedStatement传单个简单类型的值，${}中的值只能是value

### 传统方式编写Dao接口实现类

例：

```java

// dao接口

package live.ipso.mybatis.dao;

import live.ipso.mybatis.domain.Customer;

import java.util.List;

public interface CustomerDao {
   public Customer getById(int id);
   public List<Customer> getAll();
   public List<Customer> getBySearchName(String search);
   public Integer insertOne(Customer customer);
   public void update(Customer customer);
   public void deleteOne(Customer customer);
}

// dao接口实现类

package live.ipso.mybatis.dao;

import live.ipso.mybatis.domain.Customer;
import live.ipso.mybatis.uilts.MybatisUilts;
import org.apache.ibatis.session.SqlSession;

import java.util.List;

/**
 * CustomerDao实现类
 * @author ipso
 */
public class CustomerDaoImpl implements CustomerDao {

   private final String tableName = "Customer";

   @Override
   public Customer getById(int id) {
      // 获取会话 （连接数据库，相当于JDBC中的连接数据库）
      SqlSession sqlSession = MybatisUilts.openSession();

      // 执行Sql
      // selectOne只能查一条数据
      // selectList查一条或多条数据
      Customer customer = sqlSession.selectOne("queryCustomerById", id);
      System.out.println(customer);

      // 提交事务
      sqlSession.commit();

      // 关闭session(会话)
      sqlSession.close();

      return customer;
   }

   public String getTableName() {
      return tableName;
   }

   @Override
   public List<Customer> getAll() {
      // 获取会话
      SqlSession sqlSession = MybatisUilts.openSession();

      // 执行sql
      List<Customer> customers = sqlSession.selectList("queryAllCustomer");

      // 关闭session(会话)
      sqlSession.close();

      return customers;
   }

   @Override
   public Integer insertOne(Customer customer) {
      // 获取会话
      SqlSession sqlSession = MybatisUilts.openSession();

      // 执行sql
      int id = sqlSession.insert("insertCustomer", customer);

      // 提交事务当数据库数据发生更改(增删改)时一定提交事务
      sqlSession.commit();
      System.out.println(customer.getCust_id());

      // 关闭会话
      sqlSession.close();

      return id;
   }

   @Override
   public void update(Customer customer) {
      // 获取会话
      SqlSession sqlSession = MybatisUilts.openSession();

      // 执行sql语句
      sqlSession.update("updateCustomer", customer);
      sqlSession.commit();

      sqlSession.close();
   }

   @Override
   public void deleteOne(Customer customer) {
      // 获取会话
      SqlSession sqlSession = MybatisUilts.openSession();

      // 执行sql
      sqlSession.delete("deleteCustomer", customer);
      sqlSession.commit();

      // 关闭会话
      sqlSession.close();
   }

   @Override
   public List<Customer> getBySearchName(String search) {
      // 获取会话
      SqlSession sqlSession = MybatisUilts.openSession();

      // 执行sql
      List<Customer> customers = sqlSession.selectList("queryLikeCustomer", search);

      // 关闭session(会话)
      sqlSession.close();

      return customers;
   }

}

```

### 动态代理完成dao接口实现类

Mybatis可以使用动态代理自动实现dao接口实现类

**dao接口与映射xml文件要求**：

1. namespace值必须和Mapper接口类路径一致
2. id必须和Mapper接口方法名一致
3. parameterType必须和接口方法参数类型一致
4. resultType必须和接口方法返回值一致

**步骤**：

1. 按照dao接口与映射xml文件要求，编写dao接口与映射文件(还需要使用前面封装的工具类来获取数据库会话对象SqlSession)
2. 调用dao接口示例

```java

package live.ipso.mybatis.test;

import live.ipso.mybatis.dao.CustomerDao;
import live.ipso.mybatis.dao.CustomerDaoImpl;
import live.ipso.mybatis.domain.Customer;
import live.ipso.mybatis.mapper.CustomerMapper;
import live.ipso.mybatis.uilts.MybatisUilts;
import org.apache.ibatis.session.SqlSession;
import org.junit.Test;

import java.util.List;

public class MyTest2 {

   @Test
   /**
    * 传统模式dao层实现测试
    */
   public void test(){
      CustomerDao customerDao = new CustomerDaoImpl();

      System.out.println("查询一条数据");
      Customer customer = customerDao.getById(12);
      System.out.println(customer);

      /*System.out.println("查询所有");
      List<Customer> customers = customerDao.getAll();
      for (Customer customer1 : customers) {
         System.out.println(customer1);
      }*/

      System.out.println("插入一条数据");
      Customer customer1 = customerDao.getById(21);
      customer1.setCust_name("test");
      customer1.setCust_phone("123456789");
      customer1.setCust_profession("打野");
      customer1.set_email("2222@qq.com");

      customerDao.update(customer1);

      System.out.println("************** 模糊查询 ***************");

   }

   @Test
   /**
    * 使用Mybatis动态代理自动生成dao接口实现类实现dao层测试类
    */
   public void test2(){

      /*----- 获取SqlSession数据库会话和Mapper对象 ----*/
      SqlSession sqlSession = MybatisUilts.openSession();
      CustomerMapper mapper = sqlSession.getMapper(CustomerMapper.class);

      System.out.println("----- 查询一条记录 ----");
      /*----- 查询一条记录 ----*/
      // 获取会话
      Customer customer = mapper.queryCustomerById(12);
      System.out.println(customer);
      sqlSession.close();

      /*----- 查询所有记录 ----*/
      System.out.println("*----- 查询所有记录 ----*");
      List<Customer> customers = mapper.queryAllCustomer();
      for (Customer customer1 : customers) {
         System.out.println(customer1);
      }

      /*----- 模糊查询 ----*/
      System.out.println("*----- 模糊查询 ----*");
      List<Customer> customers1 = mapper.queryLikeCustomer("李");
      for (Customer customer2 : customers1) {
         System.out.println(customer2);
      }

      /*----- 插入一条数据 ----*/
      Customer customer1 = new Customer();
      customer1.setCust_name("test222");
      customer1.setCust_profession("AD");
      customer1.setCust_phone("12897385472");
      customer1.set_email("123@ipso.com");
      mapper.insertCustomer(customer1);

      /*----- 更新一条数据 ----*/
      System.out.println("*----- 更新一条数据 ----*");
      customer.setCust_name("小李子公公hhhh");
      mapper.updateCustomer(customer);
      System.out.println(customer);

      /*----- 删除一条数据 ----*/
      System.out.println("*----- 删除一条数据 ----*");
      Customer customer2 = mapper.queryCustomerById(21);
      mapper.deleteCustomer(customer2);

   }
}

```

### Mybatis-Mapper传参多个普通类型与@param

**当传单个参数时：**：

可以接收基本类型、对象类型、集合类型的值。Mybatis可以直接使用该参数，不需要经过任何处理，#{}里可以任意取名

**当传多个参数时**：

任意多个参数，都会被Mybatis重新包装成一个Map传入。Map的key就是param1、param2(注意param是从1开始)或arg0、arg1... (注意arg是从0开始), 值就是参数的值。

**当传多个参数时为什么我们可以在表达式中使用arg0,arg1或者param1,param2来接收参数**：

在package org.apache.ibatis.reflection包中有一个参数解析器ParamNameResolver类，它声明了一个map类型的数据结构```private final SortedMap<Integer, String> names```，默认存储的形式key->value为: 0->arg0; 1->arg1。**注意：**当使用@Param("")注解时key->value就会变成：0->myname1; 1->myname2（myname1、2为用户通过@Param自定义参数名，所以在使用@Param修饰参数时就不能使用arg0、1就是这个原因，使用map参数时原因类似，该类有方法```public Object getNamedParams(Object[] args)```，这个方法将参数存到数组中返回，在该方法中会判断这个names是空还是一个值或者是两个值，是空就直接返回null，一个值就直接返回```return args[names.firstKey()];```,两个值则进行以下操作：

```java

final Map<String, Object> param = new ParamMap<Object>();
int i = 0;
for (Map.Entry<Integer, String> entry : names.entrySet()) {
  param.put(entry.getValue(), args[entry.getKey()]); // 这里产生map的默认key：arg0,arg1.....(来自names默认的value)
  // add generic param names (param1, param2, ...)  这句注释意思是添加通用的参数名：param1,param2....
  final String genericParamName = GENERIC_NAME_PREFIX + String.valueOf(i + 1);
  // ensure not to overwrite parameter named with @Param
  if (!names.containsValue(genericParamName)) {
    param.put(genericParamName, args[entry.getKey()]);
  }
  i++;
}
return param;

```

这就是为什么可以使用arg0,arg1...和param1,param2...的原因。

**注意**：

传多个值时：如果不想使用原来Map的key(也就是param1、param2或arg0、arg1...)时：可以在声明Mapper接口方法时：在方法参数前使用这种形式的参数"public 返回类型 方法名(@param("myname1") 参数类型 参数名1, @param("myname2") 参数类型 参数名2, ...)",**再注意**：当使用了@Param修饰参数时，原来的arg0,arg1...就不能使用了**但是 但是 但是**param1,param2,...还能使用。

**当传Map时**：

我们可以将多个参数封装到map中直接传map
在映射文件的sql语句中接收map参数时：#{}表达式中可以使用param1,param2...但是不能使用arg0,arg1...，此外还可以使用map自身的key(建议使用)

**知识扩展**：

POJO：简单的Java对象（Plain Ordinary Java Objects）实际就是普通JavaBeans,使用POJO名称是为了避免和EJB混淆起来, 而且简称比较直接. 其中有一些属性及其getter setter方法的类,有时可以作为value object或dto(Data Transform Object)来使用.当然,如果你有一个简单的运算属性也是可以的,但不允许有业务方法,也不能携带有connection之类的方法。

POJO是Plain Ordinary Java Objects的缩写不错，但是它通指没有使用Entity Beans的普通java对象，可以把POJO作为支持业务逻辑的协助类。

我们项目中数据库的映射类就是POJO

**当传POJO对象时时**：

当我们要传的参数属于我们某个业务POJO时，我们直接传递POJO
在映射文件的sql语句中接收POJO参数时：#{}表达式中只能使用POJO对象对应的属性名

## # 配置信息

实际开发中需要根据整合的框架来做配置

注意配置文件中的标签的顺序是有约束的具体顺序可以参考[Mybatis配置文档](https://mybatis.org/mybatis-3/zh/configuration.html#settings)

### properties标签及属性文件

在配置文件中可以使用properties标签定义一条属性，也可以将该属性写到一个属性文件中

```xml

 <!-- Mybatis全局配置文件 -->

<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE configuration
        PUBLIC "-//mybatis.org//DTD config 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-config.dtd">
<configuration>

    <!-- 载入属性文件 -->
    <properties resource="db.properties" />

    <!-- 配置sql打印日志信息 -->
    <settings>
        <setting name="logImpl" value="STDOUT_LOGGING"/>
    </settings>

    <!-- spring整合后 environments配置将废除 使用spring中的连接池 -->
    <environments default="development">
        <environment id="development">
            <transactionManager type="JDBC" />
            <dataSource type="POOLED">
                <!-- 使用${}属性文件中定义的属性 -->
                <property name="driver" value="${jdbc.Driver}"/>
                <property name="url" value="${jdbc.url}"/>
                <property name="username" value="${jdbc.username}"/>
                <property name="password" value="${jdbc.password}"/>
            </dataSource>
        </environment>
    </environments>

    <!-- 加载映射文件 -->
    <mappers>
        <mapper resource="live/ipso/mybatis/domain/Customer.xml"></mapper>
    </mappers>
</configuration>

 <!-- 属性文件db.properties -->

jdbc.Driver=com.mysql.jdbc.Driver
jdbc.url=jdbc:mysql://localhost:3306/mybatis?characterEncoding=utf-8
jdbc.username=root
jdbc.password=gqm1975386453

```

**注意**：

如果在使用属性文件的情况下在配置文件的properties标签中可以继续定义属性，但是如果定义的属性存在同名情况时的执行顺序是：先执行标签中的属性，再执行属性文件中的属性，同名的属性后执行的覆盖先执行的。一般情况不会既调用在属性文件中的属性又调用在properties标签中的属性，两种方式一般情况只会出现一种。

### settings标签

这是Mybatis中极为重要的调整设置，它们会改变Mybatis的运行时行为，如采用驼峰命名法(比如数据库中采用的是'_'下划线命名法，我在做对象映射时对象属性名想要使用驼峰法，就可以在settings中设置)。更多settings设置可以参考[mybatis配置文档](https://mybatis.org/mybatis-3/zh/configuration.html#settings)

### 类型别名（typeAliases配置）

类型别名是为 Java 类型设置一个短的名字。 它只和 XML 配置有关，存在的意义仅在于用来减少类完全限定名的冗余。

```xml

<!-- 别名配置，在xml文件中使用live.ipso.mybatis.domain.Customer时可以使用别名简化成Customer -->
<typeAliases>
  <!-- 定义单个别名 -->
  <!--<typeAlias alias="Customer" type="live.ipso.mybatis.domain.Customer" />-->

  <!-- 在包live.ipso.mybatis.domain下的所有对象都可以直接使用其类名调用(即自动使用类名作为别名) -->
  <package name="live.ipso.mybatis.domain"/>
</typeAliases>

```

### typeHandlers类型处理器

无论是 MyBatis 在预处理语句（PreparedStatement）中设置一个参数时，还是从结果集中取出一个值时， 都会用类型处理器将获取的值以合适的方式转换成 Java 类型。下表描述了一些默认的类型处理器。从 3.4.5 开始，MyBatis 默认支持 JSR-310（日期和时间 API）所以项目很少用这个标签。

### 插件（plugins）

待续……

### 环境配置（environments）

```xml

<!-- spring整合后 environments配置将废除 使用spring中的连接池 -->
<environments default="development">
  
  <!-- 正式环境 -->
  <environment id="development">
    <!-- 事务控制 -->
    <transactionManager type="JDBC" />

    <!-- 数据源 -->
    <dataSource type="POOLED">
      <!-- 使用属性文件中定义的属性 -->
      <property name="driver" value="${jdbc.Driver}"/>
      <property name="url" value="${jdbc.url}"/>
      <property name="username" value="${jdbc.username}"/>
      <property name="password" value="${jdbc.password}"/>
    </dataSource>
  </environment>

  <!-- 测试所用环境 -->
  <environment id="test">
    <transactionManager type="JDBC" />
    <dataSource type="POOLED">
      <!-- 使用属性文件中定义的属性 -->
      <property name="driver" value="${jdbc.Driver}"/>
      <property name="url" value="${jdbc.url}"/>
      <property name="username" value="${jdbc.username}"/>
      <property name="password" value="${jdbc.password}"/>
    </dataSource>
  </environment>
  
</environments>

```

### 数据库厂商标识（databaseIdProvider）

MyBatis 可以根据不同的数据库厂商执行不同的语句，这种多厂商的支持是基于映射语句中的 databaseId 属性。 MyBatis 会加载不带 databaseId 属性和带有匹配当前数据库 databaseId 属性的所有语句。 如果同时找到带有 databaseId 和不带 databaseId 的相同语句，则后者会被舍弃。

### mapper加载映射文件

## # 查询输出类型

### 使用简单类型接收

上面的例子中有很多简单类型的例子这里不再重复

### 使用POJO类型接收

就是使用domain中的对象去接收，上面也很多类似例子，不再举例

### 使用Map类型接收

就是使用Map类型去接收数据库查询结果集

**接收单条记录示例与接收多条记录案例**：

```java

// CustomerMapper.java

public interface CustomerMapper {
   public Integer getCount();

   /* 接收单条记录，并使用默认key接收 */
   public Map<String, Object> getById(Integer id);

   /* 接收多条记录并使用自定义key接收 */
   @MapKey("cust_id") // 使用字段"cust_id"作为map的key值
   public Map<Integer, Customer> getByLikeName(String name);

   @MapKey("cust_id")
   public Map<Integer, Customer> getAll();
}

// MyTest.java

public class MyTest {

   @Test
   public void test(){

      /* 获取数据库连接信息 */
      SqlSession sqlSession = MyUilts.opSession(); // 获取数据库会话Session
      CustomerMapper mapper = sqlSession.getMapper(CustomerMapper.class);

      /* 获取记录数，使用普通类型接收 */
      /*int count = mapper.getCount();*/
      /*System.out.println("数据库库存：" + count + "\n");*/

      /* 根据id获取一条数据，使用map对象接收，默认情况,使用属性名为key进行存储 */
      /*Map<String, Object> customer = mapper.getById(12);
      System.out.println("查询结果：" + customer);*/
      /* 打印结果 */
      /*{cust_profession=打野, cust_name=小李子公公, cust_id=12, cust_phone=13398908928, email=197538@gmail.com}*/

      /* 使用map对象接收，默认情况 */
      /*Map<Integer, Customer> customerMap = mapper.getByLikeName("%李%");
      System.out.println(customerMap);*/
      /* 打印结果 */
      /*{
          2={cust_profession=刺客, cust_name=李白, cust_id=2, cust_phone=18977665521, email=libai@163.com},
         11={cust_profession=打野, cust_name=李四, cust_id=11, cust_phone=13398908928, email=12345@qq.com},
         12={cust_profession=打野, cust_name=小李子公公, cust_id=12, cust_phone=13398908928, email=197538@gmail.com}
      }*/

      /* 使用map对象接收，指定key，指定value */
      /*Map<Integer, Customer> allCustomer = mapper.getAll();
      System.out.println(allCustomer);*/
   }
}

```

```xml



<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<!-- 数据库映射文件CustomerMapper.xml -->
<!-- 使用传统的实现dao层, namespace可以随意赋值 -->
<!-- 使用Mybatis动态代理实现dao层，namespace的值必须是和Mapper接口类路径一致 -->
<mapper namespace="live.ipso.mybatis.mapper.CustomerMapper">

    <select id="getCount" resultType="Integer">
        select count(*) from `customer`
    </select>

    <select id="getById" parameterType="Integer" resultType="Map">
        select * from `customer` where cust_id=#{id}
    </select>

    <select id="getByLikeName" parameterType="String" resultType="java.util.Map">
        select * from `customer` where cust_name like #{name}
    </select>

    <select id="getAll" resultType="java.util.Map">
        select * from `customer`
    </select>

    <!-- 当数据库表与domain中的对象字段与属性名不相同时可以自己做表与对象的映射如下 -->
    <resultMap id="customerMap" type="Customer">
        <!-- 主键映射 -->
        <id column="cust_id" property="id" />
        <!-- 字段名对应属性名映射 -->
        <result column="cust_name" property="name" />
        <!-- 注意已经相同的属性名不用做映射也可以，但是习惯上是相同的也加上 -->
    </resultMap>

    <!-- 调用自定义映射 -->
    <select id="getNewCustomer" parameterType="Integer" resultMap="customerMap">
        select * from `customer` where cust_id=#{id}
    </select>


</mapper>

```

## # 多表操作

### 级联/关联属性赋值demo

实际开发中使用关联对象属性赋值，因为相较级联，关联有可以分步查询等优点(**注意**:通俗理解分步查询，分步查询就是使用上一条sql语句的结果作为条件再进行sql操作)

```java

// order映射类
package live.ipso.mybatis.domain;

public class Order {
   private int order_id;
   private String order_name;
   private String order_num;
   private Customer customer;

   public int getOrder_id() {
      return order_id;
   }

   public void setOrder_id(int order_id) {
      this.order_id = order_id;
   }

   public String getOrder_name() {
      return order_name;
   }

   public void setOrder_name(String order_name) {
      this.order_name = order_name;
   }

   public String getOrder_num() {
      return order_num;
   }

   public void setOrder_num(String order_num) {
      this.order_num = order_num;
   }

   public Customer getCustomer() {
      return customer;
   }

   public void setCustomer(Customer customer) {
      this.customer = customer;
   }

   @Override
   public String toString() {
      return "Order{" + "order_id=" + order_id + ", order_name='" + order_name + '\'' + ", order_num='" + order_num + '\'' + ", customer=" + customer + '}';
   }
}

```

```java

// order mapper接口类
package live.ipso.mybatis.mapper;

import live.ipso.mybatis.domain.Order;

import java.util.List;

public interface OrderMapper {

   public List<Order> getAll();
}

```

```xml

<!-- order表数据库映射文件 -->

<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<!-- 数据库映射文件 -->
<!-- 使用传统的实现dao层, namespace可以随意赋值 -->
<!-- 使用Mybatis动态代理实现dao层，namespace的值必须是和Mapper接口类路径一致 -->
<mapper namespace="live.ipso.mybatis.mapper.OrderMapper">

    <!-- 自定义映射 -->
    <resultMap id="orderMap" type="Order">
        <!-- 级联方式可以不写，关联方式必须写(习惯是都写，因为实际开发通常使用关联方式) -->
        <id property="order_id" column="order_id" />
        <result property="order_name" column="order_name" />
        <result property="order_num" column="order_num" />

        <!-- 级联映射（属性必须一一对应） -->
        <!--<result property="customer.cust_id" column="cust_id" />
        <result property="customer.email" column="email" />
        <result property="customer.cust_phone" column="cust_phone" />
        <result property="customer.cust_name" column="cust_name" />
        <result property="customer.cust_profession" column="cust_profession" />-->

        <!-- 关联对象映射 -->
        <association property="customer" javaType="Customer">
            <id property="cust_id" column="cust_id" />
            <result property="cust_name" column="cust_name" />
            <result property="email" column="email" />
            <result property="cust_phone" column="cust_phone" />
            <result property="cust_profession" column="cust_profession" />
        </association>
    </resultMap>

    <!-- 左连接查询 -->
    <select id="getAll" resultMap="orderMap">
        select * from `order` as o left join mybatis.`customer` as c on o.cust_id=c.cust_id;
    </select>
</mapper>

```

```java

// 单元测试类
package live.ipso.mybatis.test;

import live.ipso.mybatis.domain.Order;
import live.ipso.mybatis.mapper.OrderMapper;
import live.ipso.mybatis.uilts.MyUilts;
import org.apache.ibatis.session.SqlSession;
import org.junit.Test;

import java.util.List;
import java.util.Map;

public class MyTest {

   @Test
   public void test(){

      /* 获取数据库Customer表连接信息 */
      SqlSession sqlSession = MyUilts.opSession(); // 获取数据库会话Session

      /* 获取数据库order表连接信息 */
      OrderMapper mapperOrder = sqlSession.getMapper(OrderMapper.class);

      List<Order> all = mapperOrder.getAll();

      for (Order order : all) {
         System.out.println(order.getCustomer());
      }

      sqlSession.close();
   }
}

```
