# Design

## Context

本仓库当前仅有 README、选题 PDF 和规划工具。2026-10-28 只读核对课程仓库，基础版本为 `xugaopeng0/courseproject-smartteachingwebsite@a798813fe06b0657b86e0ab84285adec2fdf2875`；本次没有复制或修改其代码。其受控文件采用 React 19、TypeScript 5、Vite 7、Go 1.26、net/http、database/sql 与 MySQL 驱动；具体补丁版本以导入时锁文件为准，不以这里的主版本替代锁定。

首次提交仅含 README.md 与 .gitignore，后续提交增加常驻规则和工程细则。本次规划提交将选题 PDF、OpenSpec 草案、设计文档和锁定的规划工具共同纳入版本控制，其他成员克隆后可按 README 安装工具并校验；业务代码仍未导入。

已观察：知识库有 `knowledge_documents/generations/chunks/jobs`，字符切片使用 Go rune，定位为 `{type,start,end,label,converted}`，worker 有租约令牌、重试与 `FOR UPDATE SKIP LOCKED`，但任务外键强绑定讲义与代次，不能直接插入组卷任务。当前仓库没有已实现主规约；动机和功能边界见 proposal.md。

## Goals / Non-Goals

**Goals:** 先冻结跨模块协议，建立可在本仓库独立运行的模块化单体，并使权限、来源代次、自检修订与发布事务可验证。

**Non-Goals:** 不在本轮导入业务代码或执行迁移；不拆微服务、不更换 ORM、不另建账号系统、不提前引入消息队列、不以历史平台资料宣称新仓库已验收。

## Decisions

### D1 基础导入与版本边界

实施阶段从上述精确提交导入必要受控目录与文件：backend、frontend、scripts、apitest、seeds、Compose、脱敏环境模板及契约。逐项审查许可与种子可公开性；课程网关实际地址、模型私有配置或账号密码一律替换为占位符。保留本仓库 README、OpenSpec 规划、Git 历史和工具锁文件；不得整体覆盖仓库或导入另一仓库 `.git`、`.env`、运行数据。导入清单、哈希和排除原因记入 docs/upstream.md。使用固定快照而不是子模块，降低两人组部署复杂度；后续上游更新使用独立变更审查。

### D2 模块化单体与前端组织

沿用 React + TypeScript + Vite，不更换框架。现有 Go HTTP 入口作为适配层，新领域放在 `backend/internal/quiz/`，分别组织 generation、quality、provenance、paper、evaluation；仓储与 gateway 通过接口注入，不做全项目重构。frontend 的 `features/quiz/` 包含 pages、components、api、types；原文展示复用现有组件，通过适配器处理定位差异。不强制引入 UI 组件库或全局状态库。

```text
frontend/src/features/quiz/       教师工作台、类型、本站 API 调用
backend/internal/quiz/           新领域服务与接口
backend/migrations/             后续递增迁移，保留既有编号
backend/                        原 HTTP、身份、知识库、worker 适配
docs/openapi.json               实施前写入的唯一机器接口契约
eval/{datasets,baselines}/       公开安全的脱敏样本与冻结配置
eval/reports/                   默认忽略，脱敏摘要另行提交
scripts/                        测试、契约同步、评测入口
openspec/                       规划和验收依据
```

这是目标目录，不代表当前已存在或可运行。前端页面为 `/classes/:id/quiz/runs`、`runs/:runId`、`papers/:paperId` 和 `evaluations`；入口按权限呈现，服务端权限判断始终保留。

### D3 同库事务与任务复用

扩展现有持久化 worker 的领取、租约、退避模式，新建 `quiz_jobs` 承载 generation、recheck、evaluation，不把它们伪装成 knowledge_jobs。不开发独立通用队列服务。任务与业务记录同事务创建；后续消息中间件仅改变唤醒/消费适配，不改变幂等与权限规则。

