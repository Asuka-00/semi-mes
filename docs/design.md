# 半导体 MES 系统设计文档

## 1. 系统概述 (System Overview)

本系统是一个面向半导体制造行业的MES（Manufacturing Execution System，制造执行系统），用于管理和追踪生产过程中的各个环节。

### 1.1 技术栈

- **前端 (Frontend)**: Soybean Admin (Vue 3 + Vite + TypeScript + Naive UI)
- **后端 (Backend)**: Sponge (Go framework) + Gin + GORM
- **数据库 (Database)**: SQLite (开发环境) / MySQL / PostgreSQL (生产环境，可配置切换)
- **认证 (Authentication)**: JWT
- **授权 (Authorization)**: RBAC (Role-Based Access Control)
- **国际化 (i18n)**: 简体中文 (zh-CN) + 英文 (en-US)

### 1.2 架构原则

- **前后端分离**: 前端通过RESTful API与后端通信
- **数据库无关**: 使用GORM抽象层，支持配置切换数据库
- **权限驱动**: 所有API和前端功能基于RBAC权限控制
- **国际化优先**: 所有用户可见文本支持双语切换

## 2. 仓库结构 (Repository Layout)

```
semi-mes/
├── docs/                   # 文档目录
│   └── design.md          # 本设计文档
├── server/                # 后端代码
│   ├── cmd/               # 应用入口
│   ├── internal/          # 内部代码
│   │   ├── config/       # 配置管理
│   │   ├── model/        # 数据模型
│   │   ├── dao/          # 数据访问层
│   │   ├── service/      # 业务逻辑层
│   │   ├── handler/      # HTTP处理器
│   │   ├── middleware/   # 中间件（JWT、RBAC等）
│   │   └── router/       # 路由定义
│   ├── migrations/        # 数据库迁移
│   ├── config/            # 配置文件
│   ├── go.mod
│   └── Makefile
├── web/                   # 前端代码
│   ├── src/
│   │   ├── api/          # API调用封装
│   │   ├── views/        # 页面组件
│   │   ├── router/       # 路由配置
│   │   ├── stores/       # 状态管理
│   │   ├── locales/      # 国际化文件
│   │   └── typings/      # TypeScript类型定义
│   ├── package.json
│   └── vite.config.ts
├── scripts/               # 运行脚本
│   ├── run-dev.sh        # 开发环境启动脚本
│   └── run-prod.sh       # 生产环境启动脚本
├── Makefile              # 统一构建脚本
└── README.md             # 项目说明
```

## 3. 数据模型设计 (Data Model)

### 3.1 系统管理模块 (System Management)

#### 3.1.1 用户表 (sys_user)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| username | varchar(50) | 用户名，唯一 |
| password | varchar(255) | 密码（加密） |
| real_name | varchar(50) | 真实姓名 |
| email | varchar(100) | 邮箱 |
| phone | varchar(20) | 手机号 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.1.2 角色表 (sys_role)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| role_code | varchar(50) | 角色编码，唯一 |
| role_name | varchar(50) | 角色名称 |
| description | varchar(255) | 角色描述 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.1.3 菜单/权限表 (sys_menu)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| parent_id | bigint | 父菜单ID，0表示根菜单 |
| menu_type | tinyint | 类型：1-目录，2-菜单，3-按钮 |
| menu_name | varchar(50) | 菜单名称（i18n key） |
| permission_code | varchar(100) | 权限编码，如：base:factory:add |
| route_path | varchar(200) | 前端路由路径 |
| component_path | varchar(200) | 前端组件路径 |
| icon | varchar(50) | 图标 |
| sort_order | int | 排序号 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.1.4 用户-角色关联表 (sys_user_role)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| user_id | bigint | 用户ID |
| role_id | bigint | 角色ID |
| created_at | timestamp | 创建时间 |

#### 3.1.5 角色-菜单关联表 (sys_role_menu)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| role_id | bigint | 角色ID |
| menu_id | bigint | 菜单ID |
| created_at | timestamp | 创建时间 |

### 3.2 模块1：基础数据建模 (Base Data Modeling)

