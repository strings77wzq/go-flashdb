# TCP 服务器：goroutine-per-connection 的艺术

本章将深入讲解 go-flashdb 的网络层实现，揭示如何用 Go 语言的 goroutine 构建高性能 TCP 服务器。

## 为什么从 TCP 服务器开始？

Redis 是一个网络服务，所有操作都通过网络传输。理解 TCP 服务器的实现是理解 Redis 的第一步。

在学习本章之前，建议先了解：
- Go 语言基础语法
- TCP/IP 协议基础
- Socket 编程概念

## 架构选择：goroutine-per-connection

### 历史演进

```
阻塞 IO (Apache) → IO 多路复用 (Nginx) → 协程模型 (Go)
     ↓                    ↓                   ↓
  一个线程一个连接      epoll/kqueue       一个协程一个连接
  上下文切换开销大      编程复杂度高        轻量 + 简洁
```

### Go 的网络模型优势

Go 的 `netpoller` 基于 IO 多路复用（epoll/kqueue），但给开发者提供了**同步编程的接口**：

```go
// 看起来像阻塞代码，实际底层是异步的
conn, err := listener.Accept()  // 阻塞等待新连接
msg, err := reader.ReadString('\n')  // 阻塞等待数据
```

这种**简洁性**是 go-flashdb 选择 Go 语言的重要原因。

## 核心实现

### 1. 服务器初始化

```go
// pkg/net/server.go

type Server struct {
    mu          sync.Mutex
    addr        string
    listener    net.Listener
    running     bool
    wg          sync.WaitGroup
    closeCh     chan struct{}
    readyCh     chan struct{}
    db          *core.DB
    auth        *security.Authenticator
    rateLimiter *security.RateLimiter
    filter      *security.CommandFilter
    persistMgr  *persist.PersistManager
    tlsConfig   *tls.Config
}
```

**设计要点**：
- `sync.WaitGroup`：跟踪所有连接 goroutine
- `closeCh`：优雅关闭的信号通道
- `readyCh`：测试时等待服务启动

### 2. 监听与接受连接

```go
func (s *Server) Start() error {
    var listener net.Listener
    var err error

    // 支持 TLS
    if s.tlsConfig != nil {
        listener, err = tls.Listen("tcp", s.addr, s.tlsConfig)
    } else {
        listener, err = net.Listen("tcp", s.addr)
    }
    
    if err != nil {
        return err
    }
    
    s.listener = listener
    s.mu.Lock()
    s.running = true
    s.mu.Unlock()
    close(s.readyCh)  // 通知测试：服务已启动

    // 主循环：接受连接
    for {
        s.mu.Lock()
        running := s.running
        s.mu.Unlock()
        
        if !running {
            break
        }
        
        conn, err := listener.Accept()
        if err != nil {
            select {
            case <-s.closeCh:
                return nil  // 正常关闭
            default:
                return err  // 异常错误
            }
        }
        
        // 每个连接一个 goroutine
        s.wg.Add(1)
        go s.handleConn(conn)
    }
    return nil
}
```

**关键设计决策**：

1. **为什么用 `for` 循环而不是 `select`？**
   
   `listener.Accept()` 本身会阻塞，不需要 `select`。使用 `for` 循环更简洁。

2. **为什么先检查 `s.running` 再 Accept？**
   
   避免在关闭过程中接受新连接。通过锁保护 `running` 状态，确保线程安全。

3. **为什么用 `s.wg.Add(1)`？**
   
   每个连接 goroutine 都需要被跟踪，保证优雅关闭时等待所有连接处理完成。

### 3. 连接处理

