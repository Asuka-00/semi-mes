# 半导体 MES 系统 (Semiconductor MES System)

一个基于现代技术栈构建的半导体制造执行系统（MES），用于管理和追踪半导体生产过程。

## 技术栈

### 后端
- **框架**: [Sponge](https://github.com/go-dev-frame/sponge) v1.16.1 生成的 Web 服务（Gin + GORM）
- **数据库**: SQLite。修改 `server/configs/mes.yml` 的 `database.driver` 可切换到 MySQL 或 PostgreSQL
- **认证**: JWT
- **授权**: RBAC，按钮级权限码

### 前端
- **模板**: [Soybean Admin](https://github.com/soybeanjs/soybean-admin) 2.2.0
- **框架**: Vue 3 + Vite + TypeScript + Naive UI + Pinia + UnoCSS
- **路由**: elegant-router，后端动态路由
- **国际化**: vue-i18n（zh-CN / en-US）

## 功能模块

### 已实现
1. **系统管理**：登录、用户、角色（分配菜单）、菜单，后端按权限码拦截接口
2. **基础数据（模块1）**：工厂、车间、生产线、产品、工艺路线（含按顺序维护工序）、工序、配方。列表支持搜索和分页，按钮按权限显示

### 未实现
模块 2–5 只有菜单和占位页，没有工单、批次、WIP、设备、质量的业务接口和页面。

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
- Go 1.24（SQLite 驱动需要 CGO 和 gcc）
- Node.js 20+ 与 pnpm

### 后端启动

```bash
cd server
go mod tidy
make run
```

服务监听 `http://localhost:8080`。SQLite 文件为 `server/data/mes.db`。`make build` 关闭了 CGO，不能用来编译当前的 SQLite 版本。

### 前端启动

```bash
cd web
pnpm install
pnpm dev
```

`pnpm dev` 使用 test 模式，代理到 `http://localhost:8080/api/v1`。默认端口以终端输出为准（Soybean Admin 通常是 9527）。

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

登录后可以切换中英文，菜单由后端返回。可以维护用户、角色、菜单和全部基础数据。工艺路线页面可以按顺序维护工序。

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
