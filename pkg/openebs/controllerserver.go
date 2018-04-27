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
	"golang.org/x/net/context"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/kubernetes-csi/drivers/pkg/openebs/mayaproxy"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"
	"github.com/golang/glog"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"fmt"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func checkArguments(req *csi.CreateVolumeRequest) error {
	if len(req.GetName()) == 0 {
		return status.Error(codes.InvalidArgument, "Name missing in request")
	}
	if req.GetVolumeCapabilities() == nil {
		return status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}
	if req.Parameters["storage-class-name"] == "" {
		return status.Error(codes.InvalidArgument, "Missing storage-class-name in request")
	}
	return nil
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	glog.Infof("Received request: %v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, err
	}

	// Check arguments
	err := checkArguments(req)
	if err != nil {
		return nil, err
	}

	mayaConfig := &mayaproxy.MayaConfig{}
	err = mayaConfig.SetupMayaConfig(mayaproxy.K8sClient{})
	if err != nil {
		glog.Errorf("Error setting up MayaConfig")
		return nil, status.Error(codes.Unavailable, fmt.Sprint(err))
	}

	var volume *mayav1.Volume
	volumeSpec := mayav1.VolumeSpec{}

	// check if volume already exists
	volume, err = mayaConfig.ProxyListVolume(req.GetName())
	if err != nil {
		// REVIEW THIS. Do we need to return error here or return the value.
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Volume with the same name: %s but with different size already exist", req.GetName()))
	}
	

	// Convert from bytes to GigaBytes
	volSize := int64(req.GetCapacityRange().GetRequiredBytes() / 1e9)
	volumeSpec.Metadata.Labels.Storage = string(volSize)

	volumeSpec.Metadata.Labels.StorageClass = req.Parameters["storage-class-name"]
	volumeSpec.Metadata.Name = req.Name
	volumeSpec.Metadata.Labels.Namespace = "default"

	// Issue a request to Maya API Server to create a volume
	err = mayaConfig.ProxyCreateVolume(volumeSpec)

	if err != nil {
		return nil, status.Error(codes.Unavailable, fmt.Sprint(err))
	}

	volume, err = mayaConfig.ProxyListVolume(req.Name)
	if err != nil {
		return nil, status.Error(codes.DeadlineExceeded, fmt.Sprintf("Volume with the same name: %s but with different size already exist", req.GetName()))
	}

	// extract iscsi volume details
	var iqn, targetPortal string

	for key, value := range volume.Metadata.Annotations.(map[string]interface{}) {
		switch key {
		case "vsm.openebs.io/iqn":
			iqn = value.(string)
		case "vsm.openebs.io/targetportals":
			targetPortal = value.(string)
		}
	}

	attributes := map[string]string{"iqn": iqn, "targetPortal": targetPortal, "lun": "0"}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:            volume.Metadata.Name,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			Attributes:    attributes,
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
}
