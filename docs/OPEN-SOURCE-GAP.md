# 开源数据中台功能完备度对比（OPEN-SOURCE-GAP）

- 调研日期：2026-10-11；方法：WebSearch/WebFetch 拉取各开源项目官方功能清单 + 逐条对照本仓库代码（logic/schema/前端页面）
- 范围：**功能完备度**。布局/RBAC 菜单字段级对标见 [docs/STANDARD-BENCHMARK.md](STANDARD-BENCHMARK.md)，本文不重复。
- 我方基线：16 服务（datasource / metadata / datasvc / catalog / quality / metric / tag / job / ide / ai / standard / mdm / ops / security / theme / sys）+ Vue3 控制台 16 组视图。
- 状态标记：✅ 有且可用；🟡 简化/部分；❌ 缺失。难度：S ≤3 天 / M 1~2 周 / L 月级。

---

## 1. 数据质量/可观测性：对标 Datavines

来源：<https://github.com/datavane/datavines>（README：27 条质量规则、4 类检查、SLA 告警、错误数据存储 MySQL/本地文件、告警渠道 Email、DataProfile、插件化）、官方文档 <https://datavane.github.io/docs/user-guide/quick-start>。我们对照 `quality/rpc/internal/logic/checker.go`（6 规则：NotNull/Unique/ValueRange/RowCount/Fluctuation/Consistency）。

| 模块 | Datavines 能力 | 我们 | 状态 |
|---|---|---|---|
| 质量规则 | 27 条内置规则（空值/唯一/枚举/正则/时间比较/UUID/长度/范围/跨表比对等） | 6 条（非空/唯一/值域/行数/波动/一致性） | 🟡 |
| 检查类型 | 4 类：单表单列、单表自定义 SQL、跨表准确性、两表值对比 | 单表规则 + 一致性（跨表 1 种），无自定义 SQL 规则 | 🟡 |
| 告警通知 | SLA 达成率告警；通知插件 Email（社区教程另提及钉钉/飞书通知模块） | ❌ 无任何告警/通知出口（checker 仅落结果） | ❌ |
| 错误数据 | 错误数据存储（MySQL/本地文件）+ 页面下钻查看 | ❌ 仅记通过率，不留错误样本 | ❌ |
| 结果分析 | 数据概览报告、指标趋势、分布分析 | listResults 列表查看，无趋势/分布视图 | 🟡 |
| DataProfile | 定时剖析、自动识别列类型、行数趋势监控 | ❌ 无 profiling | ❌ |
| 调度 | 内置调度 + datavines-submit.sh 对接外部调度 | runJob 手动/API 触发，未挂 DolphinScheduler | 🟡 |
| 元数据采集 | 数据源插件 8+（MySQL/ClickHouse/Doris/StarRocks/PG/Trino/Presto/Impala） | information_schema 采集（MySQL 系） | 🟡 |
| 架构 | 执行引擎插件（Local/Spark）、Server 横向扩展 HA | 单体内嵌执行，无分布式执行引擎 | 🟡 |

## 2. 可视化：对标 DataEase

来源：<https://github.com/dataease/dataease>（License GPL v3；数据源→数据集→仪表板链路；ECharts+AntV 双图表库）、官网 <https://dataease.cn>、v2.10.13 LTS 嵌入 SQLBot 智能问数（<https://github.com/dataease/SQLBot>，LLM+RAG text2sql）。我们：**无可视化模块**（web 仅 Dashboard.vue 统计卡），README 商业版清单中"数据可视化"整块未实现。

| 模块 | DataEase 能力 | 我们 | 状态 |
|---|---|---|---|
| 数据集 | 数据库表/SQL/Excel/自定义/联合查询，字段管理+计算字段 | ❌（datasvc 有动态 SQL，但非面向分析的数据集层） | ❌ |
| 图表 | 30+ 图表（ECharts+AntV），明细表/指标卡/透视图 | ❌ 仅 Dashboard 统计卡硬编码 | ❌ |
| 仪表板 | 拖拽布局、组件联动/跳转/下钻、主题、全屏 | ❌ | ❌ |
| 数据大屏 | v2 独立大屏模块（data-visualization） | ❌ | ❌ |
| 分享/嵌入 | 公开链接/密码/嵌入 iframe/定时邮件订阅 | ❌ | ❌ |
| 模板市场 | template-market 模板一键套用 | ❌ | ❌ |
| 移动端 | v2 mobile 模块 | ❌ | ❌ |
| AI 问数 | v2.10 嵌入 SQLBot（对话式分析、图表切换、问题推荐） | ai 服务 NL2Metric+AskAndRun+日志，无 RAG/图表输出 | 🟡 |
| 权限 | 组织/角色/用户 + 资源级授权（行列细粒度部分在企业版） | RBAC 菜单/按钮权限（sys） | 🟡 |

## 3. 工作流调度：对标 DolphinScheduler（job-rpc 管控面）

来源：<https://github.com/apache/dolphinscheduler/blob/dev/README_zh_CN.md>（可视化 DAG/依赖/HA/Failover/重跑/暂停恢复/补数据/优先级）；告警插件 Email/钉钉/企业微信/飞书/Webhook（官方告警组件文档）。我们对照 `job/rpc/internal/logic/dsclient.go`：仅 ping / resolveProject / ensureWorkflow / release(上下线) / start / latestInstanceStatus 共 6 类 OpenAPI 调用。

| 模块 | DolphinScheduler | 我们 job-rpc | 状态 |
|---|---|---|---|
| 工作流定义 | 拖拽 DAG 设计器、任务节点 40+ 类型 | 同步任务登记 → ensureWorkflow 建单节点流 | 🟡 |
| 依赖管理 | 跨工作流依赖（DEPENDENT 节点）、串并联动态调整 | ❌ 单任务单流，无依赖编排 | ❌ |
| 定时调度 | cron 定时管理、上下线、日历 | ❌ 未接 /schedule 端点 | ❌ |
| 补数据 | 按日期区间批量补数生成历史实例 | ❌ | ❌ |
| 失败处理 | 失败重试次数/间隔、重跑（失败节点/当前/从头）、优先级、暂停恢复 | ❌ 仅查询最新实例状态 | ❌ |
| 告警 | 告警实例+告警组，多渠道插件 | ❌ | ❌ |
| 多环境/隔离 | 项目/租户/环境/Worker 分组 | 固定单项目（resolveProject 单值） | 🟡 |
| 运维能力 | 实例日志、任务级日志、版本管理、导入导出 | 有 latestInstanceStatus + 日志端点，无版本/导出 | 🟡 |

## 4. 元数据：对标 DataHub / OpenMetadata

来源：DataHub <https://datahubproject.io/docs/>、Business Glossary <https://docs.datahub.com>（术语/术语组挂资产）、数据契约=assertions 集合（schema/freshness/DQ 三类，<https://docs.datahub.com> Column Assertions）、表级+列级血缘；OpenMetadata <https://github.com/open-metadata/OpenMetadata>（血缘/列级血缘/DQ 信号/所有权/使用分析/术语表）、<https://docs.open-metadata.org>（Domains/Data Products/Classifications(Tier)/Test Suites/Data Contracts）。我们对照 `metadata/rpc/internal/logic/`（collect/addLineage/getLineage/compareVersions）。

| 模块 | DataHub / OpenMetadata | 我们 metadata | 状态 |
|---|---|---|---|
| 采集连接器 | 80+ 连接器（DB/数仓/BI/消息/管道），pull-push 双模式 | information_schema 直查（MySQL 系 1 类） | ❌ |
| 血缘 | 表级+列级血缘，SQL 自动解析，跨系统（表↔任务↔仪表板） | 表级、手工登记/查询，无 SQL 解析 | 🟡 |
| 影响分析 | 血缘反向传播+下游影响评估 | ❌ | ❌ |
| 业务术语表 | Glossary 术语/术语组挂资产、审批流 | ❌（standard 词根/数据元未与资产/血缘打通） | ❌ |
| 数据契约 | assertions 组合（schema/freshness/DQ），契约状态面板 | ❌ | ❌ |
| DQ 集成 | 原生 assertions + GE/dbt/Great Expectations 结果回流展示 | quality 独立服务，结果未回写资产 | 🟡 |
| 治理对象 | Domains/Data Products/Ownership/Tags/Tier 分级 | catalog 资产分类+登记+申请审批；security 分级分类 | 🟡 |
| 搜索与发现 | 全局搜索（ES）、主页推荐、使用分析热度 | listEntities 列表过滤；遗留 search 服务未接元数据 | 🟡 |
| 版本/变更 | schema 变更历史+变更通知 | 定版快照+多版本对比（亮点，对标项目也仅部分覆盖） | ✅ |

