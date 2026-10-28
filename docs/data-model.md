# 数据模型设计

状态：规划；不是已执行迁移。对应 OpenSpec 变更 `define-traceable-quiz-platform`。

## 复用基础

复用 users、sessions、classrooms、memberships、handouts、knowledge_documents、knowledge_generations、knowledge_chunks。知识正文、定位、权限与活动代次以 MySQL 为准，Qdrant 只存标识 payload 与向量。新增业务 ID 使用 UUID 字符串，所有时间 UTC，字符集 utf8mb4。

## 新增实体

| 表 | 主要字段 | 约束与索引 |
| --- | --- | --- |
| generation_runs | id,class_id,created_by,paper_id,mode,status,parameters_json,source_snapshot_json,prompt_version,model_config_version,requested_count,settled_slots,accepted_count,rejected_attempts,started_at,finished_at,error_code | class_id+created_at+id；计数不得负数；源版本不可更新 |
| quiz_prompt_versions | id,kind,version,body,content_hash,created_by,created_at | unique(kind,version)；不可变；生成与语义判定分别引用版本；不含网关密钥 |
| generation_slots | id,class_id,run_id,ordinal,type,difficulty,status,accepted_item_id,attempts | unique(run_id,ordinal)；accepted_item_id 至多一个；ordinal 从 0 开始 |
| generation_attempts | id,class_id,slot_id,attempt_no,status,reason_codes_json,output_hash,restricted_output,started_at,finished_at | unique(slot_id,attempt_no)；unknown 也计预算；受限原始输出 7 天清理 |
| question_items | id,class_id,run_id,slot_id,revision,type,stem,options_json,answer_json,explanation,difficulty,knowledge_points_json,review_state,check_state,source_validity,content_hash,created_at,updated_at | class_id+review_state；revision≥1；选项/答案依题型校验 |
| question_sources | id,class_id,item_id,item_revision,handout_id,generation_id,chunk_id,ordinal,locator_json,char_start,char_end,quote,content_hash | item_id+item_revision；0≤start&lt;end；来源 class 与题目 class 一致 |
| question_checks | id,class_id,item_id,item_revision,attempt_id,overall,rule_version,judge_version,checks_json,duplicate_of,created_at | item_id+item_revision+created_at；质量报告不可覆盖旧修订 |
| question_reviews | id,class_id,item_id,item_revision,actor_id,decision,comment,created_at | 同修订审核保留历史；审批事务条件更新当前 revision |
| question_papers | id,class_id,title,status,version,requested_count,published_by,published_at,created_at | class_id+created_at+id；version 乐观锁；published 内容不可变 |
| question_paper_items | class_id,paper_id,item_id,item_revision,position,published_snapshot_json | unique(paper_id,position)、unique(paper_id,item_id)；发布保存完整内容和来源 ID/哈希快照 |
| quiz_jobs | id,class_id,kind,business_id,status,attempts,lease_token,lease_until,next_attempt_at,last_error,created_at | unique(kind,business_id)；status+next_attempt_at+lease_until；与业务记录同事务建立 |
| request_idempotency | class_id,actor_id,operation,key,request_hash,status_code,result_json,business_id,created_at | unique(class_id,actor_id,operation,key)；业务生命周期内保留；请求重放重新鉴权 |
| eval_datasets | id,class_id,version,split,manifest_hash,source_versions_json,rubric_version,created_at | unique(class_id,id,version)；冻结后不可原地改样本 |
| eval_cases | id,class_id,dataset_id,task_kind,input_json,expected_json,topic_labels_json,source_refs_json | dataset_id+id；不可回答样本明确 answerable=false |
| eval_runs | id,class_id,dataset_id,config_json,config_hash,baseline_id,status,gate_status,metrics_json,report_json,created_at | class_id+created_at+id；保留源版本与自检开关 |
| eval_results | id,class_id,eval_run_id,case_id,repeat_no,output_ref,metrics_json,status | unique(eval_run_id,case_id,repeat_no)；不混入生产 question_items |
| eval_annotations | id,class_id,result_id,annotator_id,rubric_version,labels_json,adjudicated,created_at | 每个标注者一份当前版本；最终裁定单独留痕 |
| eval_baselines | id,class_id,source_eval_run_id,version,dataset_hash,thresholds_json,approvals_json,created_at | 不可变；class_id+version 唯一；审批来自双方记录；导出版本化文件供门禁复用 |
| audit_events | id,class_id,actor_id,action,resource_type,resource_id,revision,request_id,metadata_json,created_at | class_id+created_at；不可包含令牌、密钥、完整讲义 |

实际索引名、varchar 宽度及迁移编号在迁移任务中落实；不得改变这里的关系、版本与唯一性约束。不再以“新增四张表”作为设计约束。

## 关系与删除规则

一个 run 有多个 slot，每个 slot 有多次 attempt，至多一个 accepted item；item 有多条来源、按 revision 累积的 check/review；paper 通过 paper_items 编排题目；eval_run 对固定 dataset 产生 results 与独立 annotations。

新表的资源引用须通过 `(id,class_id)` 组合唯一键和外键阻止跨班关联。对会被基础平台硬删除的 handout/generation/chunk，不建立级联删除题目的外键：保存标识与哈希作逻辑关联，在事务中验证存在性并在读取时判失效；避免讲义删除后题目静默消失。来源正文失效后禁止对外提供，受限审计只保留必要元数据。

题卷发布的快照记录精确 item_revision；后续编辑只能发生于未发布草稿。试卷内容及来源仍需按当前授权和源代次检查，快照并非绕过授权的副本。

question_checks 对生成中的候选绑定 attempt_id，此时 item_id 可空；候选通过并创建题目时关联 item_id 与初始 revision。编辑后复检绑定 item_id+revision，attempt_id 可空；至少一种关联必须存在。原始失败记录不会因没有正式题目而丢失。所有质量报告不可原地覆盖。

## 并发与事务

锁顺序固定为 class → handout（ID 排序）→ knowledge_document → paper → item。发布、题卷替换、讲义删除与活动代次切换遵循该顺序。编辑使用 `WHERE revision=expected_revision`，影响行数为零返回冲突；领取任务通过行锁、跳过已锁记录、租约令牌与条件提交实现。不得持有业务事务等待网关。

## 分级、审计与删除

遵循 project-standards.md 的公开/内部/私有/凭证四级、审计字段及 7/14/180 天等分类留存策略。删除处理覆盖 question_sources.quote、题干/答案及发布快照等派生正文，不只删除原讲义文件；保留的审计墓碑只含必要 ID/哈希与失效原因。已发布快照的不可变性不豁免隐私删除：只能通过审计化删除操作擦除正文并永久阻断使用，不能借此编辑成另一份可发布题卷。

在线立即撤销读权限、24 小时内清理在线副本，删除清单独立于业务备份保存且至少覆盖备份保留窗口；恢复后先重放删除清单，不得恢复已撤销凭证或复活已删除正文。所有外部删除由已持久化意图驱动并可重试，不夹带在持锁数据库事务内。
