# 数据标准 / 主数据 MDM / 数据运维与系统日志 —— 行业对标调研报告

> 调研时间：2026-10；调研方式：公开资料检索（阿里云/美团/Percona/开源仓库 schema 等）。
> 对标对象：本项目三个待开发模块（对应 README 功能清单 3/12/13/14 节，草稿表已落 `docs/schema/phase5.sql`）。
> 标注约定：所有关键论断标注【事实】（有来源或可复现）或【推测】（行业经验推导，需自行验证）。
> 项目既有约定（下文简称"既有约定"）：go-zero rpc 服务（目录 model/ + rpc/internal/logic）+ `dc_` 前缀表 + `beid` 多租户字段 + `datetime(3)` 服务端维护时间戳 + utf8mb4/InnoDB。

---

## 一、数据标准管理（词根/标准字典/数据元/标准模型/发布/多版本/标准核对）

### 1.1 行业方案要点

**阿里云 Dataphin / DataWorks 的模块划分**【事实，来源：阿里云官网产品页 https://www.aliyun.com/product/dataphin ；help.aliyun.com 文档《数据标准概述》《创建及管理数据标准》《新建和管理落标映射规则》；DataWorks 文档"数据建模-数据标准"】：

- 模块层级：**词根 → 数据元/数据标准 → 标准集/标准模板 → 落标映射 → 落标监控**。
  - **词根**：命名基础词汇（中英对照），用于标准化生成字段英文名（如"金额 amount"、"编码 code"）。
  - **数据元/公共标准**：企业内通用的标准对象，含**业务属性、技术属性、管理属性**三类；标准模板可复用"公共属性"（一次配置多次引用）。
  - **码表（= 标准字典）**：值域标准，枚举取值（如性别码表、行政区划码表）。
  - **标准集**：标准的分类目录容器，支持标准集目录管理。
  - **发布**：标准有"草稿→发布→废止"生命周期，发布后版本生效；跨租户发布走"发布包导入导出"（与本项目 beid 多租户天然同构）。
  - **落标映射**：把数据标准批量映射到物理表字段；支持内置模板自动映射。
  - **落标监控（标准核对）**：通过"标准属性 ↔ 资产元数据字段"的匹配关系自动生成映射并评估**落标率**（阿里云开发者社区《数据标准应用（三）：数据标准落标监控》2023-07，developer.aliyun.com）。

**国标 GB/T 18391《信息技术 元数据注册系统(MDR)》**【事实，对应 ISO/IEC 11179，来源：openstd.samr.gov.cn 国家标准全文公开系统；cnis.ac.cn 中国标准化研究院】：

- 分部分：框架 / 分类 / 注册系统元模型与基本属性 / 数据定义的形成 / 命名和标识原则。
- 核心定义：**数据元 = 通过"定义、标识、表示、允许值"一系列属性描述的数据单元**；**值域 = 数据元允许值的集合**。这给出了数据元表的最小字段集：标识码、名称、定义、数据类型、表示形式、值域（可引用码表）、管理员、注册状态。

**Apache Atlas type system**【事实，来源：https://atlas.apache.org 官方 Type System 文档；Cloudera 博客；github.com/wforget 笔记】：

- 元模型分两层：**TypeDef（类型定义）** 与 **Entity（类型实例）**；TypeDef 有 EntityTypeDef / ClassificationDef / RelationshipTypeDef / EnumTypeDef / StructDef 五类。
- 属性用 **AttributeDefinition**（name / dataTypeName / isOptional / cardinality / constraints）声明；所有类型继承自 **Referenceable** 基类型（提供 qualifiedName 唯一标识）；支持**类型继承与版本**。
- 对本项目的启示：不必引入 Atlas/JanusGraph 这类重量级图存储，用"类型定义表 + JSON 属性 + 实例表"即可复刻其骨架——Datavines 正是这么做的（见下）。

**Datavines（开源数据可观测性平台）**【事实，本次调研直接读取仓库 schema：https://github.com/datavane/datavines （dev 分支 `scripts/sql/datavines-mysql.sql`）】：

- **没有独立的"数据标准"模块表**（schema 中无 standard 相关表）；其"标准/类型"能力内化在目录模型里：
  - `dv_catalog_entity_definition`：实体定义（properties 用 JSON 数组存属性定义 `[{name,type}]`，**`super_uuid` 实现类型继承**）——即一个轻量版 Atlas TypeDef。
  - `dv_catalog_entity_instance`：实体实例（fqn 全限定名 + version + properties JSON map + status，`(datasource_id, fqn, status)` 唯一）。
