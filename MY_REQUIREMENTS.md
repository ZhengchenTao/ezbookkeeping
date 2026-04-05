# ezBookkeeping 个人需求清单

> 基于 fork 版本的定制开发需求，持续更新。
> 标注：⚠️ 中等难度 | ❌ 难/暂缓 | 🟢 已完成

---

## 一、UI / 主题

### 1. ⚠️ 自选主题色
**描述：** 在设置页面增加主题色选择器，用户可自由选择应用主色调。

**现状：** 主题色当前硬编码为 `#c67e48`，仅支持亮色/暗色切换。

**实现思路：**
- 在设置页新增颜色选择器组件
- 将所选颜色保存至 `localStorage`
- 启动时读取颜色并注入 CSS 变量（Framework7 支持动态 primary color）
- 涉及文件：`MobileApp.vue`、`src/lib/settings.ts`、设置页面

---

## 二、账户功能

### 2. 🟢 信用卡账户：额度与剩余额度
**描述：** 为信用卡类型账户新增「信用额度」字段，在账户列表显示「可用额度」。

**已完成：**
- 后端：`AccountExtend` JSON blob 新增 `CreditLimit` 字段（无需数据库迁移）
- API：`AccountCreateRequest` / `AccountModifyRequest` / `AccountInfoResponse` 增加 `creditLimit`
- 前端 model：`Account` 类增加 `creditLimit` 字段，同步 `toCreateRequest` / `toModifyRequest`
- 移动端 EditPage：CreditCard 分类时显示信用额度输入项
- 移动端 ListPage：账户列表显示「可用额度」（= `creditLimit + balance`，因信用卡余额为负值）
- 语言包：中英文均已添加 "Credit Limit" / "Available"

### 3. ⚠️ 账户交易列表顶部显示账户信息
**描述：** 在按账户筛选的交易列表页面顶部，显示账户名称、余额（或信用卡欠款+额度）等信息卡片。

**实现思路：**
- 检测当前是否为单账户筛选状态
- 顶部插入账户信息卡片组件
- 若多账户筛选则显示多张卡片
- 涉及文件：`src/views/mobile/transactions/ListPage.vue`

### 4. 🟢 允许直接修改账户余额（自动插入调整记录）
**描述：** 在账户编辑页直接修改余额字段，保存时自动计算差值并插入一条「余额调整」类型交易记录。

**已完成：**
- 后端：移除「账户已有交易时不允许添加 ModifyBalance」限制
- 后端：`Amount` 和 `RelatedAccountAmount` 均存储 delta（而非目标余额）；删除该交易可撤销 delta
- 前端 store：新增 `adjustAccountBalance({ accountId, targetBalance, currentBalance })` 函数
- 移动端 EditPage：余额字段对已有账户解除只读，保存时若余额变化先调用 `adjustAccountBalance`；捕获 `NothingWillBeUpdated (200004)` 视为成功
- 桌面端 EditDialog：同上逻辑，支持多子账户逐一调整

---

## 三、记账页面

### 5. 🟢 记账页显示当前账户余额 / 欠款+额度
**描述：** 在记账页面选择账户后，在账户名称下方显示该账户的当前余额（普通账户）或欠款金额（信用卡/负债）。

**已完成：**
- 信用卡账户显示「可用额度: ¥xxx」（= creditLimit + balance）
- 普通负债账户显示欠款正数，普通资产账户显示余额
- 转账时源账户和目标账户均显示
- 移动端：账户行 footer；桌面端：`custom-selection-secondary-text`
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`、`src/views/desktop/transactions/list/dialogs/EditDialog.vue`

### 6. ❓ 记录上次选择的账户
**描述：** 新建交易时，默认选中上次使用的账户，而非每次重置为默认账户。

**实现思路：**
- 保存账户 ID 到 `localStorage`
- 打开记账页时读取并预选
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`、`src/lib/settings.ts`

---

## 四、小键盘

### 7. 🟢 小键盘布局调整
**描述：** 调整数字键盘布局。

**已完成：**
```
[数值显示区              ] [ ⌫ ]   ← ⌫ 在显示区右侧，左边有竖线与右列对齐
  7     8     9      ×
  4     5     6      −
  1     2     3      +
  C     0     .     OK
```
- ⌫ 位于数值显示区右侧，宽度 20% 与运算符列对齐，单击退格，长按清除
- 左下角 C：单击清除全部
- . 在 OK 左侧
- × 保留
- 涉及文件：`src/components/mobile/NumberPadSheet.vue`

---

## 五、交易详情

### 8. 🟢 交易详情菜单增加「编辑」和「删除」
**描述：** 交易详情页三点菜单中增加编辑和删除操作。

**已完成：**
- 三点菜单展开后，第一项为「Edit」（跳转编辑页）
- 最后一项为红色「Delete」（确认后删除并返回）
- 从详情编辑返回后自动刷新数据
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`

---

## 六、分类选择

### 9. ❓ 分类选择默认全部展开
**描述：** 在记账页面选择分类时，默认将所有一级分类展开显示子分类，或在设置中可配置哪些分类默认展开。

**实现思路：**
- 修改 `TreeViewSelectionSheet.vue` 中展开状态的初始值
- 可选：在设置中增加「分类默认展开」开关
- 涉及文件：`src/components/mobile/TreeViewSelectionSheet.vue`

---

## 七、性能与动画

### 10. 🟢 全局动画加速
**描述：** 全局页面跳转及各类弹层动画加速。

**已完成：**
- 全局动画时长从 300ms 加快至 150ms（页面跳转、Sheet、ActionSheet、Popup、Dialog、Popover）
- Tab 切换动画保持原样（设置中已有开关可控制）
- 涉及文件：`src/styles/mobile/global.scss`

### 11. 🟢 点击响应卡顿优化
**描述：** 点击按钮有延迟感，点不上的问题。

**已完成：**
- `tapHoldDelay` 从默认 750ms 降至 500ms，减少长按判定等待
- 涉及文件：`src/MobileApp.vue`

---

## 八、离线 / 缓存

### 12. ❌ 本地优先 / 离线数据缓存（暂缓）
**描述：** 交易数据缓存至本地，优先展示缓存数据，后台静默拉取最新数据并更新。

**现状：** Service Worker 已实现静态资源缓存（图片、字体、地图瓦片），但**交易业务数据**目前不做本地缓存。

**为何难：**
- 需引入 IndexedDB 存储交易/账户/分类数据
- 需处理本地与服务端的数据同步、冲突解决
- 属于架构级改动，工作量较大

**建议：** 先完成其他需求，此项作为后期规划。

---

## 进度总览

| # | 需求 | 状态 |
|---|------|------|
| 1 | 自选主题色 | ⚠️ 待做 |
| 2 | 信用卡额度 | 🟢 已完成 |
| 3 | 账户信息卡片 | ⚠️ 待做 |
| 4 | 调整余额入口 | 🟢 已完成 |
| 5 | 记账页显示余额 | 🟢 已完成 |
| 6 | 记住上次账户 | ❓ 待定 |
| 7 | 小键盘布局 | 🟢 已完成 |
| 8 | 详情编辑/删除 | 🟢 已完成 |
| 9 | 分类默认展开 | ❓ 待定 |
| 10 | 全局动画加速 | 🟢 已完成 |
| 11 | 点击卡顿优化 | 🟢 已完成 |
| 12 | 离线缓存 | ❌ 暂缓 |
