package config

import "github.com/zeromicro/go-zero/rest"

// Config 项目总配置
type Config struct {
	rest.RestConf
	URLPrefix string `json:",optional"` // 路由前缀（如 /ulink），用于公网反代部署；空则无前缀
	JD        JDAppConfig    // 京东 App 落地页配置
	Fallback  FallbackConfig // 唤端兜底策略
}

// JDAppConfig 京东 App 相关配置。
// 后续接入其他客户时，可把这一段抽成 map[string]AppConfig 按客户区分。
type JDAppConfig struct {
	Pkg         string // 京东包名，用于拼厂商应用市场 scheme
	Apk         string // 安卓 APK 直链（非主流厂商兜底）
	IOSAppStore string // iOS App Store 链接
	// DeepLinkTpl/ULinkTpl 中的 __LINK__ 会被 url 参数替换
	DeepLinkTpl  string
	ULinkTpl     string
	DefaultBgImg string // 默认背景图（可被 img_url 参数覆盖）
	ClickImg     string // 引导点击按钮图
}

// FallbackConfig 唤端兜底策略
type FallbackConfig struct {
	DelayMs int // 唤端后等待多久（毫秒）仍未切走，则判定未安装 → 跳应用商店/下载
}