#### 3.2.1 工厂表 (base_factory)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| factory_code | varchar(50) | 工厂编码，唯一 |
| factory_name | varchar(100) | 工厂名称 |
| address | varchar(255) | 地址 |
| contact | varchar(50) | 联系人 |
| phone | varchar(20) | 联系电话 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.2.2 车间/区域表 (base_workshop)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| factory_id | bigint | 所属工厂ID |
| workshop_code | varchar(50) | 车间编码 |
| workshop_name | varchar(100) | 车间名称 |
| workshop_type | varchar(20) | 车间类型：FAB, TEST, ASSY |
| description | varchar(255) | 描述 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.2.3 生产线表 (base_production_line)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| workshop_id | bigint | 所属车间ID |
| line_code | varchar(50) | 生产线编码 |
| line_name | varchar(100) | 生产线名称 |
| capacity | int | 产能（片/天） |
| description | varchar(255) | 描述 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.2.4 产品表 (base_product)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| product_code | varchar(50) | 产品编码，唯一 |
| product_name | varchar(100) | 产品名称 |
| product_type | varchar(50) | 产品类型：IC, DISCRETE, SENSOR等 |
| version | varchar(20) | 版本号 |
| description | varchar(500) | 产品描述 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.2.5 工艺路线表 (base_process_route)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| product_id | bigint | 产品ID |
| route_code | varchar(50) | 路线编码 |
| route_name | varchar(100) | 路线名称 |
| version | varchar(20) | 版本号 |
| is_default | tinyint | 是否默认路线：1-是，0-否 |
| description | varchar(500) | 描述 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.2.6 工序/操作表 (base_operation)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| route_id | bigint | 所属工艺路线ID |
| operation_code | varchar(50) | 工序编码 |
| operation_name | varchar(100) | 工序名称 |
| operation_type | varchar(50) | 工序类型：PHOTO, ETCH, DEPOSITION, etc. |
| sequence | int | 工序顺序 |
| standard_time | int | 标准工时（分钟） |
| description | varchar(500) | 描述 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.2.7 配方表 (base_recipe)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| operation_id | bigint | 所属工序ID |
| recipe_code | varchar(50) | 配方编码 |
| recipe_name | varchar(100) | 配方名称 |
| version | varchar(20) | 版本号 |
| parameters | text | 配方参数（JSON格式） |
| is_default | tinyint | 是否默认配方：1-是，0-否 |
| description | varchar(500) | 描述 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

### 3.3 模块2：工单和批次管理 (Work Order & Lot Management)

#### 3.3.1 工单表 (wip_work_order)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| order_no | varchar(50) | 工单号，唯一 |
| product_id | bigint | 产品ID |
| route_id | bigint | 工艺路线ID |
| planned_qty | int | 计划数量 |
| released_qty | int | 已投放数量 |
| completed_qty | int | 已完成数量 |
| priority | tinyint | 优先级：1-低，2-中，3-高，4-紧急 |
| planned_start_date | date | 计划开始日期 |
| planned_end_date | date | 计划完成日期 |
| actual_start_date | timestamp | 实际开始时间 |
| actual_end_date | timestamp | 实际完成时间 |
| status | varchar(20) | 状态：CREATED, RELEASED, IN_PROGRESS, COMPLETED, CLOSED |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.3.2 批次表 (wip_lot)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| lot_no | varchar(50) | 批次号，唯一 |
| order_id | bigint | 工单ID |
| parent_lot_id | bigint | 父批次ID（拆分/合并时使用） |
| product_id | bigint | 产品ID |
| route_id | bigint | 工艺路线ID |
| current_operation_id | bigint | 当前工序ID |
| quantity | int | 当前数量 |
| hold_flag | tinyint | Hold标志：1-Hold，0-正常 |
| hold_reason | varchar(255) | Hold原因 |
| status | varchar(20) | 状态：CREATED, WIP, HOLD, COMPLETED, SCRAPPED |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

### 3.4 模块3：WIP跟踪 (WIP Tracking)

#### 3.4.1 流转历史表 (wip_move_history)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| lot_id | bigint | 批次ID |
| operation_id | bigint | 工序ID |
| move_type | varchar(20) | 类型：TRACK_IN, TRACK_OUT |
| equipment_id | bigint | 设备ID |
| operator_id | bigint | 操作员ID |
| quantity_in | int | 进站数量 |
| quantity_out | int | 出站数量 |
| move_time | timestamp | 流转时间 |
| created_at | timestamp | 创建时间 |

