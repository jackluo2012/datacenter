-- ============================================================================
-- 数据中台演示数据包 demo_data.sql
-- ----------------------------------------------------------------------------
-- 用途    : 一键生成一套有真实感、有"故事线"的电商演示数据集（元数据采集 /
--           AI 问数 / 数据质量 / 标签圈群 / 数据服务 / 主数据等演示场景通用）。
-- 环境    : MySQL 8.4（8.0+ 均可，依赖递归 CTE / 窗口函数），utf8mb4。
-- 目标库  : datacenter（本文件不含 CREATE DATABASE / USE，
--           调用方先建库并 source sql.sql / sys.sql 与 docs/schema/phase*.sql 后再执行本文件）：
--             mysql -uroot -padmin
--             mysql > CREATE DATABASE IF NOT EXISTS datacenter DEFAULT CHARSET utf8mb4;
--             mysql > USE datacenter;
--             mysql > SOURCE sql.sql;
--             mysql > SOURCE sys.sql;
--             mysql > SOURCE docs/schema/phase1.sql;   -- 依次 phase2/3/5/6/7
--             mysql > SOURCE docs/demo/demo_data.sql;
--             （或命令行一行： mysql -uroot -padmin datacenter < docs/demo/demo_data.sql ）
-- 故事线  : 订单时间序列埋了三个"可被系统发现"的业务故事——
--           a) 大促峰值日   ：60 天前一次大促，当日订单量 48 单 = 平日 16 单的 3 倍
--                              （讲"平台扛住了峰值，报表一眼看到大促"）；
--           b) 近 7 天异常  ：3/4/5 天前订单量骤降到 8 单/天（-50%），供质量
--                              fluctuation / 趋势问数讲"数据出问题了，系统发现了"；
--           c) 一批脏数据   ：异常 3 天里各混入 4 笔 amount=-1 的脏订单（共 12 笔，
--                              order_no 带 BAD 段），`where amount<0` 可查——
--                              供质量 value_range 规则命中并抓取错误数据样本。
-- 说明    : 1) 全部数据由"递归 CTE + 算术伪随机"确定性生成，无 RAND()，
--              任意环境重复执行结果完全一致，便于演示口径对齐；
--           2) 脚本可重复执行（先 DROP 演示表、再清插演示账号）；
--           3) 演示表统一 dm_ 前缀，与平台自身表（sql.sql / phase*.sql）区分；
--           4) 表/列 COMMENT 埋了演示关键词（订单量/销售额/用户画像/用户反馈/好评率…），
--              AI 问数语义检索靠 COMMENT 命中表候选；
--           5) 已在 mysql:8.4 官方镜像上真实 source 验证（行数/关联见文末核对清单）。
-- ============================================================================

SET NAMES utf8mb4;
-- 递归 CTE 生成 2500 行，默认 cte_max_recursion_depth=1000 不够，放宽到 3000
SET SESSION cte_max_recursion_depth = 3000;

-- ============================================================================
-- 一、演示登录账号与演示站点（对应平台表，非 dm_ 前缀）
-- ============================================================================

-- 1) 演示站点：base_app（网关多租户 beid 即 base_app.id，此处固定 id=2 供演示用；
--    base_app 表无 beid 列（见 sql.sql DDL），beid=2 即指本行 id=2）
DELETE FROM `base_app` WHERE `id` = 2;
INSERT INTO `base_app`
  (`id`, `logo`, `sname`, `isclose`, `fullwebsite`, `website`,
   `create_by`, `update_by`, `created_at`, `updated_at`, `deleted_at`)
VALUES
  (2, '', '演示电商平台', 0, 'http://demo.datacenter.local', 'demo',
   'demo', 'demo', NOW(3), NOW(3), NULL);

