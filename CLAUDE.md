# CLAUDE.md — apk-landing

京东广告投放唤端落地页服务（go-zero）。媒体投放 H5 → 用户点击 → 唤起京东 App，
未装则跳应用商店/下载 APK。设备/厂商识别在客户端 JS，服务端只下发配置 + 渲染模板。

## 项目规则（踩坑沉淀，改动时务必遵守）

- **nginx `proxy_pass` 尾斜杠与 `URLPrefix` 必须配套**：带尾斜杠（`http://127.0.0.1:5001/`）会剥掉
  `/ulink` 前缀，后端收到 `/jd/landing/...`，对应 `config.yaml` 的 `URLPrefix: ""`；不带尾斜杠透传前缀，
  对应 `URLPrefix: /ulink`。配错就是 404。
  - 原因：两端对路径前缀的预期必须一致，否则路由不命中。
  - 来源：docs/engineering/session-log.md §2026-05-29 18:55

- **真实客户端 IP 取 `X-Forwarded-For` 首个**，不要用 `X-Real-IP`：本服务在多层代理后，
  `X-Real-IP` 拿到的是上游内网 IP（如 10.x），XFF 形如 `真实IP, 代理1, 代理2`，首个才是访客。
  见 `internal/handler/landing.go` 的 `clientIP()`。
  - 来源：docs/engineering/session-log.md §2026-05-29 18:55

- **模板 `templates/jd.html` 是 `//go:embed` 进二进制的**：改了模板（图片/样式/JS）必须**重新编译**
  才生效，不能只替换服务器上的 html 文件。
  - 来源：docs/engineering/session-log.md §2026-05-29 18:55

## 部署

- 服务器只有 git、无 Go：本机 `bash scripts/build-linux.sh` 交叉编译，二进制 `dist-linux/apk-landing`
  **纳入 git**，服务器 `./deploy.sh`（git pull → 装 systemd → 重启）。
- 详见 DEPLOY.md。
