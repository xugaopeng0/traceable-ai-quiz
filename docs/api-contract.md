# API 与模块协议设计

状态：规划。实现任何新路由前，将本设计转成 `docs/openapi.json` 的 OpenAPI 3.1 契约并通过契约测试。本文件不是可调用接口声明。

## 通用约定

下表路径相对于 `/api/classes/{class_id}/quiz`。全部要求当前班级教师；不存在或越班资源 404，未登录 401，本班学生 403。写 Cookie 请求要求本站 Origin；已有 Bearer 会话供脚本使用。归档班级只读，禁止新任务、编辑、审核、发布、导出。

新接口普通 JSON 响应直接返回对象；列表 `{items,next_cursor}`，limit 默认 20、1-100，cursor 为不透明的 created_at+id 游标。UTC RFC3339 时间、UUID 字符串标识。统一错误 `{code,message,request_id,details?}`；details 仅安全字段错误。

同步 API 的数据库连接与操作共享请求级 5 秒总预算，包含鉴权查询；数据库不可达在 5 秒内返回 503 DEPENDENCY_UNAVAILABLE，不误报用户会话失效。已接受异步任务中的网关失败记录在任务与尝试中，单次调用超时 60 秒、每题位最多 3 次生成，不改变已返回的 202。

所有筛选/排序/分页参数采用白名单，未知参数或非法值返回 422，不忽略筛选后返回全量。默认 created_at DESC,id DESC，游标包含全部排序键。当前未开放任意 sort；传入 sort 也属于未知参数。

异步提交、重生成、发布和基线冻结要求 `Idempotency-Key`（1-128 ASCII 字符），作用域为班级+操作者+operation。规范化 JSON SHA-256 作为请求哈希，相同键不同内容 409，重放先鉴权，永不重放失效权限下的敏感响应。同步编辑审核携带 expected_revision；题卷操作携带 expected_version。单请求最大 256 KiB，不允许任意远程 URL。

## 端点清单

| 方法与路径 | 请求要点 | 成功响应 |
| --- | --- | --- |
| GET /sources | handout_id?,cursor?,limit? | 200 已发布讲义摘要、代次、可用定位类型与边界，不暴露路径 |
| POST /runs | GenerationRequest | 202 run_id,paper_id,status |
| GET /runs | status?,cursor?,limit? | 200 任务摘要列表 |
| GET /runs/{run_id} | 无 | 200 配置、计数、当前状态、失败原因、paper_id |
| POST /runs/{run_id}/cancel | 无；幂等 | 200 当前终态，已结束保持不变 |
| GET /papers | cursor?,limit? | 200 题卷摘要列表 |
| GET /papers/{paper_id} | 无 | 200 version、题目顺序、有效性、当前题目内容及校验状态 |
| GET /items/{item_id} | 无 | 200 当前修订、来源、check/review 摘要 |
| PATCH /items/{item_id} | expected_revision,完整可编辑字段 | 200 新 revision，draft/pending；自动创建 recheck 任务 |
| POST /items/{item_id}/review | expected_revision,decision=approve/reject,comment? | 200 新审核状态与审计 ID；reject 时 comment 必填 |
| POST /papers/{paper_id}/items/{item_id}/regenerate | expected_version,expected_revision,reason | 202 run_id，成功后才替换原位置 |
| POST /papers/{paper_id}/regenerate | expected_version,reason? | 202 新 run_id,paper_id，旧卷不变 |
| GET /items/{item_id}/sources/{source_id} | 无 | 200 可信 locator、quote、全文区间、经鉴权原文相对 URL；失效 410 |
| POST /papers/{paper_id}/publish | expected_version,confirm_actual_count | 200 发布快照摘要；不足题量未确认返回 409 |
| GET /papers/{paper_id}/export | format=json/markdown,variant=student/teacher | 200 下载，Content-Disposition 附件；仅已发布且有效题卷 |
| GET /eval/datasets | cursor?,limit? | 200 该班评测集版本摘要 |
| POST /eval/runs | dataset_id,dataset_version,prompt_version,semantic_check,repeat_count=1..3,baseline_id? | 202 eval_run_id，配置白名单，服务器决定实际模型与地址 |
| GET /eval/runs | cursor?,limit? | 200 实验列表 |
| GET /eval/runs/{eval_run_id} | 无 | 200 实验状态、各指标与 gate_status |
| GET /eval/runs/{eval_run_id}/report | 无 | 200 结构化 JSON 报告；未完成附缺失项，不伪造指标 |
| POST /eval/runs/{eval_run_id}/annotations | result_id,rubric_version,labels,expected_annotation_version | 200 当前登录标注者的标签；不得冒充其他标注者 |
| POST /eval/runs/{eval_run_id}/adjudications | result_id,annotation_ids,final_labels,reason | 200 裁定记录，保留原始标签与裁定者 |
| POST /eval/baselines | completed_eval_run_id,thresholds,reason | 201 不可变基线版本；只允许已完成且标注足额的运行 |