-- 2) 演示登录账号：base_member（网关 POST /user/login 明文比对 password，
--    故 password 直接存 123456；deleted_at 置 epoch 与服务端注册逻辑
--    （user/rpc registerlogic.go time.Unix(0,0)）保持一致，避免 ORM 扫 NULL 报错）
--    注意双账号体系：本行是 C 端会员（/user/login，13800138000/123456），
--    管理台账号是 sys_user（admin/admin123，/sys/login，由 sys.sql 初始化）
DELETE FROM `base_member` WHERE `mobile` = '13800138000';
INSERT INTO `base_member`
  (`username`, `password`, `salt`, `mobile`, `icard`, `realname`, `status`, `deleted_at`)
VALUES
  ('demo', '123456', 'demo', '13800138000', NULL, '演示用户', 1, '1970-01-01 00:00:00');

-- ============================================================================
-- 二、dm_users 演示用户表（500 行）
--     user_name 手机号掩码 / level 1-5 加权 / registered_at 2023-2025 分布 /
--     province 大区分布（华东最多，便于"哪个省份用户最多"演示）
-- ============================================================================
DROP TABLE IF EXISTS `dm_users`;
CREATE TABLE `dm_users` (
  `id`            bigint       NOT NULL                COMMENT '用户ID(1-500)',
  `user_name`     varchar(32)  NOT NULL                COMMENT '用户名(手机号掩码,如138****0421,用户画像脱敏演示)',
  `level`         tinyint      NOT NULL                COMMENT '会员等级1-5(5最高,L4/L5约30%为高价值用户,标签圈群口径)',
  `registered_at` datetime     NOT NULL                COMMENT '注册时间(2023年30%/2024年40%/2025年30%,新增用户画像口径)',
  `province`      varchar(20)  NOT NULL                COMMENT '所属大区(华东40%/华北20%/华南15%/西南10%/华中10%/东北5%,省份分布口径)',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商用户主表(用户量500/用户画像/会员等级/省份分布/高价值用户,用户标签圈群与AI问数演示口径,确定性生成)';

INSERT INTO `dm_users` (`id`, `user_name`, `level`, `registered_at`, `province`)
WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 500
)
SELECT
  n,
  -- 手机号掩码：10 个真实号段轮转 + 伪随机后四位（7919 与 10000 互质，500 行不重号）
  CONCAT(
    ELT(1 + (n * 7) MOD 10,
        '133', '135', '136', '137', '138', '150', '155', '158', '186', '188'),
    '****',
    LPAD((n * 7919) MOD 10000, 4, '0')
  )                                        AS user_name,
  -- 会员等级加权：L1 15% / L2 25% / L3 30% / L4 20% / L5 10%（level>=4 约 150 人）
  CASE
    WHEN (n * 37 + FLOOR((n - 1) / 50) * 11) MOD 100 < 15 THEN 1
    WHEN (n * 37 + FLOOR((n - 1) / 50) * 11) MOD 100 < 40 THEN 2
    WHEN (n * 37 + FLOOR((n - 1) / 50) * 11) MOD 100 < 70 THEN 3
    WHEN (n * 37 + FLOOR((n - 1) / 50) * 11) MOD 100 < 90 THEN 4
    ELSE 5
  END                                      AS level,
  -- 注册时间分布：2023 30% / 2024 40% / 2025 30%，年内均匀铺开并带时刻
  TIMESTAMP(
    CASE
      WHEN (n * 53) MOD 100 < 30 THEN DATE_ADD('2023-01-01', INTERVAL (n * 53) MOD 365 DAY)
      WHEN (n * 53) MOD 100 < 70 THEN DATE_ADD('2024-01-01', INTERVAL (n * 53 + 37) MOD 366 DAY)
      ELSE                            DATE_ADD('2025-01-01', INTERVAL (n * 53 + 71) MOD 365 DAY)
    END,
    SEC_TO_TIME((n * 97) MOD 86400)
  )                                        AS registered_at,
  -- 大区分布：华东 40%（占比第一）/ 华北 20% / 华南 15% / 西南 10% / 华中 10% / 东北 5%
  CASE
    WHEN (n * 29 + 13) MOD 100 < 40 THEN '华东'
    WHEN (n * 29 + 13) MOD 100 < 60 THEN '华北'
    WHEN (n * 29 + 13) MOD 100 < 75 THEN '华南'
    WHEN (n * 29 + 13) MOD 100 < 85 THEN '西南'
    WHEN (n * 29 + 13) MOD 100 < 95 THEN '华中'
    ELSE '东北'
  END                                      AS province
