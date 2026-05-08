# ezBookkeeping 个人 fork 改动清单

> 本文件记录这个 fork 相对上游 [mayswind/ezbookkeeping](https://github.com/mayswind/ezbookkeeping) 的所有定制改动 + 进度状态。

> 关联文档：
> - [`CLAUDE.md`](CLAUDE.md) —— 仓库分支模型 / 上游同步流程 / CI 排查路径（meta 层）
> - 部署：见自家 NAS infra repo `git.zhengchentao.win/dev/nas-infra` 的 README（compose-level）
>
> 标注：❌ 难/暂缓 | ❓ 待定 | 🔍 调查中 | 🟢 已完成

---

## 一、账户功能

### 1. 🟢 信用卡账户：额度与可用额度
**描述：** 为信用卡类型账户新增「信用额度」字段，在账户列表显示可用额度。

**已完成：**
- 后端：`AccountExtend` JSON blob 新增 `CreditLimit` 字段（无需数据库迁移）
- API：`AccountCreateRequest` / `AccountModifyRequest` / `AccountInfoResponse` 增加 `creditLimit`
- 前端 model：`Account` 类增加 `creditLimit` 字段，同步序列化/反序列化
- 移动端 EditPage：CreditCard 分类时显示信用额度输入项（数字键盘）
- 桌面端 EditDialog：CreditCard 分类时显示信用额度输入框（`amount-input`）
- 移动端 ListPage：账户名下方显示「可用额度: ¥xxx」（= `creditLimit + balance`）
- 桌面端 ListPage：账户卡片余额旁显示「Available: ¥xxx」
- 语言包：中英繁均已添加 `"Credit Limit"` / `"Available"`

### 2. 🟢 按账户筛选交易时顶部显示账户信息卡
**描述：** 在按单个账户筛选的交易列表顶部，显示账户图标、名称和余额/可用额度。

**已完成：**
- 仅单账户筛选时（`queryAllFilterAccountIdsCount === 1`）显示，多账户/全量时隐藏
- 信用卡账户显示「欠款 · 可用 ¥xxx」，普通账户显示余额
- 移动端：toolbar 下方插入账户信息卡片；桌面端：日期范围行下方插入 tonal 样式账户卡片
- 多子账户（`MultiSubAccounts`）一级账户：使用 `getAccountSubAccountBalance` 获取汇总余额及正确币种，修复 `currency = '---'` 导致货币符号不显示的问题
- 涉及文件：`src/views/mobile/transactions/ListPage.vue`、`src/views/desktop/transactions/ListPage.vue`

### 3. 🟢 账户编辑页直接修改余额（自动插入调整记录）
**描述：** 在账户编辑页修改余额字段，保存时自动计算差值并插入一条「余额调整」类型交易。

**已完成：**
- 后端：移除「账户已有交易时不允许添加 ModifyBalance」的限制
- 后端：`Amount` 与 `RelatedAccountAmount` 均存储 delta；创建时 `balance += delta`，删除时 `balance -= delta`，修改时 `balance = balance - oldDelta + newDelta`
- 后端响应：`ToTransactionInfoResponse` 对 ModifyBalance 类型返回 `RelatedAccountAmount` 作为 `sourceAmount`
- 前端 store：新增 `adjustAccountBalance({ accountId, targetBalance, currentBalance })` 函数，发送 `sourceAmount = delta`
- 移动端 EditPage：余额字段对已有账户解除只读；保存时若余额变化先调 `adjustAccountBalance`；捕获 `NothingWillBeUpdated (200004)` 视为成功
- 桌面端 EditDialog：同上逻辑，支持多子账户逐一调整
- 「调整余额」入口仅在账户编辑页，已从账户列表「更多」菜单移除
- 涉及文件：`pkg/services/transactions.go`、`pkg/models/transaction.go`、`src/stores/transaction.ts`、`src/views/mobile/accounts/EditPage.vue`、`src/views/desktop/accounts/list/dialogs/EditDialog.vue`

---

## 二、记账页面

### 4. 🟢 记账页选择账户后显示余额/可用额度
**描述：** 在记账页面选择账户后，在账户行显示该账户的当前余额或信用卡可用额度。

**已完成：**
- 信用卡账户（有 `creditLimit`）：显示「欠款金额 · 可用 ¥xxx」
- 普通负债账户：显示欠款正数；普通资产账户：显示余额
- 转账类型时源账户和目标账户均显示
- 移动端：账户列表项 `footer` 字段；桌面端：`two-column-select` 的 `custom-selection-secondary-text`
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`、`src/views/desktop/transactions/list/dialogs/EditDialog.vue`

### 5. ❓ 记录上次选择的账户（待定）
**描述：** 新建交易时，默认选中上次使用的账户。

**实现思路：**
- 保存账户 ID 到 `localStorage`，打开记账页时读取并预选
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`、`src/lib/settings.ts`

---

## 三、小键盘

### 6. 🟢 小键盘布局调整
**描述：** 调整数字键盘布局。

**已完成：**
```
[数值显示区              ] [ ⌫ ]
  7     8     9      ×
  4     5     6      −
  1     2     3      +
  C     0     .     OK
```
- ⌫ 单击退格；按住不放先删一位、约 500ms 后清空全部（长按响应细节见 #11 第二阶段）
- C 一键清除全部
- 涉及文件：`src/components/mobile/NumberPadSheet.vue`

---

## 四、交易详情

### 7. 🟢 交易详情页增加「编辑」和「删除」入口（仅移动端）
**描述：** 移动端交易详情页三点菜单中增加编辑和删除操作。PC 端详情直接在编辑弹窗中展示，无需额外入口。

**已完成：**
- 三点菜单：第一项「Edit」跳转编辑页，最后一项红色「Delete」确认后删除并返回
- 从详情返回编辑页后自动刷新数据
- 涉及文件：`src/views/mobile/transactions/EditPage.vue`

---

## 五、交易时间选择

### 8. ❌ 点击交易时间标题默认打开日期选择（已回滚）
**描述：** 原想让点击「Transaction Time」标题行时默认弹日期选择器。

**为何回滚：** 改动改的是 `template #header` 那行 label 的点击 handler（`'time'` → `'date'`），实际操作中用户点的是 `template #title` 里的日期/时间文本。上游早在 commit `368322f9` 已实现"点哪走哪"的智能路由——点日期开日期选择器、点时间开时间选择器。所以这条改动**用户视角无可见差异**，纯空改，回滚到上游行为。

**留档教训：** 改 UI 行为前先把"用户实际点哪个元素"摸清楚，别只看着 DOM 结构想当然。`#header` slot 只是上方的 label 行，正常用户极少触发。

---

## 六、分类选择

### 9. 🟢 分类选择默认全部展开（仅移动端）
**描述：** 在移动端记账页选择分类时，默认展开所有一级分类。PC 端使用不同的分类选择组件，无需此设置。

**已完成：**
- 设置存 localStorage（字段 `expandCategoryTreeByDefault`，默认 `false`），即时生效无需 reload；已加入云同步白名单（`ALL_ALLOWED_CLOUD_SYNC_APP_SETTING_KEY_TYPES`），可跨设备同步
- `TreeViewSelectionSheet` 新增可选 prop `defaultExpanded`，`:opened` 改为 `props.defaultExpanded || isPrimaryItemHasSecondaryValue(item)`
- 移动端 EditPage 三个分类 sheet（支出/收入/转账）均传入 `:default-expanded`
- 设置仅在移动端设置页显示，PC 端无对应入口
- 涉及文件：`src/components/mobile/TreeViewSelectionSheet.vue`、`src/views/mobile/transactions/EditPage.vue`、`src/views/mobile/SettingsPage.vue`

---

## 七、性能与动画

### 10. 🟢 全局动画加速（仅移动端）
**描述：** 移动端全局页面跳转及各类弹层动画加速。

**已完成：**
- 页面跳转、Sheet、ActionSheet、Popup、Dialog、Popover 动画时长从 300ms → 150ms
- Tab 切换动画保持原样（设置中已有开关可控制）
- 涉及文件：`src/styles/mobile/global.scss`

### 11. 🟢 小键盘点击卡顿（三次修正）
**描述：** 移动端小键盘点击有延迟感。

**第一阶段（2026-05-02）`touch-action: none` 引发的 300ms 双击延迟：** 上游在 `.numpad-button` 上设了 `touch-action: none`（commit `e178a079` "code refactor" by MaysWind），与浏览器双击缩放检测叠加后保留了老式 300ms 点击延迟。
- 修复：`.numpad-button` 的 `touch-action: none` 改为 `touch-action: manipulation`（W3C 标准"快速点击"值，禁双击缩放）

**第二阶段（2026-05-08）退格键 `@taphold` 等待 750ms：** backspace 单点仍可感知延迟。根因是 `@click` + `@taphold` 让 F7 必须等 ~750ms 判别 tap vs hold，期间 click 被抑制。
- 修复：弃用 `@click="backspace" @taphold="clear()"`，改为原生 `pointerdown`/`pointerup`/`pointercancel`/`pointerleave` + 自管定时器
- 行为：单击立即删一位；按住不放先删一位、约 500ms 后清空全部

**第三阶段（2026-05-08）所有数字/运算键也延迟：** 第一阶段修完后用户反馈数字键仍有"等一拍"感。怀疑 F7 整套 tap 处理（含 active-state 检测、`fastClicks` 兼容代码、tap-hold 全局监听）即便不显式声明 `@taphold` 也会给 `@click` 加上判别期。
- 修复：把所有按键（数字 0-9、运算 ×−+、C 清空、小数点/双零、OK 确认）的 `@click` 全部换成 `@pointerdown.left`
- 原理：`pointerdown` 在按下瞬间触发，绕开 F7 的 tap 合成路径。`.left` 修饰符限制只响应主键（触屏 button=0 始终满足，桌面右键不会误触发）
- F7 的 `.active-state` 视觉反馈基于独立的 touchstart/touchend 监听，不依赖 `@click`，按下视觉效果保留

涉及文件：`src/components/mobile/NumberPadSheet.vue`

**附带认知：** 原 #11 假设是"全局点击响应慢"或"接口慢"，与 #12 离线缓存挂钩调研。实际诊断后跟那两条都无关——纯 F7 框架 tap 合成 + 双击缩放 + taphold 检测三者叠加。最终通过完全弃用 `@click` 改 pointer 事件解决。该认知值得记录避免后续误诊路径。

---

## 八、离线 / 缓存

### 12. ❌ 本地优先 / 离线数据缓存（暂缓）
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
| 1 | 信用卡额度 | 🟢 已完成 |
| 2 | 账户信息卡片 | 🟢 已完成 |
| 3 | 调整余额入口 | 🟢 已完成 |
| 4 | 记账页显示余额 | 🟢 已完成 |
| 5 | 记住上次账户 | ❓ 待定 |
| 6 | 小键盘布局 | 🟢 已完成 |
| 7 | 详情编辑/删除 | 🟢 已完成 |
| 8 | 点击时间默认日期 | ❌ 已回滚（无效改动） |
| 9 | 分类默认展开 | 🟢 已完成 |
| 10 | 全局动画加速 | 🟢 已完成 |
| 11 | 小键盘点击卡顿（touch-action 修复） | 🟢 已完成 |
| 12 | 离线缓存 | ❌ 暂缓 |
