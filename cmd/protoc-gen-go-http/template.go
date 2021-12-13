package main

import (
	"bytes"
	"html/template"
	"strings"
)

var tpl = `
type {{ $.Name }}HTTPServer interface {
{{range .Methods}}
   {{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{end}}
}
func Register{{ $.Name }}HTTPServer(group *gin.RouterGroup, srv {{ $.Name }}HTTPServer) {
{{range .Methods}}
   group.Handle("{{.Method}}","{{.Path}}",_{{ $.Name }}_{{.Name}}_HTTP_Handler(srv))
{{end}}
}

{{range .Methods}}
func _{{ $.Name }}_{{.Name}}_HTTP_Handler(srv {{ $.Name }}HTTPServer) func(ctx *gin.Context) {
   return func(ctx *gin.Context) {
      var in {{.Request}}
   {{if .HasPathParams }}
      if err := ctx.ShouldBindUri(&in); err != nil {
         return
      }
   {{end}}
   {{if eq .Method "GET" "DELETE" }}
      if err := ctx.ShouldBindQuery(&in); err != nil {
         return
      }
   {{else if eq .Method "POST" "PUT" }}
      if err := ctx.ShouldBindJSON(&in); err != nil {
         return
      }
   {{end}}
      out,err := srv.{{.Name}}(ctx,&in)
      if err != nil {
         return
      }
      ctx.JSON(http.StatusOK,out)
   }
}
{{end}}

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

func (m *Method) HasPathParams() bool {
	paths := strings.Split(m.Path, "/")
	for _, p := range paths {
		if len(p) > 0 && (p[0] == '{' && p[len(p)-1] == '}' || p[0] == ':') {
			return true
		}
	}
	return false
}