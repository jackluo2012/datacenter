-- ============================================================
-- Phase 5 数据中台表（数据标准 / 主数据 / 数据运维与日志）
-- 目标库: datacenter  字符集: utf8mb4  引擎: InnoDB
-- ============================================================

-- ============ 数据标准管理 ============

-- 标准词根表
CREATE TABLE IF NOT EXISTS `dc_std_word_root` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL COMMENT '词根(如 amount)',
  `cn_name` varchar(128) NOT NULL DEFAULT '' COMMENT '中文含义(如 金额)',
  `type` varchar(16) NOT NULL DEFAULT 'common' COMMENT 'common/attribute/dimension/indicator',
  `comment` varchar(512) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_name` (`beid`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标准词根表';

-- 标准字典表
CREATE TABLE IF NOT EXISTS `dc_std_dict` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `dict_code` varchar(64) NOT NULL COMMENT '字典编码(如 gender)',
  `dict_name` varchar(128) NOT NULL COMMENT '字典名(如 性别)',
  `comment` varchar(512) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1.启用 0.停用',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`dict_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标准字典表';

-- 标准字典项表
CREATE TABLE IF NOT EXISTS `dc_std_dict_item` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `dict_id` bigint NOT NULL,
  `item_code` varchar(64) NOT NULL COMMENT '项编码(如 M/F/U)',
  `item_value` varchar(255) NOT NULL COMMENT '项含义(如 男/女/未知)',
  `sort` int NOT NULL DEFAULT 0,
  `beid` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_dict_code` (`dict_id`,`item_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标准字典项表';

-- 数据元表
CREATE TABLE IF NOT EXISTS `dc_std_element` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `element_code` varchar(64) NOT NULL COMMENT '数据元编码(如 user_name)',
  `element_name` varchar(128) NOT NULL COMMENT '数据元名(如 用户姓名)',
  `data_type` varchar(32) NOT NULL COMMENT 'varchar/int/bigint/decimal/date/datetime/text',
  `data_length` int NOT NULL DEFAULT 0,
  `data_precision` int NOT NULL DEFAULT 0 COMMENT '小数位( decimal 用)',
  `dict_code` varchar(64) NOT NULL DEFAULT '' COMMENT '值域引用字典编码,空为无',
  `comment` varchar(512) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`element_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据元表';

-- 标准模型表（数据元组合成的业务标准）
CREATE TABLE IF NOT EXISTS `dc_std_model` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `model_code` varchar(64) NOT NULL COMMENT '标准模型编码',
  `model_name` varchar(128) NOT NULL COMMENT '标准模型名(如 用户标准)',
  `elements_json` text COMMENT '组成元素[{element_code, required, pk}]',
  `status` varchar(16) NOT NULL DEFAULT 'draft' COMMENT 'draft/published/deprecated',
  `version` int NOT NULL DEFAULT 1 COMMENT '发布版本号,发布+1',
  `snapshot_json` mediumtext COMMENT '最近发布快照',
  `published_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`model_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标准模型表';

-- 标准核对报告表
CREATE TABLE IF NOT EXISTS `dc_std_check_report` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `model_id` bigint NOT NULL,
  `datasource_id` bigint NOT NULL,
  `db_name` varchar(128) NOT NULL DEFAULT '',
  `table_name` varchar(128) NOT NULL,
  `diff_json` mediumtext COMMENT '差异[{column,expect_type,actual_type,result}]',
  `pass_flag` tinyint NOT NULL COMMENT '1.符合 0.存在偏差',
  `run_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `beid` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_model` (`model_id`,`run_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标准核对报告表';

-- ============ 主数据管理 ============

