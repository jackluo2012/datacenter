-- ============================================================
-- Phase 7 数据中台表（落地深度：告警 / 错误数据 / 健康分 / API 可观测 / 调度）
-- ============================================================

-- 告警通道配置
CREATE TABLE IF NOT EXISTS `dc_quality_alert_channel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `type` varchar(16) NOT NULL COMMENT 'dingtalk/feishu/email(预留)',
  `webhook_url` varchar(512) NOT NULL COMMENT '机器人 webhook',
  `secret` varchar(128) NOT NULL DEFAULT '' COMMENT '加签密钥(可选)',
  `min_level` varchar(16) NOT NULL DEFAULT 'failed' COMMENT '触发级别: failed/partial',
  `status` tinyint NOT NULL DEFAULT 1,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid` (`beid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='质量告警通道表';

-- 告警发送日志
CREATE TABLE IF NOT EXISTS `dc_quality_alert_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `channel_id` bigint NOT NULL,
  `job_id` bigint NOT NULL,
  `content` varchar(1024) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL COMMENT 'success/failed',
  `msg` varchar(512) NOT NULL DEFAULT '',
  `sent_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_job` (`job_id`,`sent_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警发送日志表';

-- 质量错误数据样本（下钻用）
CREATE TABLE IF NOT EXISTS `dc_quality_error_data` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `rule_id` bigint NOT NULL,
  `job_id` bigint NOT NULL,
  `rows_json` mediumtext COMMENT '违规行样本(limit 100)',
  `sample_count` int NOT NULL DEFAULT 0,
  `run_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_rule_job` (`rule_id`,`job_id`,`run_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='质量错误数据样本表';

-- 表剖析与健康分
CREATE TABLE IF NOT EXISTS `dc_quality_profile` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `datasource_id` bigint NOT NULL,
  `db_name` varchar(128) NOT NULL DEFAULT '',
  `table_name` varchar(128) NOT NULL,
  `rows_cnt` bigint NOT NULL DEFAULT 0,
  `cols_cnt` int NOT NULL DEFAULT 0,
  `null_cells` bigint NOT NULL DEFAULT 0 COMMENT '空值单元格数',
  `total_cells` bigint NOT NULL DEFAULT 0,
  `null_rate` decimal(6,4) NOT NULL DEFAULT 0 COMMENT '空值率',
  `dup_pk_cnt` bigint NOT NULL DEFAULT 0 COMMENT '主键重复数(有主键时)',
  `health_score` int NOT NULL DEFAULT 0 COMMENT '健康分 0-100',
  `profile_json` mediumtext COMMENT '逐列剖析明细',
  `ran_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `beid` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_tbl` (`datasource_id`,`table_name`,`ran_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='表剖析与健康分表';

-- 数据服务调用日志
CREATE TABLE IF NOT EXISTS `dc_api_call_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `api_id` bigint NOT NULL,
  `app_id` bigint NOT NULL DEFAULT 0,
  `path` varchar(128) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL DEFAULT 'ok' COMMENT 'ok/error/denied',
  `error_code` varchar(64) NOT NULL DEFAULT '',
  `rows_cnt` bigint NOT NULL DEFAULT 0,
  `cost_ms` int NOT NULL DEFAULT 0,
  `cache_hit` tinyint NOT NULL DEFAULT 0,
  `ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_api_time` (`api_id`,`created_at`),
  KEY `idx_app` (`app_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据服务调用日志表';
