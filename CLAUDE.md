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
| `main` | **锚定上游 release tag**（当前 v1.6.1）。被 `.gitea/workflows/sync-upstream.yml` `git reset --hard <tag>` 覆写。**别在 main 上做任何改动** | merge 模型下 main 是 custom 的祖先，追新 tag 通常纯 fast-forward，一般不再需要 force |
| `custom` | **所有个人改动 + workflow 文件都在这**：信用额度功能、UI 调整、个人需求清单、`.gitea/workflows/*.yml` 等。具体改动清单见 [`FORK.md`](FORK.md)。日常开发分支，**default branch** | **否**（2026-07-22 起 merge 模型，历史 append-only，普通 push） |

⚠️ **default branch 是 `custom`**。`git clone` 默认 checkout custom，直接是开发分支。

### 历史：曾经存在过的 ci 分支（已退役）

2026-05-02 之前曾经有第三个分支 `ci`，最初设计是把 `.gitea/workflows/*.yml` 单独放它上面以"meta/code 分离"。两周后发现 Gitea Actions runs 列表显示的 commit 是 workflow 文件所在 commit（即 ci 的 HEAD），不是被构建的代码 commit，UX 误导性强。

把 workflow 挪回 custom 之后：
- runs 列表 commit = 真实代码 commit ✅
- `git clone` 默认落 custom 直接是开发分支 ✅
- 同步上游时 workflow 留在 custom 不动 ✅
- 代价：失去"workflow 与代码完全独立"的设计美感 —— 这个分离原本就是过度设计

**ci 分支于 2026-05-02 删除**，仅保留这段说明给后续 Claude 会话理解 git log 里"workflow 文件迁到 custom"这条提交（commit `555ecc1a`）的来龙去脉。**workflow 改动直接在 custom 上做**。

---

## custom 分支 workflow 清单

`.gitea/workflows/` 当前有 2 个 workflow（2026-05-04 起 build+deploy 合并为单 workflow 双 job）：

| 文件 | 触发 | 干什么 | 状态 |
|---|---|---|---|
| `sync-upstream.yml` | 手动（`workflow_dispatch`，可填 tag） | 服务端把 `dev/main` 强制 reset 到 mirror 上的指定 release tag（默认最新），然后 `push --force-with-lease` + 推 tags | ✅ 在用 |
| `build-image.yml` | **自动**（push 到 custom 触发，`paths-ignore` 屏蔽 `**.md` / `.gitignore` / `LICENSE` / `screenshot/**` / `sync-upstream.yml`）+ 手动备选 | **两个 job 串联在同一 run 里**：①`build` job 装 buildkit v0.13.2 → 登录 Gitea registry → 构建镜像（带 OCI 标签 source/revision，Gitea 自动关联包到 repo）→ push 到 `git.zhengchentao.win/dev/ezbookkeeping:<hash>` 与 `:latest`，`build-args: BUILD_PIPELINE=1` 跳过活 API 测试。②`deploy` job (`needs: build`) 登录 registry → clone nas-infra → `docker compose pull && up -d` 重启 ezbookkeeping。私有 nas-infra 需要 `secrets.NAS_INFRA_TOKEN`，公开仓库不需要。UI 上 Actions 列表显示一条 run，run 详情里 dependency graph 显示 build → deploy | ✅ 日常发布通道 + 自动 CD |

**已删**：
- `docker-snapshot.yml` / `docker-release.yml`（2026-05-02，依赖未配的 `secrets.DOCKER_REPO`，永远失败）
- `deploy.yml`（2026-05-04，合并进 `build-image.yml` 作为第二个 job，理由：原先 `workflow_run` 链触发会在 Actions 列表产生两条独立 run，UX 割裂；合并后单 run + dependency graph 看 build/deploy 状态一目了然）

需要时再从 git 历史 cherry-pick 回来。

---

## 同步发布流程（merge 模型，2026-07-22 起）

> 2026-07-22 从 rebase 模型切换为 merge 模型。历史上 custom 曾经每次同步都 rebase + force-push，此后不再。

