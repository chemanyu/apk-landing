# apk-landing 部署文档

京东广告投放唤端落地页服务。本文档覆盖 **Linux（systemd，推荐生产用）** 和 **Windows VM** 两种部署方式。

---

## 0. 架构与关键概念（先读这段）

```
广告媒体投放链接（HTTPS）
   https://你的域名/ulink/jd/landing/jd.html?url=京东CPS链接
        │
        ▼
   nginx 反向代理（80/443，配 SSL）
        │  按 proxy_pass 写法决定是否剥掉 /ulink 前缀
        ▼
   apk-landing 服务（监听 127.0.0.1:5001）
```

**最容易踩的坑——`/ulink` 前缀必须 nginx 和服务端配套**：

| nginx `proxy_pass` 写法 | 后端收到的路径 | `etc/config.yaml` 里 `URLPrefix` |
|------------------------|---------------|--------------------------------|
| `http://127.0.0.1:5001/`（**带**尾斜杠，剥前缀） | `/jd/landing/jd.html` | `URLPrefix: ""` ← **当前线上** |
| `http://127.0.0.1:5001`（**无**尾斜杠，透传） | `/ulink/jd/landing/jd.html` | `URLPrefix: /ulink` |

配错就是 404。改完 `URLPrefix` 要重启服务。

服务端注册的落地页路径（前缀剥离后）：
`/jd/landing`、`/jd/landing/jd.html`、`/jd/landing/index.html`、`/jd_apk/index3.html`、健康检查 `/ping`。

---

## 1. Linux 部署（systemd，推荐）

### 1.1 在开发机交叉编译打包

```bash
cd /Users/chemanyu/workspace/apk-landing
bash scripts/build-linux.sh
```

产出 `dist-linux/`：`apk-landing`（二进制，模板已内嵌）、`etc/config.yaml`、`run.sh`、`apk-landing.service`。

> 也可以直接在 Linux 服务器上 `git clone` 后 `GOPROXY=https://goproxy.cn,direct go build -o apk-landing .`，效果相同。

### 1.2 上传到服务器

```bash
# 在开发机执行，IP/路径按实际改
scp -r dist-linux/* user@服务器IP:/opt/apk-landing/
```

### 1.3 安装 systemd 服务（在服务器上）

```bash
cd /opt/apk-landing
# 确认 config.yaml 的 URLPrefix 与你的 nginx 写法配套（见第 0 节）
vim etc/config.yaml

# 安装服务
sudo cp apk-landing.service /etc/systemd/system/
# 若用了非 www-data 用户，先改 service 里的 User/Group 和 WorkingDirectory
sudo systemctl daemon-reload
sudo systemctl enable --now apk-landing
```

### 1.4 常用运维命令

```bash
sudo systemctl status apk-landing      # 看状态
sudo systemctl restart apk-landing     # 改配置后重启
sudo systemctl stop apk-landing        # 停止
journalctl -u apk-landing -f           # 看服务进程输出（启动报错等；业务日志见下方）
```

### 1.4.1 业务日志

日志由 `etc/config.yaml` 的 `Log` 段控制，默认 `Mode: file`，写到**运行目录下的 `logs/`**
（即 `dist-linux/logs/`，因 systemd `WorkingDirectory` 指向 dist-linux）。

```bash
cd /home/sysadmin/data/apk-landing/dist-linux

tail -f logs/access.log | grep jd_landing_visit          # 实时看落地页访问
grep -c jd_landing_visit logs/access.log                 # 今日访问量
tail -1 logs/access.log | grep jd_landing_visit | jq .   # 看单条结构（需 jq）
```

每条 `jd_landing_visit` 业务日志字段：`client_ip`（真实访客 IP，取 X-Forwarded-For 首个）、
`ua`、`jd_url`（解码后京东 CPS 链接）、`media`（拆出的 rtaId/site/adPlanId 等媒体参数）、
`deep_link`（下发的唤端地址）、`trace`/`span`（链路 ID）。

