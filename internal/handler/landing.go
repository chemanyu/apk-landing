package handler

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

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
}
