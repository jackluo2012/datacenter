-- ============================================================
-- Phase 3/4 数据中台表（调度集成 / SQL IDE / AI 问数审计）
-- 目标库: datacenter  字符集: utf8mb4  引擎: InnoDB
-- ============================================================

-- 同步任务表（管控面，执行面交 DolphinScheduler/SeaTunnel）
CREATE TABLE IF NOT EXISTS `dc_sync_task` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '任务名',
  `source_ds_id` bigint NOT NULL COMMENT '源数据源id',
  `target_ds_id` bigint NOT NULL DEFAULT 0 COMMENT '目标数据源id(0=目标为平台自有库)',
  `mode` varchar(16) NOT NULL DEFAULT 'offline' COMMENT 'offline/cdc',
  `mapping_json` text COMMENT '字段/表映射(JSON)',
  `source_sql` mediumtext COMMENT '源端查询(SELECT,内置引擎)',
  `target_db` varchar(128) NOT NULL DEFAULT '' COMMENT '目标库',
  `batch_size` int NOT NULL DEFAULT 1000 COMMENT '分批大小',
  `target_table` varchar(128) NOT NULL DEFAULT '' COMMENT '目标表(内置引擎写目标库此表)',
  `cron` varchar(64) NOT NULL DEFAULT '' COMMENT '调度周期(推送 DS 时使用)',
  `ds_workflow_code` int NOT NULL DEFAULT 0 COMMENT 'DolphinScheduler 工作流编码(部署后回填)',
  `status` varchar(16) NOT NULL DEFAULT 'draft' COMMENT 'draft/deployed/running/failed',
  `last_run_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid_status` (`beid`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='同步任务表';

-- SQL 脚本表（IDE）
CREATE TABLE IF NOT EXISTS `dc_ide_script` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '脚本名',
  `sql_text` mediumtext NOT NULL COMMENT 'SQL 内容',
  `version` int NOT NULL DEFAULT 1 COMMENT '版本号,每次保存+1',
  `datasource_id` bigint NOT NULL DEFAULT 0 COMMENT '执行目标数据源',
  `db_name` varchar(128) NOT NULL DEFAULT '' COMMENT '目标库',
  `last_run_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid` (`beid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SQL 脚本表';

-- AI 问数审计表
CREATE TABLE IF NOT EXISTS `dc_ask_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `question` varchar(1024) NOT NULL COMMENT '用户问题',
  `candidates_json` text COMMENT '语义检索命中的候选(json)',
  `generated_sql` text COMMENT 'LLM 生成的 SQL(未配置 LLM 时为空)',
  `answer` text COMMENT '回复内容',
  `llm_enabled` tinyint NOT NULL DEFAULT 0 COMMENT '本次是否启用了 LLM',
  `cost_ms` int NOT NULL DEFAULT 0,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid` (`beid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 问数审计表';