```go
func (s *Server) handleConn(conn net.Conn) {
    defer func() {
        conn.Close()
        s.wg.Done()
    }()

    // 记录客户端信息
    clientAddr := conn.RemoteAddr().String()
    clientID := core.AddClient(clientAddr, 0)
    defer core.RemoveClient(clientID)

    // 使用 bufio 提高读取性能
    reader := bufio.NewReader(conn)
    parser := resp.NewParser(reader)
    
    // 认证状态
    authenticated := s.auth == nil || !s.auth.IsEnabled()

    for {
        // 检查服务器是否关闭
        s.mu.Lock()
        running := s.running
        s.mu.Unlock()
        if !running {
            break
        }

        // 解析请求
        reply, err := parser.Parse()
        if err != nil {
            return
        }

        // 转换为命令
        arrayReply, ok := reply.(*resp.ArrayReply)
        if !ok || len(arrayReply.Replies) == 0 {
            conn.Write(resp.NewErrorReply("ERR invalid command").ToBytes())
            continue
        }

        cmdName, args := parseCommand(arrayReply)
        
        // 更新客户端活跃状态
        core.UpdateClient(clientID, cmdName)

        // 限流检查
        if s.rateLimiter != nil && !s.rateLimiter.Allow(clientAddr) {
            conn.Write(resp.NewErrorReply("ERR rate limit exceeded").ToBytes())
            continue
        }

        // 认证检查
        if !authenticated {
            if cmdName == "auth" {
                authenticated = handleAuth(s.auth, args)
                if authenticated {
                    conn.Write(resp.OkReply.ToBytes())
                } else {
                    conn.Write(resp.NewErrorReply("ERR invalid password").ToBytes())
                }
            } else {
                conn.Write(resp.NewErrorReply("NOAUTH Authentication required").ToBytes())
            }
            continue
        }

        // 命令过滤
        if s.filter != nil {
            if s.filter.IsBlocked(cmdName) {
                conn.Write(resp.NewErrorReply("ERR command '" + cmdName + "' is blocked").ToBytes())
                continue
            }
            cmdName = s.filter.Rename(cmdName)
        }

        // 执行命令
        startTime := time.Now()
        result := s.db.Exec(cmdName, args)
        
        // 记录慢查询
        duration := time.Since(startTime)
        core.AddSlowLog(cmdName, argsToStrings(args), duration)

        // 处理 SHUTDOWN 命令
        if result == nil {
            go s.handleShutdown()
            return
        }

        // 发送响应
        conn.Write(result.ToBytes())
    }
}
```

**设计亮点**：

1. **`bufio.NewReader`**：
   - 减少系统调用次数
   - 内置缓冲区（默认 4KB）
   - 支持 `ReadBytes('\n')` 等便捷方法

2. **延迟关闭（defer）**：
   - 确保连接关闭
   - 确保 WaitGroup 计数减一
   - 确保客户端从列表移除

3. **分层处理**：
   - 网络层：解析请求、发送响应
   - 安全层：认证、限流、过滤
   - 核心层：执行命令

### 4. 优雅关闭

```go
func (s *Server) Close() {
    s.mu.Lock()
    if !s.running {
        s.mu.Unlock()
        return
    }
    s.running = false
    s.mu.Unlock()
    
    // 1. 通知主循环停止
    close(s.closeCh)
    
    // 2. 关闭监听器（让 Accept 返回错误）
    if s.listener != nil {
        s.listener.Close()
    }
    
    // 3. 关闭持久化管理器
    if s.persistMgr != nil {
        s.persistMgr.Close()
    }
    
    // 4. 等待所有连接 goroutine 结束
    s.wg.Wait()
}
```

**优雅关闭流程**：

```
1. 设置 running = false
       ↓
2. 关闭 closeCh（通知主循环）
       ↓
3. 关闭 listener（Accept 返回）
       ↓
4. 关闭 persistMgr（保存数据）
       ↓
5. 等待 wg.Wait（等待所有连接）
       ↓
6. 关闭完成
```

**为什么需要优雅关闭？**

- 避免数据丢失：确保正在执行的命令完成
- 避免客户端错误：确保客户端收到完整响应
- 资源清理：关闭文件描述符、释放内存

