-- ============================================================
-- RBAC 权限体系（对标 RuoYi 模型：用户/角色/菜单/按钮权限）
-- ============================================================

-- 系统用户（管理面账号，与 C 端 base_member 分离）
CREATE TABLE IF NOT EXISTS `sys_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL COMMENT '登录名',
  `password` varchar(128) NOT NULL COMMENT 'bcrypt',
  `nickname` varchar(64) NOT NULL DEFAULT '',
  `mobile` varchar(32) NOT NULL DEFAULT '',
  `email` varchar(128) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1.启用 0.停用',
  `remark` varchar(255) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统用户表';

-- 角色
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_key` varchar(64) NOT NULL COMMENT '权限字符(admin/operator/viewer)',
  `role_name` varchar(64) NOT NULL,
  `status` tinyint NOT NULL DEFAULT 1,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `beid` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_beid_key` (`beid`,`role_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 菜单（目录/菜单/按钮三级）
CREATE TABLE IF NOT EXISTS `sys_menu` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint NOT NULL DEFAULT 0,
  `name` varchar(64) NOT NULL COMMENT '菜单名',
  `path` varchar(128) NOT NULL DEFAULT '' COMMENT '路由路径',
  `component` varchar(128) NOT NULL DEFAULT '' COMMENT '组件路径',
  `sort` int NOT NULL DEFAULT 0,
  `icon` varchar(64) NOT NULL DEFAULT '',
  `menu_type` tinyint NOT NULL DEFAULT 2 COMMENT '1.目录 2.菜单 3.按钮',
  `perm` varchar(128) NOT NULL DEFAULT '' COMMENT '权限标识(模块:动作)',
  `visible` tinyint NOT NULL DEFAULT 1,
  `beid` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_parent` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单权限表';

-- 用户-角色
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `user_id` bigint NOT NULL,
  `role_id` bigint NOT NULL,
  PRIMARY KEY (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 角色-菜单
CREATE TABLE IF NOT EXISTS `sys_role_menu` (
  `role_id` bigint NOT NULL,
  `menu_id` bigint NOT NULL,
  PRIMARY KEY (`role_id`,`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';
