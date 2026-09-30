# MES Server

半导体MES系统后端服务。

## 技术栈

- Go 1.22+
- Gin (Web框架)
- GORM (ORM)
- JWT (认证)
- SQLite/MySQL/PostgreSQL (数据库)

## 快速开始

```bash
# 安装依赖
go mod tidy

# 运行
make run

# 或
go run cmd/server/main.go
```

## 配置

配置文件位于 `config/config.yaml`

## 测试

```bash
make test
```

## API文档

详见项目根目录的 `docs/design.md`
