# Jdbc数据库编程基础
注意jdbc中很多集合下标都是从1开始的如ResulSetMetaData

## # 基础
1. Java SQL操作包
- java.sql.*和javax.sql.*；这两个包只是接口类
- 根据数据库版本和Jdbc版本合理选择
- 一般数据库发行包都会提供jar包。
- 连接字符串（例：Mysql："jdbc:mysql://localhost:3306/myDB"; Oracle: "jdbc:oracle:thin:@127.0.0.1:3306:myDB"。等....）

2. java连接数据库的步骤
- 构建连接
-- 注册驱动，确定数据库(class.forName("数据库驱动")); // 数据库驱动如Mysql：com.mysql.jdbc.Driver
-- 确定连接目标，建立连接(Connection)

- 执行操作
-- Statement（执行者）
-- ResultSet（结果集）

- 释放连接
-- connection.close();

3. Statement（执行者）
- 使用executeQuery() 执行select语句，返回结果放在ResultSet中
- 使用executeUpdate() 执行insert/update/delete语句，返回修改的行数
- 一个Statement对象一次只能执行一个命令

4. ResulSet（结果集）Java使用ResulSetMetaData来获取ResulSet返回的属性（如每一行的名字、类型等）
- next()判断是否还有下一条记录
- getInt()/getSting()/getDouble()/.....
- ResultSet不能多个做笛卡尔积连接（就是不能有个ResultSet对象for循环嵌套使用）
- ResulSet最好不要超过百条，否则极其影响性能
- ResulSet也不是一次性加载所有的select结果数据
- Connection很昂贵，需要即使close
- Connection所用的jar包和数据库要匹配

## # jdbc高级
1. 事务处理
- 作为单个逻辑单元执行的一系列操作
- 事务必须满足ACID（原子性、一致性、隔离性、持久性）属性
- 事务是数据库运行中的逻辑工作单位，由DBMS中的事务管理子系统负责事务的处理
- jdbc实现事务需要先关闭自动提交connection.setAutoCommit(false); 使用connection.commit()提交事务，connection.rollback()回滚事务
- 保存点机制connection.setSavepoint()设置保存点，使用connection.rollback(savepoint)回滚到保存点


2. PreparedStatement更为安全的执行sql语句
- PreparedStatement和Statement的区别是使用"?"号代替字符串拼接
- PreparedStatement使用setXXX(int, Object)的函数来实现对于"?"的替换
- 提供addBatch批量更新功能
- select语句一样使用ResultSet接收结果集

## # 数据库连接池

1. 享元模式(23中设计模式中的一种，属结构性模式)
享元模式：一个系统中存在大量的相同的对象，由于这类对象的大量使用，会造成系统内存的消耗，可以使用享元模式来减少系统中对象的数量

2. 数据库连接的构建成本很高，单次使用成本昂贵，利用共享技术来实现数据连接池(享元模式)
- 降低系统中数据库连接Connection对象的数量
- 降低数据库服务器的连接响应消耗
- 提高Connection获取的响应速度

3. 理解池pool的概念
- 初始值、最大数量、增量、超时时间等参数

4. 常用的数据库连接池
- DBCP（Apache， 性能较差）
- C3P0 （https://mchange.com/projects/c3p0/）
- Druid (Alibaba, https://github.com/alibaba/druid)


5. C3P0为例

1. 参数
- driverClass, 驱动class，指数据库驱动，如mysql：com.mysql.jdbc.Driver
- jdbcUrl, jdbc连接
- user password 数据库用户名密码
- initiaPoolSize, 初始数量：一开始建立多少条连接
- maxPoolSize, 最大数量：最多有多少条连接
- acquireIncrement, 增量：用完每次增加多少个
- maxIdleTime, 最大空闲时间：超出的连接会被抛弃