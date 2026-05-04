# DDNS Agent 实现计划

## 项目概述

实现一个 DDNS Agent，定时检测本机 IPv4、IPv6 和局域网地址，当地址变化时更新腾讯云 DNSPod 的 DNS 解析记录。

## 技术栈

- 语言: Go
- 容器化: Docker + Docker Compose
- DNS 服务: 腾讯云 DNSPod API v3
- 通知: Telegram Bot

## 项目结构

```
ddns-agent/
├── cmd/ddns-agent/main.go      # 程序入口
├── internal/
│   ├── config/config.go        # 配置管理
│   ├── detector/               # IP 检测模块
│   │   ├── ipv4.go
│   │   ├── ipv6.go
│   │   └── lan.go
│   ├── dnspod/client.go        # DNSPod API 客户端
│   ├── notifier/
│   │   └── telegram.go         # Telegram 通知
│   └── agent/agent.go          # 主代理逻辑
├── config/config.yaml          # 配置文件示例
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## 核心功能

### 1. IP 检测模块

**IPv4 检测**
- 使用多个外部服务检测公网 IPv4：
  - https://ip.sb
  - https://api.ipify.org
  - https://icanhazip.com
- 支持备用服务，确保可用性

**IPv6 检测**
- 使用外部服务检测公网 IPv6：
  - https://ip.sb (支持 IPv6)
  - https://api6.ipify.org
  - https://icanhazip.com

**局域网地址检测**
- 读取本机网络接口（eth0, wlan0 等）
- 过滤掉 loopback 和 link-local 地址
- 过滤掉虚拟网络接口（docker0, veth, br-, virbr, tun, tap, wg 等）
- 只保留真实网卡的局域网地址
- 更新到 DNS 记录（需要配置 lan_subdomain）

### 2. DNSPod 客户端

**API 调用**
- 使用腾讯云 API v3 签名认证
- 实现以下操作：
  - `DescribeRecordList`: 查询域名记录列表
  - `CreateRecord`: 创建新记录
  - `ModifyRecord`: 修改现有记录

**记录管理**
- 自动判断记录是否存在
- 不存在则创建，存在则更新
- 支持 A (IPv4) 和 AAAA (IPv6) 记录类型

### 3. Telegram 通知

**通知内容**
- IP 变化时发送消息
- 包含旧 IP 和新 IP
- 包含变化时间

**消息格式**
```
🔄 DDNS IP 变更通知
域名: home.example.com
类型: A (IPv4)
旧 IP: 1.2.3.4
新 IP: 5.6.7.8
时间: 2026-05-04 10:30:00
```

### 4. 主代理逻辑

**执行流程**
1. 读取配置文件
2. 初始化 DNSPod 客户端
3. 定时循环：
   - 检测当前 IP
   - 与上次记录比较
   - 如有变化，更新 DNS 记录
   - 发送 Telegram 通知
4. 记录日志

**检测间隔**
- 默认 10 分钟（600 秒）
- 可通过配置文件自定义

## 配置文件格式

```yaml
# DNSPod API 配置
dnspod:
  id: "your-api-id"        # API ID
  token: "your-api-token"  # API Token

# 域名配置
domain: "example.com"      # 主域名
subdomain: "home"          # 子域名

# 检测间隔（秒）
interval: 600

# Telegram 通知配置（可选）
telegram:
  bot_token: "your-bot-token"
  chat_id: "your-chat-id"

# 日志级别: debug, info, warn, error
log_level: "info"
```

## Docker 部署

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ddns-agent ./cmd/ddns-agent

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/ddns-agent .
COPY config/config.yaml.example config.yaml
CMD ["./ddns-agent"]
```

### docker-compose.yml

```yaml
version: '3.8'
services:
  ddns-agent:
    build: .
    container_name: ddns-agent
    restart: unless-stopped
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      - TZ=Asia/Shanghai
    network_mode: host  # 需要访问宿主机网络接口
```

## 实现步骤

1. **初始化项目**
   - 创建 Go 模块
   - 添加依赖（腾讯云 SDK、日志库等）

2. **实现配置管理**
   - 定义配置结构体
   - 实现 YAML 配置文件读取

3. **实现 IP 检测模块**
   - IPv4 检测（多服务备用）
   - IPv6 检测
   - 局域网地址检测

4. **实现 DNSPod 客户端**
   - API 签名认证
   - 记录查询/创建/更新

5. **实现 Telegram 通知**
   - 发送消息功能

6. **实现主代理逻辑**
   - 整合各模块
   - 定时任务调度

7. **编写 Dockerfile 和 docker-compose.yml**

8. **测试和文档**

## 依赖库

- `github.com/tencentcloud/tencentcloud-sdk-go`: 腾讯云 SDK
- `github.com/sirupsen/logrus`: 日志库
- `gopkg.in/yaml.v3`: YAML 解析

## 注意事项

1. **网络模式**: Docker 需要使用 `network_mode: host` 以便正确检测局域网地址
2. **时区设置**: 建议设置容器时区为 Asia/Shanghai
3. **API 频率**: DNSPod API 有调用频率限制，建议检测间隔不低于 5 分钟
4. **IPv6 支持**: 确保宿主机和容器都支持 IPv6