- 结论【事实+推测】：开源侧没有可直接抄的"标准模块"成品；商业实现（Dataphin）的"词根/数据元/标准模型/落标"四层模型是最成熟蓝本，与本项目 README 目标功能一一对应。

### 1.2 表模型设计建议（字段清单）

对照 phase5.sql 草稿（dc_std_word_root / dc_std_dict / dc_std_dict_item / dc_std_element / dc_std_model / dc_std_check_report），行业对标后的**增补建议**【推测，基于上述事实】：

- `dc_std_word_root`（词根）：既有字段够用；建议补 `alias` varchar(64)（同义词根）+ `status` tinyint（词根停用后不允许新标准引用）。
- `dc_std_dict` / `dc_std_dict_item`（码表）：dict_item 建议补 `parent_id`（树形码表，如行政区划）、`ext_json`（如国标码/行业码对照）；补 beid 冗余到 item（现有 item 表已带）并加 `idx_beid_dict`。
- `dc_std_element`（数据元，对齐 GB/T 18391）：建议补
  - `word_roots` varchar(255)：组成该数据元的词根（逗号分隔），支撑"按词根自动生成 element_code/英文字段名"；
  - `definition` varchar(1024)：数据元定义（MDR 必备属性）；
  - `domain_type` varchar(16)：值域类型（dict 枚举 / range 范围 / free 自由文本），配合既有 dict_code；
  - `status` + `version` int（草稿/发布/废止生命周期，Dataphin 同款）。
- `dc_std_model`（标准模型）：既有 status/version/snapshot_json 只存"最近快照"，**多版本追溯建议新增历史表**：
  - `dc_std_model_version`：id, model_id, version, snapshot_json(mediumtext), change_log varchar(512), published_by varchar(64), published_at datetime(3), beid —— 每次发布插入一行，支持任意两版本 diff【推测，模式同 phase1 的 dc_meta_version】。
- 标准核对（落标）建议拆成两张，对齐 Dataphin"映射+监控"两步：
  - `dc_std_mapping`（落标映射，持久化）：id, model_id, element_code, meta_entity_id（关联 phase1 `dc_meta_entity`）, target_fqn, mapping_type（auto/manual）, beid, created_at —— 核对前先落映射，落标率才能统计到字段级；
  - `dc_std_check_report`（既有）：建议补 `summary_json`（缺列/类型不符/长度不符各计数）与 `pass_rate` decimal(5,2)，便于看板展示落标率。
- **标准核对实现思路**【事实+推测】：以标准模型的 elements_json 为期望 schema，对目标表执行 `SHOW FULL COLUMNS FROM t`（MySQL，复用 phase1 数据源连接器）逐列比对：缺失列 / 类型映射不一致（期望 varchar(64) vs 实际 varchar(32)）/ 可空性 / 值域不在码表（可选：抽样 `SELECT COUNT(*) WHERE col NOT IN (dict_items)`，交给 phase2 质量规则 value_range/consistency 承接更合适）。比对结果写 check_report，diff_json 结构 `[{column, expect_type, actual_type, result}]` 与草稿一致。

### 1.3 项目落地建议

- 服务形态沿用既有约定【推测】：新建 `std/` 服务（std.model + std/rpc，proto 含 WordRoot/Dict/Element/Model/Check 五组 CRUD + PublishModel + RunCheck），复用 `datasource` rpc 的连接器做 SHOW COLUMNS，复用 phase3 调度（cron 字段约定）做定期核对。
- 发布动作用 rpc 内"快照 + 版本号自增"模式，与既有 `dc_ide_script.version`、`dc_meta_version` 同构；发布时把 elements_json 冻结进 version 表。
- 多租户：标准对象（词根/码表/数据元/模型）全表带 beid 并以 `uk_beid_code` 隔离【事实：phase5 草稿已如此】；"跨租户标准共享"（Dataphin 的发布包）可后续基于 beid 导出/导入 JSON 实现，一期不做【推测】。
- 标准核对结果可作为 phase2 质量规则的数据源（落标率低的表生成质量问题工单），打通两模块【推测，Dataphin 同样把落标监控挂进治理分】。

