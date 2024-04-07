// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gengapic

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	longrunning "cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/googleapis/gapic-generator-go/internal/pbinfo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (g *generator) genExampleFile(serv *descriptorpb.ServiceDescriptorProto) error {
	pkgName := g.opts.pkgName
	servName := pbinfo.ReduceServName(serv.GetName(), pkgName)

<<<<<<< HEAD
=======
	g.exampleInitClientTemplate = newTemplate("exampleInitClient", exampleInitClient)
	g.exampleClientFactoryTemplate = newTemplate("exampleClientFactory", exampleClientFactory)
	g.exampleBidiCallTemplate = newTemplate("exampleBidiCall", exampleBidiCall)
>>>>>>> 4f0e934 (bidicall)
	if err := g.exampleClientFactory(pkgName, servName); err != nil {
		return err
	}

	methods := append(serv.GetMethod(), g.getMixinMethods()...)

	for _, m := range methods {
		if err := g.exampleMethod(pkgName, servName, m); err != nil {
			return err
		}
	}
	return nil
}

type tmpl struct {
	ServiceName string
	PackageName string
}

func (g *generator) exampleClientFactory(pkgName, servName string) error {
	p := g.printf

	for _, t := range g.opts.transports {
		if t == rest {
			servName += "REST"
		}
		out, err := execute(g.exampleClientFactoryTemplate, tmpl{
			ServiceName: servName,
			PackageName: pkgName,
		})
		if err != nil {
			return err
		}
		p(out)
		p("")
	}

	g.imports[pbinfo.ImportSpec{Path: "context"}] = true
	return nil
}

const exampleClientFactory = `func ExampleNew{{.ServiceName}}Client() {
	ctx := context.Background()
	// This snippet has been automatically generated and should be regarded as a code template only.
	// It will require modifications to work:
	// - It may require correct/in-range values for request initialization.
	// - It may require specifying regional endpoints when creating the service client as shown in:
	//   https://pkg.go.dev/cloud.google.com/go#hdr-Client_Options
	c, err := {{.PackageName}}.New{{.ServiceName}}Client(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	defer c.Close()

	// TODO: Use client.
	_ = c
}`

func (g *generator) exampleInitClient(pkgName, servName string) error {
	p := g.printf
	out, err := execute(g.exampleInitClientTemplate, tmpl{
		ServiceName: servName,
		PackageName: pkgName,
	})
	if err != nil {
		return err
	}

	p(out)
	g.imports[pbinfo.ImportSpec{Path: "context"}] = true
	return nil
}

const exampleInitClient = `
	ctx := context.Background()
	// This snippet has been automatically generated and should be regarded as a code template only.
	// It will require modifications to work:
	// - It may require correct/in-range values for request initialization.
	// - It may require specifying regional endpoints when creating the service client as shown in:
	//   https://pkg.go.dev/cloud.google.com/go#hdr-Client_Options
	c, err := {{.PackageName}}.New{{.ServiceName}}Client(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	defer c.Close()
`

func (g *generator) exampleMethod(pkgName, servName string, m *descriptorpb.MethodDescriptorProto) error {
	if m.GetClientStreaming() != m.GetServerStreaming() {
		// TODO(pongad): implement this correctly.
		return nil
	}

	p := g.printf

	p("func Example%sClient_%s() {", servName, m.GetName())
	if err := g.exampleMethodBody(pkgName, servName, m); err != nil {
		return err
	}

	p("}")
	p("")
	return nil
}

func (g *generator) exampleMethodBody(pkgName, servName string, m *descriptorpb.MethodDescriptorProto) error {
	if m.GetClientStreaming() != m.GetServerStreaming() {
		// TODO(pongad): implement this correctly.
		return nil
	}

	p := g.printf

	inType := g.descInfo.Type[m.GetInputType()]
	if inType == nil {
		return fmt.Errorf("cannot find type %q, malformed descriptor?", m.GetInputType())
	}

	inSpec, err := g.descInfo.ImportSpec(inType)
	if err != nil {
		return err
	}
	// TODO(codyoss): This if can be removed once the public protos
	// have been migrated to their new package. This should be soon after this
	// code is merged.
	if inSpec.Path == "google.golang.org/genproto/googleapis/longrunning" {
		inSpec.Path = "cloud.google.com/go/longrunning/autogen/longrunningpb"
	} else if inSpec.Path == "google.golang.org/genproto/googleapis/iam/v1" {
		inSpec.Path = "cloud.google.com/go/iam/apiv1/iampb"
	}

	httpInfo := getHTTPInfo(m)

	g.imports[inSpec] = true
	// Pick the first transport for simplicity. We don't need examples
	// of each method for both transports when they have the same surface.
	t := g.opts.transports[0]
	s := servName
	if t == rest {
		s += "REST"
	}
	g.exampleInitClient(pkgName, s)

	if !m.GetClientStreaming() && !m.GetServerStreaming() {
		p("")
		p("req := &%s.%s{", inSpec.Name, inType.GetName())
		p("  // TODO: Fill request struct fields.")
		p("  // See https://pkg.go.dev/%s#%s.", inSpec.Path, inType.GetName())
		p("}")
	}

	pf, _, err := g.getPagingFields(m)
	if err != nil {
		return err
	}
	if pf != nil {
		if err := g.examplePagingCall(m); err != nil {
			return err
		}
	} else if g.isLRO(m) || g.isCustomOp(m, httpInfo) {
		g.exampleLROCall(m)
	} else if *m.OutputType == emptyType {
		g.exampleEmptyCall(m)
	} else if m.GetClientStreaming() && m.GetServerStreaming() {
		g.exampleBidiCall(m, inType, inSpec)
	} else {
		g.exampleUnaryCall(m)
	}

	return nil
}