FROM seq;

-- ============================================================================
-- 三、dm_products 演示商品表（80 行 = 数码/家居/美妆/服饰 四类各 20 款）
--     每类目内价格分 4 个价格带（入门款/中端款/高端款/旗舰款 各 5 款），
--     全表每个价格带 20 款，价格 .99 结尾
-- ============================================================================
DROP TABLE IF EXISTS `dm_products`;
CREATE TABLE `dm_products` (
  `id`           bigint        NOT NULL              COMMENT '商品ID(1-80)',
  `product_name` varchar(128)  NOT NULL              COMMENT '商品名称(每款唯一)',
  `category`     varchar(20)   NOT NULL              COMMENT '商品类目(数码/家居/美妆/服饰四类各20款,类目销售分析口径)',
  `price_band`   varchar(8)    NOT NULL              COMMENT '价格带(入门款/中端款/高端款/旗舰款各20款,价格带分析口径)',
  `price`        decimal(10,2) NOT NULL              COMMENT '售价(元,.99结尾,商品价格口径)',
  `stock`        int           NOT NULL              COMMENT '库存(件,50-999,商品库存口径)',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商商品表(商品80款/类目四类/价格带/库存/商品销量,AI问数与数据服务演示口径,确定性生成)';

INSERT INTO `dm_products` (`id`, `product_name`, `category`, `price_band`, `price`, `stock`)
WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 80
)
SELECT
  n,
  -- 每类目 20 款真实感商品名（(n-1) DIV 20 选类目，(n-1) MOD 20 选款）
  CASE (n - 1) DIV 20
    WHEN 0 THEN ELT(1 + (n - 1) MOD 20,
        '无线蓝牙耳机', '智能手表', '移动电源', '蓝牙音箱', '机械键盘',
        '电竞鼠标', '27英寸显示器', 'NVMe固态硬盘', '行车记录仪', '智能门锁',
        '平板电脑', '千兆路由器', '电动牙刷', '数码相机', '智能体脂秤',
        '迷你无人机', '智能音箱', 'USB-C扩展坞', '电动剃须刀', '电子阅读器')
    WHEN 1 THEN ELT(1 + (n - 1) MOD 20,
        '乳胶枕头', '记忆棉床垫', '人体工学椅', '电动升降桌', '落地灯',
        '香薰机', '扫地机器人', '空气净化器', '加湿器', '电热水壶',
        '不粘锅套装', '保温杯', '收纳箱', '懒人沙发', '遮光窗帘',
        '智能台灯', '除螨仪', '手持挂烫机', '破壁机', '空气炸锅')
    WHEN 2 THEN ELT(1 + (n - 1) MOD 20,
        '保湿精华液', '补水面膜', '防晒霜', '氨基酸洗面奶', '爽肤水',
        '保湿乳液', '眼霜', '丝绒口红', '粉底液', '卸妆水',
        '淡香水', '眉笔', '睫毛膏', '腮红盘', '妆前乳',
        '修护发膜', '沐浴露', '护手霜', '身体乳', '润唇膏')
    ELSE ELT(1 + (n - 1) MOD 20,
        '男士纯棉T恤', '女士连衣裙', '直筒牛仔裤', '休闲连帽卫衣', '轻便羽绒服',
        '经典风衣', '缓震运动鞋', '小白板鞋', '棒球帽', '针织围巾',
        '羊毛大衣', '修身针织衫', '商务衬衫', '休闲短裤', '高腰瑜伽裤',
        '速干泳衣', '袜子礼盒', '真皮腰带', '双肩背包', '手提包')
  END                                      AS product_name,
  CASE (n - 1) DIV 20
    WHEN 0 THEN '数码' WHEN 1 THEN '家居' WHEN 2 THEN '美妆' ELSE '服饰'
  END                                      AS category,
  -- 价格带：类目内每 5 款一档（入门/中端/高端/旗舰），全表每带 20 款
  CASE ((n - 1) MOD 20) DIV 5
    WHEN 0 THEN '入门款' WHEN 1 THEN '中端款' WHEN 2 THEN '高端款' ELSE '旗舰款'
  END                                      AS price_band,
  -- 价格按类目×价格带分档，.99 结尾（数码 99-2999 / 家居 59-1299 / 美妆 39-699 / 服饰 49-999）
  ROUND(
    CASE (n - 1) DIV 20
      WHEN 0 THEN CASE ((n - 1) MOD 20) DIV 5   -- 数码：99-199/299-599/899-1599/1999-2999
        WHEN 0 THEN 99   + ((n * 137) MOD 100)
        WHEN 1 THEN 299  + ((n * 137) MOD 300)
        WHEN 2 THEN 899  + ((n * 137) MOD 700)
        ELSE         1999 + ((n * 137) MOD 1000) END
      WHEN 1 THEN CASE ((n - 1) MOD 20) DIV 5   -- 家居：59-99/149-299/499-699/899-1299
        WHEN 0 THEN 59  + ((n * 137) MOD 40)
        WHEN 1 THEN 149 + ((n * 137) MOD 150)
        WHEN 2 THEN 499 + ((n * 137) MOD 200)
        ELSE         899 + ((n * 137) MOD 400) END
      WHEN 2 THEN CASE ((n - 1) MOD 20) DIV 5   -- 美妆：39-69/99-199/259-399/499-699
        WHEN 0 THEN 39  + ((n * 137) MOD 30)
        WHEN 1 THEN 99  + ((n * 137) MOD 100)
        WHEN 2 THEN 259 + ((n * 137) MOD 140)
        ELSE         499 + ((n * 137) MOD 200) END
      ELSE CASE ((n - 1) MOD 20) DIV 5          -- 服饰：49-99/129-259/399-599/699-999
        WHEN 0 THEN 49  + ((n * 137) MOD 50)
        WHEN 1 THEN 129 + ((n * 137) MOD 130)
        WHEN 2 THEN 399 + ((n * 137) MOD 200)
        ELSE         699 + ((n * 137) MOD 300) END
    END + 0.99, 2)                           AS price,
  50 + (n * 29) MOD 950                      AS stock
