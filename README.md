# 可溯源 AI 出题与生成质量评测

基于班级讲义生成带有可核对出处的练习题，经质量自检与教师审核后发布，并通过评测与回归验证生成质量。

## 本机运行

当前为项目初始化阶段，尚无可运行的业务代码；启动步骤将在首个可运行版本交付时补充。

当前可以安装并校验 OpenSpec 规划工具：

```sh
npm ci
npm run spec:status
npm run spec:validate
```

## 项目规划

- [项目提案与范围](openspec/changes/define-traceable-quiz-platform/proposal.md)
- [架构与技术设计](openspec/changes/define-traceable-quiz-platform/design.md)
- [能力规约](openspec/changes/define-traceable-quiz-platform/specs/)
- [实施任务与分工](openspec/changes/define-traceable-quiz-platform/tasks.md)
- [API 与模块协议](docs/api-contract.md)
- [数据模型](docs/data-model.md)
- [评测方法与门禁](docs/evaluation.md)
- [开发流程](docs/development.md)
- [工程与产品规范](docs/project-standards.md)
- [规划验证记录](docs/verification.md)
- [选题说明 PDF](docs/小组项目选题说明.pdf)

小组成员：徐焜（生成与教师工作台）、徐高鹏（质量检查与评测）。规划工件齐全不代表功能已实现；当前 54 项实施任务均未开始。
