package dapr

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	contextPackage = protogen.GoImportPath("context")
	errorPackage   = protogen.GoImportPath("errors")
	jsonPackage    = protogen.GoImportPath("encoding/json")

	commonPackage = protogen.GoImportPath("github.com/dapr/go-sdk/service/common")
	clientPackage = protogen.GoImportPath("github.com/dapr/go-sdk/client")
)

func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_service.gen.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-dapr. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(gen, file, g)
	return g
}

func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}

	g.P()

	for _, service := range file.Services {
		genService(gen, file, g, service)
	}
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	g.P("// Service Client")
	g.P()
	genClient(gen, file, g, service)

	g.P("// Service Server")
	g.P()
	genServer(gen, file, g, service)

}

func clientSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	return method.GoName + clientParameters(g, method) + clientReturns(g, method)
}

func clientParameters(g *protogen.GeneratedFile, method *protogen.Method) string {
	s := "("
	s += "ctx " + g.QualifiedGoIdent(contextPackage.Ident("Context")) + ", "
	s += "in *" + g.QualifiedGoIdent(method.Input.GoIdent)
	s += ")"
	return s
}

func clientReturns(g *protogen.GeneratedFile, method *protogen.Method) string {
	s := "("
	s += "*" + g.QualifiedGoIdent(method.Output.GoIdent) + ", "
	s += "error"
	s += ")"
	return s
}

func genClientMethod(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, method *protogen.Method, index int) {
	service := method.Parent
	// sname := fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())

	if method.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
		g.P(deprecationComment)
	}
	g.P("func (c *", unexport(service.GoName), "Client)", clientSignature(g, method), "{")
	g.P("data, err := ", jsonPackage.Ident("Marshal"), "(in)")
	g.P("if err != nil { return nil, err }")
	g.P("content := &", clientPackage.Ident("DataContent"), `{ContentType: "application/json", Data: data}`)

	g.P(`resp, err := c.cc.InvokeMethodWithContent(ctx, c.appID,"`, method.Desc.Name(), `", "post", content)`)
	g.P("if err != nil { return nil, err }")
	g.P("out := new(", method.Output.GoIdent, ")")
	g.P("err = ", jsonPackage.Ident("Unmarshal"), "(resp, out)")
	g.P("if err != nil { return nil, err }")

	g.P("return out, nil")
	g.P("}")
	g.P()
}

func genClient(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {

	clientName := service.GoName + "Client"

	g.P("// ", clientName, " is the client API for ", service.GoName, " service.")

	// Client interface.
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	g.Annotate(clientName, service.Location)
	g.P("type ", clientName, " interface {")
	for _, method := range service.Methods {
		g.Annotate(clientName+"."+method.GoName, method.Location)
		if method.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
			g.P(deprecationComment)
		}
		g.P(method.Comments.Leading,
			clientSignature(g, method))
	}
	g.P("}")
	g.P()

	g.P("type ", unexport(clientName), " struct {")
	g.P("cc ", clientPackage.Ident("Client"))
	g.P("appID string")
	g.P("}")
	g.P()

	// NewClient factory.
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P(deprecationComment)
	}
	g.P("// Client factory")
	g.P("func New", clientName, " (cc ", clientPackage.Ident("Client"), ", appID string) ", clientName, " {")
	g.P("return &", unexport(clientName), "{cc, appID}")
	g.P("}")
	g.P()

	var methodIndex, streamIndex int
	// Client method implementations.
	for _, method := range service.Methods {
		if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
			// Unary RPC method
			genClientMethod(gen, file, g, method, methodIndex)
			methodIndex++
		} else {
			// Streaming RPC method
			genClientMethod(gen, file, g, method, streamIndex)
			streamIndex++
		}
	}
}

func invocationHandlerSignature(g *protogen.GeneratedFile) string {
	s := "func("
	s += "ctx " + g.QualifiedGoIdent(contextPackage.Ident("Context")) + ", "
	s += "in *" + g.QualifiedGoIdent(commonPackage.Ident("InvocationEvent"))
	s += ") ("
	s += "out *" + g.QualifiedGoIdent(commonPackage.Ident("Content")) + ", "
	s += "err error"
	s += ")"
	return s
}

