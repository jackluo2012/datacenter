# 前端控制台对标调研（STANDARD-BENCHMARK）

- 调研日期：2026-10-11
- 对标对象：RuoYi-Vue3（后台管理事实标准模板）、DataEase v1/v2（BI 可视化）、Datavines（数据可观测性）
- 被评对象：本仓库 `web/`（Vue3 + Vite + TS + Element Plus + Pinia + ECharts）及其 `sys_menu` 表
- 说明：`web/` 正处于 RBAC 升级中途（动态路由/面包屑/TagsView 已落地未提交），本文按工作区当前状态评估。
  约定：【事实】=来源可查；【建议】=本仓库改进项。

---

## 1. RuoYi-Vue3：Layout、权限实现与菜单表

### 1.1 Layout 组件清单

【事实】`src/layout/index.vue` 组合以下组件（来源：<https://github.com/yangzongzhuan/RuoYi-Vue3> 同路径文件）：

| 组件 | 职责 | 关键点 |
|---|---|---|
| `Sidebar/index.vue` | 侧边栏 | 递归菜单 SidebarItem、Logo、Link 外链 |
| `Navbar.vue` | 顶栏 | Hamburger、Breadcrumb（navType=1 时）/TopNav（=2）/TopBar（=3）、HeaderSearch、Screenfull、暗色开关、SizeSelect、HeaderNotice、头像下拉（个人中心/布局设置/锁定屏幕/退出） |
| `TagsView/index.vue` | 标签页 | 右键菜单：刷新页面/关闭当前/关闭其他/关闭左侧/关闭右侧/全部关闭；affix 固定标签；中键关闭 |
| `AppMain.vue` | 主内容 | router-view + keep-alive + transition |
| `Settings/index.vue` | 设置抽屉 | 主题色选择器（预置色板）、侧栏深浅主题、tagsView 开关/持久化/图标/风格、fixedHeader、sidebarLogo、dynamicTitle |

【事实】新版还内置锁屏（`store/modules/lock.js`，beforeEach 拦截跳 `/lock`）与初始/过期密码提醒（`user.js` getInfo 后 `isDefaultModifyPwd`/`isPasswordExpired` 弹窗引导去 Profile 修改密码）。

### 1.2 权限实现

【事实】三段式（来源：`src/permission.js`、`src/store/modules/{user,permission}.js`、`src/directive/permission/hasPermi.js`）：

1. 登录存 token → 路由守卫发现 `roles.length === 0` 时调 **getInfo**（注入 roles/permissions/avatar），再调 **getRouters** 拉后端菜单路由；
2. `filterAsyncRouter` 将 component 字符串经 `import.meta.glob` 的 **loadView** 映射为组件，`router.addRoute(route)` 逐条挂载（动态路由）；
3. 按钮级：**v-hasPermi** 指令，permissions 含 `*:*:*` 通配或命中任一值则保留元素，否则 `removeChild` 移除。

【事实】我们已对齐同构实现：`web/src/api/sys.ts`（getInfo/getRouters）、`web/src/stores/permission.ts`（glob 解析 + addRoute + 静态菜单降级）、`web/src/directive/perm.ts`（admin 角色与 `*:*:*` 通配）、路由守卫 `web/src/router/index.ts`（fetchInfo → generateRoutes）。

### 1.3 sys_menu 表逐字段对比

【事实】RuoYi-Vue `sys_menu` DDL（来源：<https://github.com/yangzongzhuan/RuoYi-Vue/blob/master/sql/ry_20260417.sql>）：
`menu_id, menu_name, parent_id, order_num, path, component, query, route_name, is_frame, is_cache, menu_type(M目录/C菜单/F按钮), visible, status, perms, icon, create_by, create_time, update_by, update_time, remark`

【事实】我们 `docs/schema/sys.sql` 的 `sys_menu`：`id, parent_id, name, path, component, sort, icon, menu_type(1目录/2菜单/3按钮), perm, visible, beid`（beid 为我们多租户独有，标准无）。

| RuoYi 字段 | 我们 | 缺失影响 |
|---|---|---|
| menu_id/menu_name/parent_id/order_num/path/component/menu_type/visible/perms/icon | id/name/sort/... 同义已覆盖 | 无 |
| **status** | 缺 | 无法"停用"菜单：visible 只控制显示，停用态应同时让路由拒绝访问（如公告位临时下线） |
| **is_cache** | 缺 | keep-alive 粒度失控：RuoYi 以 is_cache=0 标记 noCache；我们目前靠路由 name 猜缓存，且 List.vue 组件 name 未声明，缓存实际未生效（`MainLayout.vue` 内 TODO） |
| **is_frame** | 缺 | 菜单无法挂外链/iframe（如跳 Grafana、文档站） |
| **query** | 缺 | 菜单不能携带默认路由参数（如打开即带 ?tab=result） |
| **route_name** | 缺 | 我们以菜单 name 兼任路由 name，中文名/重名会破坏 keep-alive 的 name 匹配 |
| remark / create_by / create_time / update_by / update_time | 缺 | 菜单维护无审计与备注，排查权限变更困难 |

