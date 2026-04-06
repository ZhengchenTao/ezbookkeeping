# ezBookkeeping 个人需求清单

> 基于 fork 版本的定制开发需求，持续更新。
> 标注：⚠️ 中等难度 | ❌ 难/暂缓 | ❓ 待定 | 🔍 调查中 | 🟢 已完成

---

## 一、UI / 主题

### 1. ❌ 自选主题色（已实现后 revert）
**描述：** 在设置页面增加主题色选择器，用户可自由选择应用主色调。

**现状：** 实现后效果不佳，已全部 revert。主题色仍为硬编码 `#c67e48`。

**实现思路（留存备查）：**
- 存 localStorage + 云同步，启动时注入 `--ebk-theme-color` CSS 变量
- 移动端 f7params / 桌面端 Vuetify theme 从设置读取
- 预设颜色复用账户 14 色调板

---

## 二、账户功能

### 2. 🟢 信用卡账户：额度与可用额度
**描述：** 为信用卡类型账户新增「信用额度」字段，在账户列表显示可用额度。

**已完成：**
- 后端：`AccountExtend` JSON blob 新增 `CreditLimit` 字段（无需数据库迁移）
- API：`AccountCreateRequest` / `AccountModifyRequest` / `AccountInfoResponse` 增加 `creditLimit`
- 前端 model：`Account` 类增加 `creditLimit` 字段，同步序列化/反序列化
- 移动端 EditPage：CreditCard 分类时显示信用额度输入项
- 移动端 ListPage：账户名下方显示「可用额度: ¥xxx」（= `creditLimit + balance`）
- 桌面端 ListPage：账户卡片余额旁显示「Available: ¥xxx」
- 语言包：中英文均已添加 `"Credit Limit"` / `"Available"`

### 3. 🟢 按账户筛选交易时顶部显示账户信息卡
**描述：** 在按单个账户筛选的交易列表顶部，显示账户图标、名称和余额/可用额度。

**已完成：**
- 仅单账户筛选时（`queryAllFilterAccountIdsCount === 1`）显示，多账户/全量时隐藏
- 信用卡账户显示「欠款 · 可用 ¥xxx」，普通账户显示余额
- 移动端：toolbar 下方插入账户信息卡片
- 桌面端：日期范围行下方插入 tonal 样式账户卡片
- 多子账户（`MultiSubAccounts`）一级账户：改用 `getAccountSubAccountBalance` 获取汇总余额及正确币种，修复 `currency = '---'` 导致货币符号不显示的问题
- 涉及文件：`src/views/mobile/transactions/ListPage.vue`、`src/views/desktop/transactions/ListPage.vue`

### 4. 🟢 账户编辑页直接修改余额（自动插入调整记录）
**描述：** 在账户编辑页修改余额字段，保存时自动计算差值并插入一条「余额调整」类型交易。

**已完成：**
- 后端：移除「账户已有交易时不允许添加 ModifyBalance」的限制
- 后端：`Amount` 与 `RelatedAccountAmount` 均存储 delta；创建时 `balance += delta`，删除时 `balance -= delta`，修改时 `balance = balance - oldDelta + newDelta`
- 后端响应：`ToTransactionInfoResponse` 对 ModifyBalance 类型返回 `RelatedAccountAmount` 作为 `sourceAmount`
- 前端 store：新增 `adjustAccountBalance({ accountId, targetBalance, currentBalance })` 函数，发送 `sourceAmount = delta`
- 移动端 EditPage：余额字段对已有账户解除只读；保存时若余额变化先调 `adjustAccountBalance`；捕获 `NothingWillBeUpdated (200004)` 视为成功
- 桌面端 EditDialog：同上逻辑，支持多子账户逐一调整
- 涉及文件：`pkg/services/transactions.go`、`pkg/models/transaction.go`、`src/stores/transaction.ts`、`src/views/mobile/accounts/EditPage.vue`、`src/views/desktop/accounts/list/dialogs/EditDialog.vue`

---

## 三、记账页面

### 5. 🟢 记账页选择账户后显示余额/可用额度
**描述：** 在记账页面选择账户后，在账户行显示该账户的当前余额或信用卡可用额度。

**已完成：**
- 信用卡账户（有 `creditLimit`）：显示「欠款金额 · 可用 ¥xxx」
- 普通负债账户：显示欠款正数；普通资产账户：显示余额
- 转账类型时源账户和目标账户均显示
- 移动端：账户列表项 `footer` 字段
- 桌面端：`two-column-select` 的 `custom-selection-secondary-text`
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`、`src/views/desktop/transactions/list/dialogs/EditDialog.vue`

### 6. ❓ 记录上次选择的账户（待定）
**描述：** 新建交易时，默认选中上次使用的账户。

**实现思路：**
- 保存账户 ID 到 `localStorage`，打开记账页时读取并预选
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`、`src/lib/settings.ts`