FROM seq;

-- ============================================================================
-- 四、dm_orders 演示订单表（1500 行 = 1488 正常单 + 12 脏数据行）
--     近 90 天逐日分布 + 三段故事：
--       大促峰值日（60 天前）48 单 = 平日 3 倍，且集中在 20:00-23:59 晚高峰；
--       异常 3 天（3/4/5 天前）每天总量 8 单（-50%：正常 4 + 脏 4）；
--       脏数据 12 行 amount=-1.00（order_no 含 BAD 段，where amount<0 可查）。
--     order_no=ORD+yyyyMMdd+当日4位序号（窗口函数按天编号）/
--     user_id 均来自 dm_users / status 加权
--     （正常单 amount 先置 0，待明细生成后按订单汇总回填，保证金额=明细之和；
--       脏行无明细、不回填，amount 保持 -1）
-- ============================================================================
DROP TABLE IF EXISTS `dm_orders`;
CREATE TABLE `dm_orders` (
  `id`         bigint       NOT NULL                    COMMENT '订单ID(1-1500)',
  `order_no`   varchar(32)  NOT NULL                    COMMENT '订单号(ORD+yyyyMMdd+当日4位序号;脏数据行含BAD段)',
  `user_id`    bigint       NOT NULL                    COMMENT '下单用户ID(关联dm_users.id,用户购买画像口径)',
  `amount`     decimal(10,2) NOT NULL DEFAULT '0.00'    COMMENT '订单金额(=订单明细汇总,元;12行amount=-1为脏数据,where amount<0可查)',
  `status`     enum('paid','shipped','completed','cancelled') NOT NULL
                                                         COMMENT '订单状态(paid25%/shipped20%/completed45%/cancelled10%)',
  `created_at` datetime     NOT NULL                    COMMENT '下单时间(近90天,含60天前大促峰值与3/4/5天前异常骤降故事)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商订单主表(订单量/每日订单量趋势/销售额/GMV/客单价/订单状态,含大促峰值与异常骤降脏数据故事,AI问数与质量规则演示口径)';

-- 4.1 正常订单 1488 行：按"日配额"精确分布（确定性）
INSERT INTO `dm_orders` (`id`, `order_no`, `user_id`, `amount`, `status`, `created_at`)
WITH RECURSIVE
seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 1488
),
-- 90 天日历（day_offset 1=昨天 ... 90），逐日配额（确定性故事曲线）：
--   day 60                 → 48 单（大促峰值，= 平日 16 的 3 倍）
--   day 3/4/5              → 4 单/天（异常骤降窗口的正常单部分）
--   day 7/14/21/28/35/42/49 → 23 单（每周三会员日，约 1.4 倍）
--   day 10/40/80           → 17 单（自然波动）
--   其余 76 天             → 16 单（平日基线）
-- 配额合计 = 48 + 3*4 + 7*23 + 3*17 + 76*16 = 1488
days(day_offset, quota) AS (
  SELECT 1, 16
  UNION ALL
  SELECT day_offset + 1,
         CASE
           WHEN day_offset + 1 = 60                      THEN 48
           WHEN day_offset + 1 IN (3, 4, 5)              THEN 4
           WHEN day_offset + 1 IN (7, 14, 21, 28, 35, 42, 49) THEN 23
           WHEN day_offset + 1 IN (10, 40, 80)           THEN 17
           ELSE 16
         END
  FROM days WHERE day_offset < 90
),
span AS (
  SELECT day_offset, quota,
         COALESCE(SUM(quota) OVER (ORDER BY day_offset
            ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING), 0) AS cum_before,
         SUM(quota) OVER (ORDER BY day_offset
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)     AS cum_total
  FROM days
),
base AS (
  SELECT
    seq.n,
    span.day_offset,
    -- 下单用户：映射到 dm_users.id(1-500)，人均约 3 单
    ((seq.n * 13 + (seq.n DIV 500) * 117 + 5) MOD 500) + 1 AS user_id,
    -- 订单状态加权：cancelled 10% / paid 25% / shipped 20% / completed 45%
    CASE
      WHEN (seq.n * 17 + 3) MOD 100 < 10 THEN 'cancelled'
      WHEN (seq.n * 17 + 3) MOD 100 < 35 THEN 'paid'
      WHEN (seq.n * 17 + 3) MOD 100 < 55 THEN 'shipped'
      ELSE 'completed'
    END                                                    AS status,
    -- 下单时刻：大促日集中 20:00-23:59 晚高峰；异常日集中 10:00-15:59；
    -- 平日 08:00-21:59 内伪随机铺开（日期只取昨天及以前，杜绝未来时间）
    TIMESTAMP(
      DATE_SUB(CURDATE(), INTERVAL span.day_offset DAY),
      SEC_TO_TIME(
        CASE
          WHEN span.day_offset = 60           THEN 72000 + (seq.n * 911) MOD 14400
          WHEN span.day_offset IN (3, 4, 5)   THEN 36000 + (seq.n * 911) MOD 21600
          ELSE                                    28800 + (seq.n * 911) MOD 50400
        END)
    )                                                      AS created_at
  FROM seq
  JOIN span ON seq.n > span.cum_before AND seq.n <= span.cum_total
),
numbered AS (
  SELECT b.*,
    -- 当日流水号：同一天内按 n 顺序编号 1-4 位
    ROW_NUMBER() OVER (PARTITION BY DATE(created_at) ORDER BY n) AS rn
  FROM base b
)
SELECT
  n,
  CONCAT('ORD', DATE_FORMAT(created_at, '%Y%m%d'), LPAD(rn, 4, '0')) AS order_no,
  user_id,
  0.00,                                            -- amount 占位，稍后按明细汇总回填
  status,
  created_at