## 关键问题解析

### Q1: 如何处理粘包/拆包？

**问题描述**：

TCP 是字节流协议，不保证消息边界。可能出现：
- 粘包：收到 "*3\r\n$3\r\nSET\r\n$3\r\nkey..."（多条消息粘在一起）
- 拆包：收到 "*3\r\n$3\r\nSET"（不完整消息）

**解决方案**：

RESP 协议天然解决了这个问题：

```
*3\r\n      ← 数组长度（知道要读几个元素）
$3\r\n     ← 字符串长度（知道要读多少字节）
SET\r\n    ← 实际内容
$3\r\n     ← 下一个元素
key\r\n    ← 实际内容
...
```

解析器根据长度字段精确读取，不依赖消息边界。

### Q2: 一个连接一个 goroutine 会不会太耗资源？

**Go 协程 vs OS 线程**：

| 特性 | Goroutine | OS 线程 |
|------|-----------|---------|
| 栈大小 | 2KB（可增长） | 1-8MB |
| 切换开销 | ~200ns | ~1-2μs |
| 创建开销 | ~2μs | ~100μs |
| 数量 | 百万级 | 万级 |

**结论**：
- 10 万个连接 = 10 万个 goroutine = ~200MB 内存
- 完全可行，且性能优秀

### Q3: 如何处理大量空闲连接？

**空闲连接占用资源**：
- 每个 goroutine ~2KB 栈
- 每个连接 ~4KB bufio 缓冲区

**解决方案**：

1. **设置超时**：
```go
conn.SetReadDeadline(time.Now().Add(300 * time.Second))
```

2. **心跳检测**：
```go
// 客户端定期发送 PING
// 服务器更新 idle time
// 超过 timeout 主动断开
```

3. **连接池**（客户端）：
```go
// 复用连接，减少新建连接开销
```

## 性能优化技巧

### 1. 使用 bufio

```go
// 不使用 bufio：每次 read 都是系统调用
reader := conn
buf := make([]byte, 1024)
n, err := reader.Read(buf)  // 系统调用

// 使用 bufio：减少系统调用次数
reader := bufio.NewReader(conn)
line, err := reader.ReadBytes('\n')  // 可能从缓冲区读取
```

### 2. 复用缓冲区

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func handleConn(conn net.Conn) {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    // 使用 buf...
}
```

### 3. 设置 TCP KeepAlive

```go
tcpConn := conn.(*net.TCPConn)
tcpConn.SetKeepAlive(true)
tcpConn.SetKeepAlivePeriod(3 * time.Minute)
```

## 对比其他实现

### Redis (C 语言)

```c
// Redis 使用单线程 + IO 多路复用
// 优点：无锁编程简单
// 缺点：无法利用多核
```

### godis (Go 语言)

与 go-flashdb 类似，也使用 goroutine-per-connection。

**go-flashdb 的优势**：
1. 更清晰的代码结构
2. 更完善的安全模块
3. AI 扩展接口预留

## 总结

### 核心设计决策

1. **goroutine-per-connection**：
   - 优点：代码简洁、易于理解
   - 代价：比单线程模型略高的内存占用
   - 适用：连接数 < 10 万的场景

2. **分层架构**：
   - Network → Protocol → Security → Core
   - 每层职责单一

3. **优雅关闭**：
   - 使用 WaitGroup 跟踪连接
   - 确保资源释放

### 学到的技能

- Go 网络编程
- 并发模式
- 优雅关闭设计
- 性能优化技巧

## 参考

- [源码: pkg/net/server.go](https://github.com/strings77wzq/go-flashdb/blob/main/pkg/net/server.go)
- [Go 网络轮询器](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-netpoller/)
- [Redis 网络模型](https://redis.io/topics/internals-rediseventlib)

---

**下一章**：[RESP 协议解析器](/guide/03-resp-protocol.html) - 深入了解二进制安全协议的设计