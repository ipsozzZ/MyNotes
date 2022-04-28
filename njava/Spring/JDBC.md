# 数据库

## # JDBC （myeclipse中代码自动对齐(ctrl + i)、代码自动补全或提示(Alt + /)）

>他包含了数据库操作规范(类、接口等)，没有提供实现。只需要对应的数据库提供驱动jar包，接入驱动jar包后就可以连接对应的数据库。

### 加载驱动

>Class.forName("com.mysql.jdbc.Driver")，这里要处理异常，后面释放资源的时候也需要处理异常，加载驱动(java1.6后可以不用加载驱动，已经被实现了，但是web程序是没有被实现的，所以做web程序必须添加加载代码)。

### 连接数据库demo

>DriverManager.getConnection(url, user, password);
```

   // 1 加载数据库驱动
   Class.forName("com.mysql.jdbc.Driver");

   // url 数据库地址
   String url = "jdbc:mysql://(前面部分是连接mysql固定格式)localhost:3306(数据库主机名加端口)/jdbc_test(数据库名)"; // 注意格式
   // user 数据库用用户名
   String user = "root";
   // password 数据库用户密码
   String password= "password";
   Connection conn = DriverManager.getConnection(url, user, password); // 简单理解为从java连通数据库

   // 3 编写数据库语句
   String sql = "Create table stu(id int, name varchar(50), age int)";
   Statement st = conn.createStatement(); // 获取执行静态sql语句的对象，简单理解为从数据库连通数据库执行程序
   // 4 执行数据库语句
   int res = st.executeUpdate(sql);
   // 5 释放资源
   st.close(); // 将通道撤走
   conn.close(); // 将通道撤走

  /*----- 原始正确写法 -----*/
   Connection conn = null;
   Statement st = null;

  try{
      // 加载驱动
      Class.forName("com.mysql.jdbc.Driver");

      String url = "jdbc:mysql://localhost:3306/jdbc_db";
      String user = "root";
      String pass = "gqm1975386453";

      // 连接数据库
      conn = DriverManager.getConnection(url, user, pass);

      // 创建数据库语句
      String sql = "insert into stu values (1, 'zs', 20)";
      st = conn.createStatement();

      // 执行数据库
      int row = st.executeUpdate(sql);
      System.out.println(row);
   }catch (Exception e) {
      e.printStackTrace();
   }finally {
      // 释放资源

      if(st != null){
         try{
            st.close();
         }catch (Exception e) {
            e.printStackTrace();
         }
      }
      if(conn != null){
         try{
            conn.close();
         }catch (Exception e) {
            e.printStackTrace();
         }
      }
   }

```

### java数据库操作(CURD)

* 插入、创建使用：executeUpdate(sql),返回影响的行数int类型

* 查询：使用executeQuery(sql), 返回ResultSet结果集,object类型。获取结果集后常用的方法：
bool类型的next();判断是否有下一行数据；获取一列使用getXxx(列名)，其中(Xxx为获取列的类型如int等)。还有很多其它的方法

### 数据库操作中通俗描述几个关键的对象

1. Connection: 连接数据库对象

2. DriverManager：获取连接数据库的对象

3. Statement: 获取数据库执行程序的对象

4. ResultSet：查询结构集对象

### DAO设计(Data Access Object, 数据存取对象，位于业务逻辑和持久化数据之间，实现对持久化数据的访问)

>没有使用dao时数据库操作方面有大量的重复代码(CURD重复)，dao就是实现对持久化数据的操作(包括连接数据库，创建数据表，插入数据，查询数据，更新数据，删除数据等)

* 编写DAO组件规范
  1. 定义DAO接口 (知识扩展：面向接口编程，根据客户提出的需求定义接口，业务具体实现是通过实现类来完成，当客户提出新的需求时，只需要编写该业务逻辑新的实现类即可，例如我们可以定义数据库操作的接口(CURD等)，具体实现可以再实现类中实现，是mysql就写关于mysql的实现类，oracle就写oracle的实现类。面向接口编程的好吃非常多，例如上面的数据库接口满足不同的数据库接口实现类，还有利于代码的维护，降低耦合)
  
  2. 编写接口实现类
  
  3. 包名的规范(域名.dao.domain包放domain、域名.dao包放所有dao接口、域名.dao.impl放所有dao接口实现类、域名.dao.test放dao组件的测试类)
  
  4. 类名的规范（domain类用于描述一个对象，是一个javaBean，类名应对应数据库中的表名（如：Student）; dao接口类，用于表示某一个对象的CURD声明，取名规范：IDomainDao（Domain: 是domain的名字，如：IStudentDao）; dao接口实现类要实现dao接口，取名规范：DomainDaoImpl（如：StudentDaoImpl））

