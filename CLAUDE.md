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

本地 `git remote -v` 只有 origin 一项，**没有手工配 upstream**。上游同步通过 ci 分支上的 workflow 在服务端完成，不是本地操作。

---

## 三个分支的职责（必须先理解，否则会改错地方）

| 分支 | 职责 | force push? |
|---|---|---|
| `main` | **锚定上游 release tag**（当前 v1.4.0）。被 `.gitea/workflows/sync-upstream.yml` `git reset --hard <tag>` 覆写 | 是（由 CI 做） |
| `custom` | **所有个人改动都在这**：信用额度功能、UI 调整、个人需求清单等。具体改动清单见 [`FORK.md`](FORK.md)。日常开发分支 | 是（rebase 后人工做） |
| `ci` | **origin 的 default branch**。只放 `.gitea/workflows/*.yml`。独立于 main 和 custom | 否 |

⚠️ origin 的 default branch 是 `ci` 不是 main 也不是 custom。`git clone` 默认会落在 ci。要做开发先 `git checkout custom`。

### ci 与 custom 不需要保持一致（这是设计，不是 bug）

三分支各管各的，**ci 与 custom 内容不重叠**：

- ci：只有 `.gitea/workflows/*.yml`
- custom：fork 代码 + FORK.md / CLAUDE.md / .gitignore 等
- 理论上没有共享文件（除了 inherit 自上游的部分，例如 ci 分支也有 README、Dockerfile 等基底文件，但很少改它们）

**Gitea Actions UI 顶部显示的 commit 是 ci 的 HEAD，不是被构建的源代码 commit**。这是因为 workflow 文件在 ci 分支，dispatch 触发时 UI 引用的是 workflow 文件所在 commit。但实际：

1. `actions/checkout` 用 `ref: ${{ inputs.branch }}` 拉的是 custom
2. 镜像 tag 用 custom 的 short SHA
3. OCI label `org.opencontainers.image.revision` 是 custom 的 full SHA
4. workflow 末尾 Build summary 步骤把上述源 commit 显式写入 `$GITHUB_STEP_SUMMARY`，运行页面有突出显示

所以**真实构建的代码 commit** 在镜像本身、运行页面 summary 区都能看到，UI 顶部那个只是 dispatch 触发位置的 commit，**不是 bug，无需修复，看 summary 区即可**。

如果哪天觉得这种割裂太烦，可选两个替代方案：
- 把 workflow 挪到 custom，删 ci（UI commit 直接对得上，但失去 meta/code 分离）
- 加 `push: branches: [custom]` 自动触发（每次 push 自动 build，更直观但费 CI 时间）

当前选 `workflow_dispatch only + summary step` 是中间路径，节省 CI 又能看到真实信息。

---

## ci 分支 workflow 清单

ci 分支的 `.gitea/workflows/` 当前有 5 个 workflow：

| 文件 | 触发 | 干什么 | 状态 |
|---|---|---|---|
| `sync-upstream.yml` | 手动（`workflow_dispatch`，可填 tag） | 服务端把 `dev/main` 强制 reset 到 mirror 上的指定 release tag（默认最新），然后 `push --force-with-lease` + 推 tags | ✅ 在用 |
| `build-image.yml` | 手动（可填要打包的分支 + 镜像 tag） | checkout 指定分支（默认 `custom`）→ 装 buildkit v0.13.2（钉版本）→ 登录 Gitea registry → 构建镜像（带 OCI 标签 source/revision，Gitea 自动关联包到 repo）→ push 到 `git.zhengchentao.win/dev/ezbookkeeping:<hash>` 与 `:latest`，`build-args: BUILD_PIPELINE=1` 跳过活 API 测试 | ✅ 在用，是日常发布通道 |
| `deploy.yml` | 手动 | 跑 repo Variables 里 `CUSTOM_DEPLOY_SCRIPTS` 这条自定义脚本（通用钩子，可拼"build 完触发 NAS 端 docker compose pull/up"等） | 🟡 通用钩子，按需配 |
| `docker-snapshot.yml` | `push: branches: main` 自动 | 上游模板风格的 snapshot build（用 `secrets.DOCKER_REPO`，会改 Dockerfile 里 FROM 走私有 mirror） | ⚠️ 上游残留，main 被 sync-upstream force-push 时会触发；当前未配 secret 估计直接失败。要么补全 secret，要么删除 |
| `docker-release.yml` | `push: tags: v*` 自动 | 同上但 release 风格，tag 推送时触发 | ⚠️ 同 docker-snapshot.yml，未配 secret 会失败 |

**结论**：日常只用 `sync-upstream` + `build-image` 两个，其他三个要么按需启用要么后续清理。

---

## 同步发布流程（rebase 模型）

1. 上游出新 release（如 v1.4.0）→ Gitea pull mirror 自动把 tag 同步到 mirror
2. 人工触发 `Sync from upstream` workflow → 服务端把 dev/main reset 到该 tag
3. 本地 `git fetch && git checkout custom && git rebase origin/main`
4. 解冲突（如有）→ 验证 → `git push --force-with-lease origin custom`
5. ci 分支上的 build-image workflow 触发，构建新镜像

**为什么 rebase 不 merge**：个人项目，无团队协作语义要保留，线性历史更清爽。

---

## 给后续 Claude 会话的明确提示

- 用户说"我的分支" / "切换到我的分支" → 指 `custom`
- 用户说"rebase main" → 指 `git rebase origin/main`，目标是把 custom 的改动叠到最新上游 tag 之上
- **不要在 `main` 分支上提交任何东西**（会被 CI 覆写）
- **不要把工作流文件提交到 `custom` 或 `main`**（应该走 ci 分支）
- force-push custom 是常规操作，但每次用 `--force-with-lease`，不直接 `--force`
- 如果发现本地配了 upstream remote，那是历史遗留，不要依赖；以 origin/main 为准
- `.claude/` 在 `.gitignore` 里（个人本地配置不入库），但 `CLAUDE.md` 本身入库

---

## 同步历史

- **2026-05-01**：rebase custom → origin/main (v1.4.0)。22 个 custom-only 提交（含一个旧的 `Merge branch 'main' into myrequirement` commit）压平为 21 个线性提交。已 force-push origin/custom（`08c69042` → `fe265259`）。
- **2026-05-02**：修 Gitea Actions `Build Docker Image` 工作流。三层故障，全部不在本仓库代码里：
  - **TLS 雷**：`docker login` 走 host 进程不命中 PREROUTING REDIRECT，且 v6 撞 DSM nginx 的 CF Origin Cert。NAS 侧修：iptables 补 OUTPUT 对称规则 + `/etc/hosts` 显式 v4 兜底。详见 obsidian vault [[NAS/notes/内网证书路径]] §三.5/§三.6
  - **buildkit 内核兼容**：runc 1.2+ 撞 DSM 4.4 内核。ci 分支 `.gitea/workflows/build-image.yml` 钉 `moby/buildkit:v0.13.2`（commit `acdbb5bf`）
  - **backend 单元测试撞活 API**：`pkg/exchangerates/` 的 `TestExchangeRatesApiLatestExchangeRateHandler_*` 跑活 API（加拿大银行 / 乌兹别克央行），国内访问超时。upstream Dockerfile 已设 `ARG BUILD_PIPELINE`，测试代码看到 `BUILD_PIPELINE=1 && CHECK_3RD_API!=1` 时早退。修：workflow 加 `build-args: BUILD_PIPELINE=1`（commit `2dd8f099`），对齐上游 GH Actions（`.github/actions/build-linux-docker-and-package/action.yml` line 91）

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
