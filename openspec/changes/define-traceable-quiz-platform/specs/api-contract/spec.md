# Spec Delta

## Purpose

统一浏览器、生成模块与质量模块依赖的接口及错误处理约定，为两名成员的独立开发和联调提供明确契约，同时保证权限、幂等、并发修订和异步状态在所有接口上的表现一致。

## ADDED Requirements

### Requirement: 单一接口契约
系统 SHALL 将 docs/api-contract.md 的接口设计在实现前固化为 docs/openapi.json（OpenAPI 3.1），包括请求响应、鉴权、路径参数、错误与示例；路由与契约双向校验，嵌入副本同步检查，不允许只新增路由未补契约。

#### Scenario: 路由或副本漂移

**前置条件**：已有 1 条新增路由及其后端实现，契约检查可在本机执行。

**操作**：实现新增接口但契约缺失，或嵌入副本与根契约不一致

**可观测结果**：契约检查失败，阻止合并

### Requirement: 统一输入输出与错误
新增接口 SHALL 使用 /api/classes/{class_id}/quiz 前缀、JSON snake_case 字段、UTC RFC3339 时间与不透明字符串 ID；分页默认 20、上限 100，使用游标。错误响应包含 code、message、request_id 和可选 details，不暴露模型密钥、堆栈或内部路径。

同步接口依赖数据库不可达时 SHALL 在请求开始后 5 秒内返回 503 DEPENDENCY_UNAVAILABLE，且不产生部分业务写入；异步网关失败遵循 generation-runs 的已持久化任务协议，不把已接收任务改成同步 HTTP 失败。

#### Scenario: 服务端异常

**前置条件**：请求已完成教师鉴权，依赖数据库或课程网关的操作即将执行。

**操作**：数据库或网关请求失败

**可观测结果**：对外提供稳定错误码与请求号，日志以请求号关联且脱敏

#### Scenario: 参数非法

**前置条件**：教师拥有班级 C1 权限，讲义 H1 属于 C1 且有活动代次，当前没有活动组卷任务。

**操作**：向 POST /api/classes/C1/quiz/runs 提交 question_count=0，其他字段均合法。

**可观测结果**：返回 HTTP 422、code=VALIDATION_ERROR，details 指出 question_count 最小值为 1；generation_runs、quiz_jobs 新增行数均为 0，网关调用数为 0。

#### Scenario: 无权限

**前置条件**：学生 S1 是班级 C1 的有效成员，拥有未过期会话，C1 中有题卷 P1。

**操作**：S1 请求 GET /api/classes/C1/quiz/papers/P1。

**可观测结果**：返回 HTTP 403、code=FORBIDDEN，响应不含题干、答案和讲义正文；网关调用数为 0。

#### Scenario: 资源不存在

**前置条件**：教师 T1 拥有班级 C1 权限，题卷 ID P404 不存在，数据库连接正常。

**操作**：T1 请求 GET /api/classes/C1/quiz/papers/P404。

**可观测结果**：返回 HTTP 404、code=NOT_FOUND，响应不含题卷内容或数据库错误；新增业务记录数为 0。

#### Scenario: 依赖不可达

**前置条件**：教师鉴权已完成，持久化存储操作尚未开始；故障注入使后续 MySQL 连接与查询不可达。

**操作**：请求 GET /api/classes/C1/quiz/runs，并从 HTTP 请求进入服务端起计时。

**可观测结果**：在 5 秒内返回 HTTP 503、code=DEPENDENCY_UNAVAILABLE 和非空 request_id；不返回缓存伪造列表，新增业务记录数为 0，日志不含连接凭证。

### Requirement: 写操作保护
浏览器写请求 SHALL 校验同源 Origin，沿用 HttpOnly、SameSite Cookie，会话到期返回 401；使用 Idempotency-Key 的异步提交与发布必须持久化请求哈希和结果。编辑与审核必须提交 expected_revision。

#### Scenario: 跨站请求或重复发布

**前置条件**：存在 1 个已审核草稿卷；浏览器携带有效 Cookie，或客户端持有已成功使用的发布幂等键。

**操作**：Cookie 写请求 Origin 非本站，或同一发布幂等键重复使用

**可观测结果**：前者返回 403，后者返回第一次结果且不重复发布或审计

### Requirement: 严格筛选与稳定排序
列表接口 SHALL 校验所有筛选、排序与分页参数，非法值或未知查询参数返回 422，不忽略后返回全量。默认 created_at DESC,id DESC；其他允许排序后也必须追加唯一 id，游标包含全部排序键。

#### Scenario: 非法筛选禁止返回全量

**前置条件**：教师拥有班级 C1 权限，该班有 3 条不同状态的任务。

**操作**：请求 GET /api/classes/C1/quiz/runs?status=unknown。

**可观测结果**：返回 422 VALIDATION_ERROR，details 指出 status 非法；响应不含 3 条任务列表，网关调用和新增业务记录均为 0。

#### Scenario: 相同时间分页不漏不重

**前置条件**：固定数据集有 3 条 created_at 完全相同的任务，ID 互异，期间没有增删。

**操作**：按默认排序以 limit=1 连续读取 3 页。

**可观测结果**：每页恰好 1 条，按 id 降序排列，3 条恰好各出现 1 次，游标包含时间及 ID。

### Requirement: 有界完整导出
导出 SHALL 限定为 JSON/Markdown 的试题版/教师版，单卷最多 30 题且 UTF-8 结果最多 5 MiB。超限返回 422 EXPORT_LIMIT，不截断为成功文件；试题版不含答案/解析，导出实时复核权限和来源。

#### Scenario: 导出超限不返回半份文件

**前置条件**：隔离测试构造 1 份已发布、来源有效但序列化结果为 5 MiB+1 字节的题卷快照。

**操作**：教师导出该题卷。

**可观测结果**：返回 422 EXPORT_LIMIT，不返回 200 或可误认为完整题卷的截断附件，不更改题卷状态。
