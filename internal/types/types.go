package types

// LandingRequest 落地页查询参数。
// url 是媒体（巨量等）宏替换后的京东 CPS 推广链接，落地页原样透传。
type LandingRequest struct {
	URL    string `form:"url,optional"`     // 京东 CPS 链接（已 URL 编码）
	ImgURL string `form:"img_url,optional"` // 自定义背景图，覆盖默认
	Type   string `form:"type,optional"`    // 预留：type=1 走微信小程序（v1 暂不实现）
}