日志文件分级：`access.log`(info+业务) / `error.log` / `severe.log` / `slow.log` / `stat.log`。
`KeepDays: 14` 自动清理过期日志。

> 已通过 `Middlewares.Log: false` 关掉框架自带的 `[HTTP]` 访问日志，避免和业务日志混在一起。
> 若想改回输出到 journald（用 `journalctl` 看），把 `Log.Mode` 改成 `console`。

### 1.5 配 nginx（在服务器上）

把 `deploy/nginx.conf.sample` 里对应方案的 `location` 块加进你的站点配置，注意 SSL：

```nginx
server {
    listen 443 ssl;
    server_name 你的域名;
    ssl_certificate     /path/cert.pem;
    ssl_certificate_key /path/key.pem;

    # 方案 A：剥前缀（配套 URLPrefix: ""）
    location /ulink/ {
        proxy_pass http://127.0.0.1:5001/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
    }
}
```

```bash
sudo nginx -t && sudo systemctl reload nginx
```

### 1.6 验证

```bash
# 服务器本地（绕过 nginx，直连服务）
curl -i http://127.0.0.1:5001/ping
curl -s "http://127.0.0.1:5001/jd/landing/jd.html?url=https%3A%2F%2Fu.jd.com%2FR1TIVfS" | grep deepLink

# 走公网（经 nginx）
curl -i "https://你的域名/ulink/ping"
```

---

## 2. Windows VM 部署

### 2.1 在开发机交叉编译打包

```bash
cd /Users/chemanyu/workspace/apk-landing
bash scripts/build-windows.sh
```

产出 `dist/`：`apk-landing.exe`、`etc/config.yaml`、`run.bat`。

### 2.2 部署运行

1. 把整个 `dist/` 拷到 Windows VM（如 `D:\apk-landing\`）。
2. 确认 `etc\config.yaml` 的 `URLPrefix` 与 VM 上 nginx 写法配套（见第 0 节）。
3. 双击 `run.bat` 启动 → 服务监听 `0.0.0.0:5001`，关闭窗口即停止。

> 想让它常驻、开机自启、崩溃自拉起，用 [NSSM](https://nssm.cc/) 注册成 Windows 服务：
> ```
> nssm install apk-landing D:\apk-landing\apk-landing.exe "-f etc\config.yaml"
> nssm set apk-landing AppDirectory D:\apk-landing
> nssm start apk-landing
> ```

### 2.3 配 nginx（宝塔面板）

宝塔站点 → 反向代理，或直接编辑站点 conf 加 `deploy/nginx.conf.sample` 里的 `location` 块（注意尾斜杠与 `URLPrefix` 配套）。

### 2.4 验证

```
# VM 本地（PowerShell/CMD）
curl http://127.0.0.1:5001/ping
# 走公网
https://你的域名/ulink/jd/landing/jd.html?url=...
```

---

## 3. 改配置 / 改代码后怎么重新部署

| 改了什么 | 操作 |
|---------|------|
| 只改了 `etc/config.yaml`（端口/前缀/京东配置） | 改服务器上的 config.yaml → 重启服务（Linux `systemctl restart`，Windows 重开 run.bat / `nssm restart`） |
| 改了 Go 代码或模板 `templates/jd.html` | 重新跑 `build-linux.sh` / `build-windows.sh` → 覆盖服务器上的二进制 → 重启服务 |

> 模板是 `//go:embed` 编进二进制的，改了 `jd.html` **必须重新编译**才生效，不能只替换文件。

---

## 4. 上线检查清单

- [ ] `URLPrefix` 与 nginx `proxy_pass` 尾斜杠配套（第 0 节表格）
- [ ] 域名配了 **HTTPS**（deeplink 唤端 + 媒体投放都要求）
- [ ] `curl https://域名/ulink/ping` 返回 `{"status":"ok"}`
- [ ] 手机真机打开落地页，背景图/按钮图能正常显示（图挂了会白屏，检查 CDN 防盗链/HTTP混用）
- [ ] 装了京东 App → 能唤起；没装 → 跳应用商店/下载
