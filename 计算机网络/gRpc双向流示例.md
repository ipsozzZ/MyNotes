以下是一个使用 **TypeScript 客户端**通过 gRPC 双向流与 **Golang 服务端**通信的完整示例，涵盖 `.proto` 定义、代码生成、服务端和客户端实现：

---

### 一、环境准备

1. **安装工具链**：
   - **protoc**：Protocol Buffers 编译器。
     ```bash
     # MacOS
     brew install protobuf

     # Linux (Ubuntu)
     apt install -y protobuf-compiler
     ```
   - **Golang 插件**：
     ```bash
     go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
     go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
     ```
   - **TypeScript 插件**：
     ```bash
     npm install -g grpc-tools ts-protoc-gen
     ```

---

### 二、定义服务接口 (`chat.proto`)

```protobuf
syntax = "proto3";

package chat;

// 定义双向流式消息
message ChatMessage {
  string user = 1;
  string text = 2;
}

// 定义服务
service ChatService {
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}
```

---

### 三、生成代码

#### 1. 生成 Golang 服务端代码
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       chat.proto
```
生成文件：`chat.pb.go`, `chat_grpc.pb.go`

#### 2. 生成 TypeScript 客户端代码
```bash
protoc --plugin=protoc-gen-ts=./node_modules/.bin/protoc-gen-ts \
       --js_out=import_style=commonjs,binary:./src \
       --ts_out=service=grpc-web:./src \
       chat.proto
```
生成文件：`src/chat_pb.d.ts`, `src/chat_pb.js`

---

### 四、Golang 服务端实现 (`server.go`)

```go
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	pb "path/to/your/package" // 替换为你的包路径
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
}

// 实现双向流式 RPC
func (s *chatServer) Chat(stream pb.ChatService_ChatServer) error {
	for {
		// 接收客户端消息
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Printf("Received from client: %s: %s", msg.User, msg.Text)

		// 构造响应并发送
		response := &pb.ChatMessage{
			User: "Server",
			Text: "Echo: " + msg.Text,
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterChatServiceServer(s, &chatServer{})
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

---

### 五、TypeScript 客户端实现 (`client.ts`)

```typescript
import * as grpc from '@grpc/grpc-js';
import * as chat_pb from './chat_pb';
import { ChatServiceClient } from './chat_grpc_pb';

// 创建客户端连接
const client = new ChatServiceClient(
  'localhost:50051',
  grpc.credentials.createInsecure()
);

// 创建双向流
const stream = client.chat();

// 发送消息到服务端
stream.write(
  new chat_pb.ChatMessage()
    .setUser('Client')
    .setText('Hello from TypeScript!')
);

// 接收服务端响应
stream.on('data', (response: chat_pb.ChatMessage) => {
  console.log(`Received from server: ${response.getUser()}: ${response.getText()}`);
});

// 处理错误和结束事件
stream.on('error', (err) => {
  console.error('Stream error:', err);
});
stream.on('end', () => {
  console.log('Stream ended');
});

// 保持进程运行（示例中简单等待）
setTimeout(() => {}, 5000);
```

---

### 六、运行测试

1. **启动 Golang 服务端**：
   ```bash
   go run server.go
   ```

2. **运行 TypeScript 客户端**：
   ```bash
   ts-node client.ts
   ```

3. **观察输出**：
   - 服务端日志：
     ```
     Received from client: Client: Hello from TypeScript!
     ```
   - 客户端输出：
     ```
     Received from server: Server: Echo: Hello from TypeScript!
     ```

---

### 七、关键点说明

1. **双向流生命周期**：
   - 客户端和服务端通过 `stream` 对象独立发送/接收消息。
   - 流保持打开状态，直到任一方调用 `end()` 或连接中断。

2. **TypeScript 依赖**：
   ```bash
   npm install @grpc/grpc-js @types/node ts-node typescript
   ```

3. **扩展功能**：
   - **错误处理**：添加重试逻辑和状态码检查。
   - **心跳机制**：定期发送 Ping 消息保持连接活跃。
   - **元数据传递**：通过 `metadata` 添加认证信息。

---

通过这个示例，你可以快速实现一个支持双向实时通信的 gRPC 应用！