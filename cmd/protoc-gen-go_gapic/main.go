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

// Command protoc-gen-go_gapic provides a Generated API Client (GAPIC)
// generator that generates a Go client library from protocol buffers.
package main

import (
	"io"
	"log"
	"os"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/googleapis/gapic-generator-go/internal/gengapic"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	reqBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var genReq plugin.CodeGeneratorRequest
	if err := proto.Unmarshal(reqBytes, &genReq); err != nil {
		return err
	}

	genResp, err := gengapic.Gen(&genReq)
	if err != nil {
		genResp.Error = proto.String(err.Error())
	}

	genResp.SupportedFeatures = proto.Uint64(uint64(plugin.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))

	outBytes, err := proto.Marshal(genResp)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(outBytes); err != nil {
		return err
	}
	return nil
}