FROM numbered;

-- 4.2 脏数据行 12 笔：异常 3 天（3/4/5 天前）各混入 4 笔 amount=-1 的坏单
--     （讲法：上游系统故障那几天，单量腰斩的同时混入了 -1 元的脏单——
--       质量 value_range 规则会命中这 12 行并抓取错误样本）
INSERT INTO `dm_orders` (`id`, `order_no`, `user_id`, `amount`, `status`, `created_at`)
WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 12
)
SELECT
  1488 + n,                                        -- id 1489-1500
  CONCAT('ORD', DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 + (n MOD 3) DAY), '%Y%m%d'),
         'BAD', LPAD(n, 2, '0'))                   AS order_no,   -- 如 ORD20261009BAD01
  ((n * 89) MOD 500) + 1                           AS user_id,
  -1.00                                            AS amount,     -- 脏数据：金额为 -1
  'paid'                                           AS status,
  TIMESTAMP(DATE_SUB(CURDATE(), INTERVAL 3 + (n MOD 3) DAY),
            SEC_TO_TIME(36000 + (n * 911) MOD 21600)) AS created_at -- 10:00-15:59
FROM seq;

-- ============================================================================
-- 五、dm_order_items 演示订单明细表（2500 行）
--     每单至少 1 条明细（前 1012 单各 2 条），product_id 均存在于 dm_products，
--     price 取商品现价快照（每 5 条抽 1 条打 9 折，模拟促销价）
-- ============================================================================
DROP TABLE IF EXISTS `dm_order_items`;
CREATE TABLE `dm_order_items` (
  `id`         bigint        NOT NULL AUTO_INCREMENT COMMENT '明细ID',
  `order_id`   bigint        NOT NULL               COMMENT '订单ID(关联dm_orders.id)',
  `product_id` bigint        NOT NULL               COMMENT '商品ID(关联dm_products.id,商品销量口径)',
  `qty`        int           NOT NULL               COMMENT '购买数量(1-3件)',
  `price`      decimal(10,2) NOT NULL               COMMENT '成交单价(元,下单时商品价快照,含9折促销)',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_product_id` (`product_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商订单明细表(订单明细/商品销量/购买数量/成交价,销售额下钻与客单价血缘演示口径,2500行,确定性生成)';

INSERT INTO `dm_order_items` (`order_id`, `product_id`, `qty`, `price`)
WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 2500
)
SELECT
  ((n - 1) MOD 1488) + 1                           AS order_id,   -- 1488 正常单全覆盖,前 1012 单 2 条明细
  ((n * 31 + 7) MOD 80) + 1                        AS product_id, -- 80 款商品全覆盖(31 与 80 互质)
  1 + ((n * 7) MOD 3)                              AS qty,
  ROUND(p.price * IF((n MOD 5) = 0, 0.90, 1.00), 2) AS price
