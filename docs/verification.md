# 验证记录

日期：2026-10-28（Asia/Shanghai）。分开记录骨架实测和业务规划，不能将本页视为 AI 质量验收。

## 完整规划的历史检查

规划导入时：8 份业务能力增量、39 项需求、49 个三段式 Scenario、54 项实施任务。OpenSpec 严格校验、Scenario 格式、相对文档链接及提交文件检查通过。当前只完成长期任务 1.1 的目录骨架，其余 53 项业务/基础集成任务未实施。

## v0.0.1-skeleton 已执行

环境：Windows、Docker Desktop Linux 引擎 29.8.0；构建使用 .env.example 中锁定摘要的 Go/Node/nginx/alpine 镜像。项目运行不依赖本机 Go/Node；Node 24 用于可选验收脚本。

| 检查 | 命令/现象 | 结果 |
| --- | --- | --- |
| 规约 | npm run spec:validate；骨架与长期变更 | 严格校验通过 |
| 契约 | npm run api:check；Go 契约测试 | 仅 3 条实际 GET 路由，公开 security=[]，嵌入副本一致，双向比对通过 |
| Go | Docker backend-test 阶段的 go test ./... 与 go vet ./... | 配置边界、错误码、路由、日志及契约通过 |
| 前端 | Docker frontend-build 阶段的 TypeScript 与 Vite 构建 | 通过，无业务页面或假入口 |
| 一键检查 | npm run check | 规约/契约、Go、前端构建、隔离容器验收通过；不是完整业务门禁 |
| 干净副本 | 临时目录只复制受控构建输入，不含 .env/node_modules，再填模板执行 docker compose up --build -d --wait | 独立项目和临时端口启动 2 服务；健康、静态页面和资源可访问 |
| HTTP 失败路径 | /health 正常、未知 API、未知资源、错误方法、非法查询 | 分别为 200/401/404/405/422；JSON 错误含请求号 |
| 外部故障 | 停止隔离 API 后通过 web 请求 /health | 5 秒内 503 DEPENDENCY_UNAVAILABLE，不返回假健康 |
| 配置错误 | 单独以 API_PORT=70000 运行 API | 退出 1，单行 JSON 的 CONFIG_INVALID |
| 容器权限 | 检查 api/web 的 User 与 ReadonlyRootfs | api=65532:65532、web=101:101，均非 root、只读根 |
| JSON 日志及停机 | 隔离项目访问与故障后停止两服务，逐条 JSON.parse | 26 条非空行均通过，含请求号、状态、耗时及停机事件 |
| 页面回归 | BASE_URL 指向主实例，BROWSER_CHANNEL=msedge，npm run test:browser | 1440/768/375 px 无水平溢出；正常/错误/重试及 5 秒后仍正常，无 pageerror；桌面与手机截图已目视检查 |
| npm 审计 | 根目录及 frontend 的 npm audit --audit-level=high | 根目录 0；发现 Vite 7.3.6 高危后升级 7.3.7，前端复查 0 |

原始验收摘要和页面截图在忽略目录 tmp/，不提交运行目录或 .env。发布前仍要执行本地 Git 干净克隆验收，并将结果补入下节。

## 实施调整及失败修复

1. 本机 8080 已被课程服务占用，骨架默认改为 WEB_PORT=18080，不停止原服务。
2. 故障验收最初发现代理连接耗时可能超出 5 秒：拆出 1 秒连接超时、禁止隐式上游重试，门槛保持不变，复验通过。
3. 停机验收发现 nginx 基础镜像 SIGQUIT 会让 Go 适配器输出堆栈：覆盖 STOPSIGNAL 为 SIGTERM，同时显式捕获 SIGQUIT，复验日志均为 JSON。
4. 前端超时改由安全运行配置 PUBLIC_HEALTH_TIMEOUT_MS 注入；不暴露完整环境或网关信息。
5. Vite 按审计提示升级 7.3.7 并锁文件，未使用批量强制升级。

## 归档与发布范围

本次只归档 bootstrap-project-skeleton，主规约只同步 project-skeleton；define-traceable-quiz-platform 继续活动，不能将其 53 项未完成任务标为已交付。

本次可运行壳涉及跨层文件，按用户明确要求整体交付，超过默认 400 行/20 文件上限；分配置/契约、后端、前端、容器、克隆验收逐块核查。未声称其他组员已完成人工评审；用户本轮明确授权归档与版本标签，发布证据为实际自动检查与编码助手自测。

## 尚未实施或验收

课程平台导入、账号/班级/讲义业务、MySQL/Qdrant/worker、数据迁移、组卷、自检、审核、真实 AI、16 例快速集、完整质量门禁、数据库备份恢复及定时扫描均未实现。首次质量基线、人工标注和业务性能数据尚未建立。
