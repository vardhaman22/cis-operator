/*
Copyright 2024 Rancher Labs, Inc.

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

// Code generated by main. DO NOT EDIT.

package v1

import (
	v1 "github.com/rancher/cis-operator/pkg/apis/cis.cattle.io/v1"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/v3/pkg/schemes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	schemes.Register(v1.AddToScheme)
}

type Interface interface {
	ClusterScan() ClusterScanController
	ClusterScanBenchmark() ClusterScanBenchmarkController
	ClusterScanProfile() ClusterScanProfileController
	ClusterScanReport() ClusterScanReportController
}

func New(controllerFactory controller.SharedControllerFactory) Interface {
	return &version{
		controllerFactory: controllerFactory,
	}
}

type version struct {
	controllerFactory controller.SharedControllerFactory
}

func (c *version) ClusterScan() ClusterScanController {
	return NewClusterScanController(schema.GroupVersionKind{Group: "cis.cattle.io", Version: "v1", Kind: "ClusterScan"}, "clusterscans", false, c.controllerFactory)
}
func (c *version) ClusterScanBenchmark() ClusterScanBenchmarkController {
	return NewClusterScanBenchmarkController(schema.GroupVersionKind{Group: "cis.cattle.io", Version: "v1", Kind: "ClusterScanBenchmark"}, "clusterscanbenchmarks", false, c.controllerFactory)
}
func (c *version) ClusterScanProfile() ClusterScanProfileController {
	return NewClusterScanProfileController(schema.GroupVersionKind{Group: "cis.cattle.io", Version: "v1", Kind: "ClusterScanProfile"}, "clusterscanprofiles", false, c.controllerFactory)
}
func (c *version) ClusterScanReport() ClusterScanReportController {
	return NewClusterScanReportController(schema.GroupVersionKind{Group: "cis.cattle.io", Version: "v1", Kind: "ClusterScanReport"}, "clusterscanreports", false, c.controllerFactory)
}
