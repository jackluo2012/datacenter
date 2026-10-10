-- ============================================================
-- Phase 6 数据中台表（数仓建模 / 数据安全 / 标签圈群）
-- ============================================================

-- 主题域
CREATE TABLE IF NOT EXISTS `dc_theme_domain` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL COMMENT '域编码',
  `name` varchar(128) NOT NULL COMMENT '域名',
  `comment` varchar(512) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数仓主题域表';

-- 主题
CREATE TABLE IF NOT EXISTS `dc_theme` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `domain_id` bigint NOT NULL,
  `code` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  `comment` varchar(512) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数仓主题表';

-- 数仓模型（主题下分层表）
CREATE TABLE IF NOT EXISTS `dc_theme_model` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `theme_id` bigint NOT NULL,
  `layer` varchar(8) NOT NULL COMMENT 'ods/dwd/dws/ads/dim',
  `table_name` varchar(128) NOT NULL,
  `model_name` varchar(128) NOT NULL,
  `fields_json` text COMMENT '[{name,type,comment,pk,dim}]',
  `dim_json` text COMMENT '维度声明[{dim_table,join_key}]',
  `datasource_id` bigint NOT NULL DEFAULT 0,
  `db_name` varchar(128) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_table` (`beid`,`table_name`),
  KEY `idx_theme_layer` (`theme_id`,`layer`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数仓分层模型表';

-- 物化视图
CREATE TABLE IF NOT EXISTS `dc_theme_mv` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `datasource_id` bigint NOT NULL,
  `db_name` varchar(128) NOT NULL DEFAULT '',
  `view_name` varchar(128) NOT NULL,
  `select_sql` mediumtext NOT NULL,
  `refresh_mode` varchar(16) NOT NULL DEFAULT 'manual' COMMENT 'manual/schedule',
  `last_status` varchar(16) NOT NULL DEFAULT '',
  `last_refresh_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid` (`beid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物化视图表';

-- 安全分级目录
CREATE TABLE IF NOT EXISTS `dc_sec_level` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(8) NOT NULL COMMENT 'L1-L4',
  `name` varchar(64) NOT NULL,
  `comment` varchar(512) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='安全分级目录表';

-- 实体分级打标表
CREATE TABLE IF NOT EXISTS `dc_sec_entity_tag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `entity_id` bigint NOT NULL COMMENT '元数据实体id',
  `entity_name` varchar(255) NOT NULL DEFAULT '',
  `level_code` varchar(8) NOT NULL,
  `mask_rule_code` varchar(64) NOT NULL DEFAULT '' COMMENT '绑定脱敏规则',
  `beid` bigint NOT NULL DEFAULT 0,
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_entity` (`beid`,`entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实体分级打标表';

-- 脱敏规则库
CREATE TABLE IF NOT EXISTS `dc_sec_mask_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `rule_code` varchar(64) NOT NULL,
  `rule_name` varchar(128) NOT NULL,
  `algorithm` varchar(32) NOT NULL COMMENT 'mask/hash/replace/null',
  `params_json` text,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_code` (`beid`,`rule_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='脱敏规则库表';

-- 授权矩阵表
CREATE TABLE IF NOT EXISTS `dc_sec_grant` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `subject` varchar(128) NOT NULL COMMENT 'user:1/app:key/role:admin',
  `resource` varchar(64) NOT NULL,
  `resource_id` varchar(64) NOT NULL DEFAULT '',
  `level_code` varchar(8) NOT NULL COMMENT '许可最高密级',
  `expire_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_sub_res` (`subject`,`resource`,`resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='安全授权矩阵表';

-- 标签人群表
CREATE TABLE IF NOT EXISTS `dc_tag_group` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `object_id` bigint NOT NULL,
  `circle_rule_sql` text COMMENT '圈选 SQL: select <pk> as object_id, <v> as tag_value from ...',
  `est_count` bigint NOT NULL DEFAULT 0,
  `materialized_table` varchar(128) NOT NULL DEFAULT '' COMMENT '人群物化表',
  `last_run_at` datetime(3) DEFAULT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid` (`beid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签人群表';

-- 人群成员物化表（每人群一张，自动创建）
-- 结构: object_pk varchar(128), tag_value varchar(255), 由 RunCircle 自动建
