-- ============================================================================
-- 数据中台演示数据包 demo_data.sql
-- ----------------------------------------------------------------------------
-- 用途    : 一键生成一套有真实感的"电商"演示数据集（元数据采集 / AI 问数 /
--           数据质量 / 标签圈群 / 数据服务 / 数据标准核对等演示场景通用）。
-- 环境    : MySQL 8.4（8.0+ 均可，依赖递归 CTE / 窗口函数），utf8mb4。
-- 目标库  : datacenter（本文件不含 CREATE DATABASE / USE，
--           调用方先建库并 source sql.sql 与 docs/schema/phase*.sql 后再执行本文件）：
--             mysql -uroot -padmin
--             mysql > CREATE DATABASE IF NOT EXISTS datacenter DEFAULT CHARSET utf8mb4;
--             mysql > USE datacenter;
--             mysql > SOURCE sql.sql;
--             mysql > SOURCE docs/schema/phase1.sql;   -- 依次 phase2/3/5/6
--             mysql > SOURCE docs/demo/demo_data.sql;
--             （或命令行一行： mysql -uroot -padmin datacenter < docs/demo/demo_data.sql ）
-- 说明    : 1) 全部数据由"递归 CTE + 算术伪随机"确定性生成，无 RAND()，
--              任意环境重复执行结果完全一致，便于演示口径对齐；
--           2) 脚本可重复执行（先 DROP 演示表、再清插演示账号）；
--           3) 演示表统一 dm_ 前缀，与平台自身表（sql.sql / phase*.sql）区分；
--           4) 已在 mysql:8.4 官方镜像上真实 source 验证（行数/关联见文末核对清单）。
-- ============================================================================