-- 主数据模型表
CREATE TABLE IF NOT EXISTS `dc_mdm_model` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `model_code` varchar(64) NOT NULL COMMENT '模型编码(如 customer)',
  `model_name` varchar(128) NOT NULL COMMENT '模型名(如 客户主数据)',
  `table_name` varchar(128) NOT NULL COMMENT '物理表名(如 dc_mdm_data_customer)',
  `fields_json` text COMMENT '字段定义[{"name":"","type":"","length":0,"required":true,"comment":""}]',
  `code_strategy` varchar(16) NOT NULL DEFAULT 'segment' COMMENT '编码策略: segment码段/auto自增',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`model_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主数据模型表';

-- 编码码段表（号段模式）
CREATE TABLE IF NOT EXISTS `dc_mdm_segment` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `model_id` bigint NOT NULL COMMENT '所属主数据模型',
  `prefix` varchar(32) NOT NULL DEFAULT '' COMMENT '编码前缀(如 CUS)',
  `date_pattern` varchar(32) NOT NULL DEFAULT '' COMMENT '日期段(如 20060102,空为无)',
  `current_value` bigint NOT NULL DEFAULT 0 COMMENT '当前流水值',
  `step` int NOT NULL DEFAULT 1 COMMENT '步长',
  `beid` bigint NOT NULL DEFAULT 0,
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_model` (`model_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='编码码段表';

-- 主数据记录表
CREATE TABLE IF NOT EXISTS `dc_mdm_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `model_id` bigint NOT NULL,
  `biz_code` varchar(128) NOT NULL COMMENT '业务编码(码段生成,模型内唯一)',
  `data_json` json NOT NULL COMMENT '记录内容',
  `status` varchar(16) NOT NULL DEFAULT 'draft' COMMENT 'draft/published/archived',
  `version` int NOT NULL DEFAULT 1,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_model_code` (`model_id`,`biz_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主数据记录表';

-- 主数据分发记录表
CREATE TABLE IF NOT EXISTS `dc_mdm_distribution` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `record_id` bigint NOT NULL,
  `target_ds_id` bigint NOT NULL COMMENT '目标数据源',
  `target_db` varchar(128) NOT NULL DEFAULT '',
  `target_table` varchar(128) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/success/failed',
  `msg` varchar(1024) NOT NULL DEFAULT '',
  `distributed_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `beid` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_record` (`record_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主数据分发记录表';

-- ============ 数据运维与系统日志 ============

-- 归档任务表
CREATE TABLE IF NOT EXISTS `dc_ops_archive_task` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `datasource_id` bigint NOT NULL COMMENT '源数据源',
  `db_name` varchar(128) NOT NULL DEFAULT '',
  `table_name` varchar(128) NOT NULL COMMENT '源表',
  `condition_sql` varchar(1024) NOT NULL DEFAULT '' COMMENT '归档条件(where 片段,白名单校验)',
  `archive_table` varchar(128) NOT NULL COMMENT '归档表',
  `delete_source` tinyint NOT NULL DEFAULT 1 COMMENT '1.归档后删源 0.仅插入',
  `batch_size` int NOT NULL DEFAULT 1000 COMMENT '分批大小',
  `cron` varchar(64) NOT NULL DEFAULT '' COMMENT '(Phase 3 调度接入,本期手动)',
  `last_status` varchar(16) NOT NULL DEFAULT '',
  `last_run_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid` (`beid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='归档任务表';

-- 归档执行日志表
CREATE TABLE IF NOT EXISTS `dc_ops_archive_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint NOT NULL,
  `archived_rows` bigint NOT NULL DEFAULT 0,
  `deleted_rows` bigint NOT NULL DEFAULT 0,
  `status` varchar(16) NOT NULL COMMENT 'success/failed',
  `msg` varchar(1024) NOT NULL DEFAULT '',
  `run_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_task` (`task_id`,`run_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='归档执行日志表';

-- 登录日志表
CREATE TABLE IF NOT EXISTS `dc_ops_login_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(128) NOT NULL DEFAULT '' COMMENT '登录账号/mobile',
  `uid` bigint NOT NULL DEFAULT 0,
  `ip` varchar(64) NOT NULL DEFAULT '',
  `user_agent` varchar(512) NOT NULL DEFAULT '',
  `success` tinyint NOT NULL DEFAULT 0,
  `msg` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_created` (`created_at`),
  KEY `idx_user` (`username`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';

-- 通用审计日志表
CREATE TABLE IF NOT EXISTS `dc_ops_audit_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL DEFAULT 0,
  `username` varchar(128) NOT NULL DEFAULT '',
  `action` varchar(64) NOT NULL COMMENT 'create/update/delete/publish/login/export...',
  `resource` varchar(128) NOT NULL COMMENT '资源类型:datasource/api/rule...',
  `resource_id` varchar(64) NOT NULL DEFAULT '',
  `detail_json` text,
  `ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_res` (`resource`,`created_at`),
  KEY `idx_uid` (`uid`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通用审计日志表';
