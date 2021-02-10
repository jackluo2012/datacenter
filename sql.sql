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


--- 加入问答抽奖


-- ----------------------------
-- Table structure for app_questions
-- ----------------------------
DROP TABLE IF EXISTS `app_questions`;
CREATE TABLE `app_questions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `activity_id` int NOT NULL DEFAULT '0' COMMENT '活动的id',
  `question` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '问题',
  `options` text CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '选项',
  `corrent` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT 'A' COMMENT '正确选项（ABCD）',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态（01）',
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `questions_question_unique` (`question`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Records of app_questions
-- ----------------------------
BEGIN;
INSERT INTO `app_questions` VALUES (1, 1, 2, 1, '垃圾可以分成几类？', 'A.两类; \r\nB.三类;\r\nC.四类;\r\nD.五类;', 'C', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions` VALUES (2, 1, 2, 1, '下列哪个不属于厨余垃圾？', 'A.过期食品; \r\nB.剩饭剩菜;\r\nC.鱼刺和骨头;\r\nD.废弃的金属勺子;', 'D', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions` VALUES (3, 1, 2, 1, '联运环境的可回收垃圾箱箱体是什么颜色？', 'A.绿色; \r\nB.红色;\r\nC.蓝色;\r\nD.黄色;', 'C', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions` VALUES (4, 1, 2, 1, '下列哪些不属于可回收垃圾？', 'A.废铁丝、废铁; \r\nB.用过的餐巾纸、茶叶渣;\r\nC.玻璃瓶、废塑;\r\nD.旧衣服、废报纸;', 'B', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions` VALUES (5, 1, 2, 1, '下列哪个不是有害垃圾？', 'A.碎玻璃片; \r\nB.过期药品;\r\nC.废水银温度;\r\nD.废电池;', 'A', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions` VALUES (6, 1, 2, 1, '过期药品属于_____，需要特殊安全处理？', 'A.其他垃圾; \r\nB.有害垃圾;\r\nC.不可回收垃圾;\r\nD.餐厨垃圾;', 'B', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions` VALUES (7, 1, 2, 1, '哪种垃圾可以作为肥料滋养土壤、庄稼？', 'A.厨余垃圾; \r\nB.可回收垃圾;\r\nC.有害垃圾;\r\nD.其他垃圾;', 'A', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions` VALUES (8, 1, 2, 1, '哪种垃圾可以再造成新瓶子、再生纸和塑料玩具？', 'A.厨余垃圾; \r\nB.可回收垃圾;\r\nC.有害垃圾;\r\nD.其他垃圾;', 'B', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
COMMIT;

-- ----------------------------
-- Table structure for app_questions_activities
-- ----------------------------
DROP TABLE IF EXISTS `app_questions_activities`;
CREATE TABLE `app_questions_activities` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `title` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '活动主题',
  `start_date` int NOT NULL DEFAULT '0' COMMENT '投票活动开始时间',
  `end_date` int NOT NULL DEFAULT '0' COMMENT '投票活动结束时间',
  `get_score` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '及格分数',
  `activity_week` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '活动周期（可多选）,分隔',
  `header` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '分享标题',
  `des` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '分享文字',
  `image` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '分享图片',
  `rule` text CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '抽奖规则',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态',
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `activities_title_unique` (`title`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Records of app_questions_activities
-- ----------------------------
BEGIN;
INSERT INTO `app_questions_activities` VALUES (1, 1, 2, '转盘抽奖', 0, 0, 60.00, '6,1,2,3,4,5', '分好啦抽奖答题，小朋友快来玩呀', '分好啦抽奖答题，小朋友快来玩呀', 'image/answer_bg.png', '1.	关注“分好啦”微信服务号，回复关键词：有奖答题或点击微信号底部自定义菜单“我要答题抽奖”参与活动，每周6有1次抽奖机会，祝你好运！\n2.	每人每周六可通过垃圾分类有奖答题免费抽奖1次，抽奖机会当天有效\n3.	活动中奖结果均以页面显示为准，中奖者请与中奖24小时内点击领取奖品，中奖商品\n4.	凡中奖客户后台截图回复手机号，工作人员为您进行奖品兑换\n5.    奖项设置： 一等奖：50元话费X4；二等奖：1G流量套餐x10；三等奖：垃圾分类500积分x15', 1, '2021-01-27 16:26:44', '2021-01-27 16:26:44');
COMMIT;

-- ----------------------------
-- Table structure for app_questions_answers
-- ----------------------------
DROP TABLE IF EXISTS `app_questions_answers`;
CREATE TABLE `app_questions_answers` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '中台表用户的id',
  `auid` bigint DEFAULT NULL,
  `test_id` int unsigned NOT NULL,
  `activity_id` int unsigned NOT NULL,
  `answers` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '用户答题对错',
  `score` varchar(10) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '0.00' COMMENT '得分',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `answers_uid_foreign` (`uid`) USING BTREE,
  KEY `answers_test_id_foreign` (`test_id`) USING BTREE,
  KEY `answers_activity_id_foreign` (`activity_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for app_questions_awards
-- ----------------------------
DROP TABLE IF EXISTS `app_questions_awards`;
CREATE TABLE `app_questions_awards` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `title` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci DEFAULT NULL COMMENT '奖项名',
  `activity_id` int unsigned NOT NULL,
  `start_probability` int NOT NULL DEFAULT '0' COMMENT '开始概率',
  `end_probability` int NOT NULL DEFAULT '0' COMMENT '结束概率',
  `number` int unsigned NOT NULL DEFAULT '0' COMMENT '中奖个数',
  `is_lottery` tinyint NOT NULL DEFAULT '0' COMMENT '是否属于中奖',
  `header` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '分享标题',
  `des` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '分享文本',
  `image` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci DEFAULT NULL COMMENT '分享图片',
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `awards_title_unique` (`title`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Records of app_questions_awards
-- ----------------------------
BEGIN;
INSERT INTO `app_questions_awards` VALUES (1, 1, 2, '一等奖：50元话费', 1, 1, 4, 4, 1, '分好啦抽奖答题，小朋友快来玩呀', '太幸运了，我在联运环境抽中了一等奖：50元话费，你也一起来参与吧。', 'image/answer_bg.png', '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions_awards` VALUES (2, 1, 2, '不要灰心', 1, 30, 3030, 0, 0, '分好啦抽奖答题，小朋友快来玩呀', '好可惜，我在联运环境差一点就抽中一等奖，你也一起来参与吧。', 'image/answer_bg.png', '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions_awards` VALUES (3, 1, 2, '二等奖：1G流量套餐', 1, 5, 14, 10, 1, '分好啦抽奖答题，小朋友快来玩呀', '太幸运了，我在联运环境抽中了二等奖：1G流量套餐，你也一起来参与吧。', 'image/answer_bg.png', '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions_awards` VALUES (4, 1, 2, '谢谢参与', 1, 31, 6030, 0, 0, '分好啦抽奖答题，小朋友快来玩呀', '好可惜，我在联运环境差一点就抽中一等奖，你也一起来参与吧。', 'image/answer_bg.png', '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions_awards` VALUES (5, 1, 2, '三等奖：垃圾分类500积分', 1, 15, 29, 15, 1, '分好啦抽奖答题，小朋友快来玩呀', '太幸运了，我在联运环境抽中了三等奖：垃圾分类500积分，你也一起来参与吧。', 'image/answer_bg.png', '2021-01-27 16:26:44', '2021-01-27 16:26:44');
INSERT INTO `app_questions_awards` VALUES (6, 1, 2, '要加油哦', 1, 6031, 10000, 0, 0, '分好啦抽奖答题，小朋友快来玩呀', '好可惜，我在联运环境差一点就抽中一等奖，你也一起来参与吧。', 'image/answer_bg.png', '2021-01-27 16:26:44', '2021-01-27 16:26:44');
COMMIT;

-- ----------------------------
-- Table structure for app_questions_converts
-- ----------------------------
DROP TABLE IF EXISTS `app_questions_converts`;
CREATE TABLE `app_questions_converts` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '中台表用户的id',
  `auid` bigint NOT NULL DEFAULT '0' COMMENT '中台表用户的id',
  `lottery_id` int unsigned NOT NULL,
  `username` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '获奖名',
  `phone` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci DEFAULT NULL COMMENT '手机号',
  `status` int NOT NULL DEFAULT '0' COMMENT '处理（0/1）',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for app_questions_lotteries
-- ----------------------------
DROP TABLE IF EXISTS `app_questions_lotteries`;
CREATE TABLE `app_questions_lotteries` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '中台表用户的id',
  `auid` bigint DEFAULT NULL,
  `activity_id` int unsigned NOT NULL,
  `answer_id` int unsigned NOT NULL,
  `award_id` int unsigned NOT NULL,
  `is_winning` int NOT NULL DEFAULT '0' COMMENT '是否中奖（0/1）',
  `is_convert` int NOT NULL DEFAULT '0' COMMENT '兑奖名称',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `lotteries_uid_foreign` (`uid`) USING BTREE,
  KEY `lotteries_activity_id_foreign` (`activity_id`) USING BTREE,
  KEY `lotteries_answer_id_foreign` (`answer_id`) USING BTREE,
  KEY `lotteries_award_id_foreign` (`award_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for app_questions_tests
-- ----------------------------
DROP TABLE IF EXISTS `app_questions_tests`;
CREATE TABLE `app_questions_tests` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `beid` int NOT NULL DEFAULT '0' COMMENT '对应的平台',
  `ptyid` int NOT NULL COMMENT '平台id: 1.微信公众号，2.微信小程序，3.支付宝',
  `title` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT '题库类名',
  `question_ids` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '题库题编号',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态（0/1）',
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `tests_title_unique` (`title`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;








SET FOREIGN_KEY_CHECKS = 1;