---

## 二、主数据管理 MDM（主数据模型/编码码段/主数据管理/分发）

### 2.1 概念界定（一句话版）【事实】

来源：Precisely（precisely.com/blog "Metadata Management vs MDM"）、Stibo Systems（stibosystems.com）、CloverDX（cloverdx.com）：

- **主数据（Master Data）**：驱动业务流程的核心业务实体数据（客户、产品、供应商、物料）——"业务名词"，相对稳定但会变更。
- **参考数据（Reference Data）**：跨系统共用的稳定查表/分类值（国家码、币种、状态码）——比主数据更稳定，正好对应模块一的"标准字典/码表"，两模块共用 `dc_std_dict` 即可，不必另建。
- **元数据（Metadata）**：描述数据的数据（结构、定义、血缘），对应本项目 phase1 `dc_meta_entity`。

### 2.2 码段设计最佳实践（前缀+日期+流水号）

**美团 Leaf 号段模式**【事实，来源：美团技术团队《Leaf——美团点评分布式ID生成系统》2017-04 https://tech.meituan.com/2017/04/21/mt-leaf.html ；官方仓库 https://github.com/Meituan-Dianping/Leaf （本次调研读取 `scripts/leaf_alloc.sql` 原文）】：

- 官方号段表（照录）：
  ```sql
  CREATE TABLE `leaf_alloc` (
    `biz_tag` varchar(128) NOT NULL DEFAULT '',
    `max_id`  bigint(20) NOT NULL DEFAULT '1',
    `step`    int(11) NOT NULL,
    `description` varchar(256) DEFAULT NULL,
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`biz_tag`)
  ) ENGINE=InnoDB;
  ```
- 取号段为**单条原子 UPDATE**（不再 SELECT 后回写）：`UPDATE leaf_alloc SET max_id=max_id+step WHERE biz_tag=?`，拿回 `[max_id-step+1, max_id]` 区间在内存中发放——并发安全由数据库行内原子自增保证，DB 压力从"每单一次"降为"每 step 一次"。
- **双 buffer 优化**：当前号段消耗到约 90%（剩余 10% 水位）时异步预加载下一号段到备用 buffer，用尽无缝切换；官方数据：单节点约 5 万 QPS、TP999 1ms，可容忍 DB 宕机 10~20 分钟。
- **动态 step**：根据号段消耗周期自适应（上一次更新时间间隔越短，step 越大）。
- 业务编码=前缀+日期+流水号的组合：号段表只负责"流水号段"，前缀/日期在应用层拼接；**按日重置流水**用"日期作为号段 key 的一部分"（key=`model:20261010`）实现，避免回改 current_value【推测，Leaf 无此场景，属常见扩展做法】。

**开源/商业 MDM 的码段与数据模型**【事实，来源：https://pimcore.com （MDM/PIM 一体，GPL 社区版）；qData 数据资产管理平台（github.com/642933588 ，社区版文档 community.qdata.tech）：其 MDM 模块支持"主数据实体建模 + 编码规则生成（流水码/日期码/自定义码段）+ 订阅分发与推送记录追溯"】：

- 通用模型四件套（各家一致）【事实+推测】：模型定义（实体+字段 schema）→ 编码规则（码段）→ 主数据实例（记录+版本+状态）→ 分发记录（订阅方+推送结果）。phase5.sql 草稿的 dc_mdm_model/segment/record/distribution 恰好就是这个结构。

### 2.3 分发（下游同步）常见模式【事实】

来源：知乎《如何制定主数据治理策略》2023-12；腾讯云开发者社区《KPaaS：主数据平台下游系统过多如何下发数据》；Solace《Modernizing Enterprise MDM》2025-08；Profisee/Stibo（注册表式 Registry vs 轴辐式 Hub-and-Spoke）；Informatica MDM（出站 JMS 队列）：

1. **增量推送**（主流）：主数据平台在记录发布/变更后主动调下游接口推送，实时性高；
2. **全量拉取**：下游按需从平台拉全量对齐，一般用于初始化与月度校准；
3. **发布订阅解耦**（下游多于 3~5 个系统时）：变更事件写 Kafka/RocketMQ topic，下游自订阅，点对点网状同步不可维护（Solace 称之为从 batch/请求响应走向事件驱动 pub/sub）；
4. **注册表式 vs 集中托管式**：Registry 只聚合各源的黄金记录引用，Hub/托管式把黄金记录统一存储再分发回各系统。