### ORM(对象关系映射)

>对象关系映射(ORM)，就是将关系数据库中表的记录映射成为对象，以对象的形式展现，因此ORM的目的是为了方便开发人员以面向对象的思想来实现对数据库的操作。只是一种规范，一种概念
对应关系：
面向对象概念：    面向关系概念：
类               表
对象             记录(行)
属性             字段(列)

用户调用数据查询语句时：ORM会先创建一个与数据库名对象的类(表)，包括其中的属性(字段)，然后new一个刚创建的ORM类类对象(记录)出来，将数据库返回的数据赋给对象，最后返回给用户。

### domain

>domain就是一个类，是ORM的一个类，就是这个类的属性等和数据表的字段相对应，这个类符合javaBean规范(就是一个类当中有属性并且有该属性的get和set方法)；其作用就是用户与数据库交互的核心中转站；例如用户在查入数据时，我们就new一个domain类，然后将插入的数据0赋给domain类对象的对应属性(前面已经介绍过domain类是符合ORM关系映射的类，所以它的属性就和数据库里对应表的字段对应)，然后将domain类的对象传给DAO进行数据库语句拼接，再执行数据语句，最后反馈给用户就完成了一次数据库操

### util(工具类)

>将实现类中重复的属性提取出来作为工具类中的static静态属性，将重复代码提取出来作为工具类中的静态方法

### 静态代码块static

>将每个类都重复执行(实则只需要执行一次)的程序放入工具类的static(静态代码块中)即可。(注意：静态代码块不能在属性声明或初始化之前，否则属性声明将或初始化不会被执行，虽然语法检查时没有报错，但是执行时程序将会报错(重要))

```

语法：
static {
  // 静态代码块执行的代码……
}

静态代码实例：加载数据库驱动
static {
  try{
    Class.forName("com.mysql.jdbc.Driver");
  }catch(Exception e){
    e.printStack();
  }
}

```

### 预编译

#### Statement接口

* 接口：用于java程序和数据库之间的数据传输

* 3个实现类

  1. Statement :用于对数据库进行通用访问，使用的是静态sql语句
  2. PreparedStatement:
  3. CallableStatement:用于预编译模板sql语句，在运行时接受sql输入参数

#### 预编译语句

PreparedStatment:用于预编译sql语句；
在性能和代码灵活性上有显著的提升；
PreparedStatement对象使用?作为占位符，即参数标记；
使用setXxx(index, value)方法将值绑定到参数中，每个参数标记是其顺序位置引用，注意index从1开始；
PreparedStatement对象执行sql语句

* 当不适用预编译语句是(使用字符串与参数拼接的形式，即静态sql语句)，语句的执行过程：sql语句通过Connection传到数据库服务器 -> 创建Statement语句 -> 分析sql语句(做安全分析和语法分析) -> 到预编译语句池中查看有没有相同(参数不同也为不同语句)的语句，没有就要将语句加入到预编译语句池中，有就直接执行语句池中的语句 -> Mysql.exe执行语句 -> 返回ResultSet结果集。

* Mysql是不支持预编译语句的，但是写法是支持的，即用?号拼接的语句和静态sql语句在性能上没有太大差别。但是用预编译语句传参数比较方便。虽然性能没有提升，但是更安全，比如防sql注入。

* Oracal数据库支持预编译语句，预编译语句的效率明显高于静态sql语句，而且更安全，如防sql注入。

* sql注入：在传统数据库语句拼接过程中(静态sql语句)，用户在表单提交参数时将sql语句作为参数传给数据库服务器，数据库将作为参数的sql语句字符串于原字符串的sql语句拼接成为新语句从而非法获取到数据库中的信息。

