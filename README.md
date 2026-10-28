# 可溯源 AI 出题与生成质量评测

基于班级讲义生成带有可核对出处的练习题，经质量自检与教师审核后发布，并通过评测与回归验证生成质量。

## 当前版本

`v0.0.1-skeleton` 是可启动工程壳：React + TypeScript 页面、Go 分层 API、nginx 代理、外置配置及单行 JSON 日志。登录、讲义、数据库、组卷、审核、评测和 AI 调用尚未实现，没有可用业务入口。

## 本机启动

前置条件：Git，以及已启动的 Docker Desktop（Linux 容器，提供支持 `--wait` 的 Docker Compose）。首次构建需要下载锁定镜像和 npm 依赖的网络；运行壳不需要本机 Go、Node、数据库或模型密钥。

1. 克隆并进入目录：

   ```sh
   git clone https://github.com/xugaopeng0/traceable-ai-quiz.git
   cd traceable-ai-quiz
   ```

2. 复制配置模板。Windows PowerShell：

   ```powershell
   Copy-Item .env.example .env
   ```

   macOS / Linux：

   ```sh
   cp .env.example .env
   ```

   打开 `.env`，检查 `WEB_PORT=18080` 是否可用；如被占用只修改该值。模板其余值可直接启动壳，不需要填写真实密钥。`.env` 不提交 Git。

3. 用一条命令构建并启动：

   ```sh
   docker compose up --build -d --wait
   ```

4. 打开 [首页](http://127.0.0.1:18080) 和免认证的 [健康检查](http://127.0.0.1:18080/health)。若修改端口，URL 同步修改。健康检查返回 HTTP 200：

   ```json
   {"status":"ok","service":"traceable-ai-quiz","version":"0.0.1-skeleton"}
   ```

   页面显示“API 已连接”。这只证明进程可访问，不表示数据库或 AI 就绪。

查看状态与 JSON 运行日志：

```sh
docker compose ps
docker compose logs -f --no-log-prefix api web
```

构建进度与 Compose 自身提示不是应用日志；api/web 的每条非空运行日志均为一个 JSON。停止：

```sh
docker compose down
```

## 配置

所有部署及运行配置均在 `.env`，镜像以摘要锁定；修改后重新执行启动命令使容器读取新值。

| 配置 | 用途与约束 |
| --- | --- |
| COMPOSE_PROJECT_NAME | Compose 实例前缀；同机运行多份时分别填写 |
| WEB_BIND_HOST / WEB_PORT | 本机绑定地址与访问端口，默认 127.0.0.1:18080 |
| WEB_LISTEN_PORT / API_HOST / API_PORT / API_LISTEN_HOST | 容器网络与监听配置；通常保留模板，API_HOST 对应 api 服务 |
| APP_ENV / APP_NAME / APP_VERSION / BUILD_COMMIT | 环境 dev/test/demo、服务标识、版本与提交标识 |
| LOG_LEVEL | DEBUG / INFO / WARN / ERROR；验收日志使用 INFO |
| HTTP_READ_TIMEOUT_SECONDS / HTTP_WRITE_TIMEOUT_SECONDS / HTTP_IDLE_TIMEOUT_SECONDS | API 请求与连接超时（秒） |
| PROXY_CONNECT_TIMEOUT_SECONDS / PROXY_TIMEOUT_SECONDS | nginx 连接超时默认 1 秒，读写超时默认 5 秒 |
| PUBLIC_HEALTH_TIMEOUT_MS | 页面健康请求超时，100-30000 ms，默认 5000；是唯一公开页面运行配置 |
| SHUTDOWN_TIMEOUT_SECONDS / STOP_GRACE_PERIOD | 应用优雅退出默认 15 秒，容器宽限默认 30 秒 |
| HEALTHCHECK_* | 容器探针周期、超时、重试与启动宽限 |
| GO_IMAGE / NODE_IMAGE / NGINX_IMAGE / ALPINE_IMAGE | 构建镜像引用，模板包含精确摘要，升级需审查 |

公开接口只有 GET `/health`、GET `/api/health`、GET `/api/openapi.json`。前端只请求本站 `/api/health`。健康入口不接收查询参数，非法参数为 422；未知 API 为 401，其他未知资源为 404。web 连接不上 API 时返回 503，不返回假健康。参数不合法导致启动失败时查看日志的 CONFIG_INVALID。不要将模板中的运行配置当作已有数据库或网关配置。

## 目录与分层

```text
backend/cmd/                         进程装配与信号处理
backend/internal/transport/httpapi/  HTTP 路由、错误和契约
backend/internal/application/health/ 健康应用服务
backend/internal/domain/             与 HTTP/SQL 无关的领域类型
backend/internal/platform/           配置与 JSON 日志
backend/internal/quiz/               后续业务预留
frontend/src/                       React 页面
frontend/src/features/quiz/          后续教师模块预留
scripts/                            骨架检查与验收
openspec/                           规约、活动变更及归档
```

依赖方向为 transport → application → domain，平台能力通过装配注入。当前只有 web/api 两服务，后续业务变更再接 MySQL/Qdrant/worker，未导入课程平台业务代码。

## 开发检查

可选开发工具：Node.js 24+ 和 npm。下面检查使用容器内 Go/Node，不需要本机 Go。

```sh
npm ci
npm run check
```

`check` 执行 OpenSpec 严格校验、OpenAPI 副本检查、容器内 Go test/vet、前端类型/构建，再复制干净工程至独立临时目录，创建独立 Compose 项目，验证健康、401/404/405/422/503、非法配置、非 root/只读根、优雅退出和逐行 JSON 日志。失败非零退出；只清理本轮临时目录与容器。运行中的主实例不受影响。

这不是完整业务门禁：数据库、真实 AI、16 例快速集与完整质量评测仍待后续任务实施。

可单独执行：

```sh
npm run test:skeleton
npm run api:check
npm run spec:validate
```

浏览器回归另需 Playwright 浏览器，保持主实例运行后：

```sh
npx playwright install chromium
```

Windows PowerShell：

```powershell
$env:BASE_URL='http://127.0.0.1:18080'
npm run test:browser
```

macOS / Linux：

```sh
BASE_URL=http://127.0.0.1:18080 npm run test:browser
```

也可设置 `BROWSER_CHANNEL=msedge` 使用本机 Edge。回归检查 1440/768/375 px、正常/错误/重试及无水平溢出；截图写入被忽略的 tmp/。浏览器测试不会操作业务数据。

## 项目规划

- [常驻规则](AGENTS.md)与[工程规范](docs/project-standards.md)
- [已交付骨架规约](openspec/specs/project-skeleton/spec.md)与[骨架归档](openspec/changes/archive/2026-10-28-bootstrap-project-skeleton/design.md)
- [完整业务提案](openspec/changes/define-traceable-quiz-platform/proposal.md)、[技术设计](openspec/changes/define-traceable-quiz-platform/design.md)、[后续任务](openspec/changes/define-traceable-quiz-platform/tasks.md)
- [业务 API 设计](docs/api-contract.md)、[实际机器契约](docs/openapi.json)、[数据模型](docs/data-model.md)
- [评测设计](docs/evaluation.md)、[开发流程](docs/development.md)、[验证记录](docs/verification.md)、[选题 PDF](docs/小组项目选题说明.pdf)

小组成员：徐焜（生成与教师工作台）、徐高鹏（质量检查与评测）。
