# Netty 第三方网络编程包
JDK原生的TCP、UDP编程即BIO、NIO、AIO编程功能虽然全面但是不够强大，Java使用最广泛的网络编程包是第三方提供的netty包
### 关键技术：
1. 通道 Channel
- ServerSocketChannel/NioServerSocketChannel/...     (在NIO和AIO编程中ServerSocketChannel是用来接收客户端请求的，这里类似)
- SocketChannel/NioSocketChannel        (真正用于通讯的Channel，这里也和NIO、AIO相似)

2. 事件驱动 EventLoop
- 为每一个通道定义一个EventLoop，处理所有的I/O事件
- EventLoop事件注册
- EventLoop将事件配发给ChannelHandler
- EventLoop安排进一步工作

3. 事件
- 事件按照数据流向进行分类
- 入站事件：连接激活/数据读取/...
- 出站事件：打开到远程连接/写数据/...

4. 事件处理 ChannelHandler
- Channel通道发生数据或状态改变
- EventLoop会将事件分类，并调用ChannelHandler的回调函数
- 程序员需要实现ChannelHandler内的回调函数
- ChannelInboundHandler/ChannelOutboundHandler  (对应入站事件和出站事件)

5. ChannelHandler工作模式：责任链模式
- 责任链
-- 将请求的接收者连成一条链
-- 在链上传递请求，直到有一个接收者处理该请求
-- 避免请求者和接收者的耦合
- ChannelHandler可以有多个，依次进行调用
- ChannelPipeline作为容器，承载多个ChannelHandler

6. ByteBuf
- 强大的字节容器，提供丰富的API进行操作，并没有使用JDK的ByteBuffer，而是netty自己实现的功能更为强大的字节buffer