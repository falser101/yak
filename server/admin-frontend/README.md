# 管理后台前端 (Vue 3 + Element Plus)

## 安装和运行

```bash
cd admin-frontend

# 安装依赖
npm install

# 开发模式运行
npm run dev
# 访问 http://localhost:5173/admin/login

# 构建生产版本
npm run build
# 构建产物在 dist/ 目录
```

## 构建后端 Go 服务

构建 Vue 后，需要重新构建 Go 服务以集成前端：

```bash
cd ..
go build -o yak-server .
./yak-server
# 访问 http://localhost:8080/admin/login
```

## 功能模块

- **活动管理** - 活动列表、发布、编辑、删除、查看报名
- **品牌管理** - 品牌增删改、车型管理
- **自行车** - 用户自行车列表查看

## API 接口

前端通过 `/admin/api/*` 调用后端接口：
- `/admin/api/activities` - 活动列表
- `/admin/api/brands` - 品牌列表
- `/admin/api/bikes` - 用户自行车列表