### 3.5 模块4：设备管理 (Equipment Management)

#### 3.5.1 设备台账表 (eqp_equipment)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| equipment_code | varchar(50) | 设备编码，唯一 |
| equipment_name | varchar(100) | 设备名称 |
| line_id | bigint | 所属生产线ID |
| equipment_type | varchar(50) | 设备类型 |
| model | varchar(50) | 型号 |
| manufacturer | varchar(100) | 厂商 |
| install_date | date | 安装日期 |
| status | varchar(20) | 状态：IDLE, RUNNING, DOWN, PM, MAINTENANCE |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

#### 3.5.2 预防性维护表 (eqp_pm_plan)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| equipment_id | bigint | 设备ID |
| pm_type | varchar(50) | PM类型：DAILY, WEEKLY, MONTHLY, QUARTERLY |
| pm_name | varchar(100) | PM名称 |
| interval_days | int | 间隔天数 |
| last_pm_date | date | 上次PM日期 |
| next_pm_date | date | 下次PM日期 |
| status | tinyint | 状态：1-启用，0-禁用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除时间 |

### 3.6 模块5：质量管理 (Quality Management)

#### 3.6.1 检验记录表 (qc_inspection)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| inspection_no | varchar(50) | 检验单号，唯一 |
| lot_id | bigint | 批次ID |
| operation_id | bigint | 工序ID |
| inspection_type | varchar(20) | 检验类型：IQC, PQC, FQC, OQC |
| inspector_id | bigint | 检验员ID |
| sample_size | int | 抽样数量 |
| defect_count | int | 不良数量 |
| result | varchar(20) | 结果：PASS, FAIL, PENDING |
| inspection_time | timestamp | 检验时间 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

#### 3.6.2 缺陷记录表 (qc_defect)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint | 主键ID |
| inspection_id | bigint | 检验记录ID |
| defect_code | varchar(50) | 缺陷代码 |
| defect_name | varchar(100) | 缺陷名称 |
| defect_count | int | 缺陷数量 |
| severity | varchar(20) | 严重性：CRITICAL, MAJOR, MINOR |
| description | varchar(500) | 描述 |
| created_at | timestamp | 创建时间 |

### 3.7 ER关系图

```
[sys_user] 1---N [sys_user_role] N---1 [sys_role]
[sys_role] 1---N [sys_role_menu] N---1 [sys_menu]

[base_factory] 1---N [base_workshop] 1---N [base_production_line]
[base_product] 1---N [base_process_route] 1---N [base_operation] 1---N [base_recipe]

[wip_work_order] 1---N [wip_lot]
[base_product] 1---N [wip_work_order]
[base_process_route] 1---N [wip_work_order]
[wip_lot] 1---N [wip_move_history]

[base_production_line] 1---N [eqp_equipment]
[eqp_equipment] 1---N [eqp_pm_plan]
[eqp_equipment] 1---N [wip_move_history]

[wip_lot] 1---N [qc_inspection]
[qc_inspection] 1---N [qc_defect]
```

## 4. API设计规范 (API Conventions)

### 4.1 RESTful接口规范

- **基础路径**: `/api/v1`
- **认证**: 使用 `Authorization: Bearer <token>` Header
- **命名规则**: 使用小写字母和连字符，如 `/api/v1/base-data/factories`

### 4.2 统一响应格式

#### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 业务数据
  }
}
```

#### 错误响应
```json
{
  "code": 40001,
  "message": "error.auth.invalid_token",
  "data": null
}
```

#### 分页响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

### 4.3 错误码设计

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40001 | 认证失败 |
| 40003 | 权限不足 |
| 40004 | 资源不存在 |
| 40009 | 参数错误 |
| 50000 | 服务器内部错误 |

### 4.4 主要API端点

#### 4.4.1 系统管理
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/logout` - 登出
- `GET /api/v1/auth/user-info` - 获取当前用户信息
- `GET /api/v1/auth/menus` - 获取当前用户菜单权限
- `GET /api/v1/users` - 用户列表
- `POST /api/v1/users` - 创建用户
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户
- `GET /api/v1/roles` - 角色列表
- `POST /api/v1/roles` - 创建角色
- `PUT /api/v1/roles/:id` - 更新角色
- `DELETE /api/v1/roles/:id` - 删除角色
- `GET /api/v1/menus` - 菜单列表（树形）
- `POST /api/v1/menus` - 创建菜单
- `PUT /api/v1/menus/:id` - 更新菜单
- `DELETE /api/v1/menus/:id` - 删除菜单

