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
	"path/filepath"
	"testing"

	"github.com/googleapis/gapic-generator-go/internal/pbinfo"
	"github.com/googleapis/gapic-generator-go/internal/snippets"
	"github.com/googleapis/gapic-generator-go/internal/testing/sample"
	"github.com/googleapis/gapic-generator-go/internal/txtdiff"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/genproto/googleapis/api/serviceconfig"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/apipb"
)

// TestExampleMethodBody_Pattern tests
// https://github.com/googleapis/gapic-generator-go/issues/1372, using the
// example in
// https://github.com/googleapis/gapic-generator-go/issues/1372#issuecomment-1633101248.
func TestExampleMethodBody_Pattern(t *testing.T) {
	g := generator{
		serviceConfig: &serviceconfig.Service{
			Apis: []*apipb.Api{
				{
					Name: fmt.Sprintf("%s.%s", sample.ProtoPackage, sample.ProtoService),
				},
			},
		},
		imports: map[pbinfo.ImportSpec]bool{
			{Path: "context"}: true,
			{Name: sample.ProtoPackage, Path: sample.GoPackagePath}: true,
		},
		snippetMetadata: snippets.NewMetadata(sample.ProtoPackage, sample.GoPackagePath, sample.GoPackageName),
		descInfo:        pbinfo.Of([]*descriptorpb.FileDescriptorProto{}),
		protoPackage:    sample.ProtoPackage,
		opts: &options{
			pkgName:    sample.GoPackageName,
			transports: []transport{grpc, rest},
		},
	}
	g.snippetMetadata.AddService(sample.ProtoService, sample.ProtoPackage)

	serv := sample.Service()
	inputType := createRequestInputWithResourceReferenceField()
	inputType2 := sample.InputType(sample.GetRequest)
	outputType := resourceOutputWithResourceField()

	for _, typ := range []*descriptorpb.DescriptorProto{
		inputType,
		inputType2,
		outputType,
	} {
		typName := sample.DescriptorInfoTypeName(typ.GetName())
		g.descInfo.Type[typName] = typ
		g.descInfo.ParentFile[typ] = &descriptorpb.FileDescriptorProto{
			Options: &descriptorpb.FileOptions{
				GoPackage: proto.String(sample.GoPackagePath),
			},
			Package: proto.String(sample.ProtoPackage),
		}
	}

	for _, m := range serv.Method {
		if err := g.genSnippetFile(serv, m); err != nil {
			t.Fatal(err)
		}
	}
	g.commit(filepath.Join(sample.SnippetsDirectory, "main.go"), "main")
	got := *g.resp.File[1].Content
	fmt.Println(got)
	txtdiff.Diff(t, got, filepath.Join("testdata", "snippet_Library.want"))
}

func createRequestInputWithResourceReferenceField() *descriptorpb.DescriptorProto {
	name := "name"
	inputType := &descriptorpb.DescriptorProto{
		Name: proto.String(sample.CreateRequest),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:    &name,
				Options: &descriptorpb.FieldOptions{},
			},
		},
	}
	proto.SetExtension(
		inputType.Field[0].GetOptions(),
		annotations.E_ResourceReference,
		&annotations.ResourceReference{
			Type: sample.ResourceType,
		},
	)
	return inputType
}

func resourceOutputWithResourceField() *descriptorpb.DescriptorProto {
	name := "name"
	outputType := &descriptorpb.DescriptorProto{
		Name: proto.String(sample.Resource),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name: &name,
			},
		},
		Options: &descriptorpb.MessageOptions{},
	}
	proto.SetExtension(
		outputType.GetOptions(),
		annotations.E_Resource,
		&annotations.ResourceDescriptor{
			Type:    sample.ResourceType,
			Pattern: []string{"projects/*/secrets/*", "projects/*/locations/*/secrets/*"},
		})
	return outputType
}