```
sql注入示例：

数据库中的信息：id=1, name="ipso", pass="123456"

String login(String name, String pass) throws SQLException {
  Connection conn = JDBCUtil.getConnection();
  String sql = "select * from user where name = '"+name+"' and pass = '"+pass+"'";
  System.out.println(sql); // 打印拼接后的sql语句
  Statement st = conn.createStatement();
  ResultSet res = st.executeQuery(sql);

  if (res.next()){
     JDBCUtil.close(conn, st, res);
     return "登录成功";
  }
  else {
     JDBCUtil.close(conn, st, res);
     return "登录失败";
  }
}

单元测试：
@Test
public void test() throws SQLException {
  将字符串"' OR 1=1 OR '"当做name参数去login，密码随意给
   System.out.println(login("' OR 1=1 OR '", "123456s"));
}

执行结果：
select * from user where name = '' OR 1=1 OR '' and pass = '123456s'
登录成功

使用预编译sql打印的拼接sql语句：select * from user where name = '\' OR 1=1 OR \'' and pass = '123456s'

执行结果分析：

我们本来的sql语句："select * from user where name = '"+name+"' and pass = '"+pass+"'";

用户传入参数后打印出来的sql语句：select * from user where name = '' OR 1=1 OR '' and pass = '123456s'

可以看到用户传入参数后将我们原本的sql语句彻底改变，改变后的语句where的前两个条件为true后面的条件将不会考虑，所以不管密码为什么都会登录成功。这就是sql注入，静态sql语句的不安全性。可以看到用预编译的sql语句时，会将参数中的特殊字符进行转译所以拼接后不会构成新的sql语句.比较安全。

```

```

预编译语句代码示例：数据库删除操作(这里不展示util工具类)

void delete(int id)
{
  Connection conn = null;
  PreparedStatement ps = null;

  try {
     // 通过工具类获取Connection对象
     conn = JDBCUtil.getConnection();

     String sql = "delete from stu where id = ?";
     // PreparedStatment预编译sql语句
     ps = conn.prepareStatement(sql);
     ps.setInt(1, id); //将值绑定到参数中，每个参数标记是其顺序位置引用，注意index从1开始

     // 执行数据库语句
     ps.executeUpdate();

     // 查看执行的sql语句
     System.out.println(((JDBC4PreparedStatement)ps).asSql());

  }catch (Exception e){
     e.printStackTrace();
     return false;
  }finally {
     // 释放资源因为ps的PreparedStatement对象时Statement的子接口所以可以传ps给Statement对象
     JDBCUtil.close(conn, ps,null);
  }
}

// 单元测试
@Test
void delete(int id){
  delete(id);
  System.out.println("删除成功");
}

```

### 调用存储过程

* 在数据库中定义存储过程

```

// 创建一个参数的存储过程
use jdbc_db;
delimiter $$    // 定义新的结束标识
create procedure getStudent(IN n varchar(50))
  begin
    select * from stu where name=n;
  end;
delimiter ;     // 改回默认的结束标识


// 创建带输出参数的存储过程
use jdbc_db;
delimiter $$    // 定义新的结束标识
create procedure getName(in i int , out n varchar(50))
  begin
    select name into n from stu where  id = i;

  end$$
delimiter ;     // 改回默认的结束标识

```

* JDBC调用一个参数的存储过程

```
// 连接数据库
Connection conn = JDBCUtil.getConnection();
// 调用存储过程
CallableStatement cs =  conn.prepareCall("{ call getStudent(?)}");
// 设置参数
cs.setString(1, "ipso1");
// 执行存储过程
ResultSet res = cs.executeQuery();
if (res.next()){
   Student stu = new Student();
   stu.setId(res.getInt("id"));
   stu.setName(res.getString("name"));
   stu.setAge(res.getInt("age"));
   System.out.println(stu);
}
```

* 编写输入参数和输出参数的存储过程

```
// 编写输入参数和输出参数的存储过程
// 连接数据库
Connection conn = JDBCUtil.getConnection();
// 调用存储过程
CallableStatement cs =  conn.prepareCall("{ call getName(?,?)}");
// 设置参数
cs.setInt(1, 6);
cs.registerOutParameter(2, Types.VARCHAR);
// 执行存储过程
cs.execute();
String name = cs.getString(2);
System.out.println(name);

```

* JDBC调用两个参数的存储过程

### 事务

>不可分割的操作，假设有abcd四个步骤组成
系统默认每条语句都是一个事务，所以系统默认是每一句sql语句就提交一次；
事务只对DML语句有效，对DQL无效。还有就是MyIsAM不支持外键和事务，InnoDB才支持外键和事务。

myisam支持全文检索(fulltext),操作表的时候是表级锁，不支持事务，日志，外键等。
innodb则支持事务处理，日志，外键。但不支持全文检索。操作表的时候是行级锁。
由于有事务和日志，所以innodb在添加和修改的时候数据更安全，但是读取速度较慢。
InnoDB的AUTOCOMMIT默认是打开的，即每条SQL语句会默认被封装成一个事务，
自动提交，这样会影响速度，所以最好是把多条SQL语句显示放在begin和commit之间，组成一个事务去提交。