### 2.4 表模型设计建议（字段清单）

对照 phase5 草稿的增补建议【推测】：

- `dc_mdm_model`：够用；建议补 `status`（draft/published）与 `publish_snapshot_json`（模型发布快照，防模型改版后历史记录无法解释）。
- `dc_mdm_segment`（关键修订）：
  - 现有 `UNIQUE KEY uk_model (model_id)` 只支持一模型一段，**无法按日期滚动重置**；建议改为唯一键 `(model_id, prefix, date_key)`，新增 `date_key` varchar(16)（当期日期串，如 20261010；date_pattern 为空则恒为 'ALL'）；
  - 新增 `buffer` 相关字段不需要——号段缓存在应用内存（rpc 进程内 double buffer），表只存 `[current_value, step]` 语义；
  - 取段 SQL 定式：`UPDATE dc_mdm_segment SET current_value=current_value+step WHERE id=?` 后 `SELECT current_value,step`，事务内完成（同 Leaf）；step 建议 ≥50（高频模型 200+）【推测，经验值】。
- `dc_mdm_record`：建议补 `source_system` varchar(64)（数据来源系统，黄金记录追溯）、`hash` char(32)（内容指纹，分发时幂等判重）、`prev_version_json`（变更前快照，轻量审计）或约定变更史进 dc_ops_audit_log。
- `dc_mdm_distribution`：建议补
  - `subscriber` varchar(64)（订阅方系统名）、`mode` varchar(16)（push api / mq / pull）；
  - `retry_count` int + `next_retry_at`（失败重试，指数退避）；
  - `payload_json`/`hash`（推送内容指纹，支持对账）；
  - 索引 `idx_status_next(status,next_retry_at)` 供后台重试扫描。

### 2.5 项目落地建议

- 服务形态【推测】：新建 `mdm/` 服务（model + rpc），proto：ModelCrud / SegmentAlloc / RecordCrud(Publish) / Distribute / DistributionQuery；分发推送复用 `datasource` rpc 写目标库（草稿已用 target_ds_id/target_db/target_table，即"写库式分发"），HTTP 推送与 MQ 订阅二期扩展。
- 码段服务做成 rpc 内独立 usecase：进程内双 buffer（预加载阈值 20% 剩余）+ 原子 UPDATE 取段；go-zero sqlx 事务即可，无需引入 Leaf 本体（Java 栈）【事实：Leaf 为 Java 项目】。
- 与数据标准联动：mdm_model 的字段 schema 可引用 dc_std_element（字段类型/长度/码表约束从标准继承），建模时校验"落标"【推测，呼应模块一】。
- 多租户：全部表带 beid；号段唯一键需含 beid（`(beid, model_id, prefix, date_key)`）——草稿 uk_model 漏了 beid，跨租户会互相顶住【事实：草稿索引定义如此；修订为推测建议】。

---

## 三、数据运维 / 系统日志（登录日志/审计日志/数据归档恢复）

### 3.1 审计日志表设计最佳实践【事实】

来源：Redgate《Database Design for Audit Logging》2021-10（red-gate.com/blog/database-design-for-audit-logging）；Martin Fowler《Audit Log》（martinfowler.com/eaaAudit.html）；SonarSource 审计日志最佳实践；Bytebase《Database Audit Logging》；OneUptime《MySQL audit schema》2026-03：

- 核心五元组 **who / when / what / where / result**：操作者、时间戳、动作+对象（含 before/after 值）、来源 IP/端/trace、成功与否。
- 三种落库形态（Redgate 归纳）：行内版本号列 / 影子表（每表一张镜像） / **通用审计表（两张表即可记全 schema）**——通用表适合管理面操作审计，影子表适合细粒度数据变更。
- 行业共识：
  - 审计日志与业务/应用日志**分离存储**（受众、留存、不可篡改要求不同；SonarSource 明确建议）；
  - **append-only**，不做 update/delete（日志是证据不是数据）；
  - 对"审计什么"做数据最小化，绝不记密钥/口令明文。

### 3.2 数据归档策略【事实】

