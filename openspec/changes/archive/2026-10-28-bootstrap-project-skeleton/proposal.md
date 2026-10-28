# Proposal

## Why

现有完整规划尚不能运行，组员需要先获得按 README 可独立启动和验证的分层工程壳。将骨架从长期业务变更中单独交付、验收和归档，避免将未实现业务标为完成。

## What Changes

- 只建立 React/TypeScript 前端、Go 分层 API、nginx 代理及测试/评测目录。
- 通过 .env 与 Docker Compose 一条命令启动 web/api；配置严格校验，镜像按摘要固定。
- 提供免认证 GET /health、本站 /api/health 别名及只包含已实现路由的 OpenAPI。
- API 与 nginx 访问/错误输出统一为逐行 JSON，支持请求号及优雅停机。
- 验证干净副本启动、失败路径、容器权限、日志及前端构建，完成后归档并标记 v0.0.1-skeleton。

## Capabilities

### New Capabilities

- `project-skeleton`：外置配置、一键启动、分层工程、公开健康入口、JSON 日志与可复现验收。

### Modified Capabilities

无。完整业务变更仅将目录骨架任务关联本变更，其他业务任务不执行。

## Impact

新增 backend、frontend、Dockerfile、compose.yaml、.env.example、scripts 与最小机器契约；更新 README 和规范的当前阶段描述。

## Non-goals

不导入课程平台业务代码；不实现账号、数据库连接、worker、Qdrant、讲义、组卷、审核、评测和真实 AI 调用。不对未实现业务建立假接口、假成功或假门禁。