#### 4.4.2 基础数据（模块1）
- `GET /api/v1/base-data/factories` - 工厂列表
- `POST /api/v1/base-data/factories` - 创建工厂
- `PUT /api/v1/base-data/factories/:id` - 更新工厂
- `DELETE /api/v1/base-data/factories/:id` - 删除工厂
- `GET /api/v1/base-data/workshops` - 车间列表
- `POST /api/v1/base-data/workshops` - 创建车间
- `PUT /api/v1/base-data/workshops/:id` - 更新车间
- `DELETE /api/v1/base-data/workshops/:id` - 删除车间
- `GET /api/v1/base-data/production-lines` - 生产线列表
- `POST /api/v1/base-data/production-lines` - 创建生产线
- `PUT /api/v1/base-data/production-lines/:id` - 更新生产线
- `DELETE /api/v1/base-data/production-lines/:id` - 删除生产线
- `GET /api/v1/base-data/products` - 产品列表
- `POST /api/v1/base-data/products` - 创建产品
- `PUT /api/v1/base-data/products/:id` - 更新产品
- `DELETE /api/v1/base-data/products/:id` - 删除产品
- `GET /api/v1/base-data/process-routes` - 工艺路线列表
- `POST /api/v1/base-data/process-routes` - 创建工艺路线
- `PUT /api/v1/base-data/process-routes/:id` - 更新工艺路线
- `DELETE /api/v1/base-data/process-routes/:id` - 删除工艺路线
- `GET /api/v1/base-data/operations` - 工序列表
- `POST /api/v1/base-data/operations` - 创建工序
- `PUT /api/v1/base-data/operations/:id` - 更新工序
- `DELETE /api/v1/base-data/operations/:id` - 删除工序
- `GET /api/v1/base-data/recipes` - 配方列表
- `POST /api/v1/base-data/recipes` - 创建配方
- `PUT /api/v1/base-data/recipes/:id` - 更新配方
- `DELETE /api/v1/base-data/recipes/:id` - 删除配方

## 5. 国际化方案 (i18n Approach)

### 5.1 前端国际化

使用 `vue-i18n` 实现前端国际化：

```typescript
// locales/zh-CN.ts
export default {
  menu: {
    dashboard: '仪表板',
    baseData: '基础数据',
    factory: '工厂管理',
    workshop: '车间管理',
    // ...
  },
  baseData: {
    factory: {
      title: '工厂管理',
      code: '工厂编码',
      name: '工厂名称',
      // ...
    }
  }
}

// locales/en-US.ts
export default {
  menu: {
    dashboard: 'Dashboard',
    baseData: 'Base Data',
    factory: 'Factory',
    workshop: 'Workshop',
    // ...
  },
  baseData: {
    factory: {
      title: 'Factory Management',
      code: 'Factory Code',
      name: 'Factory Name',
      // ...
    }
  }
}
```

### 5.2 后端国际化

后端错误消息和枚举值返回i18n key，由前端根据当前语言渲染：

```go
// 后端返回错误消息的key
{
  "code": 40009,
  "message": "error.validation.required_field",
  "data": {
    "field": "factory_code"
  }
}
```

前端根据message key和当前语言显示对应文本。

### 5.3 菜单国际化

菜单名称在数据库中存储i18n key（如 `menu.baseData.factory`），前端根据key渲染对应语言文本。

## 6. 数据库切换方案 (Database Switching)

### 6.1 配置文件

```yaml
# config/config.yaml
database:
  type: sqlite  # 可选: sqlite, mysql, postgres
  
  # SQLite配置
  sqlite:
    path: ./data/mes.db
  
  # MySQL配置
  mysql:
    host: localhost
    port: 3306
    database: mes
    username: root
    password: password
    charset: utf8mb4
  
  # PostgreSQL配置
  postgres:
    host: localhost
    port: 5432
    database: mes
    username: postgres
    password: password
    sslmode: disable
```