## 5. 后台完备度：对标 RuoYi-Vue3

来源：<https://github.com/yangzongzhuan/RuoYi-Vue3> README 内置功能 19 项、<https://doc.ruoyi.vip>。我们对照 `sys/rpc/internal/logic/` 与 `web/src/views/system/`（5 页：用户/角色/菜单/登录日志/操作日志）。

| 模块 | RuoYi 内置功能 | 我们 sys | 状态 |
|---|---|---|---|
| 用户/菜单/角色 | 用户、菜单（按钮权限）、角色 | 同左，另有动态路由 getRouters | ✅ |
| 部门管理 | 公司/部门/小组树 + 按机构分配数据权限 | ❌ 无部门树 | ❌ |
| 岗位管理 | 用户岗位配置 | ❌ | ❌ |
| 数据权限 | 角色按机构划数据范围（全部/自定义/本部门及以下等 5 档） | ❌ 仅功能权限，无行级数据范围 | ❌ |
| 字典管理 | 固定枚举数据统一维护 | ❌（枚举散落各服务建表） | ❌ |
| 参数管理 | 动态系统参数 | ❌（各服务 yaml 手改重启） | ❌ |
| 通知公告 | 公告发布维护 | ❌ | ❌ |
| 日志 | 操作日志（含异常）+ 登录日志（含异常） | listAuditLogs + listLoginLogs | ✅ |
| 在线用户 | 会话监控、强退 | ❌ | ❌ |
| 定时任务 | 在线增删改任务调度（Quartz） | job 走 DS，本系统内无 Quartz 管理页 | 🟡 |
| 代码生成 | 前后端 CRUD 代码生成 | ❌（有 goctl，但无在线生成页） | ❌ |
| 监控 | 服务监控（CPU/内存/磁盘）、缓存监控、连接池监视、接口文档 | ❌（ops 仅归档/日志） | ❌ |
| 差异项 | — | 多租户字段 beid、防越权中间件（RuoYi 无） | ➕ |

## 6. 横向空白（README 商业版清单 vs 开源版）

| 模块 | 商业版对标（README 前 56 行） | 开源版 | 状态 |
|---|---|---|---|
| 数据集成 | 实时 CDC/离线接入（csv/excel/ftp/kafka/mqtt/mongodb）、拖拉拽集成任务 | ❌ 无采集引擎（job 仅登记同步任务转 DS） | ❌ |
| 数据可视化 | 数据集/报表设计/知识图谱 | ❌（见第 2 节） | ❌ |
| 数据仓库 | 主题域/维度建模/物化视图/模型审计 | theme 服务：主题域/模型/物化视图/部署调度（`theme/rpc/internal/logic/`），无建模画布与模型数据查询 | 🟡 |
| 数据安全 | 分级分类/脱敏/授权 | security：级别/脱敏规则/授权/预览（maskPreview） | ✅ |
| 数据运维 | 微服务/中间件/服务器监控、备份恢复 | ops：归档/恢复+登录/审计日志，无指标监控 | 🟡 |
| 前端补齐 | — | theme 有 api 无页面、security 无 api/页面 | 🟡 |

---

## Top 10 缺口清单（按"数据中台完整度"影响降序）

