-- ============================================================
-- 「数据可视化」菜单种子（DataEase 集成，前端组件 web/src/views/dataease/DataeaseEmbed.vue）
--
-- 背景：sys/model/seed.go 的内置种子树暂不含本菜单（主线合入前不改种子代码），
--       本脚本按 sys_menu 现有结构手工插入，可重复执行（幂等）。
--
-- 用法（仓库根目录执行，二选一）：
--   平台容器编排：docker compose exec -T mysql mysql -uroot -padmin datacenter < docs/dataease-menu-seed.sql
--   宿主机直连：  mysql -h127.0.0.1 -uroot -padmin datacenter < docs/dataease-menu-seed.sql
--
-- 注意：
--   1. sys_menu 按 beid 隔离，菜单 beid 必须与登录用户的 beid 一致才会出现在
--      动态菜单里。先查现有菜单用的 beid，再改下方 @beid：
--        SELECT DISTINCT beid FROM sys_menu;
--   2. 插入后重新登录（或刷新页面重拉 /sys/getrouters）即可看到「数据可视化」。
-- ============================================================

-- 0) 租户 beid：与现有菜单保持一致（常见 0/1/2，见文件头查询）
SET @beid = 0;

-- 1) 定位父目录：数据中台目录的 id 需查询得到（menu_type=1 为目录）。
--    默认挂在「数据运维」目录下；也可换成 '数据源'/'数据服务' 等任一目录名。
--    想做成与「仪表盘」同级的顶层菜单见文末备选脚本。
SET @parent_id = (
  SELECT id FROM sys_menu
  WHERE name = '数据运维' AND menu_type = 1 AND beid = @beid
  ORDER BY id
  LIMIT 1
);

-- 2) 排序号：排在父目录现有子菜单之后
SET @next_sort = (SELECT COALESCE(MAX(`sort`), 0) + 1 FROM sys_menu WHERE parent_id = @parent_id);

-- 3) 插入菜单（幂等：已存在同名菜单则跳过）。
--    path 以 / 开头 → 动态路由注册为顶层路径 /dataease（前端 joinPath 对绝对路径直接采用）；
--    component 'dataease/DataeaseEmbed' → stores/permission.ts 解析为
--    /src/views/dataease/DataeaseEmbed.vue。
INSERT INTO sys_menu (parent_id, name, path, component, `sort`, icon, menu_type, perm, visible, beid)
SELECT @parent_id, '数据可视化', '/dataease', 'dataease/DataeaseEmbed', @next_sort, 'DataBoard', 2, '', 1, @beid
FROM DUAL
WHERE @parent_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE parent_id = @parent_id AND name = '数据可视化' AND beid = @beid
  );

-- 4) 承接新菜单自增 id（id 为 AUTO_INCREMENT，用会话变量取回给角色绑定用；
--    若第 3 步因已存在/父目录缺失跳过，@menu_id 为 0，第 5 步自动不绑定）
SET @menu_id = LAST_INSERT_ID();

-- 5) 绑定到内置角色（admin=全量 / viewer=只读；菜单可见性走
--    用户→角色→菜单链路，不绑定则 /sys/getrouters 不下发）。
--    sys_role_menu 主键 (role_id, menu_id) 防重，IGNORE 保证幂等。
INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT r.id, @menu_id
FROM sys_role r
WHERE r.beid = @beid AND r.role_key IN ('admin', 'viewer') AND @menu_id > 0;

-- 6) 自检：应返回一行 component=dataease/DataeaseEmbed 的菜单；
--    查询为空说明 @beid 不对或父目录名不存在（按第 0/1 步修正后重跑）
SELECT m.id, m.parent_id, m.name, m.path, m.component, m.`sort`, m.beid
FROM sys_menu m
WHERE m.name = '数据可视化' AND m.beid = @beid;

-- ============================================================
-- 备选：顶层独立菜单（与「仪表盘」同级，parent_id=0；替换第 1/3 步使用）
-- ============================================================
-- INSERT INTO sys_menu (parent_id, name, path, component, `sort`, icon, menu_type, perm, visible, beid)
-- SELECT 0, '数据可视化', '/dataease', 'dataease/DataeaseEmbed', 11, 'DataBoard', 2, '', 1, @beid
-- FROM DUAL
-- WHERE NOT EXISTS (
--   SELECT 1 FROM sys_menu WHERE parent_id = 0 AND name = '数据可视化' AND beid = @beid
-- );
-- SET @menu_id = LAST_INSERT_ID();
-- INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
-- SELECT r.id, @menu_id FROM sys_role r
-- WHERE r.beid = @beid AND r.role_key IN ('admin', 'viewer') AND @menu_id > 0;
