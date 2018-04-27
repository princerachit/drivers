/*
Copyright 2017 The Kubernetes Authors.

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

package openebs

import (
	"testing"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"fmt"
)

func TestCheckArguments(t *testing.T) {
	req := &csi.CreateVolumeRequest{Name: "csi-volume-1"}
	fmt.Printf("Test 1: req=%v", req)
	if err := checkArguments(req); err == nil {
		t.Errorf("Expected error when req=%v", req)
	}else {
		fmt.Println("passed\n")
	}

	req = &csi.CreateVolumeRequest{Name: "csi-volume-1", VolumeCapabilities: []*csi.VolumeCapability{&csi.VolumeCapability{AccessMode: nil}}}
	fmt.Printf("Test 2: req=%v", req)
	if err := checkArguments(req); err == nil {
		t.Errorf("Expected error when req=%v", req)
	}else {
		fmt.Println("passed\n")
	}

	req = &csi.CreateVolumeRequest{Name: "csi-volume-1", VolumeCapabilities: []*csi.VolumeCapability{&csi.VolumeCapability{AccessMode: nil}}, Parameters: map[string]string{"storage-class-name": "openebs"}}
	fmt.Printf("Test 3: req=%v", req)
	if err := checkArguments(req); err != nil {
		t.Errorf("Expected success when req=%v", req)
	}else {
		fmt.Println("passed\n")
	}
}
