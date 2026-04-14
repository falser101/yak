# 本地后端开发指南

## 前置要求

- Docker & Docker Compose
- Go 1.21+

## 启动步骤

### 1. 启动 PostgreSQL 数据库

在项目根目录执行：

```bash
cd /home/falser/projects/yak
docker-compose up -d
```

等待数据库启动，查看日志：

```bash
docker-compose logs -f postgres
```

### 2. 启动 Go 后端

```bash
cd server
go mod download
go run main.go
```

服务器会在 `http://localhost:8080` 启动

### 3. 配置小程序

小程序已经配置好调用 `http://127.0.0.1:8080/api`

在微信开发者工具中：
1. 导入 `miniprogram` 目录
2. 点击 **详情** → **本地设置**
3. 勾选 **不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书**

### 4. 访问后台管理

后台地址：`http://localhost:8080/admin/login`

默认账号：
- 用户名：`admin`
- 密码：`admin123`

### 5. 测试 API

```bash
# 获取活动列表
curl http://localhost:8080/api/activities

# 获取活动详情
curl http://localhost:8080/api/activities/1

# 创建活动
curl -X POST http://localhost:8080/api/activities \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试活动",
    "cover": "",
    "date": "2026-04-15 10:00:00",
    "location": "测试地点",
    "maxParticipants": 30,
    "price": 50,
    "description": "测试描述",
    "createdBy": "test"
  }'

# 报名活动
curl -X POST http://localhost:8080/api/activities/1/signup \
  -H "Content-Type: application/json" \
  -d '{"userId": "user123"}'
```

## 数据库信息

- 主机：localhost
- 端口：5432
- 用户：yak
- 密码：yak123456
- 数据库：yak

## 停止服务

```bash
docker-compose down
```

## 项目结构

```
yak/
├── miniprogram/          # 小程序前端
│   ├── app.js
│   ├── app.json
│   └── pages/
├── server/               # Go 后端
│   ├── main.go
│   ├── handlers/
│   │   ├── activity.go   # 小程序 API
│   │   └── admin.go      # 后台管理
│   ├── templates/        # 后台 HTML 模板
│   └── init.sql
├── docker-compose.yml
└── README.md
```

## 后台功能

- ✅ 登录验证
- ✅ 活动列表管理
- ✅ 新建/编辑/删除活动
- ✅ 查看报名名单
