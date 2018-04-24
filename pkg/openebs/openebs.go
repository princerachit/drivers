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
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/golang/glog"
)

type openEbs struct {
	driver   *csicommon.CSIDriver
	endpoint string

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability


}

const (
	driverName = "openebs"
)

var (
	version = "0.0.1"
)

func GetOpenEbsDriver() *openEbs {
	return &openEbs{}
}

func (oe *openEbs) Run(nodeID, endpoint string) {
	glog.Infof("Driver: %v ", driverName)

	// Initialize with default driver
	oe.driver = csicommon.NewCSIDriver(driverName, version, nodeID)
	if oe.driver == nil {
		glog.Fatalln("Failed to initialize CSI Driver.")
	}

	// Add capabilities
	oe.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	oe.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	oe.ids = NewIdentityServer(oe.driver)
	oe.ns = NewNodeServer(oe.driver)
	oe.cs = NewControllerServer(oe.driver)

	// Create GRPC server and starts it
	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(endpoint, oe.ids, oe.cs, oe.ns)
	s.Wait()
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func NewControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}

func NewNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
	}
}