来源：OceanBase 社区博客《为什么应该使用 ob-archiver》2023-12（open.oceanbase.com）；dbaplus 社区《归档常见疑虑解答》2022-07（dbaplus.cn/news-11-4666-1）；博客园/LearnKu《MySQL 安全批量 delete 数据行及数据归档》2021-06；Percona pt-archiver 文档与实践（腾讯云开发者社区）：

- **大事务 insert-select+delete 一次跑**的问题：长事务、大 undo、锁范围大、主从延迟尖峰、失败回滚代价高——仅适合小表/一次性迁移。
- **pt-archiver 式分批**是事实标准：按 `--limit`（每批行数）+ `--commit-each`（每批一提交）循环"先归档后删除"；先 INSERT INTO archive SELECT 再 DELETE 同一批，保证"删了但没归上"不会发生。
- 约束与经验值：
  - 源表必须有主键或至少一个索引；WHERE 条件列无索引会退化成锁全表扫描【事实，OceanBase 博客明确指出】；
  - 分批大小常用 **1000~2000 行/批**，批间 sleep（pt-archiver `--sleep` 常配 0.1~1s）【推测：社区常用经验值，非官方规范】；
  - 更稳的分批用**游标法**：`WHERE id > last_id ORDER BY id LIMIT n`（pt-archiver 内部即按主键游标推进），避免 OFFSET 深翻页；
  - GTID 实例注意：GTID 不支持 CTAS 建归档表，需预先建表再 insert-select【事实，博客园/LearnKu 指出】。
- 归档表建议与源表同构 + 归档时间列（archived_at），恢复 = 反向 insert-select 回源表【事实+推测，README"数据生命周期：归档/恢复"的常规实现】。

### 3.3 登录日志与合规要求【事实】

来源：GB/T 22239-2019《网络安全等级保护基本要求》第三级"安全审计"控制点；腾讯云开发者社区《等保2.0 合规必备：满足六个月日志留存》2026-07；《网络安全法》第 21 条（国家法律法规数据库 flk.npc.gov.cn）：

- 三级"安全审计"要求：审计覆盖到**每个用户**；审计记录包括**日期和时间、用户、事件类型、事件是否成功**及其他审计相关信息；对审计记录**保护（定期备份，防未预期的删除、修改、覆盖）**；对审计进程保护（防未授权中断）。
- **留存 ≥6 个月**的直接法律依据是《网络安全法》第 21 条第(三)项："留存相关的网络日志不少于六个月"；等保测评将"日志留存不满足法规要求"列为高风险判例（二级及以上适用）。
- 落到表设计：登录日志至少含 账号/uid、IP、UA、成功与否、时间；建议补登录方式（password/wechat/sso）与失败原因枚举，连续失败检测（防爆破）在应用层做。

### 3.4 表模型设计建议（字段清单）

对照 phase5 草稿的增补建议【推测】：

- `dc_ops_audit_log`（对齐五元组）：
  - 补 `result` varchar(16)（success/failed——等保"事件是否成功"是硬要求）、`beid`（草稿缺，违反既有约定）、`trace_id` varchar(64)（串联网关日志）、`method` varchar(16)/`path` varchar(255)（HTTP 面操作）、`app` varchar(32)（哪个服务写入）；
  - detail_json 中 before/after 分开成 `before_json`/`after_json` 更利于 diff 查询（可选）；
  - 索引：保留 `idx_uid`、`idx_res`，加 `idx_beid_created(beid, created_at)`。
- `dc_ops_login_log`：补 `beid`、`login_type` varchar(32)、`fail_reason` varchar(64)、`logout_at`（可选）；索引已有 idx_user/idx_created 够用。
- `dc_ops_archive_task`：建议补
  - `cursor_col` varchar(64)（游标列，默认主键 id）、`last_cursor` bigint（断点续跑）；
  - `sleep_ms` int DEFAULT 200（批间隔节流）；
  - `max_rows_per_run` bigint（单次运行上限，防归档窗口过长）；
  - condition_sql 的**白名单校验**（只允许 WHERE 片段、禁 DML 关键字、EXPLAIN 验证走索引）——这是自研归档工具最大的安全点【事实：业界共识；实现方案为推测】。
- `dc_ops_archive_log`：补 `batch_count` int、`beid`、`duration_ms`；建议按 `task_id+run_at` 已有索引够用。
- 留存策略：登录/审计日志保留 ≥183 天（合规下限），过期数据用 dc_ops_archive_task 自归档（archive_table=历史归档库），即"日志模块自己吃自己的归档能力"【推测】。

