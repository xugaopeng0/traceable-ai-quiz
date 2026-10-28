# 开发流程与验收映射

## 开始阅读

1. [项目提案](../openspec/changes/define-traceable-quiz-platform/proposal.md)：范围及非目标。
2. [技术设计](../openspec/changes/define-traceable-quiz-platform/design.md)：框架、基础导入、模块、状态、事务、运行与迁移。
3. [能力规约](../openspec/changes/define-traceable-quiz-platform/specs/)：八个能力的需求及场景。
4. [API 与模块协议](api-contract.md)、[数据模型](data-model.md)、[评测设计](evaluation.md)：跨成员的开发契约。
5. [实施任务](../openspec/changes/define-traceable-quiz-platform/tasks.md)：顺序、负责人和完成证据。
6. [项目工程与产品规范](project-standards.md)：交互、接口、安全、测试、运维、数据与完成定义细则。

## 当前工具

```sh
npm ci
npm run spec:list
npm run spec:status
npm run spec:validate
```

项目锁定 OpenSpec 1.13.1；本机工具可用 Node 22.12+ 或兼容的更新版本，依赖安装以 package-lock.json 为准。这里不提供未实现的 npm dev 或 Docker 启动命令。

OpenSpec 四份规划工件齐全只表示可进入实施，不表示功能实现。当前变更仍放在 changes 中；不要提前归档或把全部增量规约同步成已交付主规约。

起草新变更使用 `/opsx-propose` 工作流，在本环境对应 `openspec-propose` 技能；本轮修订沿用已有变更。proposal 写原因与边界，design 写选型与取舍，specs 按能力组织需求及 Scenario，tasks 给出可执行清单且骨架排第一。每个 Scenario 分为“前置条件 / 操作 / 可观测结果”三段，至少分别覆盖参数非法、无权限、资源不存在和依赖不可达；数值写成可验收的数量、区间、状态码和时限。

## 分工和依赖

徐焜负责生成、来源、任务编排和教师工作台；徐高鹏负责质量规则、语义判定、去重和评测。数据库迁移、接口、状态枚举、Candidate/QualityReport 及发布事务双方共同审查。

先完成基础导入和共同契约，再按 tasks 分阶段推进。W1-2 的出题基线为隔离实验最小驱动，不是生产组卷模块；正式自检与生成接入后，用相同评测样本继续对比。涉及跨模块协议的修改先修改规约和相关设计文件，再改代码。

## 协作约定

- 每项改动使用短分支，例如 `feat/generation-runs`、`feat/quality-checks`、`docs/spec-update`。
- PR 说明对应能力、任务编号、行为变化与验证；另一成员重点审核公共契约与安全边界。
- 提交可采用 `docs:`、`feat:`、`fix:`、`test:` 前缀，信息具体可读。
- TypeScript 开启 strict；Go 使用 gofmt、显式错误处理与有超时的 context；SQL 使用参数绑定。
- mock 只覆盖边界与单元规则，事务、外键、租约和竞态用 MySQL 集成测试。
- 生成质量改动必须附对照报告；依赖更新单独记录原因，不在功能改动中无目的升级全栈。

## 验收映射

| 能力 | 关键验证 | 主要任务 |
| --- | --- | --- |
| platform-foundation | 空环境复现、旧功能回归、教师/跨班/归档权限 | 1、3.1、8 |
| generation-runs | 参数、配额、幂等、崩溃恢复、取消、真实统计 | 3.2-3.3、4 |
| question-quality | 三题型负例、语义不确定、原文不支持、同义题 | 5 |
| question-provenance | Unicode、伪造来源、原文定位、重建/删除竞态 | 3.4-3.5、6.5 |
| teacher-workbench | 断网恢复、修订冲突、审核、重生成、发布导出 | 6 |
| quality-evaluation | 标注、手算指标、实验隔离、基线与退出码 | 2、7 |
| api-contract | OpenAPI/路由/鉴权/副本、错误与请求限制 | 1.5、3.1、8.1 |
| delivery-operations | Compose、日志脱敏、私有数据、真实证据 | 1.4、5.5、8 |

性能数值、真实网关质量、部署通过状态以验收阶段的实测证据为准，不从文档存在推导。