func genServer(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {

	// Server interface.
	serverType := service.GoName + "Server"
	g.P("// ", serverType, " is the server API for ", service.GoName, " service.")
	g.P("// All implementations must embed Unimplemented", serverType)
	g.P("// for forward compatibility")
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	g.Annotate(serverType, service.Location)
	g.P("type ", serverType, " interface {")
	for _, method := range service.Methods {
		g.Annotate(serverType+"."+method.GoName, method.Location)
		if method.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
			g.P(deprecationComment)
		}
		g.P(method.Comments.Leading,
			serverSignature(g, method))
	}
	g.P("mustEmbedUnimplemented", serverType, "()")
	g.P("}")
	g.P()

	// Server Unimplemented struct for forward compatibility.
	g.P("// Unimplemented", serverType, " must be embedded to have forward compatible implementations.")
	g.P("type Unimplemented", serverType, " struct {")
	g.P("}")
	g.P()

	for _, method := range service.Methods {
		g.P("func (Unimplemented", serverType, ") ", serverSignature(g, method), "{")
		g.P("return nil, ", errorPackage.Ident("New"), `("method `, method.GoName, ` not implemented")`)
		g.P("}")
	}
	g.P("func (Unimplemented", serverType, ") mustEmbedUnimplemented", serverType, "() {}")
	g.P()

	// Unsafe Server interface to opt-out of forward compatibility.
	g.P("// Unsafe", serverType, " may be embedded to opt out of forward compatibility for this service.")
	g.P("// Use of this interface is not recommended, as added methods to ", serverType, " will")
	g.P("// result in compilation errors.")
	g.P("type Unsafe", serverType, " interface {")
	g.P("mustEmbedUnimplemented", serverType, "()")
	g.P("}")

	g.P("type InvocationHandlerFunc ", invocationHandlerSignature(g))

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Methods {
		hname := genServerMethod(gen, file, g, method)
		handlerNames = append(handlerNames, hname)
	}

	genRegister(g, service, handlerNames)
}

func genRegister(g *protogen.GeneratedFile, service *protogen.Service, handlerNames []string) {
	g.P("func Register(s ", commonPackage.Ident("Service"), ", srv ", service.GoName, "Server) {")
	for i, method := range service.Methods {
		g.P(`s.AddServiceInvocationHandler("`, method.GoName, `",`, handlerNames[i], "(srv))")
	}
	g.P("}")
}

func serverSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	var reqArgs []string
	ret := "error"
	if !method.Desc.IsStreamingClient() && !method.Desc.IsStreamingServer() {
		reqArgs = append(reqArgs, "ctx "+g.QualifiedGoIdent(contextPackage.Ident("Context")))
		ret = "(*" + g.QualifiedGoIdent(method.Output.GoIdent) + ", error)"
	}
	if !method.Desc.IsStreamingClient() {
		reqArgs = append(reqArgs, "in *"+g.QualifiedGoIdent(method.Input.GoIdent))
	}
	if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
		reqArgs = append(reqArgs, method.Parent.GoName+"_"+method.GoName+"Server")
	}
	return method.GoName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
}

func genServerMethod(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, method *protogen.Method) string {
	service := method.Parent
	hname := fmt.Sprintf("_%s_%s_Handler", service.GoName, method.GoName)

	g.P("func ", hname, "(srv ", service.GoName, "Server) InvocationHandlerFunc {")
	g.P("return ", invocationHandlerSignature(g), "{")
	g.P("req := &", method.Input.GoIdent, "{}")
	g.P("decErr := ", jsonPackage.Ident("Unmarshal"), "(in.Data, req)")

	g.P("if decErr != nil {")
	g.P("err = decErr")
	g.P("return }")

	g.P("resp, mErr := srv.", method.GoName, "(ctx, req)")

	g.P("if mErr != nil {")
	g.P("err = mErr")
	g.P("return }")

	g.P("data, encErr := ", jsonPackage.Ident("Marshal"), "(resp)")

	g.P("if encErr != nil {")
	g.P("err = encErr")
	g.P("return }")

	g.P("out = &", commonPackage.Ident("Content"), "{")
	g.P(`ContentType: "application/json",`)
	g.P("Data: data,")
	g.P("}")

	g.P("return")

	g.P("}")
	g.P("}")

	// g.P("in := new(", method.Input.GoIdent, ")")
	// g.P("if err := dec(in); err != nil { return nil, err }")
	// g.P("if interceptor == nil { return srv.(", service.GoName, "Server).", method.GoName, "(ctx, in) }")
	// g.P("info := &", grpcPackage.Ident("UnaryServerInfo"), "{")
	// g.P("Server: srv,")
	// g.P("FullMethod: ", strconv.Quote(fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())), ",")
	// g.P("}")
	// g.P("handler := func(ctx ", contextPackage.Ident("Context"), ", req interface{}) (interface{}, error) {")
	// g.P("return srv.(", service.GoName, "Server).", method.GoName, "(ctx, req.(*", method.Input.GoIdent, "))")
	// g.P("}")
	// g.P("return interceptor(ctx, in, info, handler)")
	// g.P("}")
	// g.P()

	return hname
}

const deprecationComment = "// Deprecated: Do not use."

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }