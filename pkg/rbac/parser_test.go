/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rbac

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"golang.org/x/tools/go/packages"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "todo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gen := Generator{}

			reg := &markers.Registry{}

			if err := gen.RegisterMarkers(reg); err != nil {
				t.Fatalf("unable to register markers: %v", err)
			}

			input := inputFromMemory{
				map[string]string{
					"foo": "bar",
				},
			}

			roots, err := loader.LoadRootsWithConfig(
				&packages.Config{BuildFlags: []string{"-tags", "test"}},
				"foo",
			)
			if err != nil {
				t.Fatalf("unable to laod roots with config: %v", err)
			}

			ctx := genall.GenerationContext{
				Collector: &markers.Collector{
					Registry: &markers.Registry{},
				},
				Roots:     roots,
				InputRule: input,
				Checker:   &loader.TypeChecker{},
			}
			ctx.OutputRule = outputToMemory{}
			if err := gen.Generate(&ctx); err != nil {
				t.Fatalf("unable to genrate: %v", err)
			}
		})
	}
}

type outputToMemory struct{}

func (o outputToMemory) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return nopCloser{ioutil.Discard}, nil
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}

type memoryReaderCloser struct {
	data []byte
	read int

	nopCloser
}

func (r *memoryReaderCloser) Read(p []byte) (int, error) {
	if len(r.data) == r.read {
		return 0, io.EOF
	}

	size := len(p)
	high := r.read + size
	if high > len(r.data) {
		high = len(r.data)
		size = high - r.read
	}

	for i := 0; i < size; i++ {
		p[i] = r.data[r.read+i]
	}

	r.read += size

	return size, nil
}

type inputFromMemory struct {
	files map[string]string
}

func (in inputFromMemory) OpenForRead(path string) (io.ReadCloser, error) {
	content, ok := in.files[path]
	if !ok {
		return nil, fmt.Errorf("unable to open %s for read", path)
	}
	return &memoryReaderCloser{data: []byte(content)}, nil
}

// InputFromFileSystem reads from the filesystem as normal.
var InputFromFileSystem = inputFromMemory{}