func (g *generator) exampleLROCall(m *descriptorpb.MethodDescriptorProto) {
	p := g.printf
	retVars := "resp, err :="

	// if response_type is google.protobuf.Empty, don't generate a "resp" var
	eLRO := proto.GetExtension(m.Options, longrunning.E_OperationInfo)
	opInfo := eLRO.(*longrunning.OperationInfo)
	if opInfo.GetResponseType() == emptyValue || opInfo == nil {
		// no new variables when this is used
		// therefore don't attempt to delcare it
		retVars = "err ="
	}

	p("op, err := c.%s(ctx, req)", *m.Name)
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")
	p("")

	p("%s op.Wait(ctx)", retVars)
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")
	// generate response handling snippet
	if strings.Contains(retVars, "resp") {
		p("// TODO: Use resp.")
		p("_ = resp")
	}
}

func (g *generator) exampleUnaryCall(m *descriptorpb.MethodDescriptorProto) {
	p := g.printf

	p("resp, err := c.%s(ctx, req)", *m.Name)
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")
	p("// TODO: Use resp.")
	p("_ = resp")
}

func (g *generator) exampleEmptyCall(m *descriptorpb.MethodDescriptorProto) {
	p := g.printf

	p("err = c.%s(ctx, req)", *m.Name)
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")
}

func (g *generator) examplePagingCall(m *descriptorpb.MethodDescriptorProto) error {
	outType := g.descInfo.Type[m.GetOutputType()]
	if outType == nil {
		return fmt.Errorf("cannot find type %q, malformed descriptor?", m.GetOutputType())
	}

	outSpec, err := g.descInfo.ImportSpec(outType)
	if err != nil {
		return err
	}

	p := g.printf

	p("it := c.%s(ctx, req)", m.GetName())
	p("for {")
	p("  resp, err := it.Next()")
	p("  if err == iterator.Done {")
	p("    break")
	p("  }")
	p("  if err != nil {")
	p("    // TODO: Handle error.")
	p("  }")
	p("  // TODO: Use resp.")
	p("  _ = resp")
	p("")
	p("  // If you need to access the underlying RPC response,")
	p("  // you can do so by casting the `Response` as below.")
	p("  // Otherwise, remove this line. Only populated after")
	p("  // first call to Next(). Not safe for concurrent access.")
	p("  _ = it.Response.(*%s.%s)", outSpec.Name, outType.GetName())
	p("}")

	g.imports[pbinfo.ImportSpec{Path: "google.golang.org/api/iterator"}] = true
	g.imports[outSpec] = true
	return nil
}

func (g *generator) exampleBidiCall(m *descriptorpb.MethodDescriptorProto, inType pbinfo.ProtoType, inSpec pbinfo.ImportSpec) error {
	p := g.printf

	out, err := execute(g.exampleBidiCallTemplate, struct {
		MethodName string
		InSpecName string
		InTypeName string
	}{
		MethodName: m.GetName(),
		InSpecName: inSpec.Name,
		InTypeName: inType.GetName(),
	})
	if err != nil {
		return err
	}
	p(out)
	g.imports[pbinfo.ImportSpec{Path: "io"}] = true
	return nil
}

const exampleBidiCall = `
stream, err := c.{{.MethodName}}(ctx)
if err != nil {
	// TODO: Handle error.
}

go func() {
	reqs := []*{{.InSpecName}}.{{.InTypeName}}{
	// TODO: Create requests.")
	}
	for _, req := range reqs {
	if err := stream.Send(req); err != nil {
			// TODO: Handle error.
	}
	}
	stream.CloseSend()
}()

for {
	resp, err := stream.Recv()
	if err == io.EOF {
	break
	}
	if err != nil {
	// TODO: handle error.
	}
	// TODO: Use resp.
	_ = resp
}
`

func newTemplate(name string, body ...string) *template.Template {
	t := template.Must(template.New(name).Parse(""))
	for _, b := range body {
		t.Parse(b)
	}
	return t
}

func execute(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