【建议】P0 补 `status`、`is_cache` 两列（直接影响权限正确性与缓存正确性）；P1 补 `route_name`（keep-alive 依赖）；P2 补 `is_frame/query/审计字段`。

### 1.4 布局差距速览

【事实】我们 `web/src/layouts/MainLayout.vue`：暗色侧栏 #001529 + 顶栏（折叠钮/面包屑/全屏/头像下拉）+ TagsView + keep-alive 主区，支持动态菜单递归渲染（`SidebarItem.vue`）。
【建议】对标 Navbar 仍缺：全局搜索、暗色/主题开关、消息通知、尺寸切换、"布局设置"抽屉、个人中心入口；TagsView 缺"刷新当前页/关闭左侧/关闭右侧/标签持久化"。

---

## 2. DataEase v1/v2 前端：视觉与信息架构

【事实】v2 前端为 Vue3 + TS + Element Plus（仓库 `core/core-frontend`，来源：<https://github.com/dataease/dataease> dev-v2 分支）。布局为**顶部导航 + 侧栏 + 主区**：`layout/components` 含 `Header.vue`、`HeaderMenuItem/HeaderSystem`（顶栏模块菜单）、`Sidebar.vue`、`Menu/MenuItem`、`CollapseBar.vue`（侧栏折叠）、`AccountOperator.vue`（账号）、`LangSelector.vue`（语言）、`AiComponent.vue`（AI 助手常驻入口）。

【事实】views 按模块划分：`workbranch`（工作台）、`dashboard`（仪表板）、`data-visualization`（数据大屏）、`panel`、`visualized`、`template`/`template-market`、`toolbox`、`system`、`share`、`home`、`mobile`、`sqlbot`（AI 问数）——v2 在 v1"仪表板"之上新增工作台与数据大屏（v2.0 发布报道，来源：<https://www.sohu.com> 搜索"DataEase v2.0 发布"；官方文档 <https://dataease.cn>、《DataEase v2.10 文档》 <https://www.bookstack.cn/read/dataease-2>）。

【事实】数据组织链路固定为：**数据源 → 数据集 → 仪表板**（视图为最小可复用单元，来源：<https://dataease.cn> 快速入门；知乎《开源软件 DataEase 的体验》）；图表底层为 ECharts + AntV 双库（来源：<https://dataease.cn> 视图样式设计文档）。

【建议】给 `web/` 的视觉与信息架构改进（3-5 条）：

1. **一级菜单分组**：我们 14 个菜单平铺过长；参照 DataEase 按域分组（如 数据集成：数据源/元数据/调度｜治理：标准/质量/主数据｜服务：数据服务/资产目录/标签/指标｜智能：AI 问数/SQL IDE｜系统管理），依赖已落地的目录型动态菜单即可配置化。
2. **两栏资源页布局**：数据源/资产目录/主题看板等页面采用 DataEase 式"左侧列表（文件夹树 + 搜索框）+ 右侧卡片/详情"，替代整页表格。
3. **列表卡片/表格双视图**：工作台与模板类页面提供卡片网格视图（带封面、 Owner、更新时间），表格视图承载管理操作。
4. **AI 常驻入口**：AI 问数目前是独立页；参照 `AiComponent` 在顶栏加全局抽屉入口，任意页面可唤起。
5. **工作台首页**：以"常用功能 + 最近访问 + 我负责的资产"组织 Dashboard，而非单一统计卡。

---

## 3. Datavines：quality 页面改进

【事实】Datavines 是一站式开源数据可观测性平台：元数据管理、数据概览报告、数据质量管理、数据分布查询、数据趋势洞察；内置 27 个质量检查规则、4 种规则类型；Web 页面可配置/运行检查作业、查看执行日志、**查看错误数据**与检查结果；另有独立"数据目录"（定时采集元数据、监听变更、标签管理）与"告警（warning）"模块（来源：<https://github.com/datavane/datavines> README.zh-CN.md）。

【事实】其 UI 主导航为：数据源（home）、告警、错误数据管理、用户管理、标签、配置、令牌管理；**数据源详情页固定四页签**：数据概览（dashboard）→ 元数据/目录（editor）→ 作业（jobs）→ 作业执行日志（jobExecutionLogs）（来源：同仓库 `datavines-ui/src/router/mainRouter.tsx`、`detailRouter.tsx`）。

【建议】对我们 `web/src/views/quality/List.vue`（现为 规则/任务/结果 三页签平铺表格）的改进（2-3 条）：

