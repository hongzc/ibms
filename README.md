# ibms

基于 Gin 的后端服务，采用 route / service / store 三层架构。

## 分层

依赖方向单向：`route → service → store → model`

```
cmd/server/main.go    入口，串联三层
internal/
  config/   配置加载（Viper + config.yaml）
  model/    实体定义（跨层共享）
  store/    数据访问层：接口 + sqlite 实现
  service/  业务逻辑层：依赖 store 接口
  route/    HTTP 层：gin handler，依赖 service
```

- **store** 定义接口（如 `UserStore`）并提供 sqlite 实现，service 依赖接口便于替换/测试。
- **service** 编排业务逻辑，输入输出用自定义类型，不直接暴露 HTTP/SQL 细节。
- **route** 只负责参数绑定、调用 service、组织响应。

## 运行

```bash
go run ./cmd/server
```

默认监听 `:8080`，sqlite 数据库文件 `ibms.db`。

## 接口示例

```bash
# 健康检查
curl localhost:8080/health

# 创建用户
curl -X POST localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"alice","email":"alice@example.com"}'

# 查询用户
curl localhost:8080/api/v1/users/1

# 用户列表
curl localhost:8080/api/v1/users
```

## 新增一个模块

照着 user 的模式，在每层各加一个文件：

1. `model/xxx.go` — 实体
2. `store/xxx.go` — `XxxStore` 接口 + 实现，并在 `store.go` 的 `Store` 里挂上
3. `service/xxx.go` — `XxxService`，在 `service.go` 的 `Service` 里挂上
4. `route/xxx.go` — handler，在 `route.go` 里注册路由