| # | 缺口 | 影响 | 难度 | 建议 |
|---|---|---|---|---|
| 1 | BI 可视化整块（数据集/图表/仪表板/大屏/分享订阅） | 中台"消费端"为零：治理成果无处呈现，演示与落地价值折半 | L | 集成优先：旁路部署 DataEase（GPLv3，注意与 Apache-2.0 主仓库隔离）或 AntV 自研最小集——基于 datasource 建"数据集"表 + 图表配置 + ECharts 渲染器（M 起步） |
| 2 | 数据集成/同步引擎（离线+CDC） | 没有集成，中台无数据可治，全靠外部灌入 | L | 接 SeaTunnel/DataX 为 job 新任务类型：登记任务→生成 DS 流→执行回读状态；拖拉拽后置 |
| 3 | 质量告警通知 | 质量问题发现不了人，Datavines/DS 均为标配 | M | 新增 alert_channel 表 + 通知器（SMTP/钉钉/飞书 webhook），失败率/波动超阈值触发；复用 datasvc 出口脱敏经验 |
| 4 | 错误数据留存与下钻 | 告警后无法定位坏数据，质量闭环断在最后一步 | S | checker 失败时执行 sample SQL 存错误样本表（Datavines 同范式 MySQL/文件），结果页加"查看错误数据" |
| 5 | 调度管控面深化（定时/依赖/补数据/重试/告警组） | 管控面只有"登记+启动"，替代不了 DS 原生界面 | M | DS OpenAPI 均有现成端点：/schedule、补数、/alert-group 逐个封装，前端 job 页加三个 Tab |
| 6 | 列级血缘 + 影响分析 | 元数据停留在表级手工登记，达不到 DataHub/OpenMetadata 底线 | M(影响) / L(列级) | 第一步零成本：基于已有表级血缘做反向影响分析接口+前端血缘图；列级再引 SQL 解析（sqllineage 思路） |
| 7 | 质量剖析 Profiling + 健康分 | 无表级"健康画像"，资产页缺质量分display | M | 复用 quality 执行器加 profile 规则（行数/空值率/基数/分布），按表聚合健康分，catalog 资产详情展示趋势 |
| 8 | RBAC 部门/岗位/数据权限 | 多租户有了但没有组织树，行级授权无处安放 | M | 补 sys_dept/sys_post/sys_role.data_scope（RuoYi 5 档），网关防越权中间件扩展部门过滤 |
| 9 | 业务术语表（glossary）+ 标准落地联动 | standard 词根/数据元与资产血缘脱节，治理"标准"悬空 | M | standard 升级：数据元→术语表实体，挂接 metadata 实体与 catalog 资产，对标 DataHub Glossary |
| 10 | 数据服务 API 可观测（调用量/成功率/文档门户） | OneService 有"管控"没"运营"，API 治理不完整 | M | datasvc 加调用日志聚合表+统计接口+前端大盘；再补在线 API 文档页（对标商业版接口测试/发布） |

次级缺口（本轮不入榜）：后台横切件字典/参数/公告/在线用户（S）、AI 问数补 RAG 与图表输出对标 SQLBot（M）、theme/security 前端页面补齐（S）、元数据采集连接器扩展 ClickHouse/PG（M）。

## 结论（如实定位）

1. **广度已达标**：16 服务覆盖"集成缺口外的治理主干"（源/元/标/质/指/标/服/资/调/IDE/AI/主数/安全/数仓/运维/RBAC），模块清单与商业版 README 一一对应，骨架形态对齐"Datavines + DolphinScheduler 管控面 + OneService + RuoYi 权限"的开源组合。
2. **深度约为对标的 20%~40%**：质量规则 6/27≈22%、DQ 告警/错误下钻 0%；DS OpenAPI 仅用 6 类端点（定时/补数/依赖/告警全缺）；元数据无列级血缘/术语表/契约，采集仅 1 类连接器；可视化 0%。单模块成熟度大致相当于各对标项目 2021~2023 早期版本形态。
3. **短板集中在"闭环"**：发现问题（质量告警）、定位问题（错误下钻/影响分析）、消费数据（BI/集成）三段闭环均断开；已实现的"管控类"能力（签名/限流/脱敏/版本对比/分级分类/多租户）反而是部分对标开源不具备的差异化。
4. **路径建议**：可视化与集成两块（缺口 1/2）坚决集成开源（DataEase/SQLBot/SeaTunnel）而非自研；缺口 3~10 全部可在现有服务内增量实现，合计约 2~3 人月可把"深度百分比"从 20~40% 抬到 60%+。