评测集初期通过版本化文件及受控导入命令装载，不提供任意网页上传/执行脚本功能。已有身份、讲义下载、检索和问答接口保留原路径，禁止臆造或替换既有契约。

导入导出边界遵循 project-standards.md：评测集 JSONL 最多 5 MiB/500 行，整批校验后事务写入；错误提供 1 起始行号。单份题卷导出最多 30 题/5 MiB，报告下载同为 5 MiB，上限违反返回 422 EXPORT_LIMIT 而非截断成功。

独立运维路径不使用班级前缀：GET /api/healthz（200/503 进程存活）、GET /api/readyz（200/503 依赖与迁移就绪），单次 2 秒预算，无需用户会话，仅返回最小状态和请求号，不暴露版本细节或凭证。不会发起付费网关推理请求。

## GenerationRequest

```json
{
  "title": "第一讲巩固练习",
  "sources": [{"handout_id":"示例标识","scope":{"kind":"pdf_page","start":1,"end":3}}],
  "question_count": 6,
  "type_counts": {"single_choice":2,"true_false":2,"fill_blank":2},
  "difficulty_counts": {"easy":3,"medium":2,"hard":1},
  "include_explanation": true
}
```

title 1-100 码点；sources 为 1-30 条讲义范围记录，去重后涉及 1-10 份讲义，同一讲义可重复出现以表达多个区间；全篇 scope 为 `{kind:"all"}`，不得与同讲义范围混用。kind 支持 all/pdf_page/ppt_slide/word_paragraph/txt_line；接口适配基础 locator.type，不要求修改基础存储枚举。配额为非负整数且各自之和等于 1-30 的 question_count。服务端冻结实际模型、prompt_version、source_snapshot；客户端不提交任意模型或系统提示词。

## 逐题模型输出 Candidate

```json
{
  "type":"single_choice",
  "stem":"示例题干",
  "options":[{"id":"A","text":"选项一"},{"id":"B","text":"选项二"},{"id":"C","text":"选项三"},{"id":"D","text":"选项四"}],
  "answer":"A",
  "explanation":"依据引用解释答案。",
  "difficulty":"easy",
  "knowledge_points":["示例知识点"],
  "evidence":[{"chunk_id":"服务端候选标识","quote_start":0,"quote_end":4,"quote":"示例原文"}]
}
```

这是格式示例，不是已验证题目。实现使用 JSON Schema 2020-12、oneOf 和 additionalProperties=false；single_choice 四项各 1-300 码点，answer 为 A/B/C/D；true_false 的 options=[] 且 answer 为 boolean；fill_blank 的 options=[]，stem 仅一次 `{{blank}}`，answer 为 1-200 码点字符串。stem 1-1000，explanation 为 null 或 1-2000（是否必填取决于请求）；difficulty 必须等于题位计划；knowledge_points 1-5 个，每个 1-50 码点；evidence 1-3 条，区间不可越出允许范围。只检查核心题型，未知字段不透传。

## 生成线与质量线协议

输入 `QualityInput={candidate,slot_plan,allowed_sources,existing_items,rule_version,judge_config,mode}`；allowed_sources 由服务端加载，mode 只能是 production 或 evaluation。质量线不能自行读取其他班级来源。

输出 `QualityReport={overall,checks,duplicate_of,rule_version,judge_version,revision,latency_ms}`，overall=pass/fail/inconclusive；checks 为 `{code,status,reason,evidence_refs}` 数组，status=pass/fail/inconclusive。duplicate_of 是该任务/目标卷内候选或题目 ID，不允许跨班引用。quality 判断不直接修改 item/paper，由编排层在同一事务中写入 attempt、check、item 与进度。

## 错误码

400 MALFORMED_JSON；401 UNAUTHENTICATED；403 FORBIDDEN/ORIGIN_REJECTED；404 NOT_FOUND；409 IDEMPOTENCY_CONFLICT/REVISION_CONFLICT/PAPER_LOCKED/RUN_ACTIVE/REVIEW_REQUIRED/COUNT_CONFIRMATION_REQUIRED；410 SOURCE_INVALID；413 PAYLOAD_TOO_LARGE；422 VALIDATION_ERROR/SOURCE_EMPTY/UNSUPPORTED_SCOPE/EXPORT_LIMIT；429 RUN_LIMIT；503 DEPENDENCY_UNAVAILABLE；500 INTERNAL_ERROR。

质量失败码（写入检查/尝试，不等同于 HTTP 错误）：SCHEMA_INVALID、SOURCE_OUT_OF_SCOPE、SOURCE_INVALID、QUOTE_MISMATCH、ANSWER_NOT_IN_SOURCE、ANSWER_UNSUPPORTED、AMBIGUOUS、OPTIONS_INVALID、LOW_QUALITY、DIFFICULTY_MISMATCH、DUPLICATE、JUDGE_INCONCLUSIVE、GATEWAY_TIMEOUT、LEASE_LOST。