SET NAMES utf8mb4;
-- 递归 CTE 生成 1500/2500 行，默认 cte_max_recursion_depth=1000 不够，放宽到 3000
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
  `user_name`     varchar(32)  NOT NULL                COMMENT '用户名(手机号掩码,如138****0421)',
  `level`         tinyint      NOT NULL                COMMENT '会员等级1-5(5最高,L4/L5约30%为高价值用户)',
  `registered_at` datetime     NOT NULL                COMMENT '注册时间(2023年30%/2024年40%/2025年30%)',
  `province`      varchar(20)  NOT NULL                COMMENT '所属大区(华东40%/华北20%/华南15%/西南10%/华中10%/东北5%)',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商用户表(用户量/省份分布等问数与标签圈群演示口径,500行,确定性生成)';

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
-- 三、dm_products 演示商品表（80 行 = 数码/家居/美妆/服饰 各 20 款）
-- ============================================================================
DROP TABLE IF EXISTS `dm_products`;
CREATE TABLE `dm_products` (
  `id`           bigint        NOT NULL              COMMENT '商品ID(1-80)',
  `product_name` varchar(128)  NOT NULL              COMMENT '商品名称(每款唯一)',
  `category`     varchar(20)   NOT NULL              COMMENT '商品类目(数码/家居/美妆/服饰)',
  `price`        decimal(10,2) NOT NULL              COMMENT '售价(元,.99结尾)',
  `stock`        int           NOT NULL              COMMENT '库存(件,50-999)',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商商品表(商品/价格/库存等问数与数据服务演示口径,80行,确定性生成)';

INSERT INTO `dm_products` (`id`, `product_name`, `category`, `price`, `stock`)
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
  -- 价格按类目分档：数码 99-2999 / 家居 59-899 / 美妆 39-599 / 服饰 79-799，.99 结尾
  ROUND(
    CASE (n - 1) DIV 20
      WHEN 0 THEN 99  + ((n * 137) MOD 2900) + 0.99
      WHEN 1 THEN 59  + ((n * 137) MOD 840)  + 0.99
      WHEN 2 THEN 39  + ((n * 137) MOD 560)  + 0.99
      ELSE         79  + ((n * 137) MOD 720)  + 0.99
    END, 2)                                AS price,
  50 + (n * 29) MOD 950                    AS stock
FROM seq;

-- ============================================================================
-- 四、dm_orders 演示订单表（1500 行）
--     order_no=ORD+yyyyMMdd+当日4位序号（窗口函数按天编号）/
--     user_id 均来自 dm_users / status 加权 / created_at 近 90 天逐日分布
--     （amount 先置 0，待明细生成后按订单汇总回填，保证金额=明细之和）
-- ============================================================================
DROP TABLE IF EXISTS `dm_orders`;
CREATE TABLE `dm_orders` (
  `id`         bigint       NOT NULL                    COMMENT '订单ID(1-1500)',
  `order_no`   varchar(32)  NOT NULL                    COMMENT '订单号(ORD+yyyyMMdd+当日4位序号)',
  `user_id`    bigint       NOT NULL                    COMMENT '下单用户ID(关联dm_users.id)',
  `amount`     decimal(10,2) NOT NULL DEFAULT '0.00'    COMMENT '订单金额(=订单明细汇总,元)',
  `status`     enum('paid','shipped','completed','cancelled') NOT NULL
                                                         COMMENT '订单状态(paid25%/shipped20%/completed45%/cancelled10%)',
  `created_at` datetime     NOT NULL                    COMMENT '下单时间(近90天逐日分布,昨天起往前90天)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-电商订单表(订单量/销售额/订单状态等问数与质量规则演示口径,1500行,确定性生成)';

INSERT INTO `dm_orders` (`id`, `order_no`, `user_id`, `amount`, `status`, `created_at`)
WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 1500
),
base AS (
  SELECT
    n,
    -- 下单用户：映射到 dm_users.id(1-500)，人均约 3 单
    ((n * 13 + (n DIV 500) * 117 + 5) MOD 500) + 1 AS user_id,
    -- 下单日期（距今天数 1-90）：每 7 单抽 1 单集中于近两周（模拟大促小高峰），
    -- 其余在 90 天内均匀分布（37 与 90 互质，90 天每天均有 >=13 单，逐日曲线不断点）
    CASE WHEN n MOD 7 = 0 THEN 1 + (n * 11) MOD 14
                         ELSE 1 + (n * 37) MOD 90
    END                                            AS day_offset,
    -- 订单状态加权：cancelled 10% / paid 25% / shipped 20% / completed 45%
    CASE
      WHEN (n * 17 + 3) MOD 100 < 10 THEN 'cancelled'
      WHEN (n * 17 + 3) MOD 100 < 35 THEN 'paid'
      WHEN (n * 17 + 3) MOD 100 < 55 THEN 'shipped'
      ELSE 'completed'
    END                                            AS status
  FROM seq
),
numbered AS (
  SELECT
    n, user_id, status,
    -- 下单时刻：08:00:00-21:59:59 内伪随机铺开（日期只取昨天及以前，杜绝未来时间）
    TIMESTAMP(DATE_SUB(CURDATE(), INTERVAL day_offset DAY),
              SEC_TO_TIME(28800 + (n * 911) MOD 50400)) AS created_at,
    -- 当日流水号：同一天内按 n 顺序编号 1-4 位（递归 CTE 内禁用窗口函数，
    -- 故在非递归 CTE numbered 中做 ROW_NUMBER）
    ROW_NUMBER() OVER (
      PARTITION BY DATE_SUB(CURDATE(), INTERVAL day_offset DAY)
      ORDER BY n)                                  AS rn
  FROM base
)
SELECT
  n,
  CONCAT('ORD', DATE_FORMAT(created_at, '%Y%m%d'), LPAD(rn, 4, '0')) AS order_no,
  user_id,
  0.00,                                            -- amount 占位，稍后按明细汇总回填
  status,
  created_at
FROM numbered;

-- ============================================================================
-- 五、dm_order_items 演示订单明细表（2500 行）
--     每单至少 1 条明细（前 1000 单各 2 条），product_id 均存在于 dm_products，
--     price 取商品现价快照（每 5 条抽 1 条打 9 折，模拟促销价）
-- ============================================================================
DROP TABLE IF EXISTS `dm_order_items`;
CREATE TABLE `dm_order_items` (
  `id`         bigint        NOT NULL AUTO_INCREMENT COMMENT '明细ID',
  `order_id`   bigint        NOT NULL               COMMENT '订单ID(关联dm_orders.id)',
  `product_id` bigint        NOT NULL               COMMENT '商品ID(关联dm_products.id)',
  `qty`        int           NOT NULL               COMMENT '购买数量(1-3件)',
  `price`      decimal(10,2) NOT NULL               COMMENT '成交单价(元,下单时商品价快照,含9折促销)',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_product_id` (`product_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4
  COMMENT = '演示-订单明细表(明细/单价/购买数量等血缘与数据服务演示口径,2500行,确定性生成)';

INSERT INTO `dm_order_items` (`order_id`, `product_id`, `qty`, `price`)
WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 2500
)
SELECT
  ((n - 1) MOD 1500) + 1                           AS order_id,   -- 1500 单全覆盖,前 1000 单 2 条明细
  ((n * 31 + 7) MOD 80) + 1                        AS product_id, -- 80 款商品全覆盖(31 与 80 互质)
  1 + ((n * 7) MOD 3)                              AS qty,
  ROUND(p.price * IF((n MOD 5) = 0, 0.90, 1.00), 2) AS price
FROM seq
JOIN dm_products p ON p.id = ((n * 31 + 7) MOD 80) + 1;

-- 订单金额回填：amount = 该单全部明细 price*qty 之和（严格一致，供"总销售额"演示）
UPDATE `dm_orders` o
JOIN (
  SELECT order_id, SUM(price * qty) AS total_amount
  FROM `dm_order_items`
  GROUP BY order_id
) t ON t.order_id = o.id
SET o.amount = t.total_amount;

-- ============================================================================
-- 六、数据自检（本文件默认不执行，演示/验证时可手工放开注释）
-- ============================================================================
-- SELECT COUNT(*) AS users_500      FROM dm_users;        -- 预期 500
-- SELECT COUNT(*) AS products_80    FROM dm_products;     -- 预期 80
-- SELECT COUNT(*) AS orders_1500    FROM dm_orders;       -- 预期 1500
-- SELECT COUNT(*) AS order_items_2500 FROM dm_order_items;-- 预期 2500
-- -- 外键完整性抽查：应均为 0
-- SELECT COUNT(*) AS bad_user_fk    FROM dm_orders o LEFT JOIN dm_users u ON u.id = o.user_id WHERE u.id IS NULL;
-- SELECT COUNT(*) AS bad_item_fk    FROM dm_order_items i LEFT JOIN dm_orders o ON o.id = i.order_id WHERE o.id IS NULL;
-- SELECT COUNT(*) AS bad_product_fk FROM dm_order_items i LEFT JOIN dm_products p ON p.id = i.product_id WHERE p.id IS NULL;
-- -- 金额一致性抽查：应均为 0（amount = 明细汇总）
-- SELECT COUNT(*) FROM dm_orders o
--   JOIN (SELECT order_id, SUM(price * qty) s FROM dm_order_items GROUP BY order_id) t ON t.order_id = o.id
--  WHERE o.amount <> t.s;
-- -- 业务抽查
-- SELECT province, COUNT(*) FROM dm_users GROUP BY province ORDER BY 2 DESC;      -- 华东第一
-- SELECT level, COUNT(*) FROM dm_users GROUP BY level ORDER BY 1;                 -- 1-5 加权分布
-- SELECT status, COUNT(*) FROM dm_orders GROUP BY status ORDER BY 2 DESC;         -- completed 最多
-- SELECT COUNT(DISTINCT DATE(created_at)) AS active_days FROM dm_orders;          -- 预期 90
-- SELECT SUM(amount) AS gmv FROM dm_orders WHERE status <> 'cancelled';           -- 有效 GMV
