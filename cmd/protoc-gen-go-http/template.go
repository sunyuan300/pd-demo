package main

import (
	"bytes"
	"html/template"
	"strings"
)

var tpl = `
type {{ $.Name }}HTTPServer interface {
{{range .MethodSet}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{end}}
}
func Register{{ $.Name }}HTTPServer(r gin.IRouter, srv {{ $.Name }}) {
	s := {{.Name}}{
		server: srv,
		router:     r,
	}
	s.RegisterService()
}

type {{$.Name}} struct{
	server {{ $.Name }}
	router gin.IRouter
}


{{range .Methods}}
func (s *{{$.Name}}) {{ .Name }} (ctx *gin.Context) {
	var in {{.Request}}
	if err := ctx.BindJSON(&in); err != nil {
		return
	}
	out, err := s.server.({{ $.Name }}).{{.Name}}(ctx, &in)
	if err != nil {
		return
	}
}
{{end}}

func (s *{{$.Name}}) RegisterService() {
{{range .Methods}}
		s.router.Handle("{{.Method}}", "{{.Path}}", s.{{ .Name }})
{{end}}
}
`

type Service struct {
	Name     string
	FullName string
	FilePath string

	Methods   []*Method
	MethodSet map[string]*Method
}

func (s *Service) execute() string {
	if s.MethodSet == nil {
		s.MethodSet = map[string]*Method{}
		for _, m := range s.Methods {
			m := m
			s.MethodSet[m.Name] = m
		}
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(tpl))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return buf.String()
}

type Method struct {
	Name    string // SayHello
	Num     int    // 一个 rpc 方法可以对应多个 http 请求
	Request string // SayHelloReq
	Reply   string // SayHelloResp
	// http_rule
	Path         string // 路由
	Method       string // HTTP Method
	Body         string
	ResponseBody string
}
