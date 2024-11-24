import concurrent.futures
import pymysql
 
# 连接MySQL数据库的配置
mysql_config = {
    'host': '127.0.0.1',
    'user': 'root',
    'password': 'pwd',
    'database': 'test'
}
 
# 创建数据库连接
def create_connection():
    connection = pymysql.connect(**mysql_config)
    return connection
 
# 执行事务
def execute_transaction(connection, transaction):
    with connection.cursor() as cursor:
        cursor.execute(transaction)
        connection.commit()
 
# 主程序
def main():
    # 创建连接
    connection1 = create_connection()
    connection2 = create_connection()
 
    # 定义两个事务
    transaction1 = "select id from t1 where c in(5,20,10) lock in share mode;"
    transaction2 = "select id from t1 where c in(5,20,10) order by c desc for update;"
 
    # 使用ThreadPoolExecutor并行提交
    with concurrent.futures.ThreadPoolExecutor() as executor:
        future1 = executor.submit(execute_transaction, connection1, transaction1)
        future2 = executor.submit(execute_transaction, connection2, transaction2)
 
        # 等待任务完成
        concurrent.futures.wait([future1, future2])
 
# 运行主程序
if __name__ == '__main__':
    main()