### 6.2 数据库连接代码

```go
func InitDB(cfg *Config) (*gorm.DB, error) {
    var dialector gorm.Dialector
    
    switch cfg.Database.Type {
    case "sqlite":
        dialector = sqlite.Open(cfg.Database.SQLite.Path)
    case "mysql":
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
            cfg.Database.MySQL.Username,
            cfg.Database.MySQL.Password,
            cfg.Database.MySQL.Host,
            cfg.Database.MySQL.Port,
            cfg.Database.MySQL.Database,
            cfg.Database.MySQL.Charset,
        )
        dialector = mysql.Open(dsn)
    case "postgres":
        dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
            cfg.Database.Postgres.Host,
            cfg.Database.Postgres.Port,
            cfg.Database.Postgres.Username,
            cfg.Database.Postgres.Password,
            cfg.Database.Postgres.Database,
            cfg.Database.Postgres.SSLMode,
        )
        dialector = postgres.Open(dsn)
    default:
        return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
    }
    
    return gorm.Open(dialector, &gorm.Config{})
}
```

### 6.3 数据库无关SQL

- 使用GORM的Model方法，避免手写SQL
- 时间字段使用`time.Time`类型，由GORM自动转换
- 避免使用数据库特定函数（如MySQL的`GROUP_CONCAT`）
- 使用GORM的AutoMigrate进行数据库迁移

## 7. RBAC权限控制

### 7.1 权限模型

采用标准RBAC模型：
- 用户 (User) → 角色 (Role) → 权限 (Permission)
- 权限以菜单/按钮的形式存在
- 权限编码格式：`模块:功能:操作`，如 `base:factory:add`

### 7.2 后端权限检查

使用中间件检查API权限：

```go
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt64("user_id")
        hasPermission := checkUserPermission(userID, permission)
        if !hasPermission {
            c.JSON(403, Response{
                Code: 40003,
                Message: "error.auth.insufficient_permission",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}

// 路由定义
router.POST("/api/v1/base-data/factories", 
    middleware.RequireAuth(),
    middleware.RequirePermission("base:factory:add"),
    handler.CreateFactory,
)
```

### 7.3 前端权限控制

- 菜单：通过后端接口获取当前用户可访问的菜单树，动态生成路由
- 按钮：使用指令或组件检查权限码，控制按钮显隐

```vue
<script setup lang="ts">
import { usePermission } from '@/hooks/use-permission'

const { hasPermission } = usePermission()
</script>

<template>
  <n-button v-if="hasPermission('base:factory:add')" @click="handleAdd">
    {{ $t('common.add') }}
  </n-button>
</template>
```

## 8. 开发和部署

### 8.1 开发环境

```bash
# 启动后端
cd server
make run

# 启动前端
cd web
npm run dev
```

### 8.2 生产环境

```bash
# 构建后端
cd server
make build

# 构建前端
cd web
npm run build

# 使用Docker部署
docker-compose up -d
```

## 9. 初始数据

### 9.1 默认管理员

- 用户名: `admin`
- 密码: `admin123`
- 角色: 超级管理员

### 9.2 默认角色

- 超级管理员 (super_admin): 拥有所有权限
- 系统管理员 (sys_admin): 拥有系统管理权限
- 操作员 (operator): 拥有基础数据录入权限
- 只读用户 (viewer): 只有查看权限

### 9.3 菜单权限树

