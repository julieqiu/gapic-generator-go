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

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/googleapis/gapic-generator-go/internal/pbinfo"
	"github.com/googleapis/gapic-generator-go/internal/snippets"
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
	const (
		serviceName     = "LibraryService"
		serviceEndpoint = "library.googleapis.com"

		protoVersion     = "v1"
		protoPackagePath = "google.cloud.library.v1"

		goPackagePath      = "cloud.google.com/go/library/apiv1"
		goPackageName      = "library"
		goProtoPackagePath = "cloud.google.com/go/library/apiv1/librarypb"
		goProtoPackageName = "librarypb"
	)
	g := generator{
		serviceConfig: &serviceconfig.Service{
			Apis: []*apipb.Api{
				{
					Name: fmt.Sprintf("%s.%s", protoPackagePath, serviceName),
				},
			},
		},
		imports: map[pbinfo.ImportSpec]bool{
			{Path: "context"}: true,
			{Name: goProtoPackageName, Path: goProtoPackagePath}: true,
		},
		snippetMetadata: snippets.NewMetadata(protoPackagePath, goPackagePath, goPackageName),
		descInfo:        pbinfo.Of([]*descriptor.FileDescriptorProto{}),
		opts: &options{
			pkgName:    goPackageName,
			transports: []transport{grpc, rest},
		},
	}
	g.snippetMetadata.AddService(serviceName, protoPackagePath)

	/*
		service LibraryService {
			rpc GetBook (GetBookRequest) returns (Book) {
				option (google.api.http) = {
					get: "/v1/{name=books/*}"
				};
			}
		}
	*/
	serv := &descriptor.ServiceDescriptorProto{
		Name: proto.String(serviceName),
		Method: []*descriptor.MethodDescriptorProto{
			{
				Name:       proto.String("GetBookRequest"),
				InputType:  proto.String(fmt.Sprintf(".%s.GetBookRequest", protoPackagePath)),
				OutputType: proto.String(fmt.Sprintf(".%s.Book", protoPackagePath)),
			},
		},
	}

	/*
		  message GetBookRequest {
				string name = 1 [(google.api.resource_reference).type = "library.googleapis.com/Book"];
		  }
	*/
	name := "name"
	inputType := &descriptor.DescriptorProto{
		Name: proto.String("GetBookRequest"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:    &name,
				Options: &descriptor.FieldOptions{},
			},
		},
	}
	proto.SetExtension(
		inputType.Field[0].GetOptions(),
		annotations.E_ResourceReference,
		&annotations.ResourceReference{
			Type: fmt.Sprintf("%s/Book", serviceEndpoint),
		},
	)

	/*
		message Book {
		  option (google.api.resource) = {
		    type: "library.googleapis.com/Book"
		    pattern: "books/{book}"
		  };

		  string name = 1;
		}
	*/
	outputType := &descriptor.DescriptorProto{
		Name: proto.String("Book"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name: &name,
			},
		},
		Options: &descriptor.MessageOptions{},
	}
	proto.SetExtension(
		outputType.GetOptions(),
		annotations.E_Resource,
		&annotations.ResourceDescriptor{
			Type:    fmt.Sprintf("%s/Book", serviceEndpoint),
			Pattern: []string{"books/{book}"},
		})

	for _, typ := range []*descriptor.DescriptorProto{
		inputType,
		outputType,
	} {
		typName := fmt.Sprintf(".%s.%s", protoPackagePath, typ.GetName())
		g.descInfo.Type[typName] = typ
		g.descInfo.ParentFile[typ] = &descriptor.FileDescriptorProto{
			Options: &descriptor.FileOptions{
				GoPackage: proto.String(goProtoPackagePath),
			},
			Package: proto.String(protoPackagePath),
		}
	}

	for _, m := range serv.Method {
		if err := g.genSnippetFile(serv, m); err != nil {
			t.Fatal(err)
		}
	}
	g.commit(filepath.Join("cloud.google.com/go", "internal", "generated", "snippets", goPackageName, "main.go"), "main")
	got := *g.resp.File[1].Content
	txtdiff.Diff(t, got, filepath.Join("testdata", "snippet_Library.want"))
}
