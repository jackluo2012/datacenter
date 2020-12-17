/*
 Navicat MySQL Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 80021
 Source Host           : localhost:3306
 Source Schema         : datacenter

 Target Server Type    : MySQL
 Target Server Version : 80021
 File Encoding         : 65001

 Date: 17/12/2020 11:29:26
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for app_config
-- ----------------------------
DROP TABLE IF EXISTS `app_config`;
CREATE TABLE `app_config` (
  `id` int NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `appid` varchar(100) NOT NULL COMMENT 'appid',
  `appsecret` varchar(200) NOT NULL COMMENT '配置密钥',
  `title` varchar(100) NOT NULL COMMENT '社交描述',
  `create_by` varchar(128) DEFAULT NULL,
  `update_by` varchar(128) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT='中台APP社交属性表';

-- ----------------------------
-- Records of app_config
-- ----------------------------
BEGIN;
INSERT INTO `app_config` VALUES (1, 1, 1, '你的appi', '你的secret', '测试', '1', '1', '2020-11-23 17:44:05.357', '2020-11-26 13:59:18.267', NULL);
COMMIT;

-- ----------------------------
-- Table structure for app_enroll
-- ----------------------------
DROP TABLE IF EXISTS `app_enroll`;
CREATE TABLE `app_enroll` (
  `aeid` bigint NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '中台表用户的id',
  `auid` bigint NOT NULL DEFAULT '0' COMMENT '中台表appuser的id',
  `actid` bigint NOT NULL DEFAULT '0' COMMENT '投票活动的id',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '名字',
  `address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '地址',
  `images` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '图片',
  `descr` varchar(2500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '介绍',
  `votecount` bigint NOT NULL DEFAULT '0' COMMENT '投票数',
  `viewcount` bigint NOT NULL DEFAULT '0' COMMENT '浏览数',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '0,未审核，1.审核通过，2.删除',
  `update_by` int DEFAULT '0',
  `create_by` int DEFAULT '0',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`aeid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=68 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报名表';

-- ----------------------------
-- Table structure for app_user
-- ----------------------------
DROP TABLE IF EXISTS `app_user`;
CREATE TABLE `app_user` (
  `auid` bigint NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '对应中台表中的id',
  `openid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '社交属性的openid',
  `nickname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '昵称',
  `avator` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像',
  `sex` tinyint unsigned DEFAULT '0' COMMENT '性别',
  `city` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `province` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `country` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `privilege` text COLLATE utf8mb4_unicode_ci,
  `unionid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`auid`) USING BTREE,
  KEY `openid` (`openid`(191)) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='中台具体用户表';

-- ----------------------------
-- Table structure for app_votes
-- ----------------------------
DROP TABLE IF EXISTS `app_votes`;
CREATE TABLE `app_votes` (
  `avid` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '投票人序号：自增',
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '中台表用户的id',
  `auid` bigint NOT NULL DEFAULT '0' COMMENT '中台表appuser的id',
  `actid` bigint NOT NULL DEFAULT '0' COMMENT '投票活动的id',
  `ip` varchar(15) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '投票人IP',
  `aeid` bigint unsigned NOT NULL COMMENT '投票的id',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`avid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8 COMMENT='投票记录表';

-- ----------------------------
-- Table structure for app_votes_activity
-- ----------------------------
DROP TABLE IF EXISTS `app_votes_activity`;
CREATE TABLE `app_votes_activity` (
  `actid` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '投票活动的id',
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '投票活动名称',
  `descr` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '投票活动描述',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0无效，1.是有效',
  `start_date` int NOT NULL DEFAULT '0' COMMENT '投票活动开始时间',
  `enroll_date` int NOT NULL DEFAULT '0' COMMENT '开始投票时间',
  `end_date` int NOT NULL DEFAULT '0' COMMENT '投票活动结束时间',
  `votecount` bigint NOT NULL DEFAULT '0' COMMENT '投票活动的总票数',
  `enrollcount` bigint NOT NULL DEFAULT '0' COMMENT '报名人数',
  `viewcount` bigint NOT NULL DEFAULT '0' COMMENT '投票活动的总浏览量',
  `create_time` timestamp NULL DEFAULT NULL,
  `create_by` int DEFAULT '0',
  `update_by` int DEFAULT '0',
  `update_time` timestamp NULL DEFAULT NULL,
  `type` tinyint DEFAULT NULL COMMENT '投票的方式:1一次性，2.按天来',
  `num` int DEFAULT NULL COMMENT '单位',
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`actid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COMMENT='投票活动表';

-- ----------------------------
-- Records of app_votes_activity
-- ----------------------------
BEGIN;
INSERT INTO `app_votes_activity` VALUES (1, 1, 1, '测试xxx', '测试xxxx', 1, 1606565600, 1607565600, 1609322231, 458, 63, 548, NULL, NULL, 1, NULL, 2, 104, NULL);
COMMIT;

-- ----------------------------
-- Table structure for base_app
-- ----------------------------
DROP TABLE IF EXISTS `base_app`;
CREATE TABLE `base_app` (
  `id` int NOT NULL AUTO_INCREMENT,
  `logo` varchar(1000) DEFAULT NULL COMMENT '应用login',
  `sname` varchar(100) NOT NULL COMMENT '应用名称',
  `isclose` int NOT NULL COMMENT '站点是否关闭',
  `fullwebsite` varchar(200) NOT NULL COMMENT '完整的域名',
  `website` varchar(100) NOT NULL COMMENT '站点名称',
  `create_by` varchar(128) DEFAULT NULL,
  `update_by` varchar(128) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT='中台APP表';

-- ----------------------------
-- Records of base_app
-- ----------------------------
BEGIN;
INSERT INTO `base_app` VALUES (1, 'xx', 'xxx投票xx', 0, 'http://xxxxxx', 'xxxxx', '1', '1', '2020-11-23 17:33:48.916', '2020-12-05 13:31:08.519', NULL);
COMMIT;

-- ----------------------------
-- Table structure for base_member
-- ----------------------------
DROP TABLE IF EXISTS `base_member`;
CREATE TABLE `base_member` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户id',
  `username` varchar(20) NOT NULL COMMENT '帐号',
  `password` varchar(50) NOT NULL COMMENT '密码',
  `salt` varchar(50) NOT NULL COMMENT '密码加盐',
  `mobile` varchar(20) NOT NULL COMMENT '手机号',
  `icard` varchar(50) DEFAULT NULL COMMENT '身份证号码',
  `realname` varchar(20) NOT NULL COMMENT '真实姓名',
  `status` int NOT NULL COMMENT '状态',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mobile_index` (`mobile`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT='中台用户表';


SET FOREIGN_KEY_CHECKS = 1;
