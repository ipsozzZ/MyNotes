

## # 唯一索引和普通索引的区别（9）

由于唯一索引用不上 change buffer 的优化机制，因此如果业务可以接受，从性能角度出发我建议你优先考虑非唯一索引。

## # mysql 选错索引的情况，索引统计更新的机制（10）

优化器的扫描行数是不准的，是采样统计的结果

对于其他优化器误判的情况，可以在应用端用 force index 来强行指定索引，也可以通过修改语句来引导优化器，还可以通过增加或者删除索引来绕过这个问题。


## # 字符串索引怎么建？(11)
字符串字段创建索引的场景。你可以使用的方式有：
1. 直接创建完整索引，这样可能比较占用空间；
2. 创建前缀索引，节省空间，但会增加查询扫描次数，并且不能使用覆盖索引；
3. 倒序存储，再创建前缀索引，用于绕过字符串本身前缀的区分度不够的问题；
4. 创建 hash 字段索引，查询性能稳定，有额外的存储和计算消耗，跟第三种方式一样，都不支持范围扫描。

在实际应用中，要根据业务字段的特点选择使用哪种方式。


## # flush 刷脏页导致 mysql 性能抖动 (12)
利用 WAL 技术，数据库将随机写转换成了顺序写，大大提升了数据库的性能。但是，由此也带来了内存脏页的问题。脏页会被后台线程自动 flush，也会由于数据页淘汰而触发 flush，而刷脏页的过程由于会占用资源，可能会让你的更新和查询语句的响应时间长一些。在文章里，介绍了控制刷脏页的方法和对应的监控方式。



## # 删除数据 (13)
如果要收缩一个表，只是 delete 掉表里面不用的数据的话，表文件的大小是不会变的，还要通过 alter table 命令重建表，才能达到表文件变小的目的。重建表的两种实现方式，Online DDL 的方式是可以考虑在业务低峰期使用的，而 MySQL 5.5 及之前的版本，这个命令是会阻塞 DML 的，这个你需要特别小心。

## # count相关语句 (14)
按照效率排序的话，```count(字段)<count(主键 id)<count(1)≈count(*)，所以我建议你，尽量使用 count(*)。```


## # redo log 和 bin log （15）
两阶段提交逻辑

mysql crash 后的完整恢复流程

redo log 和 binlog 如何关联 （redolog 和 binlog 通过 xid 关联），当 mysql crash 后恢复时，会按顺序扫描redo log：
1. 如果碰到 prepare、又有 commit 的 log 就直接提交；
2. 如果碰到只有 prepare 而没有 commit 的 log 就拿着XID去binlog找对应的事务，并判断是否完整，如果不完整则事务回滚，否则提交并更新内存中的数据页
3. binlog 判断事务完整：statement格式的binlog最后会有COMMIT；row格式的binlog最后会有一个XID event。总之mysql 自己会判断binlog是否完整。


redo log 更新 buffer pool 的过程是：
1. redolog 并没有完整保存 buffer pool 中的每个数据页，而是保存了对每个数据页做了什么改变
2. 所以恢复时需要先从磁盘把数据页读到 buffer pool 中，然后根据 redolog 恢复 crash 之前的脏页，之后就和正常流程一样定期刷新脏页就行

## # 排序 （16~17）

1. 普通排序

mysql 会给排序分配一个 sort buffer，用来给 order by 排序使用，使用sort_buffer_size参数控制buffer大小

**如果排序内容小于sort_buffer_size，则排序使用快速排序在内存中进行，否则使用归并排序将排序内容存储到磁盘文件中，sort_beffer_size越小，排序所需文件数量越多，每个文件都存放有序的结果集，最终实行一次归并排序就能得到最终的排序结果**

2. rowid 排序

**如果查询的字段太多的话，sort buffer 中存储的无用字段太多，导致buufer承载的行数变少，为了优化，mysql实现了rowid 算法，就是将排序的列和主键列先取出来排序，然后根据主键字段遍历主键索引，获得其它没有参与排序的字段，得到最终排序结果**

注意，这种方式每行会多一次回表，假设语句 limit 1000，则会多1000 扫描1000次

所以万不得已 mysql 不会使用这个算法

3. optimizer_trace 可以将mysql执行的堆栈重要信息保存在 optimizer_trace 临时表，查询临时表就能看到这些信息，其中包括了排序所需的文件数量，排序是否使用了 rowid算法等，用法如下
```sql
/* 打开optimizer_trace，只对本线程有效 */
SET optimizer_trace='enabled=on'; 

/* @a保存Innodb_rows_read的初始值 */
select VARIABLE_VALUE into @a from  performance_schema.session_status where variable_name = 'Innodb_rows_read';

/* 执行语句 */
select city, name,age from t where city='杭州' order by name limit 1000; 

/* 查看 OPTIMIZER_TRACE 输出 */
SELECT * FROM `information_schema`.`OPTIMIZER_TRACE`\G

/* @b保存Innodb_rows_read的当前值 */
select VARIABLE_VALUE into @b from performance_schema.session_status where variable_name = 'Innodb_rows_read';

/* 计算Innodb_rows_read差值 */
select @b-@a;
```

