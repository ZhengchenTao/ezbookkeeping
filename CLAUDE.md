# CLAUDE.md

本项目是 [mayswind/ezbookkeeping](https://github.com/mayswind/ezbookkeeping) 的个人 fork。

> **本文件**：仓库分支模型、上游同步流程、CI 故障排查 —— meta 层
> **[`FORK.md`](FORK.md)**：fork 相对上游的具体改动清单（feature 维度 + 进度状态）
> **个人笔记**：通用 fork 工作流决策框架在 `fork-工作流决策框架.md`（不入库）

本文件只记录**这个仓库的具体事实**，避免 Claude 会话误判。

---

## 仓库拓扑

```
github.com/mayswind/ezbookkeeping       （上游）
        │  Gitea pull mirror（后台异步）
        ▼
git.zhengchentao.win/mirror/ezbookkeeping   （只读镜像）
        │  CI workflow 拉过来
        ▼
git.zhengchentao.win/dev/ezbookkeeping     （origin，本地唯一 remote）
```

本地 `git remote -v` 只有 origin 一项，**没有手工配 upstream**。上游同步通过 custom 分支上的 workflow 在服务端完成，不是本地操作。

---

## 两个分支的职责（必须先理解，否则会改错地方）

| 分支 | 职责 | force push? |
|---|---|---|
| `main` | **锚定上游 release tag**（当前 v1.4.0）。被 `.gitea/workflows/sync-upstream.yml` `git reset --hard <tag>` 覆写。**别在 main 上做任何改动** | 是（由 CI 做） |
| `custom` | **所有个人改动 + workflow 文件都在这**：信用额度功能、UI 调整、个人需求清单、`.gitea/workflows/*.yml` 等。具体改动清单见 [`FORK.md`](FORK.md)。日常开发分支，**default branch** | 是（rebase 后人工做） |

⚠️ **default branch 是 `custom`**。`git clone` 默认 checkout custom，直接是开发分支。

### 历史：曾经存在过的 ci 分支（已退役）

2026-05-02 之前曾经有第三个分支 `ci`，最初设计是把 `.gitea/workflows/*.yml` 单独放它上面以"meta/code 分离"。两周后发现 Gitea Actions runs 列表显示的 commit 是 workflow 文件所在 commit（即 ci 的 HEAD），不是被构建的代码 commit，UX 误导性强。

把 workflow 挪回 custom 之后：
- runs 列表 commit = 真实代码 commit ✅
- `git clone` 默认落 custom 直接是开发分支 ✅
- rebase 上游时 workflow 跟 custom 一起平移 ✅
- 代价：失去"workflow 与代码完全独立"的设计美感 —— 这个分离原本就是过度设计

**ci 分支于 2026-05-02 删除**，仅保留这段说明给后续 Claude 会话理解 git log 里"workflow 文件迁到 custom"这条提交（commit `555ecc1a`）的来龙去脉。**workflow 改动直接在 custom 上做**。

---

## custom 分支 workflow 清单

`.gitea/workflows/` 当前有 3 个 workflow（2026-05-02 起精简，删了上游残留的 docker-snapshot/docker-release）：

| 文件 | 触发 | 干什么 | 状态 |
|---|---|---|---|
| `sync-upstream.yml` | 手动（`workflow_dispatch`，可填 tag） | 服务端把 `dev/main` 强制 reset 到 mirror 上的指定 release tag（默认最新），然后 `push --force-with-lease` + 推 tags | ✅ 在用 |
| `build-image.yml` | **自动**（push 到 custom 触发，`paths-ignore` 屏蔽 `**.md` / `.gitignore` / `LICENSE` / `screenshot/**`）+ 手动备选 | checkout 触发分支（push 时即 custom；手动时用 `inputs.branch` 默认 custom）→ 装 buildkit v0.13.2（钉版本）→ 登录 Gitea registry → 构建镜像（带 OCI 标签 source/revision，Gitea 自动关联包到 repo）→ push 到 `git.zhengchentao.win/dev/ezbookkeeping:<hash>` 与 `:latest`，`build-args: BUILD_PIPELINE=1` 跳过活 API 测试 | ✅ 在用，是日常发布通道 |
| `deploy.yml` | **自动**（`workflow_run` 在 build-image 成功后跑）+ 手动备选 | clone nas-infra → `docker compose pull && up -d` 重启 ezbookkeeping。脚本内联在 yml 里（早期版本走 `vars.CUSTOM_DEPLOY_SCRIPTS` 是过度设计，2026-05-02 移除）。私有 nas-infra 需要 `secrets.NAS_INFRA_TOKEN`，公开仓库不需要 | ✅ 自动 CD |

**已删**：`docker-snapshot.yml`（push main 自动触发，未配 secrets.DOCKER_REPO 永远失败）、`docker-release.yml`（push tag 同样问题）。需要时再从 git 历史 cherry-pick 回来。

---

## 同步发布流程（rebase 模型）

1. 上游出新 release（如 v1.4.0）→ Gitea pull mirror 自动把 tag 同步到 mirror
2. 人工触发 `Sync from upstream` workflow → 服务端把 dev/main reset 到该 tag
3. 本地 `git fetch && git checkout custom && git rebase origin/main`
4. 解冲突（如有）→ 验证 → `git push --force-with-lease origin custom`
5. **build-image workflow 自动触发**（force-push 也算 push 事件），构建新镜像；不需要手动点

日常 feature commit 流程（全自动 CD）：

1. 在 custom 上改代码 → commit → push
2. **自动触发 build**（除非只改了 `**.md` / `.gitengine` / `LICENSE` / `screenshot/**`）
3. build 成功 → **自动触发 deploy**（跑 repo Variables 里的 `CUSTOM_DEPLOY_SCRIPTS`）
4. 整条 push → build → deploy 链路无人工介入

如果想跳过 build/deploy（例如手动多次 push 调试），commit 时只改文档相关文件即可（落在 paths-ignore 范围内）。如果想强制重打某个旧 commit，去 Actions UI 手动触发 `Build Docker Image`，填要打包的 branch / tag。如果想只重新部署当前镜像（不重新 build），手动触发 `Deploy Docker Image` workflow。

**为什么 rebase 不 merge**：个人项目，无团队协作语义要保留，线性历史更清爽。

---

## 给后续 Claude 会话的明确提示

- 用户说"我的分支" / "切换到我的分支" → 指 `custom`
- 用户说"rebase main" → 指 `git rebase origin/main`，目标是把 custom 的改动叠到最新上游 tag 之上
- **不要在 `main` 分支上提交任何东西**（会被 CI 覆写）
- **workflow 文件改动直接在 custom 上做**（2026-05-02 起，不再是 ci 分支）
- force-push custom 是常规操作，但每次用 `--force-with-lease`，不直接 `--force`
- 如果发现本地配了 upstream remote，那是历史遗留，不要依赖；以 origin/main 为准
- `.claude/` 在 `.gitignore` 里（个人本地配置不入库），但 `CLAUDE.md` 本身入库

---

## 同步历史

- **2026-05-01**：rebase custom → origin/main (v1.4.0)。22 个 custom-only 提交（含一个旧的 `Merge branch 'main' into myrequirement` commit）压平为 21 个线性提交。已 force-push origin/custom（`08c69042` → `fe265259`）。
- **2026-05-02**：修 Gitea Actions `Build Docker Image` 工作流。三层故障，全部不在本仓库代码里：
  - **TLS 雷**：`docker login` 走 host 进程不命中 PREROUTING REDIRECT，且 v6 撞 DSM nginx 的 CF Origin Cert。NAS 侧修：iptables 补 OUTPUT 对称规则 + `/etc/hosts` 显式 v4 兜底。详见 obsidian vault [[NAS/notes/内网证书路径]] §三.5/§三.6
  - **buildkit 内核兼容**：runc 1.2+ 撞 DSM 4.4 内核。`.gitea/workflows/build-image.yml` 钉 `moby/buildkit:v0.13.2`（commit `acdbb5bf`）
  - **backend 单元测试撞活 API**：`pkg/exchangerates/` 的 `TestExchangeRatesApiLatestExchangeRateHandler_*` 跑活 API（加拿大银行 / 乌兹别克央行），国内访问超时。upstream Dockerfile 已设 `ARG BUILD_PIPELINE`，测试代码看到 `BUILD_PIPELINE=1 && CHECK_3RD_API!=1` 时早退。修：workflow 加 `build-args: BUILD_PIPELINE=1`（commit `2dd8f099`），对齐上游 GH Actions
- **2026-05-02 (后续)**：workflow 文件从 ci 分支迁到 custom，default branch 切到 custom（commit `555ecc1a`），随后**删掉 ci 分支**。原因：Gitea Actions runs 列表的 commit 字段一直显示 ci 的 workflow commit，不是被构建的代码 commit，UX 误导性强。挪到 custom 后列表直接显示真实代码 commit。同时清理上游残留的 `docker-release.yml` / `docker-snapshot.yml`（依赖未配的 `secrets.DOCKER_REPO`，永远失败）。仓库回到朴素的 main + custom 双分支模型
- **2026-05-02 (numpad fix)**：FORK.md #11 定位 + 修复。小键盘点击卡顿真因是 `.numpad-button` 的 `touch-action: none`（上游 e178a079 引入）与 F7 tap 处理叠加，改为 `touch-action: manipulation`（commit `75b4d78d`）

## 给后续 Claude 会话：CI 故障排查路径

如果 Gitea Actions build 又炸，按 NAS 域问题 vs 仓库代码问题分别排查：

| 现象 | 大概率位置 | 文档 |
|---|---|---|
| `Login to Gitea Container Registry` 步骤报 `x509: certificate signed by unknown authority` | NAS 网络层（iptables / dnsmasq / DSM nginx 占 443） | obsidian vault `NAS/notes/内网证书路径.md` + `NAS/notes/IPv6 设计.md` |
| `Build and push` 步骤里 `RUN ...` 在第二条之内就炸 `unsafe procfs detected` 之类 | buildkit/runc 与 DSM 内核版本 | `.gitea/workflows/build-image.yml` 的 `driver-opts` |
| `Failed to pass unit testing` / `Failed to pass lint checking`（build.sh 报） | **先看 Dockerfile 顶部 `ARG`**，多半是 CI 跳过开关没传（如 `BUILD_PIPELINE` / `CHECK_3RD_API` / `SKIP_TESTS`）。**不要先去改测试代码** | `Dockerfile` 顶部 ARG + `.gitea/workflows/build-image.yml` 的 `build-args` |
| `actions/checkout` 报 fetch 失败 | Gitea SSH/HTTPS 路径或 token 权限 | gitea-runner 的 `GITEA_RUNNER_REGISTRATION_TOKEN` + NPM `git.zhengchentao.win` 的 Advanced 配置 |
| Dockerfile 里某条指令业务逻辑报错 | 真正的代码问题 | 本仓库 `Dockerfile` |

**通用排查原则**：build.sh 报的"测试失败 / lint 失败"先看是不是上游已经设计了 CI 跳过路径。Dockerfile 的 `ARG` + `build.sh` 内的 `os.Getenv()` 检查通常成对出现（如 `BUILD_PIPELINE=1` → 跳过 3rd API 测试，`SKIP_TESTS=...` → 跳过指定测试名）。对齐上游 `.github/actions/` 下的传参，绝大多数情况能直接对齐。
