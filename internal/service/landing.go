package service

import (
	"strings"

	"apk-landing/internal/config"
)

// RenderData 传给模板 / 客户端 JS 的全部数据。
// 设备/厂商识别、商店 scheme 拼接均在客户端完成，这里只下发配置。
type RenderData struct {
	DeepLink     string            `json:"deepLink"`     // 已用 url 替换 __LINK__ 的京东 deeplink
	ULink        string            `json:"uLink"`        // 已替换的 ulink
	Pkg          string            `json:"pkg"`          // 京东包名
	Apk          string            `json:"apk"`          // 安卓 APK 直链（非主流厂商兜底）
	IOSAppStore  string            `json:"iosAppStore"`  // iOS App Store
	VendorStores map[string]string `json:"vendorStores"` // 厂商 → 商店 scheme 模板（含 {pkg} 占位）
	DelayMs      int               `json:"delayMs"`      // 唤端兜底等待时长
	BgImg        string            `json:"-"`            // 背景图（模板直接用，不进 JSON）
	ClickImg     string            `json:"-"`            // 点击按钮图
}

// LandingService 落地页业务逻辑。无状态，便于后续加埋点/多客户。
type LandingService struct {
	jd       config.JDAppConfig
	fallback config.FallbackConfig
}

// NewLandingService 创建 service
func NewLandingService(jd config.JDAppConfig, fallback config.FallbackConfig) *LandingService {
	return &LandingService{jd: jd, fallback: fallback}
}

// vendorStoreTpls 各厂商应用市场 scheme 模板，{pkg} 在客户端替换为包名。
// 与参考落地页一致：华为/小米/vivo/OPPO/三星，其余厂商下 APK。
func vendorStoreTpls() map[string]string {
	return map[string]string{
		"huawei":  "appmarket://details?id={pkg}",
		"xiaomi":  "mimarket://details?id={pkg}",
		"vivo":    "vivomarket://details?packagename={pkg}",
		"oppo":    "oppomarket://details?id={pkg}",
		"samsung": "samsungapps://ProductDetail/{pkg}",
	}
}

// BuildRenderData 根据 url 参数组装下发数据。
func (s *LandingService) BuildRenderData(rawURL, imgURL string) RenderData {
	deepLink := strings.ReplaceAll(s.jd.DeepLinkTpl, "__LINK__", rawURL)
	uLink := strings.ReplaceAll(s.jd.ULinkTpl, "__LINK__", rawURL)

	bg := s.jd.DefaultBgImg
	if imgURL != "" {
		bg = imgURL
	}

	return RenderData{
		DeepLink:     deepLink,
		ULink:        uLink,
		Pkg:          s.jd.Pkg,
		Apk:          s.jd.Apk,
		IOSAppStore:  s.jd.IOSAppStore,
		VendorStores: vendorStoreTpls(),
		DelayMs:      s.fallback.DelayMs,
		BgImg:        bg,
		ClickImg:     s.jd.ClickImg,
	}
}