**注意1：sort buffer 会对varchar(30)字段做“紧凑”处理，就是没有30个字符，会安装实际长度来分配空间**


## # 查询不会使用索引的情况（18）

1. 使用聚合函数（可能会扫描索引树，但是也是全索引扫描）

2. 查询条件中使用了计算公式（如 select * from t where a+1=3，但是改成 select * from t where a=3+1 之后是可以使用索引的）

3. 隐式类型转换（如：select * from t where a=300 如果a字段是varchar(32)，查询条件是整形的话就需要类型转换，这个时候也是不能用索引的；如果要使用索引，需要把a=300 改成 a='300'，这个时候就可以使用索引了），这里其实mysql内部是调用了一个类型转换的聚合函数，所以又回到了第一条，使用聚合函数查询，不会使用索引

**注意再mysql中，使用数字和字符串做比较的话，会将字符串转换成数字**

**注意，如果将上面的逻辑反过来：select * from t where a='300'，a字段类型是int，这个时候是会使用索引的，因为这个是查询条件字符串转int，索引字段不需要转，所以可以使用索引，优化器会帮助我们把查询条件中的a='300',优化为a=300**

4. 隐式字符编码转换（联表查询是，表字符编码不一致时，就算关联的字段在被驱动的表上有索引，也不能使用，因为需要先转格式，转格式的时候又会用到聚合函数）


## # 导致语句变慢的可能情况：（19）

1. 等待MDL读锁（DDL语句占用了表的MDL写锁，所有的DML语句都只能阻塞等待该表的MDL读锁）

2. 等待flush（刷脏页，redo log 满了，或者主动调用 flush tables 语句，导致mysql 停下所有工作去执行 flush， mysql 的 buffer pool 此时不可用，所有语句均不能再执行）

3. 等行锁

4. 慢查询（比如，undo log 太长，导致以来比较早的版本的事务，消耗比较长的时间来获取旧版本，导致undo log 边长的主要原因就是长事务；此外还有语句本身的问题，如没有使用正确的索引等不合理的语句）


## # 表锁复习 （之幻读） （20）

#### 表级锁分为：表锁和MDL（mata data lock）

1. 表锁的语法是 lock tables … read/write。锁粒度太大，一般不用

2. MDL读锁：正常执行DML（增删改查）语句时，mysql 会自动帮我们加上MDL读锁，读锁之间不会互相阻塞

3. MDL写锁：当对表结构调整时，会对表加MDL写锁，这个时候所有DML语句阻塞等待MDL读锁

#### 幻读有什么问题？
1. 首先是语义上，比如我要把所有 d=5 的行锁住，不准别的事务进行读写操作”。而实际上，这个语义被破坏了。

2. 其次，是数据一致性的问题。会导致备库在使用binlog写数据的时候，备库数据与主库数据不一致，这个问题很严重，是不行的。

我们知道，锁的设计是为了保证数据的一致性。而这个一致性，不止是数据库内部数据状态在此刻的一致性，还包含了数据和日志在逻辑上的一致性。

#### 幻读情况

1. 在RR隔离级别下：

**快照读不会产生幻读**

**当前读会产生幻读, innodb的解决办法引入GAP锁（间隙锁+行锁）**

2. 在RC隔离级别下：

**会产生幻读, binlog_format需要使用row格式，否则会出现数据和日志不一致的情况**

#### 解决幻读
现在知道了，产生幻读的原因是，行锁只能锁住行，但是新插入记录这个动作，要更新的是记录之间的“间隙”。因此，为了解决幻读问题，InnoDB 只好引入新的锁，也就是间隙锁 (Gap Lock)。

间隙锁和行锁合称 next-key lock，每个 next-key lock 是前开后闭区间。也就是说，我们的表 t 初始化以后，如果用 select * from t for update 要把整个表所有记录锁起来，就形成了 7 个 next-key lock，分别是 ```(-∞,0]、(0,5]、(5,10]、(10,15]、(15,20]、(20, 25]、(25, +supremum]```。

这个 supremum 从哪儿来的呢？这是因为 +∞是开区间。实现上，InnoDB 给每个索引加了一个不存在的最大值 supremum，这样才符合我们前面说的“都是前开后闭区间”。


间隙锁和 next-key lock 的引入，帮我们解决了幻读的问题，但同时也带来了一些“困扰”。

#### 间隙锁的问题




