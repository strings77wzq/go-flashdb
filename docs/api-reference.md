# GoSwiftKV API 参考手册

## 目录

- [连接命令](#连接命令)
- [字符串命令](#字符串命令)
- [键命令](#键命令)
- [服务器命令](#服务器命令)

---

## 连接命令

### PING

测试服务器连接是否正常。

**语法**
```
PING [message]
```

**返回值**
- 未指定 message：`PONG`
- 指定 message：返回 message 内容

**示例**
```bash
redis-cli ping
# PONG

redis-cli ping "hello"
# "hello"
```

---

## 字符串命令

### SET

设置指定 key 的值。

**语法**
```
SET key value
```

**返回值**
- 成功：`OK`

**示例**
```bash
redis-cli set mykey "myvalue"
# OK
```

---

### GET

获取指定 key 的值。

**语法**
```
GET key
```

**返回值**
- key 存在：返回 value
- key 不存在：返回 `nil`

**示例**
```bash
redis-cli get mykey
# "myvalue"

redis-cli get nonexistent
# (nil)
```

---

### SETNX

仅在 key 不存在时设置值。

**语法**
```
SETNX key value
```

**返回值**
- 设置成功：`1`
- key 已存在：`0`

**示例**
```bash
redis-cli setnx mykey "newvalue"
# (integer) 0  # 已存在，未设置

redis-cli setnx newkey "value"
# (integer) 1  # 设置成功
```

---

### SETEX

设置 key-value 并指定过期时间（秒）。

**语法**
```
SETEX key seconds value
```

**返回值**
- 成功：`OK`

**示例**
```bash
redis-cli setex tempkey 60 "expires in 60s"
# OK
```

---

### PSETEX

设置 key-value 并指定过期时间（毫秒）。

**语法**
```
PSETEX key milliseconds value
```

**返回值**
- 成功：`OK`

**示例**
```bash
redis-cli psetex tempkey 5000 "expires in 5s"
# OK
```

---

### MSET

批量设置多个 key-value。

**语法**
```
MSET key1 value1 [key2 value2 ...]
```

**返回值**
- 成功：`OK`

**示例**
```bash
redis-cli mset key1 value1 key2 value2 key3 value3
# OK
```

---

### MGET

批量获取多个 key 的值。

**语法**
```
MGET key1 [key2 ...]
```

**返回值**
- 返回数组，key 不存在则对应位置为 `nil`

**示例**
```bash
redis-cli mget key1 key2 key3 nonexistent
# 1) "value1"
# 2) "value2"
# 3) "value3"
# 4) (nil)
```

---

### INCR

将 key 中存储的数字值加一。

**语法**
```
INCR key
```

**返回值**
- 返回递增后的值
- key 不存在则初始化为 0 后递增
- 值不是整数则返回错误

**示例**
```bash
redis-cli set counter 10
redis-cli incr counter
# (integer) 11

redis-cli incr newcounter
# (integer) 1  # 不存在，从 0 递增
```

---

### DECR

将 key 中存储的数字值减一。

**语法**
```
DECR key
```

**返回值**
- 返回递减后的值

**示例**
```bash
redis-cli set counter 10
redis-cli decr counter
# (integer) 9
```

---

### INCRBY

将 key 中存储的数字值增加指定增量。

**语法**
```
INCRBY key increment
```

**返回值**
- 返回递增后的值

**示例**
```bash
redis-cli set counter 10
redis-cli incrby counter 5
# (integer) 15
```

---

### DECRBY

将 key 中存储的数字值减少指定减量。

**语法**
```
DECRBY key decrement
```

**返回值**
- 返回递减后的值

**示例**
```bash
redis-cli set counter 10
redis-cli decrby counter 3
# (integer) 7
```

---

### APPEND

将 value 追加到 key 原有值的末尾。

**语法**
```
APPEND key value
```

**返回值**
- 返回追加后的字符串长度

**示例**
```bash
redis-cli set mykey "Hello"
redis-cli append mykey " World"
# (integer) 11

redis-cli get mykey
# "Hello World"
```

---

### STRLEN

返回 key 所储存的字符串值的长度。

**语法**
```
STRLEN key
```

**返回值**
- 返回字符串长度
- key 不存在返回 0

**示例**
```bash
redis-cli set mykey "Hello World"
redis-cli strlen mykey
# (integer) 11

redis-cli strlen nonexistent
# (integer) 0
```

---

## 键命令

### DEL

删除一个或多个 key。

**语法**
```
DEL key1 [key2 ...]
```

**返回值**
- 返回被删除 key 的数量

**示例**
```bash
redis-cli set key1 value1
redis-cli set key2 value2
redis-cli del key1 key2 key3
# (integer) 2  # 删除了 key1 和 key2
```

---

### EXISTS

检查一个或多个 key 是否存在。

**语法**
```
EXISTS key1 [key2 ...]
```

**返回值**
- 返回存在的 key 数量

**示例**
```bash
redis-cli set key1 value1
redis-cli exists key1 key2 key3
# (integer) 1  # 只有 key1 存在
```

---

## 服务器命令

### BGSAVE

在后台异步保存当前数据库到磁盘。

**语法**
```
BGSAVE
```

**返回值**
- 成功：`OK`

**示例**
```bash
redis-cli bgsave
# OK
```

---

## 错误响应格式

### 常见错误

| 错误信息 | 说明 |
|---------|------|
| `ERR unknown command 'xxx'` | 未知命令 |
| `ERR wrong number of arguments for 'xxx' command` | 参数数量错误 |
| `WRONGTYPE Operation against a key holding the wrong kind of value` | 类型错误 |
| `ERR value is not an integer` | 值不是整数 |

---

## 兼容性说明

GoSwiftKV 当前支持以下 Redis 命令：

| 类别 | 命令 |
|------|------|
| 字符串 | SET, GET, SETNX, SETEX, PSETEX, MSET, MGET, INCR, DECR, INCRBY, DECRBY, APPEND, STRLEN |
| 键 | DEL, EXISTS |
| 连接 | PING |
| 服务器 | BGSAVE |

**计划支持**：Hash, List, Set, ZSet, Pub/Sub, 事务等。

---

## 版本

- 文档版本：v0.1.0
- 兼容 Redis 协议版本：RESP2
