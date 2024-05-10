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
	"fmt"
	"strings"

	longrunning "cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/googleapis/gapic-generator-go/internal/pbinfo"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
)

func (g *generator) genExampleFile(serv *descriptor.ServiceDescriptorProto) error {

	pkgName := g.opts.pkgName
	servName := pbinfo.ReduceServName(serv.GetName(), pkgName)

	g.exampleClientFactory(pkgName, servName)

	methods := append(serv.GetMethod(), g.getMixinMethods()...)

	for _, m := range methods {
		if err := g.exampleMethod(pkgName, servName, m); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) exampleClientFactory(pkgName, servName string) {
	p := g.printf
	for _, t := range g.opts.transports {
		s := servName
		if t == rest {
			s += "REST"
		}

		p("func ExampleNew%sClient() {", s)
		g.exampleInitClient(pkgName, s)
		p("")
		p("  // TODO: Use client.")
		p("  _ = c")
		p("}")
		p("")
	}

	g.imports[pbinfo.ImportSpec{Path: "context"}] = true
}

func (g *generator) exampleInitClient(pkgName, servName string) {
	p := g.printf

	p("ctx := context.Background()")
	p("// This snippet has been automatically generated and should be regarded as a code template only.")
	p("// It will require modifications to work:")
	p("// - It may require correct/in-range values for request initialization.")
	p("// - It may require specifying regional endpoints when creating the service client as shown in:")
	p("//   https://pkg.go.dev/cloud.google.com/go#hdr-Client_Options")
	p("c, err := %s.New%sClient(ctx)", pkgName, servName)
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")
	p("defer c.Close()")

	g.imports[pbinfo.ImportSpec{Path: "context"}] = true
}

func (g *generator) exampleMethod(pkgName, servName string, m *descriptor.MethodDescriptorProto) error {
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

func (g *generator) exampleMethodBody(pkgName, servName string, m *descriptor.MethodDescriptorProto) error {
	p := g.printf
	if m.GetClientStreaming() != m.GetServerStreaming() {
		// TODO(pongad): implement this correctly.
		return nil
	}
	/*
		m.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			how to get fielddescriptorproto from fielddescriptor golang		w := v.Message().Get(fd.Message().Fields().ByNumber(1))
			fmt.Println(w) // prints "text"
			return false
		})
		p := g.printf

		m.GetOptions()
	*/

	// inType = library.google.apis.com/BookRequestMessage
	mt := m.GetInputType()
	inType := g.descInfo.Type[mt]

	x := inType.(*descriptor.DescriptorProto)
	fmt.Println(x)

	y := x.Field[0]
	z := proto.GetExtension(y.GetOptions(), annotations.E_ResourceReference)
	a := z.(*annotations.ResourceReference)
	fmt.Println(a.Type) // "library.googleapis.com/Book"

	// 	inType2 := g.descInfo.Type[fmt.Sprintf(".google.cloud.%s", a.Type)]

	inType2 := g.descInfo.Type[".google.cloud.library.v1.Book"]
	b := inType2.(*descriptor.DescriptorProto)
	c := proto.GetExtension(b.GetOptions(), annotations.E_Resource)
	d := c.(*annotations.ResourceDescriptor)
	fmt.Println(d.Type)

	/*
		inType.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			fmt.Println(fd)
			fmt.Println(v)
			// Have fd protoreflect.FieldDescriptor
			// Want x descriptorpb.FieldDescriptorProto
			z := v.Message().Get(fd)
			fmt.Println(z)

			x := proto.GetExtension(fd.Options(), annotations.E_ResourceReference)
			y := proto.GetExtension(v.Message().Descriptor().Options(), annotations.E_ResourceReference)

			fmt.Println(x)
			fmt.Println(y)
			return false
		})
	*/

	// proto.GetExtension(x.Get(0).GetOptions(), annotations.E_ResourceReference)

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
		p("  // ")
		p("  // The resource name (either the Parent or Name field) should")
		p("  // be in the format `projects/<Project ID>` or")
		p("  // `projects/<Project ID>/locations/<Location ID>`")
		p("  // ")
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

func (g *generator) exampleLROCall(m *descriptor.MethodDescriptorProto) {
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

func (g *generator) exampleUnaryCall(m *descriptor.MethodDescriptorProto) {
	p := g.printf

	p("resp, err := c.%s(ctx, req)", *m.Name)
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")
	p("// TODO: Use resp.")
	p("_ = resp")
}

func (g *generator) exampleEmptyCall(m *descriptor.MethodDescriptorProto) {
	p := g.printf

	p("err = c.%s(ctx, req)", *m.Name)
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")
}

func (g *generator) examplePagingCall(m *descriptor.MethodDescriptorProto) error {
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

func (g *generator) exampleBidiCall(m *descriptor.MethodDescriptorProto, inType pbinfo.ProtoType, inSpec pbinfo.ImportSpec) {
	p := g.printf

	p("stream, err := c.%s(ctx)", m.GetName())
	p("if err != nil {")
	p("  // TODO: Handle error.")
	p("}")

	p("go func() {")
	p("  reqs := []*%s.%s{", inSpec.Name, inType.GetName())
	p("    // TODO: Create requests.")
	p("  }")
	p("  for _, req := range reqs {")
	p("    if err := stream.Send(req); err != nil {")
	p("            // TODO: Handle error.")
	p("    }")
	p("  }")
	p("  stream.CloseSend()")
	p("}()")

	p("for {")
	p("  resp, err := stream.Recv()")
	p("  if err == io.EOF {")
	p("    break")
	p("  }")
	p("  if err != nil {")
	p("    // TODO: handle error.")
	p("  }")
	p("  // TODO: Use resp.")
	p("  _ = resp")
	p("}")

	g.imports[pbinfo.ImportSpec{Path: "io"}] = true
}