---

## 四、小键盘

### 7. 🟢 小键盘布局调整
**描述：** 调整数字键盘布局。

**已完成：**
```
[数值显示区              ] [ ⌫ ]
  7     8     9      ×
  4     5     6      −
  1     2     3      +
  C     0     .     OK
```
- ⌫ 单击退格，长按清除；C 清除全部
- 涉及文件：`src/components/mobile/NumberPadSheet.vue`

---

## 五、交易详情

### 8. 🟢 交易详情页增加「编辑」和「删除」入口
**描述：** 交易详情页三点菜单中增加编辑和删除操作。

**已完成：**
- 三点菜单：第一项「Edit」跳转编辑页，最后一项红色「Delete」确认后删除并返回
- 从详情返回编辑页后自动刷新数据
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`

---

## 六、交易时间选择

### 9. 🟢 点击交易时间默认打开日期选择
**描述：** 在记账/编辑页面点击「Transaction Time」标题行时，默认弹出日期选择器而非时间选择器。

**已完成：**
- 点击标题行（`transaction-edit-datetime-header`）改为以 `'date'` 模式打开
- 点击日期部分 → 日期选择器；点击时间部分 → 时间选择器（保持不变）
- 仅移动端有此交互（PC 端使用统一的 `date-time-select` 组件，无需修改）
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`

---

## 七、分类选择

### 10. 🟢 分类选择默认全部展开（仅移动端）
**描述：** 在移动端记账页选择分类时，默认展开所有一级分类。PC 端使用不同的分类选择组件，无需此设置。

**已完成：**
- 设置存 localStorage（字段 `expandCategoryTreeByDefault`，默认 `false`），即时生效无需 reload
- `TreeViewSelectionSheet` 新增可选 prop `defaultExpanded`，`:opened` 改为 `props.defaultExpanded || isPrimaryItemHasSecondaryValue(item)`
- 移动端 EditPage 三个分类 sheet（支出/收入/转账）均传入 `:default-expanded`
- 设置仅在移动端设置页显示，PC 端无对应入口
- 涉及文件：`src/components/mobile/TreeViewSelectionSheet.vue`、`src/views/mobile/transactions/EditPage.vue`、`src/views/mobile/SettingsPage.vue`

---

## 八、性能与动画

### 11. 🟢 全局动画加速
**描述：** 全局页面跳转及各类弹层动画加速。

**已完成：**
- 页面跳转、Sheet、ActionSheet、Popup、Dialog、Popover 动画时长从 300ms → 150ms
- Tab 切换动画保持原样（设置中已有开关可控制）
- 涉及文件：`src/styles/mobile/global.scss`

### 12. 🔍 点击响应卡顿（暂调查）
**描述：** 点击按钮有延迟感，点不上的问题。

**初步判断：** 可能是接口响应慢导致，非前端交互延迟。等 #12 离线缓存完成后再评估。

---

## 九、离线 / 缓存

### 13. ❌ 本地优先 / 离线数据缓存（暂缓）
**描述：** 交易数据本地缓存，优先展示缓存数据，后台静默拉取更新。

**现状：** Service Worker 已实现静态资源缓存，但交易业务数据目前不做本地缓存。

**为何难：**
- 需引入 IndexedDB 存储交易/账户/分类数据
- 需处理本地与服务端的数据同步、冲突解决
- 属于架构级改动，工作量较大

---

## 进度总览

| # | 需求 | 状态 |
|---|------|------|
| 1 | 自选主题色 | ❌ 已 revert |
| 2 | 信用卡额度 | 🟢 已完成 |
| 3 | 账户信息卡片 | 🟢 已完成 |
| 4 | 调整余额入口 | 🟢 已完成 |
| 5 | 记账页显示余额 | 🟢 已完成 |
| 6 | 记住上次账户 | ❓ 待定 |
| 7 | 小键盘布局 | 🟢 已完成 |
| 8 | 详情编辑/删除 | 🟢 已完成 |
| 9 | 点击时间默认日期选择器 | 🟢 已完成 |
| 10 | 分类默认展开 | 🟢 已完成 |
| 11 | 全局动画加速 | 🟢 已完成 |
| 12 | 点击卡顿优化 | 🔍 暂调查 |
| 13 | 离线缓存 | ❌ 暂缓 |
