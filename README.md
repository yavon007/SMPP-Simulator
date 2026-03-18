# SMPP Simulator

SMPP短信模拟器，为短信平台提供短信发送功能测试环境。

## 功能特性

- SMPP协议服务端（TCP 2775端口）
- REST API管理接口
- WebSocket实时推送
- 可配置的模拟行为
- Docker部署支持

## 快速开始

### Docker部署（推荐）

```bash
docker-compose up -d
```

服务启动后：
- SMPP端口：2775
- HTTP API：8080
- 前端界面：80

### 手动部署（生产环境）

#### 环境要求

- Go 1.21+
- Node.js 18+ & pnpm
- Nginx（可选，用于反向代理）

#### 1. 编译后端

```bash
cd backend

# 下载依赖
go mod download

# 编译（Linux）
CGO_ENABLED=1 go build -o smpp-simulator ./cmd/server

# 编译（Windows）
go build -o smpp-simulator.exe ./cmd/server
```

#### 2. 构建前端

```bash
cd frontend

# 安装依赖
pnpm install

# 构建生产版本
pnpm build

# 构建产物在 dist/ 目录
```

#### 3. 运行后端

```bash
# 直接运行
./smpp-simulator

# 或使用环境变量配置
SMPP_HOST=0.0.0.0 SMPP_PORT=2775 HTTP_HOST=0.0.0.0 HTTP_PORT=8080 ./smpp-simulator
```

#### 4. 配置 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /path/to/smpp-simulator/frontend/dist;
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # WebSocket 代理
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

#### 5. 配置 Systemd 服务（Linux）

创建服务文件 `/etc/systemd/system/smpp-simulator.service`：

```ini
[Unit]
Description=SMPP Simulator
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/smpp-simulator
ExecStart=/opt/smpp-simulator/smpp-simulator
Restart=on-failure
RestartSec=5

# 环境变量
Environment=SMPP_HOST=0.0.0.0
Environment=SMPP_PORT=2775
Environment=HTTP_HOST=0.0.0.0
Environment=HTTP_PORT=8080
Environment=DB_PATH=/opt/smpp-simulator/data/smpp.db

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
# 创建数据目录
mkdir -p /opt/smpp-simulator/data

# 重载配置
systemctl daemon-reload

# 启动服务
systemctl start smpp-simulator

# 设置开机自启
systemctl enable smpp-simulator

# 查看状态
systemctl status smpp-simulator
```

### 本地开发

**后端：**
```bash
cd backend
go mod download
go run cmd/server/main.go
```

**前端：**
```bash
cd frontend
pnpm install
pnpm dev
```

## 使用说明

1. 配置短信平台连接到 `localhost:2775`
2. 使用任意 system_id 和 password 进行绑定
3. 发送 submit_sm 请求进行测试
4. 通过前端界面查看消息和连接状态

## API接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sessions | 获取连接列表 |
| DELETE | /api/sessions/:id | 断开连接 |
| GET | /api/messages | 消息列表 |
| GET | /api/messages/:id | 消息详情 |
| POST | /api/messages/:id/deliver | 触发状态报告 |
| GET | /api/stats | 统计数据 |
| GET/PUT | /api/mock/config | 模拟配置 |

## 技术栈

- 后端：Go + Gin + SQLite
- 前端：Vue 3 + TypeScript + Element Plus
- 部署：Docker + docker-compose

## License

MIT
