# 质量评测与门禁设计

状态：规划，尚无真实基线数据。本文件定义计算和决策方法，不声明任何质量阈值已达成。

16 例 AI 快速集的分类、受控 JSONL 导入上限、缺陷判定信号与一键门禁遵循 project-standards.md。快速集是完整问答/出题评测的前置补充，不降低下述人工样本目标；固定输出回放只验证确定性规则。

## 数据与标注

问答开发/验收总目标 40-50 例；出题人工样本总目标 100-150 题，覆盖三类核心题型和三个难度。样本按固定 split（建议开发 70%、保留验收 30%，按题型难度分层）预先分割，保留集不得用于提示词调优。同源近似题放在同一 split，避免泄漏。

人工标签包括 answer_correct、source_correct、unambiguous、difficulty_match、duplicate_group、topic_labels。徐焜和徐高鹏独立标注，再记录分歧和最终裁定；只有完整裁定样本进入正式人工指标。模型判定仅作辅助标签，不能填入人工字段。

问答 case 包含 question、answerable、relevant_chunk_ids、reference_answer、expected_facts；出题 case 包含讲义版本、范围、题量/配额、target_topics。评测清单记录版本、source hash、parser/chunker/prompt/judge/model 配置版本、split 和样本哈希。真实讲义只存本地私有数据，公开仓库仅合成或已获授权脱敏样本。

## 指标定义

| 指标 | 精确定义 | 方向 |
| --- | --- | --- |
| delivery_completion | 人工审核通过的最终题数 / 请求题数，跨任务微平均；工作台自检接受率另名 check_acceptance | 高 |
| candidate_usability | 人工判为可用的候选 / 已生成候选；候选包括返回了原始输出但解析失败者，后者必不可用；无响应调用不算候选但计失败调用 | 高 |
| source_accuracy | 全部引用位置正确且支持考查内容的题数 / 已完整标注题数 | 高 |
| answer_accuracy | 标准答案正确的题数 / 已完整标注题数 | 高 |
| duplicate_rate | 每个语义重复组除第一题外的题数之和 / 参与判重的题数 | 低 |
| topic_coverage | 有效题目覆盖的预设目标知识点数 / 目标知识点总数；未预设知识点则 null | 高 |
| duration_ms | 创建至终态，另报排队与执行时间、任务中位数、p95、总执行时间/有效题数 | 低 |
| recall_at_4 | 每个可回答样本的 top4 与相关切片集合交集数量 / 相关切片集合数量，再宏平均 | 高 |
| mrr_at_4 | top4 首个相关结果排名倒数，未命中为 0，再宏平均；不宣称全列表 MRR | 高 |
| refusal_accuracy | 可回答且未拒答 + 不可回答且拒答 的样本数 / 全部样本数；单独报告误拒答率和漏拒答率 | 高 |
| faithfulness | 有来源支持的可核验断言数 / 全部可核验断言数；拒答与无断言单独计数，不赋 1 | 高 |

每项记录 numerator、denominator、sample_count（适用时）、value 和 missing_reason。空分母为 null，不能当作 0 或 100%。人工抽样无法代表全部候选时 candidate_usability 明确标为 sampled estimate，报告抽样方法和样本量；自动门禁只比较相同抽样设计的估计值。API 成功率、失败调用数、重试次数和网关 token 用量（网关提供时）另列运行指标。

## 实验设计

核心比较：提示词基线/改进版本，以及语义自检开/关；保持相同来源、任务、样本和参数，默认每配置执行 3 次独立重复，记录 seed（若网关支持）与实际模型版本。不同模型版本变化必须记录，不宣称完全确定性复现。结构、范围和权限校验在所有配置中始终开启；实验输出不得进入生产题卷。

报告每次结果与汇总，避免只选择最好一次。语义自检关闭配置仍由人工评价最终质量，不能用“检查没失败”推断质量。切片策略实验是扩展目标，需要独立索引版本。

## 冻结基线

首次测量后，两人共同审阅，在版本化 baseline 文件中记录 baseline_id、dataset/split/manifest_hash、rubric_version、model_config、prompt_version、sample_policy、每项 direction/absolute_limit/allowed_regression/min_samples、审批人与时间。阈值由实测确定，不在规划中填猜测数值。自动门禁不得写回基线；改基线需独立审查并说明原因。

比较时数据集、split、来源、标注规范和抽样政策必须一致；提示词等实验自变量允许不同，差异明确记录。质量下限用 min，耗时与重复率上限用 max；相对基线允许偏差也必须满足。样本不足或必需指标为 null 即 blocked；任何质量指标越界即 failed；全部满足才 passed。

确定性安全测试（跨班、伪造来源、发布竞态、无效结构）必须全部通过，不允许用质量统计容差豁免。无密钥的 CI 执行固定 fixture/契约/规则回归，真实网关质量评测为受控运行；影响生成质量的变更合并前附对应提交的完整评测报告，否则不得宣称通过。若基线尚未建立，首轮实验仅用于建基线，门禁状态为 blocked（baseline_missing）。

## 报告与退出码

报告结构：`{schema_version,run_id,commit,dataset_manifest,config,repeats,metrics,annotation_status,gate:{status,baseline_id,violations,missing},started_at,finished_at}`。planned/running/awaiting_annotation 不返回虚构完成时间。门禁命令退出码：0 passed、1 failed、2 blocked 或无效配置。共享报告去除原文、凭证和个人资料；原始报告默认存被忽略目录。
