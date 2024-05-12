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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/googleapis/gapic-generator-go/internal/testing/sample"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestGenSecret(t *testing.T) {
	fileSet, err := run()
	if err != nil {
		t.Fatal(err)
	}
	var toGen []string
	for _, f := range fileSet {
		toGen = append(toGen, *f.Name)
	}

	parameter := fmt.Sprintf("go-gapic-package=%s;%s", sample.GoProtoPackagePath, sample.GoProtoPackageName)
	output, err := Gen(&pluginpb.CodeGeneratorRequest{
		FileToGenerate: toGen,
		ProtoFile:      fileSet,
		Parameter:      &parameter,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range output.File {
		if f.Name == nil {
			continue
		}
		if *f.Name == "cloud.google.com/go/internal/generated/snippets/secretmanager/apiv1/secretmanagerpb/SecretManagerClient/CreateSecret/main.go" {
			fmt.Println(*f.Content)
			/*
				for _, a := range f.GeneratedCodeInfo.Annotation {
					fmt.Println(a.Path)
				}
			*/
		}
	}
}

func run() (set []*descriptorpb.FileDescriptorProto, err error) {
	files := []string{
		"google/api/annotations",
		"google/api/client",
		"google/api/field_behavior",
		"google/api/resource",
		"google/cloud/secretmanager/v1/resources",
		"google/cloud/secretmanager/v1/service",
		"google/iam/v1/iam_policy",
		"google/iam/v1/policy",
	}
	for _, f := range files {
		protoFile, err := readFile(f)
		if err != nil {
			return nil, err
		}

		pb_set := new(descriptorpb.FileDescriptorSet)
		if err := proto.Unmarshal(protoFile, pb_set); err != nil {
			return nil, err
		}
		set = append(set, pb_set.File...)
	}
	return set, nil
}

func readFile(f string) (_ []byte, err error) {
	// First, convert the .proto file to a file descriptor set
	tmp_file := filepath.Base(f) + "_tmp.pb"
	srcDir := "protodata"
	protofile := srcDir + "/" + f + ".proto"
	cmd := exec.Command(
		"protoc",
		"--descriptor_set_out="+tmp_file,
		"-I"+srcDir,
		protofile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//	fmt.Println(cmd.String())
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			err = os.Remove(tmp_file)
		}
	}()
	protoFile, err := os.ReadFile(tmp_file)
	if err != nil {
		return nil, err
	}
	return protoFile, nil
}