租约 120 秒、每 30 秒续期、单次网关超时 60 秒，失去租约不得提交。领取或恢复最多 3 次；每题位最多 3 次生成，attempt 序号在发起请求前持久化；崩溃时 unknown 调用同样占预算，不伪称外部调用 exactly-once。瞬态失败退避 2 秒、8 秒（附抖动），永久结构/权限/来源错误不重复请求同一无效输入。网关最多 2 个并发调用/worker，班级活动组卷上限 2。配置变更需记录版本。

使用租约 token + 当前任务状态的条件写与唯一键保障逻辑幂等。取消和完成都通过任务行锁串行化；任务计数从题位/尝试记录归约。API 2 秒轮询进度，首期不增加组卷 SSE，保留已有问答 SSE 行为。

同步 API 的数据库操作与连接使用请求级 5 秒总预算，依赖不可达时在请求开始后 5 秒内返回 503；此预算包含鉴权查询，数据库不可达不能误报 401。异步网关调用的 60 秒超时与题位 3 次预算独立计算。

### D4 配额、选材与来源协议

题型和难度均使用整数配额，按固定顺序 single_choice、true_false、fill_blank 与 easy、medium、hard 生成两个序列，逐位配对为题位计划并持久化，重试不改变配额。选材按 handout_id、ordinal 稳定排序、轮转分配；不保证模型输出正确，只保证候选不越界。候选范围、内容哈希、代次与提示词版本在任务创建时冻结。知识点由输出标注或教师编辑，首期不保证按知识点精确补题。

来源全文与区间以课程解析器规范化结果为准，不使用原文件字节偏移。模型接收每条候选的 chunk_id、文本及允许引用区间，仅返回 chunk_id、quote_start、quote_end、quote。服务端回查后派生可信身份与全局区间；不能映射到选定页段的跨界切片排除。对无法提取文本的扫描件明确不支持 OCR。

### D5 生成与质量协议

输入输出结构见 docs/api-contract.md；结构校验后依次进行来源/填空/选项规则、语义判定和去重。质量线提供稳定 `QualityReport`，包含 revision、rule_version、judge_version、overall、checks、duplicate_of；overall 为 pass/fail/inconclusive。语义返回必须结构化且理由有依据，不使用模型私有思维过程。仅 pass 可创建审核草稿。

题量上限 30，文本规范化精确去重后对当前任务/目标卷所有候选做轻量文本相似比较；同来源与 Qdrant 邻近来源补充候选，最后语义判定重复。不同来源也走文本候选检查。阈值由标注基线校准并冻结，不能把向量距离当成重复率。首期不新增题目向量集合。Qdrant 故障使需要其检查的任务进入有限恢复，不静默跳过检查。

编辑复检与逐题重生成的判重集合排除被编辑或被替换的旧题，但仍检查同卷其他题；并发候选接受时锁定题卷并核对判重时的版本，版本变化则重新检查，避免两个 worker 同时接受重复题。

### D6 状态、修订与发布

- run：queued → running → succeeded/partial/failed；queued/running 可进入 cancelled。
- item：draft → approved/rejected；内容编辑增加 revision 并回到 draft，check_state=pending；审核只接受当前 revision 的 pass。
- source_validity：valid/invalid，与 item 审核状态正交，避免污染审核历史。
- paper：draft → published；发布内容不可变，effective_validity 随来源变化计算，失效后不自动恢复为可用，需新建草稿。
- eval：queued/running/awaiting_annotation/completed/failed/cancelled；gate_status 独立为 passed/failed/blocked。

首次组卷创建草稿卷，题位通过即加入；整卷重生成创建新卷；逐题重生成锁定旧 revision 与题卷 version，成功后替换，失败或并发变更保留旧题。发布锁顺序为班级、按 ID 排序的讲义、knowledge_documents、题卷、题目；讲义删除与代次切换沿用相同前缀顺序。事务内复核活动代次及 revision，写发布快照和审计；如果导入的平台代码未采用协议，必须改适配并测试后才可发布。外部模型调用不放在数据库事务内。

读取与导出实时回查来源，不依赖批量失效任务的完成时机。删除可保留只含 ID/哈希的审计墓碑，不向普通题目接口返回失效来源正文。发布快照不授予永久访问原文的权限。