```
仪表板 (dashboard)
系统管理 (system)
  ├── 用户管理 (system:user)
  │   ├── 查询 (system:user:query)
  │   ├── 新增 (system:user:add)
  │   ├── 编辑 (system:user:edit)
  │   └── 删除 (system:user:delete)
  ├── 角色管理 (system:role)
  │   ├── 查询 (system:role:query)
  │   ├── 新增 (system:role:add)
  │   ├── 编辑 (system:role:edit)
  │   └── 删除 (system:role:delete)
  └── 菜单管理 (system:menu)
      ├── 查询 (system:menu:query)
      ├── 新增 (system:menu:add)
      ├── 编辑 (system:menu:edit)
      └── 删除 (system:menu:delete)
基础数据 (baseData)
  ├── 工厂管理 (base:factory)
  │   ├── 查询 (base:factory:query)
  │   ├── 新增 (base:factory:add)
  │   ├── 编辑 (base:factory:edit)
  │   └── 删除 (base:factory:delete)
  ├── 车间管理 (base:workshop)
  │   ├── 查询 (base:workshop:query)
  │   ├── 新增 (base:workshop:add)
  │   ├── 编辑 (base:workshop:edit)
  │   └── 删除 (base:workshop:delete)
  ├── 生产线管理 (base:line)
  │   ├── 查询 (base:line:query)
  │   ├── 新增 (base:line:add)
  │   ├── 编辑 (base:line:edit)
  │   └── 删除 (base:line:delete)
  ├── 产品管理 (base:product)
  │   ├── 查询 (base:product:query)
  │   ├── 新增 (base:product:add)
  │   ├── 编辑 (base:product:edit)
  │   └── 删除 (base:product:delete)
  ├── 工艺路线管理 (base:route)
  │   ├── 查询 (base:route:query)
  │   ├── 新增 (base:route:add)
  │   ├── 编辑 (base:route:edit)
  │   └── 删除 (base:route:delete)
  ├── 工序管理 (base:operation)
  │   ├── 查询 (base:operation:query)
  │   ├── 新增 (base:operation:add)
  │   ├── 编辑 (base:operation:edit)
  │   └── 删除 (base:operation:delete)
  └── 配方管理 (base:recipe)
      ├── 查询 (base:recipe:query)
      ├── 新增 (base:recipe:add)
      ├── 编辑 (base:recipe:edit)
      └── 删除 (base:recipe:delete)
工单管理 (workOrder) - 待实现
  └── 工单列表 (wo:order:query)
批次管理 (lot) - 待实现
  └── 批次列表 (lot:lot:query)
WIP跟踪 (wip) - 待实现
  └── 流转记录 (wip:move:query)
设备管理 (equipment) - 待实现
  └── 设备台账 (eqp:equipment:query)
质量管理 (quality) - 待实现
  └── 检验记录 (qc:inspection:query)
```

## 10. 技术要点

### 10.1 后端技术要点

1. **使用Sponge框架生成代码骨架**：基于protobuf定义生成model、dao、service、handler代码
2. **JWT认证**：使用`golang-jwt/jwt`库实现token生成和验证
3. **GORM**：ORM框架，支持多数据库
4. **Gin**：Web框架，提供路由、中间件等功能
5. **配置管理**：使用`viper`读取YAML配置
6. **日志**：使用`zap`结构化日志
7. **参数验证**：使用`validator`验证请求参数

### 10.2 前端技术要点

1. **Soybean Admin**：基于Vue 3的现代化管理后台模板
2. **Naive UI**：Vue 3组件库
3. **Pinia**：Vue 3状态管理
4. **TypeScript**：类型安全
5. **Vue Router**：路由管理，支持动态路由
6. **Axios**：HTTP客户端
7. **vue-i18n**：国际化
8. **UnoCSS**：原子化CSS

## 11. 后续开发计划

### 阶段1：基础架构（本次完成）
- ✅ 设计文档
- ✅ 项目脚手架
- ✅ RBAC系统
- ✅ 模块1：基础数据建模

### 阶段2：工单和批次管理
- 工单CRUD
- 批次投放
- 批次拆分/合并
- Hold/Release功能

### 阶段3：WIP跟踪
- Track In/Track Out
- 流转历史查询
- 在制品报表

### 阶段4：设备管理
- 设备台账
- 设备状态监控
- PM计划和执行

### 阶段5：质量管理
- 检验计划
- 检验记录
- 缺陷分析
- SPC控制图

### 阶段6：报表和看板
- 生产进度看板
- WIP统计
- 质量分析报表
- 设备稼动率

## 12. 参考资料

- [Sponge文档](https://github.com/zhufuyi/sponge)
- [Soybean Admin文档](https://github.com/soybeanjs/soybean-admin)
- [GORM文档](https://gorm.io/)
- [Gin文档](https://gin-gonic.com/)
- [Naive UI文档](https://www.naiveui.com/)
