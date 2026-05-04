# DDNS Agent

定时检测本机 IPv4 / IPv6 / 局域网地址，自动更新腾讯云 DNSPod 解析记录。支持 Telegram 通知。

## 功能

- 定时检测公网 IPv4、IPv6 和局域网地址
- 地址变化时自动更新 DNSPod DNS 记录
- IPv4、IPv6、局域网分别使用独立子域名
- 过滤虚拟网卡（Docker、veth、bridge 等），只保留真实网卡地址
- 支持自建 IP 检测服务（附带 ip-echo 服务）
- IP 变化时 Telegram 机器人通知
- Docker Compose 一键部署

## 项目结构

```
ddns-agent/
├── cmd/
│   ├── ddns-agent/main.go     # DDNS Agent 主程序
│   └── ip-echo/main.go        # 自建 IP 检测服务
├── internal/
│   ├── config/config.go       # 配置管理
│   ├── detector/
│   │   ├── ip.go              # IPv4/IPv6 检测
│   │   └── lan.go             # 局域网地址检测
│   ├── dnspod/client.go       # DNSPod API 客户端
│   ├── notifier/telegram.go   # Telegram 通知
│   └── agent/agent.go         # 主代理逻辑
├── config/config.yaml.example # 配置文件示例
├── Dockerfile                 # DDNS Agent 镜像
├── Dockerfile.ip-echo         # ip-echo 镜像
├── docker-compose.yml
└── go.mod
```

## 快速开始

### 1. 获取 DNSPod Token

登录 [DNSPod 控制台](https://console.dnspod.cn/) -> 密钥管理 -> 创建密钥，获取 `ID,Token` 格式的 Token。

### 2. 创建 Telegram Bot（可选）

1. 在 Telegram 中找 [@BotFather](https://t.me/BotFather)，创建 Bot 获取 Token
2. 找 [@userinfobot](https://t.me/userinfobot) 获取你的 Chat ID

### 3. 配置

```bash
cp config/config.yaml.example config.yaml
```

编辑 `config.yaml`：

```yaml
# DNSPod Token（格式：ID,Token）
dnspod_token: "12345,abcdefghijklmnopqrstuvwxyz"

# 域名
domain: "example.com"

# 子域名（留空则不更新该类型记录）
ipv4_subdomain: "home"      # A 记录
ipv6_subdomain: "home-v6"   # AAAA 记录
lan_subdomain: "lan"        # 局域网 A 记录

# 自定义 IP 检测服务（留空使用公共默认服务）
ipv4_url: ""
ipv6_url: ""

# 检测间隔（秒）
interval: 600

# Telegram 通知
telegram:
  bot_token: "your-bot-token"
  chat_id: "your-chat-id"
```

### 4. 部署

#### Docker Compose（推荐）

```bash
docker compose up -d
```

#### 直接运行

```bash
go build -o ddns-agent ./cmd/ddns-agent
./ddns-agent                # 默认读取 ./config.yaml
./ddns-agent /path/to.yaml  # 指定配置文件
```

## 自建 IP 检测服务

默认使用公共检测服务（ip.sb、ipify.org 等），如果需要自建：

ip-echo 监听两个端口：IPv4 端口只绑定 `tcp4`，IPv6 端口只绑定 `tcp6`。
域名只添加 A 记录指向 IPv4 端口、只添加 AAAA 记录指向 IPv6 端口，即可区分。

### 部署 ip-echo

> 前提：VPS 需要有 IPv6 地址。验证：`curl -6 ip.sb`

```bash
# 直接运行
go build -o ip-echo ./cmd/ip-echo
IPV4_PORT=8080 IPV6_PORT=8081 ./ip-echo

# Docker（必须用 host 网络，否则容器内没有 IPv6）
docker build -f Dockerfile.ip-echo -t ip-echo .
docker run -d --network=host --restart=unless-stopped --name ip-echo ip-echo
```

### DNS 配置

假设 VPS 的 IPv4 是 `1.2.3.4`，IPv6 是 `2001:db8::1`：

| 记录类型 | 主机记录 | 值 | 端口 |
|---------|---------|-----|------|
| A | `ip` | `1.2.3.4` | 8080 |
| AAAA | `ip6` | `2001:db8::1` | 8081 |

> 或者用同一个主机记录 `ip`，但客户端访问时可能优先走 IPv6，所以建议分开。

### 验证

```bash
curl http://ip.example.com:8080    # 返回 IPv4
curl http://ip6.example.com:8081   # 返回 IPv6
```

### ddns-agent 配置

```yaml
ipv4_url: "http://ip.example.com:8080"
ipv6_url: "http://ip6.example.com:8081"
```

### 配置

在 `config.yaml` 中填入自建服务地址：

```yaml
ipv4_url: "http://your-vps-ipv4:8080"
ipv6_url: "http://your-vps-ipv6:8080"
```

> 注意：检测 IPv6 需要 VPS 本身支持 IPv6。

## 配置说明

| 字段 | 必填 | 说明 |
|------|------|------|
| `dnspod_token` | 是 | DNSPod API Token，格式：`ID,Token` |
| `domain` | 是 | 主域名 |
| `ipv4_subdomain` | 否 | IPv4 子域名（A 记录），留空不更新 |
| `ipv6_subdomain` | 否 | IPv6 子域名（AAAA 记录），留空不更新 |
| `lan_subdomain` | 否 | 局域网子域名（A 记录），留空不更新 |
| `ipv4_url` | 否 | 自定义 IPv4 检测服务地址 |
| `ipv6_url` | 否 | 自定义 IPv6 检测服务地址 |
| `interval` | 否 | 检测间隔，默认 600 秒 |
| `telegram.bot_token` | 否 | Telegram Bot Token |
| `telegram.chat_id` | 否 | Telegram Chat ID |
| `log_level` | 否 | 日志级别：debug/info/warn/error |

## 运行效果

```
2026/05/04 10:00:00 Starting DDNS agent for example.com
2026/05/04 10:00:00 IPv4: home.example.com (A)
2026/05/04 10:00:00 IPv6: home-v6.example.com (AAAA)
2026/05/04 10:00:00 LAN:  lan.example.com (A)
2026/05/04 10:00:00 Interval: 600s
2026/05/04 10:00:01 [IPv4] updating:  -> 203.0.113.1
2026/05/04 10:00:01 [IPv4] updated to 203.0.113.1
2026/05/04 10:00:02 [IPv6] updating:  -> 2001:db8::1
2026/05/04 10:00:02 [IPv6] updated to 2001:db8::1
2026/05/04 10:00:03 [LAN] updating:  -> 192.168.1.100
2026/05/04 10:00:03 [LAN] updated to 192.168.1.100
```

## License

MIT