### D7 接口、数据与评测

实体关系及持久化约束见 docs/data-model.md，外部接口和错误见 docs/api-contract.md，评测定义及冻结规则见 docs/evaluation.md。这些文件是本变更设计的组成部分，变更时须同时更新相应能力规约和任务。机器契约 docs/openapi.json 在实现首批路由前完成，当前不创建空壳契约冒充完整 API。

### D8 运行与安全

沿用 web/nginx → api 与独立 worker → MySQL/Qdrant/课程网关的 Compose 拓扑，开发入口 localhost:8080，数据库和 Qdrant 仅内部网络。配置提供模板；实际网关配置只由服务端读取。Cookie 写请求校验 Origin；已有 Bearer 供脚本使用，浏览器不把网关凭证存储到 localStorage。

正常日志只含请求号、资源 ID、状态、耗时和错误码；候选原始输出放受限存储，默认保留 7 天后清理，审核/自检记录在课程验收期保留。公开仓库只提交脱敏且有权公开的合成评测样本；私有讲义、完整原始响应、真实学生数据不提交。模型输出及讲义片段按文本渲染，Markdown 禁止原始 HTML 和 javascript 链接。

### D9 产品与工程细则

以 docs/project-standards.md 固定页面四状态、术语、表单校验、视觉尺度和操作矩阵；实现按其正常/失败/边界测试。一键门禁、探针、停机和备份目标都在 tasks 中落地，不因写入规则而当作已实现。

严格筛选和唯一排序键避免误取全量及分页漏重；逐资源业务授权覆盖附件，出站 origin/端口白名单阻断模型指定地址。受控 JSONL 导入采用先整批验证再单事务提交，而非逐行部分成功；单卷导出先完整序列化、验证权限与大小再返回附件。限制与错误码同步至 docs/api-contract.md。

提示词源码模板受版本控制，发布时写入不可变 quiz_prompt_versions；生成和评审分别引用 ID/版本/哈希。16 例快速集用于安全与协议回归，不替代人工质量基线。新增案例不会改变既有完整样本目标。

本机 dev/test/demo 采用独立配置/数据卷，存活与就绪探针分离，探针不调用付费推理。迁移兼容回退优先于 down SQL；RPO/RTO 使用 24 小时/2 小时目标，以演练验证。删除立即撤销读权限，后台清理在线副本，恢复重放删除清单，避免备份绕过来源失效规则。

## Risks / Trade-offs

- [基础平台尚未导入] → 将基线导入及旧流程测试设为首个实施门槛，不直接开发一套平行系统。
- [同库锁协议涉及旧接口] → 用真实 MySQL 并发测试验证发布/重建/删除交错，不能仅 mock。
- [语义判断非确定] → 固定配置、保留人工标签和样本量，失败与 inconclusive 均不自动放行。
- [两人七周范围] → 核心三题型、轮询、JSON/Markdown 导出；扩展能力需新 proposal。
- [独立标注成本] → 在开发中持续积累样本，标注不足时报告 blocked，禁止临时降低门槛。
- [网关质量与耗时不确定] → 阈值在基线实测后冻结，不提前承诺质量/耗时数值。

## Migration Plan

1. 先建立 frontend、backend、scripts、eval、docs 工程骨架和接口边界，再审查固定基础版本，按清单导入、补环境模板并验证原有流程。
2. 补 OpenAPI 与协议测试，按现有最大迁移编号顺序增加 quiz 表；不编辑已应用迁移。
3. 用功能开关保持新入口关闭，完成生成、自检、审核、失效、评测和安全测试。
4. 在本机 Compose 验证后开放入口，记录版本与证据。
5. 回退时停止领取新任务、关闭入口、部署兼容旧版本，保留新表及数据；不以破坏性 down migration 回滚。失败的非事务 DDL 使用幂等迁移修复并验证，数据库升级前备份。

## Open Questions

无影响首期范围或架构的未决问题。网关实际地址/模型名、人工基线数值与运行环境性能作为实施阶段配置及测量产物，不是本轮可填写的事实。