### 3.5 项目落地建议

- 服务形态【推测】：新建 `ops/` 服务（model + rpc），proto：AuditLogQuery / LoginLogQuery / ArchiveTaskCrud / ArchiveRun / RestoreRun；审计写入用 **go-zero 中间件**（http 网关 + rpc 拦截）旁路异步写（channel + 单写协程批量 insert），避免阻塞主链路；登录日志在 user 服务登录逻辑内嵌点埋，通过 rpc 推给 ops 或直接写库（一期写库简单）。
- 归档执行器复用 phase3 的 job/sync_task 执行面模式：管控面存 dc_ops_archive_task，执行面在 job 服务内跑"游标分批 insert-select + delete"，每批一事务，失败记录 archive_log 并可从 last_cursor 续跑。
- 等保对照：登录/审计日志天然满足"记录日期时间/用户/事件类型/是否成功"五要素（草稿补 result/beid 后完全对齐）；"定期备份防篡改"可由归档任务 + 归档库只读账号实现，一期不做异地备份【推测】。
- 与模块一/二联动：标准核对报告、主数据分发记录本身也是审计对象，统一走 dc_ops_audit_log 记录发布/分发动作【推测】。

---

## 四、三模块对标速查表

| 维度 | 数据标准 | 主数据 MDM | 数据运维/日志 |
|---|---|---|---|
| 行业蓝本 | Dataphin 词根→数据元→标准集→落标映射→落标监控 | Leaf 号段 + Pimcore/qData 四件套模型 | Redgate 通用审计表 + pt-archiver 分批 |
| 国标/规范 | GB/T 18391 (ISO/IEC 11179) 数据元属性集 | 参考数据=码表（与标准字典共用） | 网安法 21 条(≥6个月) + GB/T 22239-2019 三级审计 |
| 开源可抄 | Datavines 类型系统（definition+instance+JSON） | Leaf leaf_alloc 表结构（已验证原文） | pt-archiver 游标分批参数 |
| 关键并发点 | 标准发布版本冻结 | 号段原子 UPDATE + 双 buffer | 归档每批一事务、防锁表 |
| 对 phase5 草稿最大修订 | 增 model_version 历史表 + std_mapping 落标映射表 | segment 唯一键改 (beid,model,prefix,date_key) | audit/login_log 补 beid+result；task 补游标+节流 |

## 主要参考来源

- Dataphin 数据标准/落标：https://www.aliyun.com/product/dataphin ；help.aliyun.com《数据标准概述》《新建和管理落标映射规则》；developer.aliyun.com《数据标准落标监控》2023
- DataWorks 数据标准：alibabacloud.com 帮助文档"数据建模-数据标准"
- GB/T 18391 系列：openstd.samr.gov.cn（国家标准全文公开系统）；cnis.ac.cn
- Apache Atlas Type System：https://atlas.apache.org
- Datavines（schema 已核验）：https://github.com/datavane/datavines （dev/scripts/sql/datavines-mysql.sql）
- 美团 Leaf：https://tech.meituan.com/2017/04/21/mt-leaf.html ；https://github.com/Meituan-Dianping/Leaf （scripts/leaf_alloc.sql 已核验）
- MDM 概念与产品：https://www.precisely.com ；https://www.stibosystems.com ；https://www.cloverdx.com ；https://pimcore.com ；qData：community.qdata.tech
- 主数据分发：zhuanlan.zhihu.com《如何制定主数据治理策略》；cloud.tencent.com《KPaaS 主数据下发》；https://solace.com 《Modernizing Enterprise MDM》
- 审计日志：https://www.red-gate.com/blog/database-design-for-audit-logging ；https://martinfowler.com/eaaAudit.html ；sonarsource.com；bytebase.com/blog/database-audit-logging
- 数据归档：open.oceanbase.com 博客 2023-12；https://www.dbaplus.cn/news-11-4666-1 ；腾讯云开发者社区 pt-archiver 实践
- 等保/留痕：GB/T 22239-2019；flk.npc.gov.cn（网络安全法第 21 条）；developer.cloud.tencent.com《等保2.0 六个月日志留存》2026
- 注：部分中文站点 URL 为搜索结果摘录的域名级链接，站内按文档标题可精确定位【事实+推测混合，已尽量核验到可复现层级】。
