# project-skeleton Specification

## Purpose

提供可克隆、可配置、可一键启动并通过公开健康接口验证的最小分层工程，作为后续教学平台与组卷功能的共同起点；严格区分进程存活与业务可用，所有运行日志保持单行 JSON。

## Requirements

### Requirement: 外置配置与一键启动
系统 SHALL 从 .env/环境读取运行配置，提供无真实凭证的 .env.example；仅需 Docker Compose 按 README 一条启动命令即可运行，前端为中文白红壳，不显示可用业务入口。

#### Scenario: 干净副本启动

**前置条件**：独立目录含全部受控文件，安装了 Docker Compose，可下载锁定镜像与依赖；未安装本机 Go/Node，未配置网关。

**操作**：复制 .env.example 为 .env，按需修改本机端口，执行 docker compose up --build -d --wait。

**可观测结果**：web/api 健康运行，GET /health 返回 200 且 status=ok，GET / 返回中文骨架页面；不要求登录、数据库或模型配置。

#### Scenario: 参数非法

**前置条件**：其余配置合法，API_PORT=70000。

**操作**：启动 API。

**可观测结果**：退出码 1，输出单行 JSON 且 error_code=CONFIG_INVALID，不监听端口，不把非法值转换为默认值。

### Requirement: 公开健康与默认拒绝
系统 SHALL 提供免认证 GET /health、GET /api/health 与 GET /api/openapi.json；健康不返回密钥或数据库/模型就绪断言。未知 /api 路径拒绝访问，未知非 API 路径 404，健康接口不接收查询参数。

#### Scenario: 无权限

**前置条件**：服务已启动，骨架没有账号或教师业务实现。

**操作**：无凭证请求 /api/classes/C1/quiz/runs。

**可观测结果**：返回 401 UNAUTHENTICATED，不返回伪造任务列表，不写业务记录，不调用网关。

#### Scenario: 资源不存在

**前置条件**：API 健康运行。

**操作**：请求不存在的 /missing-resource。

**可观测结果**：返回 404 NOT_FOUND 的 JSON，不重定向为成功页面。

#### Scenario: 非法健康筛选

**前置条件**：健康接口没有定义任何查询参数。

**操作**：请求 /health?status=unknown。

**可观测结果**：返回 422 VALIDATION_ERROR，不忽略参数后返回 200。

#### Scenario: 依赖不可达

**前置条件**：web 健康运行，API 容器已停止，代理超时配置为 5 秒。

**操作**：通过 web 请求 /health。

**可观测结果**：在 5 秒内返回 503 DEPENDENCY_UNAVAILABLE 的 JSON，不返回静态伪健康；web 错误日志仍为单行 JSON。

### Requirement: 分层与日志契约
代码 SHALL 遵守 handler → application → domain 的单向依赖；配置和日志为基础设施。API 和 web 每行输出一个合法 JSON，包含 timestamp、level、service、env、commit、event，请求日志另含 request_id、duration_ms、status；不得记录凭证。

#### Scenario: 请求与异常诊断均可解析

**前置条件**：web/api 已运行，日志级别 INFO。

**操作**：访问健康、未知路径并停止 API 后触发一次代理错误，读取两服务日志。

**可观测结果**：每条非空日志行均可独立 JSON 解析，能按 service/event 区分启动、访问及错误；日志无 Cookie/Authorization，领域代码不依赖 HTTP/SQL。

### Requirement: 验证与独立归档
骨架 SHALL 有 Go test/vet、前端类型/构建、契约双向/副本检查和独立容器验收；仅本变更任务完成后同步主规约并归档，标签为 v0.0.1-skeleton，长期业务任务不得自动勾选。

#### Scenario: 完成范围保持为骨架

**前置条件**：本变更全部骨架检查实际通过，完整业务规划仍有未完成任务。

**操作**：归档 bootstrap-project-skeleton 并创建版本标签。

**可观测结果**：主规约只新增 project-skeleton，本变更移入日期归档目录，长期业务规划继续活动；标签指向含代码、文档和验收证据的提交。
