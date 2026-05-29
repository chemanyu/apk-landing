package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"apk-landing/internal/config"
	"apk-landing/internal/handler"
	"apk-landing/internal/service"
)

//go:embed templates/*.html
var templatesFS embed.FS

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化 service（无状态，落地页 v1 不依赖 DB）
	landingSvc := service.NewLandingService(c.JD, c.Fallback)

	// 解析内嵌的落地页模板
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/jd.html"))

	// 启动 HTTP 服务
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	landingHandler := handler.NewLandingHandler(landingSvc, tmpl)

	// 路由前缀：公网经反向代理以 /ulink 前缀转发到本服务时，路由需带同样前缀。
	prefix := strings.TrimRight(c.URLPrefix, "/")

	// 京东落地页：新路径 + 兼容旧投放链接的别名
	for _, path := range []string{"/jd/landing", "/jd_apk/index3.html"} {
		server.AddRoute(rest.Route{
			Method:  http.MethodGet,
			Path:    prefix + path,
			Handler: landingHandler.Handle,
		})
	}

	// 健康检查
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   prefix + "/ping",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			httpx.OkJson(w, map[string]string{"status": "ok"})
		},
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
