-- ============================================================
-- Phase 1 数据中台基础表（数据源/元数据/数据服务/资产目录）
-- 目标库: datacenter  字符集: utf8mb4  引擎: InnoDB
-- 约定: 所有表带 beid 多租户字段; created_at/updated_at 由服务端维护
-- ============================================================

-- 数据源登记表
CREATE TABLE IF NOT EXISTS `dc_datasource` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '数据源id',
  `name` varchar(128) NOT NULL COMMENT '名称',
  `type` varchar(32) NOT NULL COMMENT '类型: mysql/redis/es/http/kafka',
  `host` varchar(255) NOT NULL DEFAULT '' COMMENT '主机',
  `port` int NOT NULL DEFAULT 0 COMMENT '端口',
  `db_name` varchar(128) NOT NULL DEFAULT '' COMMENT '库名',
  `username` varchar(128) NOT NULL DEFAULT '' COMMENT '用户名',
  `password_enc` varbinary(512) NOT NULL COMMENT 'AES-GCM 加密后的密码',
  `params_json` text COMMENT '扩展参数(json)',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1.启用 0.停用',
  `owner` varchar(64) NOT NULL DEFAULT '' COMMENT '负责人',
  `beid` bigint NOT NULL DEFAULT 0 COMMENT '平台id',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid_type` (`beid`,`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据源登记表';

-- 元数据实体表（database/table/column）
CREATE TABLE IF NOT EXISTS `dc_meta_entity` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '实体id',
  `entity_type` varchar(32) NOT NULL COMMENT 'database/table/column',
  `fqn` varchar(512) NOT NULL COMMENT '全限定名,如 mysql://db.table.column',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `comment` varchar(1024) NOT NULL DEFAULT '' COMMENT '注释',
  `engine` varchar(64) NOT NULL DEFAULT '' COMMENT '存储引擎',
  `rows_cnt` bigint NOT NULL DEFAULT 0 COMMENT '行数',
  `props_json` text COMMENT '扩展属性(json)',
  `classification` varchar(8) NOT NULL DEFAULT 'L2' COMMENT '分级:L1公开/L2内部/L3敏感/L4机密',
  `owner` varchar(64) NOT NULL DEFAULT '' COMMENT '负责人',
  `ds_id` bigint NOT NULL DEFAULT 0 COMMENT '来源数据源id',
  `beid` bigint NOT NULL DEFAULT 0 COMMENT '平台id',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_fqn` (`beid`,`fqn`),
  KEY `idx_beid_type` (`beid`,`entity_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='元数据实体表';

-- 元数据关系表（contains/lineage/depends）
CREATE TABLE IF NOT EXISTS `dc_meta_relation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `rel_type` varchar(32) NOT NULL COMMENT 'contains/lineage/depends',
  `from_id` bigint NOT NULL COMMENT '起点实体id',
  `to_id` bigint NOT NULL COMMENT '终点实体id',
  `props_json` text COMMENT '关系属性(json,如转换SQL摘要)',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_edge` (`rel_type`,`from_id`,`to_id`),
  KEY `idx_to_id` (`to_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='元数据关系/血缘表';

-- 元数据版本表（定版快照）
CREATE TABLE IF NOT EXISTS `dc_meta_version` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `entity_id` bigint NOT NULL COMMENT '实体id',
  `version` int NOT NULL COMMENT '版本号,实体内自增',
  `snapshot_json` mediumtext COMMENT '实体快照(json)',
  `operator` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人',
  `published_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_entity_version` (`entity_id`,`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='元数据版本表';

-- 数据服务 API 表
CREATE TABLE IF NOT EXISTS `dc_api` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '名称',
  `path` varchar(128) NOT NULL COMMENT '执行路径标识',
  `mode` varchar(16) NOT NULL DEFAULT 'sql' COMMENT 'wizard/sql/register',
  `datasource_id` bigint NOT NULL COMMENT '绑定的数据源id',
  `sql_text` text COMMENT 'SQL模板,参数用 :name 占位',
  `params_json` text COMMENT '参数定义[{"name":"","type":"","required":true,"default":""}]',
  `cache_ttl` int NOT NULL DEFAULT 0 COMMENT '结果缓存秒数,0不缓存',
  `status` varchar(16) NOT NULL DEFAULT 'draft' COMMENT 'draft/published/offline',
  `resp_sample_json` text COMMENT '响应样例',
  `owner` varchar(64) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_path` (`beid`,`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据服务API表';

-- 数据服务应用表
CREATE TABLE IF NOT EXISTS `dc_api_app` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_key` varchar(64) NOT NULL COMMENT '应用标识',
  `app_secret_enc` varbinary(512) NOT NULL COMMENT 'AES-GCM 加密的密钥',
  `name` varchar(128) NOT NULL,
  `rate_limit_qps` int NOT NULL DEFAULT 10 COMMENT '限流QPS',
  `ip_whitelist` varchar(512) NOT NULL DEFAULT '' COMMENT '白名单,逗号分隔,空为不限',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_app_key` (`app_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据服务应用表';

-- API 授权表
CREATE TABLE IF NOT EXISTS `dc_api_grant` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `api_id` bigint NOT NULL,
  `app_id` bigint NOT NULL,
  `expire_at` datetime(3) DEFAULT NULL COMMENT '过期时间,NULL永久',
  `field_mask` varchar(512) NOT NULL DEFAULT '' COMMENT '脱敏字段,逗号分隔',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_api_app` (`api_id`,`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API授权表';

-- 资产目录分类表
CREATE TABLE IF NOT EXISTS `dc_asset_category` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint NOT NULL DEFAULT 0 COMMENT '父分类,0为根',
  `name` varchar(128) NOT NULL,
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid_parent` (`beid`,`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资产目录分类表';

-- 资产表（元数据实体入目录,冗余快照避免跨服务联查）
CREATE TABLE IF NOT EXISTS `dc_asset` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `entity_id` bigint NOT NULL COMMENT '元数据实体id',
  `category_id` bigint NOT NULL DEFAULT 0,
  `name` varchar(255) NOT NULL COMMENT '资产名(冗余快照)',
  `entity_type` varchar(32) NOT NULL DEFAULT '' COMMENT '冗余快照',
  `comment` varchar(1024) NOT NULL DEFAULT '' COMMENT '冗余快照',
  `classification` varchar(8) NOT NULL DEFAULT 'L2' COMMENT '冗余快照',
  `asset_tags_json` text COMMENT '资产标签(json)',
  `visibility` varchar(16) NOT NULL DEFAULT 'private' COMMENT 'public/private',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_entity` (`beid`,`entity_id`),
  KEY `idx_category` (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资产表';

-- 资产申请使用表
CREATE TABLE IF NOT EXISTS `dc_asset_apply` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `asset_id` bigint NOT NULL,
  `reason` varchar(1024) NOT NULL DEFAULT '' COMMENT '申请理由',
  `applicant` varchar(64) NOT NULL DEFAULT '' COMMENT '申请人',
  `status` varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/approved/rejected',
  `reviewer` varchar(64) NOT NULL DEFAULT '' COMMENT '审批人',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_beid_status` (`beid`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资产申请使用表';