1. 上游出新 release（如 v1.6.1）→ Gitea pull mirror 自动把 tag 同步到 mirror
2. 人工触发 `Sync from upstream` workflow → 服务端把 dev/main reset 到该 tag（v1.6.1 起 main 是 custom 的祖先，reset 到新 tag 通常是纯 fast-forward，本地直接 `git push origin <tag>:refs/heads/main` 也行）
3. 本地 `git fetch && git checkout custom && git merge origin/main`
4. 解冲突（**对着最终状态一次性解**，rerere 已开启会自动复用历史解法）→ 验证（`go vet/test` + `npx vue-tsc --noEmit` + `npx eslint src` + `npm test` + `npm run build`）→ 普通 `git push`
5. **build-image workflow 自动触发**，构建新镜像；不需要手动点

日常 feature commit 流程（全自动 CD）：

1. 在 custom 上改代码 → commit → push
2. **自动触发 build job**（除非只改了 `**.md` / `.gitignore` / `LICENSE` / `screenshot/**` / `sync-upstream.yml`）
3. build 成功 → **同 run 内 deploy job 接力跑**（`needs: build` 串联）：clone nas-infra → docker compose pull → up -d
4. 整条 push → build → deploy 链路无人工介入，UI 上是单条 run

**并发取消策略**：`build-image.yml` 设了 `concurrency.cancel-in-progress: true`，连续多次 push 时**只构建+部署最新那一次**（同 run 里 build/deploy 是原子单元，一起取消）。例：连续 3 次 push 间隔 30 秒，第 1 次 build 跑到 30%、第 2 次到来取消它、第 3 次又取消第 2，最终只 build + deploy 第 3 次的代码。省 CI 时间又保证最终一致性。

如果想跳过 build/deploy（例如手动多次 push 调试），commit 时只改文档相关文件即可（落在 paths-ignore 范围内）。如果想强制重打某个旧 commit，去 Actions UI 手动触发 `Build Docker Image`，填要打包的 branch / tag —— 注意手动触发也会跑 deploy job，**没有"只重新部署不重新 build"的单点入口了**（合并的代价，原 `deploy.yml` 那条路径已废）。临时只想重启容器：直接到 NAS 上 `docker compose up -d` 或在 Actions UI 临时禁用 deploy job。

**为什么 merge 不 rebase**（2026-07-22 决策）：rebase 的唯一收益是"干净的 patch 序列"，而 delta 的事实源是 FORK.md 文档不是 git 历史。rebase 成本 = O(commit 数 × 冲突面) 且逐次同步递增，还要 force-push（Actions 历史 run 指向的 commit 变孤儿、其他 checkout 需 hard reset）；merge 成本 = O(冲突面) 一次性，历史 append-only。看 fork 全部 delta 用 `git diff main...custom`。

---

## 给后续 Claude 会话的明确提示

- 用户说"我的分支" / "切换到我的分支" → 指 `custom`
- 用户说"同步上游" → 指 `git merge origin/main`（2026-07-22 起 merge 模型；**不要 rebase custom**，那是旧模型）
- **不要在 `main` 分支上提交任何东西**（会被 CI 覆写）
- **workflow 文件改动直接在 custom 上做**（2026-05-02 起，不再是 ci 分支）
- **不要 force-push custom**（merge 模型下历史 append-only；旧文档/记忆里"force-with-lease 是常规操作"已过时）
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
- **2026-07-22**：**同步模型从 rebase 切换为 merge** + 首次 merge 同步 v1.4.0-dev(422f1844) → v1.6.1（193 个上游 commit，merge commit `57cafadd`）。main fast-forward 到 v1.6.1（本地直接 push，没走 workflow）。开启 `rerere.enabled`。冲突 22 个文件，解决要点见 merge commit message；关键决策：ModifyBalance 保留 fork delta 语义（上游"更新期末余额"走普通收支交易，与 fork 不冲突）；NumberPadSheet `touch-action: manipulation` 已被上游吸收（FORK.md #11 第一阶段不再是 delta）。验证 go vet/test + vue-tsc + eslint + vitest + 生产 build 全过
- **2026-05-04**：把 `deploy.yml` 合并进 `build-image.yml` 作为第二个 job（`needs: build`），删除 `deploy.yml`。原先 `workflow_run` 链路会在 Actions 列表产生两条独立 run（build 完一条、deploy 又一条），用户视角割裂；合并后 UI 列表单条 run，run 详情里 dependency graph 显示 build → deploy 串联。代价：失去"不 rebuild 只 redeploy"的 UI 单点触发，临时只想重启容器需直接 ssh NAS 跑 compose。`paths-ignore` 移除已不存在的 `deploy.yml` 项

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
