# 半导体 MES

基于 Sponge（Gin + GORM）和 Soybean Admin 的半导体制造执行系统。默认使用纯 Go SQLite，不需要 CGO。

## 已实现

- 登录、用户、角色、菜单，按钮级权限
- 基础数据：工厂、车间、产线、产品、工序、配方、工艺路线流程图
- 工单、开批、拆批、合批、Hold / 解除 Hold
- 过站：Track In / Track Out、取消进站、工位、在制总览
- 设备：台账、E10 风格状态、PM 计划和任务
- 质量：检验计划、量测、规格判定、缺陷和柏拉图
- 站内通知：批次 Hold、返工超限、PM 到期/超期、设备非计划停机。顶栏铃铛按当前用户区分已读/未读，中英文都返回
- SPC 代码仍在仓库里，默认关闭

## 环境

- Go 1.24（`CGO_ENABLED=0`）
- Node.js 22 与 pnpm 9
- 可选：MySQL 8 或 PostgreSQL 16

## 启动

在 `server` 目录启动后端，配置文件是 `server/configs/mes.yml`：

```bash
cd server
CGO_ENABLED=0 go run ./cmd/mes
```

服务监听 `http://127.0.0.1:8080`。首次启动会在 `server/data/mes.db` 建库并写入演示数据。

前端：

```bash
cd web
pnpm install
pnpm dev
```

`pnpm dev` 使用 test 模式，请求 `http://localhost:8080/api/v1`。页面端口以终端输出为准，通常是 9527。

默认账号 `admin` / `admin123`。登录页可以点「超级管理员」，账号已经填好。

## 功能开关

`server/configs/mes.yml`：

```yaml
features:
  spc: false
```

`spc: false` 时不播种 SPC 菜单和权限，顶栏路由里没有控制图，`/api/v1/qcSpc` 返回无权限，量测也不会做 OOC/OOS 反应。把 `spc` 改成 `true` 后重启，会显示 SPC 菜单并启用规则 1–4。测试套件在启动时把该开关打开，以便覆盖现有 SPC 用例。

## 切换数据库

只改 `database.driver`。SQLite 用 `sqlite.dbFile`，MySQL / PostgreSQL 用下面的 DSN。

```yaml
database:
  driver: sqlite # sqlite | mysql | postgresql
  mysql:
    dsn: "root:password@(127.0.0.1:3306)/mes?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_general_ci"
  postgresql:
    dsn: "postgres:password@127.0.0.1:5432/mes?sslmode=disable"
  sqlite:
    dbFile: "data/mes.db"
```

本地起库：

```bash
docker run -d --name mes-mysql -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=mes -p 3306:3306 mysql:8.4
docker run -d --name mes-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=mes -p 5432:5432 postgres:16
```

没有 Docker 时，用系统里的 MySQL 8 和 PostgreSQL 16，建好空库 `mes` 后改 DSN 即可。改完 `driver` 再重启后端。换库后请使用新的空库，种子数据会在第一次启动时写入。

用户名和角色编码的唯一性按「未删除的行」判断。软删除后可以再用同一个用户名或角色编码，已删除账号不能登录。不要给 `username` 加普通唯一索引，否则软删除的行会挡住重建。

## 测试

```bash
cd server
CGO_ENABLED=0 go test -count=1 ./internal/...

# MySQL
MES_TEST_DRIVER=mysql \
MES_TEST_DSN='root:password@(127.0.0.1:3306)/mes?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_general_ci' \
CGO_ENABLED=0 go test -count=1 ./internal/...

# PostgreSQL
MES_TEST_DRIVER=postgresql \
MES_TEST_DSN='postgres:password@127.0.0.1:5432/mes?sslmode=disable' \
CGO_ENABLED=0 go test -count=1 ./internal/...
```

`MES_TEST_DRIVER` 只影响 `internal/handler` 里的集成测试。MySQL 和 PostgreSQL 会在测试开始时清空 public / 当前库的表再播种。

前端：

```bash
cd web
pnpm typecheck
pnpm lint:ci
pnpm build
```

`pnpm lint` 会自动改文件。CI 用 `pnpm lint:ci`，只检查不改。

## CI

`.github/workflows/ci.yml` 在 push 和 pull request 上跑四件事：SQLite、MySQL 8.4、PostgreSQL 16 的后端测试，以及前端 typecheck、lint、build。

## 还没做

- 腔体 / 端口状态，PM 按晶圆片数计数
- 单独的 R 图（SPC 默认关闭，不再继续做图种）
- 邮件通知。目前只有站内通知
- 出站当时批次仍在加工，SPC 的 Hold 反应要等批次回到等待才会扣留
- 报表和稼动率看板
