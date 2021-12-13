package main

import (
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	contextPkg         = protogen.GoImportPath("context")
	ginPkg             = protogen.GoImportPath("github.com/gin-gonic/gin")
	httpPkg             = protogen.GoImportPath("net/http")
	errPkg             = protogen.GoImportPath("errors")
	metadataPkg        = protogen.GoImportPath("google.golang.org/grpc/metadata")
	deprecationComment = "// Deprecated: Do not use."
)

var methodSets = make(map[string]int)

// generateFile generates a _gin.pb.go file.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_http.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by github.com/sunyuan300/protoc-gen-go-http. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	g.P("// ", contextPkg.Ident(""), ginPkg.Ident(""),httpPkg.Ident(""))
	//g.P("// ", contextPkg.Ident(""), metadataPkg.Ident(""))
	//g.P("//", ginPkg.Ident(""), errPkg.Ident(""))
	g.P()

	// 遍历每个proto文件的service
	for _, service := range file.Services {
		// 生成service的Go代码
		genService(gen, file, g, service)
	}
	return g
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, s *protogen.Service) {
	if s.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	// HTTP Server.
	sd := &Service{
		Name:     s.GoName,
		FullName: string(s.Desc.FullName()),
		FilePath: file.Desc.Path(),
	}

	// 遍历service的rpc方法
	for _, m := range s.Methods {
		sd.Methods = append(sd.Methods, genMethod(m)...)
	}
	g.P(sd.execute())
}

func genMethod(m *protogen.Method) []*Method {
	var methods []*Method
	rule, ok := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
	if rule != nil && ok {
		methods = append(methods, buildHTTPRule(m, rule))
		return methods
	}
	return methods
}

func buildHTTPRule(m *protogen.Method, rule *annotations.HttpRule) *Method {
	var (
		path   string
		method string
	)
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = "GET"
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = "PUT"
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = "POST"
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = "DELETE"
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = "PATCH"
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	md := buildMethodDesc(m, method, path)
	return md
}

func buildMethodDesc(m *protogen.Method, httpMethod, path string) *Method {
	defer func() { methodSets[m.GoName]++ }()
	md := &Method{
		Name:    m.GoName,
		Num:     methodSets[m.GoName],
		Request: m.Input.GoIdent.GoName,
		Reply:   m.Output.GoIdent.GoName,
		Path:    path,
		Method:  httpMethod,
	}
	return md
}