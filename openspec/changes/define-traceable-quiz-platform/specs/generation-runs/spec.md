# Spec Delta

## Purpose

定义教师从班级已发布讲义提交组卷任务后可以依赖的范围、数量、状态与失败恢复行为，确保长耗时生成过程可观察、可恢复且不会因为重试产生重复题目或伪装已达到要求的题量。

## ADDED Requirements

### Requirement: 范围与配额
系统 SHALL 支持 1-10 份当前班级已发布讲义，按整份或已有定位区间选材；题量 1-30，题型为 single_choice、true_false、fill_blank，难度为 easy、medium、hard。客户端提交题型与难度的整数配额，各自之和均等于题量。区间上下界为正整数且 start≤end，同一讲义重叠范围合并。不支持的定位或题型返回 422。

#### Scenario: 参数不合法

**前置条件**：教师已登录并拥有该班权限，所选 Word 讲义有可用已发布代次。

**操作**：请求 10 题但难度配额之和为 9，或 Word 讲义请求 PDF 页码

**可观测结果**：返回 422，指出字段错误且不创建生成任务

### Requirement: 来源版本快照
系统 SHALL 在接收任务时冻结讲义代次、内容哈希、提示词与模型配置版本，按请求范围限制证据；不得因素材不足自动扩大范围。跨范围切片只能使用范围内可定位的文本，无法精确映射时排除该切片。

#### Scenario: 范围没有可用素材

**前置条件**：本班讲义已发布，但 PDF 第 2-3 页不含可用文本切片。

**操作**：提交 PDF 第 2-3 页的组卷请求

**可观测结果**：返回 422 SOURCE_EMPTY，且不创建任务或消耗模型调用

### Requirement: 持久化任务与幂等
系统 SHALL 接收后返回 202 和任务 ID；同一班级、操作者、操作与 Idempotency-Key 的相同请求返回原结果，不同请求体返回 409。每个班级最多 2 个未终结组卷任务，超限返回 429。每个题位固定题型与难度，首次生成加重生成合计最多 3 次；整任务基础设施重领最多 3 次，不重置题位尝试预算。

#### Scenario: 重复提交或 worker 中断

**前置条件**：已保存 1 个任务及其题位计划，题位唯一键和租约机制已启用。

**操作**：同一请求重复提交，或 worker 取得租约后中断并被接管

**可观测结果**：只有一个逻辑任务，每个题位最多一个有效题目，失去租约的 worker 不得提交结果

### Requirement: 真实进度与终态
任务 SHALL 提供 queued、running、succeeded、partial、failed、cancelled 状态，以及 requested_count、settled_slots、accepted_count、rejected_attempts。全部题位结束时，足额通过自检为 succeeded，部分通过为 partial，零通过为 failed；accepted_count 只代表自检通过，不代表教师通过。

#### Scenario: 达到尝试上限

**前置条件**：请求题量为 10，每题位最多尝试 3 次，其中 8 个题位已通过自检。

**操作**：10 个题位中 8 个通过，另 2 个耗尽预算

**可观测结果**：终态为 partial，settled_slots=10、accepted_count=8，并报告失败原因，不伪装完成 10 题

#### Scenario: 教师取消任务

**前置条件**：教师拥有班级权限，存在 1 个 queued 或 running 的组卷任务。

**操作**：教师取消 queued 或 running 的任务

**可观测结果**：终止后续生成，迟到结果不再写入，已接受草稿保留且任务为 cancelled
