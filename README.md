# 半导体 MES 系统 (Semiconductor MES System)

一个基于现代技术栈构建的半导体制造执行系统（MES），用于管理和追踪半导体生产过程。

## 技术栈

### 后端
- **框架**: Sponge (Go) + Gin + GORM
- **数据库**: SQLite (开发环境) / MySQL / PostgreSQL (生产环境)
- **认证**: JWT
- **授权**: RBAC (基于角色的访问控制)

### 前端
- **框架**: Vue 3 + Vite + TypeScript
- **UI库**: Naive UI
- **状态管理**: Pinia
- **路由**: Vue Router
- **国际化**: vue-i18n (简体中文/英文)

## 功能模块

### 已实现
1. **系统管理**
   - 用户管理
   - 角色管理
   - 菜单/权限管理
   - JWT登录/登出
   - RBAC权限控制

2. **基础数据管理**（模块1）
   - 工厂管理 ✅
   - 车间管理
   - 生产线管理
   - 产品管理
   - 工艺路线管理
   - 工序管理
   - 配方管理

### 待实现（已预留菜单和权限）
3. **工单管理**（模块2）
   - 工单CRUD
   - 批次投放
   - 批次拆分/合并
   - Hold/Release

4. **WIP跟踪**（模块3）
   - Track In/Track Out
   - 流转历史

5. **设备管理**（模块4）
   - 设备台账
   - 设备状态
   - 预防性维护

6. **质量管理**（模块5）
   - 检验记录
   - 缺陷记录
   - SPC

## 快速开始

### 前置要求
- Go 1.22+
- Node.js 18+
- npm 或 yarn

### 后端启动

```bash
# 进入后端目录
cd server

# 安装依赖
go mod tidy

# 运行（开发模式）
make run

# 或直接运行
go run cmd/server/main.go
```

后端默认运行在 `http://localhost:8080`

### 前端启动

```bash
# 进入前端目录
cd web

# 安装依赖
npm install

# 运行（开发模式）
npm run dev
```

前端默认运行在 `http://localhost:3000`

### 使用 Makefile 一键启动

```bash
# 同时启动后端和前端
make run

# 或者分别启动
make run-server  # 启动后端
make run-web     # 启动前端
```

## 默认账号

- 用户名: `admin`
- 密码: `admin123`

登录后可以：
- 切换中英文界面
- 管理用户、角色、菜单
- 管理基础数据（工厂、车间、生产线、产品、工艺路线、工序、配方）

## 项目结构

```
semi-mes/
├── docs/               # 文档
│   └── design.md      # 详细设计文档
├── server/            # 后端代码
│   ├── cmd/           # 应用入口
│   ├── internal/      # 内部代码
│   │   ├── config/    # 配置管理
│   │   ├── model/     # 数据模型
│   │   ├── handler/   # HTTP处理器
│   │   ├── middleware/# 中间件
│   │   └── router/    # 路由
│   ├── migrations/    # 数据库迁移和种子数据
│   ├── pkg/           # 公共包
│   ├── config/        # 配置文件
│   └── Makefile       # 构建脚本
├── web/               # 前端代码
│   ├── src/
│   │   ├── api/       # API调用
│   │   ├── views/     # 页面组件
│   │   ├── router/    # 路由配置
│   │   ├── stores/    # 状态管理
│   │   ├── locales/   # 国际化
│   │   └── layouts/   # 布局组件
│   └── package.json
├── Makefile           # 项目级构建脚本
└── README.md          # 本文件
```

## 数据库

### SQLite（开发环境，默认）
- 数据文件：`server/data/mes.db`
- 自动创建和迁移
- 包含默认管理员账号和权限数据

### 切换到 MySQL/PostgreSQL
编辑 `server/config/config.yaml`：

```yaml
database:
  type: mysql  # 或 postgres
  mysql:
    host: localhost
    port: 3306
    database: mes
    username: root
    password: password
    charset: utf8mb4
```

## 测试

### 后端测试
```bash
cd server
go test -v ./...
```

### 前端类型检查和构建
```bash
cd web
npm run build
```

## API文档

后端API遵循RESTful规范：

- **基础路径**: `/api/v1`
- **认证**: `Authorization: Bearer <token>` Header
- **响应格式**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {...}
  }
  ```

详细API文档请参考 [docs/design.md](docs/design.md)

## 开发指南

### 添加新的CRUD模块

1. **后端**：
   - 在 `internal/model/` 添加数据模型
   - 在 `internal/handler/` 添加处理器
   - 在 `internal/router/router.go` 添加路由
   - 在 `migrations/seed.go` 添加菜单权限

2. **前端**：
   - 在 `src/api/` 添加API调用
   - 在 `src/views/` 添加页面组件
   - 在 `src/router/` 添加路由
   - 在 `src/locales/` 添加国际化文本

3. **参考示例**：
   - 后端：`server/internal/handler/basedata.go` （工厂管理）
   - 前端：`web/src/views/base-data/Factory.vue` （工厂管理页面）

## 国际化

系统支持简体中文和英文：

- 前端：使用 vue-i18n，配置文件在 `web/src/locales/`
- 菜单：菜单名称存储为i18n key（如 `menu.baseData.factory`）
- 错误消息：后端返回i18n key，前端根据语言渲染

## 权限控制

### 权限编码格式
`模块:功能:操作`，例如：
- `base:factory:add` - 添加工厂
- `base:factory:edit` - 编辑工厂
- `base:factory:delete` - 删除工厂
- `base:factory:query` - 查询工厂

### 前端权限检查
```vue
<n-button v-if="userStore.hasPermission('base:factory:add')" @click="handleAdd">
  添加
</n-button>
```

### 后端权限检查
```go
router.POST("/api/v1/base-data/factories",
    middleware.RequirePermission("base:factory:add"),
    handler.CreateFactory,
)
```

## 部署

### 开发环境
```bash
make run
```

### 生产环境
```bash
# 构建后端
cd server
make build

# 构建前端
cd web
npm run build

# 部署
# 1. 复制 server/bin/server 到服务器
# 2. 复制 server/config/ 到服务器
# 3. 复制 web/dist/ 到 Nginx/Apache 静态文件目录
# 4. 配置反向代理将 /api 请求代理到后端服务
```

## 后续开发计划

详细的模块设计和开发计划请参考 [docs/design.md](docs/design.md)

- [ ] 完善系统管理模块的前端页面（用户、角色、菜单）
- [ ] 完善基础数据模块的前端页面（车间、生产线等）
- [ ] 实现工单管理模块（模块2）
- [ ] 实现WIP跟踪模块（模块3）
- [ ] 实现设备管理模块（模块4）
- [ ] 实现质量管理模块（模块5）
- [ ] 添加报表和看板
- [ ] 添加更多测试用例
- [ ] 添加API文档（Swagger）
- [ ] 添加Docker支持

## 许可

MIT License

## 贡献

欢迎提交Issue和Pull Request！
