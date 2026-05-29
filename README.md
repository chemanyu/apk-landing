# apk-landing

京东广告投放唤端落地页服务。用我们自己的域名托管落地页，投放在巨量等媒体的 H5 位：
用户点击后尝试唤起京东 App（deeplink/DP），**未安装则跳应用商店或下载 APK，安装完成后吊起 DP**。

参考自研落地页：`https://landing.domob.cn/jd_apk/index3.html`。

## 技术栈

- Go 1.21 + [go-zero](https://github.com/zeromicro/go-zero) `rest`（与公司栈一致，自带 OTEL/logx）
- 模板：标准库 `html/template` + `//go:embed`，单二进制部署
- v1 无状态，不依赖 DB

## 架构

```
main.go                     启动：加载配置 → 注册路由 → server.Start()
internal/config/config.go   Config{rest.RestConf; JD; Fallback}
internal/types/types.go     LandingRequest 查询参数
internal/service/landing.go 组装下发数据（url 透传 + 京东配置），不做 UA 判断
internal/handler/landing.go GET 落地页：parse → service → 渲染模板
templates/jd.html           背景图+按钮 + 客户端 UA识别/厂商商店/唤端/兜底 JS
etc/config.yaml             端口 + 京东 App 配置
```

**设备/厂商识别在客户端 JS 完成**（与参考页一致）：服务端只注入京东配置（包名、APK、App Store、
各厂商商店 scheme 模板、deeplink），JS 按 UA 选 iOS App Store / 华为·小米·vivo·OPPO·三星应用市场 /
APK 直链，唤端后 `DelayMs` 毫秒内页面未切走即判定未安装，跳兜底地址。

## 运行

```bash
go mod tidy
go build -o apk-landing .
./apk-landing -f etc/config.yaml
```

> 内网无法访问默认 GOPROXY 时：`GOPROXY=https://goproxy.cn,direct go mod tidy`

## 路由

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/jd/landing` | 京东落地页 |
| GET | `/jd_apk/index3.html` | 旧投放链接兼容别名 |
| GET | `/ping` | 健康检查 |

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `url` | 是 | 媒体宏替换后的京东 CPS 推广链接（URL 编码）。落地页原样透传，不处理 `__REQUEST_ID__` 等媒体宏 |
| `img_url` | 否 | 自定义背景图，覆盖默认 |
| `type` | 否 | 预留：`type=1` 走微信小程序（v1 暂未实现） |

## 投放链接示例

```
https://<你的域名>/jd/landing?url=https%3A%2F%2Fu.jd.com%2FR1TIVfS%3Fe%3DCPS-11417-__REQUEST_ID__-CPS-DMCID_NEW--__TS__--6000035425%26adPlanId%3D__PROJECT_ID__
```

## 配置说明（etc/config.yaml）

京东 App 配置全部走 `JD` 段，后续接入其他客户时可把这段抽成 `map[string]AppConfig` 按客户区分。
`Fallback.DelayMs` 控制唤端兜底等待时长（默认 1000ms）。

## 验证

```bash
curl localhost:8888/ping                                          # {"status":"ok"}
curl "localhost:8888/jd/landing?url=https%3A%2F%2Fu.jd.com%2Fxxx"  # 返回 HTML，#cfg 含拼好的 deeplink
```

浏览器 DevTools 切换 UA（华为/小米/iOS/普通）可验证 JS 选出的兜底地址。
# apk-landing