FROM seq
JOIN dm_products p ON p.id = ((n * 31 + 7) MOD 80) + 1;

-- 订单金额回填：正常单 amount = 该单全部明细 price*qty 之和（严格一致，供"总销售额"演示）
-- （脏行无明细，JOIN 不中，amount 保持 -1）
UPDATE `dm_orders` o
JOIN (
  SELECT order_id, SUM(price * qty) AS total_amount
  FROM `dm_order_items`
  GROUP BY order_id
) t ON t.order_id = o.id
SET o.amount = t.total_amount;

-- ============================================================================
-- 六、dm_user_feedback 演示用户反馈表（200 行，近 30 天）
--     score 1-5 满意度评分（4/5 好评约 70%）/ content 词典组合短句 /
--     末尾 3 行为脏行（content=''、score=0，供 not_null / 值域演示命中）——
--     供 AI 问数"用户反馈怎么样 / 好评率多少"演示
-- ============================================================================
DROP TABLE IF EXISTS `dm_user_feedback`;
CREATE TABLE `dm_user_feedback` (
  `id`         bigint       NOT NULL              COMMENT '反馈ID(1-200)',
  `user_id`    bigint       NOT NULL              COMMENT '反馈用户ID(关联dm_users.id)',
  `score`      tinyint      NOT NULL              COMMENT '满意度评分1-5(4/5好评约70%,好评率口径;末尾3行0分为脏数据)',
  `content`    varchar(255) NOT NULL              COMMENT '反馈内容(词典组合短句,口碑舆情口径;末尾3行空串为脏数据)',
  `created_at` datetime     NOT NULL              COMMENT '反馈时间(近30天)',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商用户反馈表(用户反馈/好评率/满意度评分/口碑舆情/客服体验,AI问数"用户反馈怎么样"演示口径,200行,确定性生成)';

INSERT INTO `dm_user_feedback` (`id`, `user_id`, `score`, `content`, `created_at`)
WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 200
)
SELECT
  n,
  ((n * 67) MOD 500) + 1                           AS user_id,    -- 67 与 500 互质
  -- 满意度加权：5 星 40% / 4 星 30% / 3 星 15% / 2 星 10% / 1 星 5%；末尾 3 行 0 分（脏行）
  CASE WHEN n > 197                       THEN 0
       WHEN (n * 23) MOD 100 < 40         THEN 5
       WHEN (n * 23) MOD 100 < 70         THEN 4
       WHEN (n * 23) MOD 100 < 85         THEN 3
       WHEN (n * 23) MOD 100 < 95         THEN 2
       ELSE 1
  END                                              AS score,
  -- 内容 =（按评分选正/中/负词典词）+（后缀），末尾 3 行空串（脏行）
  CASE WHEN n > 197 THEN ''
       ELSE CONCAT(
         CASE
           WHEN (n * 23) MOD 100 < 70 THEN          -- 好评（4/5 星）
             ELT(1 + (n * 7) MOD 8, '物流很快', '包装严实', '客服态度很好', '性价比超高',
                                 '质量超出预期', '发货速度快', '客服很耐心', '用过还想买')
           WHEN (n * 23) MOD 100 < 85 THEN          -- 中评（3 星）
             ELT(1 + (n * 7) MOD 3, '整体一般', '中规中矩', '符合预期')
           ELSE                                      -- 差评（1/2 星）
             ELT(1 + (n * 7) MOD 5, '发货太慢', '有色差', '客服响应慢', '包装破损', '质量一般')
         END,
         ELT(1 + (n * 11) MOD 5, '，推荐购买', '，整体满意', '，还会再来', '，感觉一般', '。')
       )
  END                                              AS content,
  -- 反馈时间：近 30 天（1-30 天前），全天时段伪随机铺开
  TIMESTAMP(DATE_SUB(CURDATE(), INTERVAL 1 + (n * 17) MOD 30 DAY),
            SEC_TO_TIME((n * 97) MOD 86400))       AS created_at