1. **按"数据源/库表"为入口重组信息架构**：将三页签改为"数据源详情（概览/作业/执行日志）+ 全局错误数据页"，与 Datavines 心智一致；至少在结果页支持按表聚合（表 → 规则 → 最近趋势）。
2. **结果页加监控视图**：顶部加健康分/通过率趋势卡（ECharts 已在依赖中，当前仅 Dashboard.vue 使用）；失败规则 TopN 一眼可见，替代纯明细表。
3. **错误数据下钻与告警**：失败结果支持查看样例错误行（Datavines 核心卖点"查看错误数据"），并为任务挂接告警通知（独立告警配置页）。

---

## 4. 通用后台体验 checklist（对标 RuoYi-Vue3）

| 能力 | 标准做法（RuoYi-Vue3） | 我们现状 | 结论 |
|---|---|---|---|
| 面包屑 | Navbar 内 Breadcrumb，菜单树层级 | `components/Breadcrumb.vue`，菜单树优先/matched 兜底 | 已有 |
| TagsView | 右键 6 项操作 + affix + 持久化 | 右键 关闭当前/其他/全部 + affix；无刷新当前、关闭左/右侧、持久化 | 部分有 |
| 全屏 | Screenfull（@vueuse useFullscreen） | 手写 requestFullscreen（MainLayout） | 已有 |
| 主题切换/暗色 | 暗色开关 + 主题色选择器 + 设置抽屉 | 无 | 缺 |
| 全局搜索快捷键 | HeaderSearch + fuse.js 模糊 + ↑↓/Enter | 无 | 缺 |
| 修改密码页 | Profile 页 resetPwd 页签 + 初始/过期密码提醒 | 后端 sys-rpc 已有 ResetPwd，前端无页面 | 缺 |
| 个人中心 | /user/profile 基本资料+头像 | 头像下拉仅"退出登录" | 缺 |
| 锁屏 | lock store + /lock 页 | 无 | 缺 |
| 消息通知 | HeaderNotice 铃铛 | 无 | 缺 |
| 组件尺寸切换 | SizeSelect（large/default/small） | 无 | 缺 |
| 国际化 | LangSelector（DataEase 亦内置） | 无 | 缺 |
| 字体图标方案 | @element-plus/icons-vue + vite-plugin-svg-icons 自定义 sprite | @element-plus/icons-vue 全量全局注册；无自定义 svg sprite 通道 | 部分有 |
| 动态标题 | dynamicTitle 设置 + route meta | document.title 跟随路由（固定前缀"数据中台"） | 已有（简化） |
| 404/降级 | 兜底路由 | 404 + getRouters 失败降级静态菜单（超出标准） | 已有 |

---

## 5. 我们 vs 标准 差距表

| 功能 | 已实现 | 建议（对标来源） | 优先级 |
|---|---|---|---|
| getInfo/getRouters + 动态路由 + 静态降级 | 是（api/sys.ts、stores/permission.ts） | 对齐 RuoYi，无需动作 | — |
| v-perm 按钮权限 | 是 | 无需动作；页面按钮逐步补 perm 标识 | P1 |
| keep-alive 页面缓存 | 形式有、实际未生效 | views 页面补 `defineOptions({name: 路由name})`；sys_menu 加 is_cache | P0 |
| sys_menu.status 停用态 | 否 | DDL + getRouters 过滤 | P0 |
| 用户/角色/菜单管理页面 | api 已预留，无 views | 补 `system/user|role|menu` 三页（RBAC 闭环） | P0 |
| 个人中心 + 修改密码 | 否 | Profile 页（资料+resetPwd），对接 ResetPwd；加初始/过期密码提醒 | P0 |
| 一级菜单分组（目录） | 机制已支持 | 菜单种子按域分组配置 | P1 |
| 主题切换/暗色模式 | 否 | Element Plus dark CSS + 设置抽屉（对标 RuoYi Settings） | P1 |
| TagsView 增强（刷新/关闭左侧/右侧/持久化） | 部分 | 补右键"刷新当前页/关闭左侧/右侧"，localStorage 持久化 | P1 |
| 全局搜索快捷键 | 否 | HeaderSearch + fuse.js（package.json 可直接加依赖） | P1 |
| quality 监控视图（趋势/健康分/错误数据下钻/告警） | 否 | 对标 Datavines（§3 建议） | P1 |
| 数据源/目录类页面两栏布局与卡片视图 | 否 | 对标 DataEase（§2 建议 2/3） | P2 |
| sys_menu.route_name / is_frame / query / 审计字段 | 否 | 按需追加 | P2 |
| 布局设置抽屉 / 锁屏 / 消息通知 / 尺寸切换 / 国际化 | 否 | RuoYi 均有；按需引入 | P2 |
| AI 全局抽屉入口 | 否 | 对标 DataEase AiComponent | P2 |

> 结论：权限链路（getInfo/getRouters/动态路由/v-perm）已与 RuoYi 同构；当前短板集中在 **RBAC 管理页闭环、菜单表语义字段（status/is_cache/route_name）、以及主题/搜索/个人中心等通用体验件**；quality 页对标 Datavines 需从"表格平铺"升级为"监控视图 + 错误数据下钻"。
