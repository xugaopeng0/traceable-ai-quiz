# Design

## Context

从 define-traceable-quiz-platform 的目录骨架任务拆出本次可运行交付。业务任务、五服务拓扑和完整质量门禁仍属于后续变更；用户明确要求本轮只完成壳。

## Goals / Non-Goals

**Goals:** clone → 复制并填写 .env → docker compose up --build -d --wait → GET /health=200；无需本机 Go/Node 即可启动。日志为每行一个 JSON。

**Non-Goals:** 不导入业务平台，不增加数据库、索引与模型依赖，不执行长期规划中任何业务任务。

## Decisions

1. 结构使用 cmd/api → internal/transport/httpapi → internal/application/health → internal/domain；平台配置与日志在 internal/platform。领域无 HTTP/SQL 依赖，健康应用只报告进程存活。预留 internal/quiz、frontend/src/features/quiz、eval 和 scripts。
2. web/api 两个非 root、只读根容器是本轮最小拓扑。Compose 仅暴露 web 的本机端口；Go、Node、nginx、alpine 镜像用摘要锁定并由 .env 配置，依赖以锁文件构建。数据库、worker 和 Qdrant 待业务依赖变更引入，避免导入整套业务。
3. 运行端口、监听地址、版本、日志级别、超时与镜像引用均外置。应用配置不合法以 CONFIG_INVALID 的 JSON 日志退出 1，不静默回退。示例不需要任何账号/模型密钥；.env 不提交。
4. /health 与 /api/health 均免认证；前者用于文档验收，后者供前端遵守本站 /api 边界。健康只证明进程可访问，不声称数据库/模型就绪；后续 readyz 单独实施。未知 /api 路径默认 401，不伪造业务；其他未知路径 404，非法健康查询 422。
5. docs/openapi.json 只列已实现 GET 路由与实际响应。同步到 Go embed 副本，测试检查副本、路由及公开 security=[]；不把规划中的二十多个接口标为已实现。
6. 使用 Go JSON 日志工具统一字段。nginx 由轻量 Go 启动适配器托管，访问日志按 JSON 解析，原生 stderr 诊断转义后写 JSON，避免混入非 JSON 行；收到终止信号转发给 nginx 并等候退出。请求不记录查询串、Cookie 或 Authorization。
7. HTTP 请求/响应超时、nginx 代理超时、API 关闭时限均从环境载入；API 15 秒内优雅关闭，web 适配器转发退出，容器宽限 30 秒。/health 检查不使用业务网关。
8. 骨架门禁只跑实际存在的规约、Go test/vet、契约、前端类型/构建和容器验收；数据库、AI 和浏览器业务用例保持待实施，不能用空命令伪装全业务门禁通过。容器验收使用独立项目/临时目录，清理只删除本轮容器，无业务数据卷。

### 实施中确认的调整

- 本机 8080 已被原课程服务占用，骨架外部端口模板改为 18080；保留 WEB_PORT 可配，不停止其他项目。
- API 停止时，单一 5 秒代理连接超时无法保证端到端小于 5 秒；拆出 PROXY_CONNECT_TIMEOUT_SECONDS=1，关闭隐式上游重试，读写超时仍为 5 秒，故障验收保持原 5 秒门槛。
- nginx 基础镜像继承 SIGQUIT 停机信号，Go 默认会输出堆栈；web 镜像改用 SIGTERM，适配器同时捕获 SIGTERM/SIGQUIT 并向 nginx 转发优雅退出，实测停机日志全部 JSON。
- 前端健康超时通过 PUBLIC_HEALTH_TIMEOUT_MS 外置，在 web 启动时只生成这一项安全的 runtime-config.js；不把完整环境或网关配置发往浏览器。
- 依赖审计发现 Vite 7.3.6 高危项，升级到 7.3.7 并锁定；根及前端 npm 审计重新执行。原 React/TypeScript 与 Go 选型不变。
- 开发验收用干净副本检查未提交变动；发布前另用 --from-git 克隆本地已提交版本，验证没有 .env/node_modules 也能通过 README 的 Docker 命令启动。

## Risks / Trade-offs

- [首次镜像和依赖下载依赖网络] → README 明确 Docker 与网络前置条件，镜像摘要/锁文件保证内容一致，失败不归档为已验收。
- [骨架与原基础导入目录重叠] → 后续导入保留本轮分层入口，通过适配整合，不整体覆盖。
- [整体规划未完成] → 仅归档本变更，主规约只增加 project-skeleton；完整变更继续活动。
- [首轮跨层文件多于默认上限] → 用户明确授权完整可运行壳；验证按配置/契约 → 后端 → 前端 → 容器 → 干净副本分块进行，不混入业务。

## Migration Plan

无需数据库迁移。启动前复制 .env；停止使用 docker compose down。回退至原提交只停止骨架容器，不删除文档或重写历史。完成实测后归档本变更、提交并打 v0.0.1-skeleton 标签。