mysql数据库，默认的存储引擎是myisam,不支持事务和外键，我们需要去改变数据库的存储引擎(改为innodb)来支持事务。

事务的ACID特性:
原子性(Atomicity)：
一致性(Consistency)：
隔离性(Isolation)：
持久性(Durability):

事务处理过程：
关闭自动提交: connection.setAutoCommit(false);
没有问题时提交事务；
出现异常时进行回滚操作；
只要增、删、改(即DML)需要事务，查询不需要事务(即DQL)
可以设置事务隔离级别，一般使用默认的事务隔离级别。

```java

// 银行转账问题
描述：A申请转账给B：A将申请提交银行，银行收到A的申请，查询余额，充足
对A账户进行扣钱处理，并进行B账户加钱处理。通知B,否则驳回申请。完成交易。

<!-- 无事务处理 -->

// 连接数据库
Connection conn = JDBCUtil.getConnection();

// 1. 检查zs账户余额
String sql = "select * from account where name = ? and money > ?";
PreparedStatement ps = conn.prepareStatement(sql);
ps.setString(1, "zs");
ps.setInt(2, 1000);
ResultSet res = ps.executeQuery();
if (!res.next()){
   throw new RuntimeException("余额不足");
}

// 开启事务
// 开启事务，将自动提交事务关闭就是开启事务，默认是一句sql语句就是一个事务，
// 手动提交后遇到connection.commit()时为一个事务。
conn.setAutoCommit(false); // 将自动提交关闭，下面手动提交。就是开启事务

// 2. 减少zs账户1000
sql = "update account set money = money - ? where name = ?";
ps = conn.prepareStatement(sql);
ps.setInt(1, 1000);
ps.setString(2, "zs");
ps.executeUpdate();

// 这里让程序抛出一个算术异常，如果没有开启事务，上面减钱的操作正常执行，而下面加钱操作将不会执行，并且资源还会被一直占用着。
// 开启事务后上面执行的语句我们可以通过commit提交事务，出现异常时用rollback回滚，并释放资源。
int a = 1/0;

// 3. 增加ls账户1000
sql = "update account set money = money + ? where name = ?";
ps = conn.prepareStatement(sql);
ps.setInt(1, 1000);
ps.setString(2, "ls");
ps.executeUpdate();
conn.commit(); // 手动提交事务，从'conn.setAutoCommit(false);'开始到此处为一个事务。
// 释放资源
JDBCUtil.close(conn, ps, res);


 /* 事实上我们在项目中都会捕获事务中的异常，正常的代码格式如下 */

 /**
  * 事务方式处理转账问题
  * @throws SQLException
  */
 void transaction() throws SQLException {

    Connection conn = null;
    PreparedStatement ps = null;
    ResultSet res = null;

    // 连接数据库
    conn = JDBCUtil.getConnection();

    // 1. 检查zs账户余额
    String sql = "select * from account where name = ? and money > ?";
    ps = conn.prepareStatement(sql);
    ps.setString(1, "zs");
    ps.setInt(2, 1000);
    res = ps.executeQuery();
    if (!res.next()){
       throw new RuntimeException("余额不足");
    }

    try{

       // 开启事务，将自动提交事务关闭就是开启事务，默认是一句sql语句就是一个事务，
       // 手动提交后遇到connection.commit()时为一个事务。
       conn.setAutoCommit(false);

       // 2. 减少zs账户1000
       sql = "update account set money = money - ? where name = ?";
       ps = conn.prepareStatement(sql);
       ps.setInt(1, 1000);
       ps.setString(2, "zs");
       ps.executeUpdate();

       // 这里让程序抛出一个算术异常，如果没有开启事务，上面减钱的操作正常执行，
       // 而下面加钱操作将不会执行，开启事务后上面执行的语句将被回滚。
       int a = 1/0;

       // 3. 增加ls账户1000
       sql = "update account set money = money + ? where name = ?";
       ps = conn.prepareStatement(sql);
       ps.setInt(1, 1000);
       ps.setString(2, "ls");
       ps.executeUpdate();
       conn.commit(); // 提交事务

    }catch (Exception e){
       e.printStackTrace();
       conn.rollback(); // 出现异常时回滚
    }finally {
       JDBCUtil.close(conn, ps, res);
    }
 }

```

### 批处理

待续……

## #源码仓库

[JDBC-DEMO](https://github.com/ipsozzZ/JDBC-DEMO)
