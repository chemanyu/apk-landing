package handler

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"net/url"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"apk-landing/internal/service"
	"apk-landing/internal/types"
)

// LandingHandler 京东落地页 handler
type LandingHandler struct {
	svc  *service.LandingService
	tmpl *template.Template
}

// NewLandingHandler 创建 handler。tmpl 为已解析的 jd.html 模板。
func NewLandingHandler(svc *service.LandingService, tmpl *template.Template) *LandingHandler {
	return &LandingHandler{svc: svc, tmpl: tmpl}
}

// pageData 模板渲染上下文
type pageData struct {
	BgImg      string
	ClickImg   string
	ConfigJSON template.JS // 注入 <script> 的配置 JSON
}

// Handle 渲染京东落地页
func (h *LandingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req types.LandingRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	data := h.svc.BuildRenderData(req.URL, req.ImgURL)

	cfgJSON, err := json.Marshal(data)
	if err != nil {
		logx.WithContext(r.Context()).Errorf("marshal render data failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, pageData{
		BgImg:      data.BgImg,
		ClickImg:   data.ClickImg,
		ConfigJSON: template.JS(cfgJSON),
	}); err != nil {
		logx.WithContext(r.Context()).Errorf("render template failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())

	// 业务访问日志：每次落地页访问打一条结构化日志，便于投放排查与统计。
	logx.WithContext(r.Context()).Infow("jd_landing_visit",
		logx.Field("client_ip", clientIP(r)),
		logx.Field("ua", r.UserAgent()),
		logx.Field("referer", r.Referer()),
		logx.Field("path", r.URL.Path),
		logx.Field("jd_url", req.URL),          // 解码后的京东 CPS 链接
		logx.Field("media", parseMediaParams(req.URL)), // 从 url 拆出的媒体参数
		logx.Field("deep_link", data.DeepLink), // 本次下发的 deeplink
	)
}

// clientIP 取真实客户端 IP，优先信任反向代理传的头（nginx 已配 X-Real-IP / X-Forwarded-For）。
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff // 可能是逗号分隔的链路，首个为最初客户端
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// parseMediaParams 从京东 CPS 链接里提取常见的媒体投放参数（已被媒体宏替换后的真实值），
// 便于按 rtaId / requestId 等维度排查与统计。解析失败返回空 map，不影响主流程。
func parseMediaParams(jdURL string) map[string]string {
	out := map[string]string{}
	if jdURL == "" {
		return out
	}
	u, err := url.Parse(jdURL)
	if err != nil {
		return out
	}
	q := u.Query()
	// 关注的媒体参数键（巨量等），存在才记录
	for _, k := range []string{"rtaId", "rtaExpId", "adPlanId", "adUserId", "adCreativityId", "site", "os", "materialId"} {
		if v := q.Get(k); v != "" {
			out[k] = v
		}
	}
	return out
}

