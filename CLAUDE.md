# CLAUDE.md

本项目是 [mayswind/ezbookkeeping](https://github.com/mayswind/ezbookkeeping) 的个人 fork。

工作流详细决策框架见个人笔记 `fork-工作流决策框架.md`。本文件只记录**这个仓库的具体事实**，避免 Claude 会话误判。

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
| `custom` | **所有个人改动都在这**：信用额度功能、UI 调整、个人需求清单等。日常开发分支 | 是（rebase 后人工做） |
| `ci` | **origin 的 default branch**。只放 `.gitea/workflows/*.yml`。独立于 main 和 custom | 否 |

⚠️ origin 的 default branch 是 `ci` 不是 main 也不是 custom。`git clone` 默认会落在 ci。要做开发先 `git checkout custom`。

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
- **2026-05-02**：修 Gitea Actions `Build Docker Image` 工作流。两层 NAS 网络问题 + 一层 buildkit/内核问题，全部不在本仓库代码里：
  - **TLS 雷**：`docker login` 走 host 进程不命中 PREROUTING REDIRECT，且 v6 撞 DSM nginx 的 CF Origin Cert。NAS 侧修：iptables 补 OUTPUT 对称规则 + `/etc/hosts` 显式 v4 兜底。详见 obsidian vault [[NAS/notes/内网证书路径]] §三.5/§三.6
  - **buildkit 内核兼容**：runc 1.2+ 撞 DSM 4.4 内核。本仓库 ci 分支 `.gitea/workflows/build-image.yml` 已钉 `moby/buildkit:v0.13.2`（commit `acdbb5bf`）

## 给后续 Claude 会话：CI 故障排查路径

如果 Gitea Actions build 又炸，按 NAS 域问题 vs 仓库代码问题分别排查：

| 现象 | 大概率位置 | 文档 |
|---|---|---|
| `Login to Gitea Container Registry` 步骤报 `x509: certificate signed by unknown authority` | NAS 网络层（iptables / dnsmasq / DSM nginx 占 443） | obsidian vault `NAS/notes/内网证书路径.md` + `NAS/notes/IPv6 设计.md` |
| `Build and push` 步骤里 `RUN ...` 在第二条之内就炸 `unsafe procfs detected` 之类 | buildkit/runc 与 DSM 内核版本 | `.gitea/workflows/build-image.yml` 的 `driver-opts` |
| `actions/checkout` 报 fetch 失败 | Gitea SSH/HTTPS 路径或 token 权限 | gitea-runner 的 `GITEA_RUNNER_REGISTRATION_TOKEN` + NPM `git.zhengchentao.win` 的 Advanced 配置 |
| Dockerfile 里某条指令业务逻辑报错 | 真正的代码问题 | 本仓库 `Dockerfile` |