FROM seq;

-- ============================================================================
-- 七、数据自检（本文件默认不执行，演示/验证时可手工放开注释）
-- ============================================================================
-- SELECT COUNT(*) AS users_500      FROM dm_users;          -- 预期 500
-- SELECT COUNT(*) AS products_80    FROM dm_products;       -- 预期 80（四类各 20、四价格带各 20）
-- SELECT COUNT(*) AS orders_1500    FROM dm_orders;         -- 预期 1500（1488 正常 + 12 脏）
-- SELECT COUNT(*) AS order_items_2500 FROM dm_order_items;  -- 预期 2500
-- SELECT COUNT(*) AS feedback_200   FROM dm_user_feedback;  -- 预期 200（好评率(4+星)≈70%）
--
-- -- 故事数据核对（核心三查）：
-- -- 1) 大促峰值日 ≈ 平日 3 倍（配额 48 vs 16）
-- SELECT DATE_SUB(CURDATE(), INTERVAL 60 DAY) AS promo_day, COUNT(*) AS promo_cnt
--   FROM dm_orders WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 60 DAY)
--    AND created_at <  DATE_SUB(CURDATE(), INTERVAL 59 DAY);
-- SELECT ROUND(AVG(c),1) AS normal_daily FROM (
--   SELECT DATE(created_at) d, COUNT(*) c FROM dm_orders
--   WHERE created_at < DATE_SUB(CURDATE(), INTERVAL 59 DAY)
--     AND created_at >= DATE_SUB(CURDATE(), INTERVAL 90 DAY)
--     AND DATE(created_at) <> DATE_SUB(CURDATE(), INTERVAL 60 DAY)
--   GROUP BY DATE(created_at)) t;                       -- 平日均值≈17（含会员日/波动日）
-- -- 2) 异常 3 天骤降 50%（总量 8 vs 16）
-- SELECT DATE(created_at) d, COUNT(*) cnt, SUM(amount<0) dirty_cnt
--   FROM dm_orders WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 5 DAY)
--   GROUP BY DATE(created_at) ORDER BY d;               -- 3/4/5 天前各 8 单（其中 4 笔脏）
-- -- 3) 脏数据行 >= 10
-- SELECT COUNT(*) AS dirty_rows FROM dm_orders WHERE amount < 0;   -- 预期 12
-- SELECT order_no, amount, created_at FROM dm_orders WHERE amount < 0 LIMIT 5;
--
-- -- 外键完整性抽查：应均为 0
-- SELECT COUNT(*) AS bad_user_fk    FROM dm_orders o LEFT JOIN dm_users u ON u.id = o.user_id WHERE u.id IS NULL;
-- SELECT COUNT(*) AS bad_item_fk    FROM dm_order_items i LEFT JOIN dm_orders o ON o.id = i.order_id WHERE o.id IS NULL;
-- SELECT COUNT(*) AS bad_product_fk FROM dm_order_items i LEFT JOIN dm_products p ON p.id = i.product_id WHERE p.id IS NULL;
-- SELECT COUNT(*) AS bad_fb_user_fk FROM dm_user_feedback f LEFT JOIN dm_users u ON u.id = f.user_id WHERE u.id IS NULL;
-- -- 金额一致性抽查：应均为 0（正常单 amount = 明细汇总）
-- SELECT COUNT(*) FROM dm_orders o
--   JOIN (SELECT order_id, SUM(price * qty) s FROM dm_order_items GROUP BY order_id) t ON t.order_id = o.id
--  WHERE o.amount <> t.s;
-- -- 业务抽查
-- SELECT province, COUNT(*) FROM dm_users GROUP BY province ORDER BY 2 DESC;      -- 华东第一
-- SELECT level, COUNT(*) FROM dm_users GROUP BY level ORDER BY 1;                 -- 1-5 加权分布
-- SELECT status, COUNT(*) FROM dm_orders GROUP BY status ORDER BY 2 DESC;         -- completed 最多
-- SELECT category, price_band, COUNT(*), MIN(price), MAX(price) FROM dm_products
--   GROUP BY category, price_band;                                                -- 4 类 × 4 带 各 5 款
-- SELECT COUNT(DISTINCT DATE(created_at)) AS active_days FROM dm_orders;          -- 预期 90
-- SELECT SUM(amount) AS gmv FROM dm_orders WHERE status <> 'cancelled' AND amount >= 0;  -- 有效 GMV
-- SELECT ROUND(100*SUM(score>=4)/COUNT(*),1) AS good_rate FROM dm_user_feedback WHERE score > 0;  -- 好评率≈70
