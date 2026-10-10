-- ============================================================
-- Phase 2 数据中台表（数据质量 / 数据指标 / 数据标签）
-- 目标库: datacenter  字符集: utf8mb4  引擎: InnoDB
-- 约定: 所有表带 beid 多租户字段
-- ============================================================

-- 质量规则表
CREATE TABLE IF NOT EXISTS `dc_quality_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '规则名',
  `rule_type` varchar(32) NOT NULL COMMENT 'row_count/not_null/unique/value_range/consistency/fluctuation',
  `target_table` varchar(128) NOT NULL COMMENT '目标表',
  `target_column` varchar(128) NOT NULL DEFAULT '' COMMENT '目标列(not_null/unique/value_range 用)',
  `params_json` text COMMENT '规则参数: value_range={min,max}; consistency={ref_table,ref_column}; fluctuation={baseline,ratio_pct}',
  `threshold_json` text COMMENT '期望阈值: {op:gt|lt|eq|between, value:..}',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1.启用 0.停用',
  `owner` varchar(64) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid_status` (`beid`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='质量规则表';

-- 质量任务表（规则组合 + 数据源）
CREATE TABLE IF NOT EXISTS `dc_quality_job` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `datasource_id` bigint NOT NULL COMMENT '目标数据源',
  `db_name` varchar(128) NOT NULL DEFAULT '' COMMENT '目标库',
  `rule_ids` varchar(1024) NOT NULL COMMENT '规则id,逗号分隔',
  `cron` varchar(64) NOT NULL DEFAULT '' COMMENT 'cron 表达式(Phase 3 调度接入,本期仅存档)',
  `last_status` varchar(16) NOT NULL DEFAULT '' COMMENT 'success/failed/partial',
  `last_run_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid` (`beid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='质量任务表';

-- 质量结果表
CREATE TABLE IF NOT EXISTS `dc_quality_result` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `job_id` bigint NOT NULL,
  `rule_id` bigint NOT NULL,
  `rule_type` varchar(32) NOT NULL,
  `check_value` varchar(255) NOT NULL DEFAULT '' COMMENT '检查实测值',
  `pass_flag` tinyint NOT NULL COMMENT '1.通过 0.不通过',
  `detail_json` text COMMENT '详情(阈值/实测/差异)',
  `run_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_job_run` (`job_id`,`run_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='质量检查结果表';

-- 指标表（原子指标 + 派生指标）
CREATE TABLE IF NOT EXISTS `dc_metric` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name_en` varchar(128) NOT NULL COMMENT '英文名/标识',
  `name_cn` varchar(128) NOT NULL COMMENT '中文名',
  `metric_type` varchar(16) NOT NULL COMMENT 'atom/derive',
  `biz_field` varchar(128) NOT NULL DEFAULT '' COMMENT '原子指标业务口径字段(如 pay_amount)',
  `agg_type` varchar(16) NOT NULL DEFAULT '' COMMENT '聚合: sum/count/avg/max/min/distinct',
  `expr` varchar(1024) NOT NULL DEFAULT '' COMMENT '派生表达式: 原子指标+修饰+周期(如 sum(pay_amount) where status=1 by day)',
  `dimension_json` text COMMENT '维度[{name,comment}]',
  `comment` varchar(1024) NOT NULL DEFAULT '' COMMENT '口径说明',
  `owner` varchar(64) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_name` (`beid`,`name_en`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据指标表';

-- 指标关系表（派生指标 → 原子指标引用）
CREATE TABLE IF NOT EXISTS `dc_metric_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `derive_metric_id` bigint NOT NULL COMMENT '派生指标id',
  `atom_metric_id` bigint NOT NULL COMMENT '被引用原子指标id',
  `beid` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_rel` (`derive_metric_id`,`atom_metric_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='指标关系表';

-- 标签对象表
CREATE TABLE IF NOT EXISTS `dc_tag_object` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `object_type` varchar(32) NOT NULL COMMENT 'user/member/goods',
  `name` varchar(128) NOT NULL COMMENT '对象名',
  `ds_id` bigint NOT NULL DEFAULT 0 COMMENT '对象主数据所在数据源',
  `db_name` varchar(128) NOT NULL DEFAULT '' COMMENT '对象库',
  `table_name` varchar(128) NOT NULL DEFAULT '' COMMENT '对象主表',
  `pk_column` varchar(64) NOT NULL DEFAULT 'id' COMMENT '主键列',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid_type` (`beid`,`object_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签对象表';

-- 标签定义表
CREATE TABLE IF NOT EXISTS `dc_tag_def` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `object_id` bigint NOT NULL COMMENT '所属对象',
  `tag_key` varchar(64) NOT NULL COMMENT '标签键(如 vip_level)',
  `tag_name` varchar(128) NOT NULL COMMENT '标签名',
  `tag_type` varchar(16) NOT NULL DEFAULT 'rule' COMMENT 'rule规则/fact事实/model模型',
  `rule_sql` text COMMENT '规则SQL: select <pk> as object_id, <expr> as tag_value from ... where ...',
  `refresh_cron` varchar(64) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_obj_key` (`object_id`,`tag_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签定义表';

-- 标签值表（打标结果）
CREATE TABLE IF NOT EXISTS `dc_tag_value` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tag_id` bigint NOT NULL COMMENT '标签定义id',
  `object_pk` varchar(128) NOT NULL COMMENT '对象主键值',
  `tag_value` varchar(255) NOT NULL DEFAULT '' COMMENT '标签值',
  `beid` bigint NOT NULL DEFAULT 0,
  `tagged_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '打标时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_tag_pk` (`tag_id`,`object_pk`),
  KEY `idx_pk` (`object_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签值表';
