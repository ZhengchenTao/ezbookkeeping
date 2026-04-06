# 部署说明

## 镜像地址

```
ghcr.io/zhengchentao/ezbookkeeping:latest
```

每次向 `myrequirement` 分支推送代码，GitHub Actions 自动构建并推送新镜像。

---

## 首次迁移（从官方镜像换成自定义镜像）

### 1. 备份容器内配置文件到宿主机

```bash
sudo docker cp ezbookkeeping:/ezbookkeeping/conf/ezbookkeeping.ini /opt/ezbookkeeping/ezbookkeeping.ini
```

> 这样之后删容器也不会丢配置。

### 2. 停止并删除旧容器

```bash
docker stop ezbookkeeping && docker rm ezbookkeeping
```

> 只删容器本身，数据目录不受影响。

### 3. 登录 GitHub Container Registry（只需一次）

在 GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic) 生成 token，勾选 `read:packages`，然后：

```bash
echo 你的TOKEN | docker login ghcr.io -u zhengchentao --password-stdin
```

### 4. 拉取新镜像

```bash
docker pull ghcr.io/zhengchentao/ezbookkeeping:latest
```

### 5. 启动容器

```bash
docker run -d \
  --name ezbookkeeping \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /opt/ezbookkeeping/data:/ezbookkeeping/data \
  -v /opt/ezbookkeeping/ezbookkeeping.ini:/ezbookkeeping/conf/ezbookkeeping.ini \
  -e EBK_MCP_ENABLE_MCP=true \
  -e EBK_SECURITY_ENABLE_API_TOKEN=true \
  ghcr.io/zhengchentao/ezbookkeeping:latest
```

**参数说明：**
| 参数 | 含义 |
|------|------|
| `-d` | 后台运行 |
| `--restart unless-stopped` | 服务器重启后自动启动 |
| `-p 8080:8080` | 端口映射 |
| `-v .../data:...` | 挂载数据目录（数据库、图片等） |
| `-v .../ezbookkeeping.ini:...` | 挂载配置文件 |
| `-e EBK_*` | 环境变量覆盖配置 |

### 6. 确认运行正常

```bash
docker ps                        # 确认容器在运行
docker logs ezbookkeeping        # 查看启动日志，确认无报错
```

---

## 后续更新（代码有改动时）

```bash
# 拉取最新镜像
docker pull ghcr.io/zhengchentao/ezbookkeeping:latest

# 停止并删除旧容器
docker stop ezbookkeeping && docker rm ezbookkeeping

# 重新启动（与首次启动命令相同）
docker run -d \
  --name ezbookkeeping \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /opt/ezbookkeeping/data:/ezbookkeeping/data \
  -v /opt/ezbookkeeping/ezbookkeeping.ini:/ezbookkeeping/conf/ezbookkeeping.ini \
  -e EBK_MCP_ENABLE_MCP=true \
  -e EBK_SECURITY_ENABLE_API_TOKEN=true \
  ghcr.io/zhengchentao/ezbookkeeping:latest
```

---

## 常用运维命令

```bash
# 查看运行中的容器
docker ps

# 查看容器实时日志（Ctrl+C 退出）
docker logs -f ezbookkeeping

# 进入容器内部排查问题
docker exec -it ezbookkeeping sh

# 查看磁盘占用
docker system df
```
