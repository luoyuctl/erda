// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package localvolume

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v6/controller"

	"github.com/erda-project/erda/modules/volume-provisioner/exec"
	"github.com/erda-project/erda/pkg/strutil"
)

type localVolumeProvisioner struct {
	client      kubernetes.Interface
	restClient  rest.Interface
	config      *rest.Config
	cmdExecutor *exec.CmdExecutor
}

func NewLocalVolumeProvisioner(config *rest.Config, client kubernetes.Interface, namespace string) *localVolumeProvisioner {
	return &localVolumeProvisioner{
		client:      client,
		restClient:  client.CoreV1().RESTClient(),
		config:      config,
		cmdExecutor: exec.NewCmdExecutor(config, client, namespace),
	}
}

func (p *localVolumeProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	logrus.Infof("Start provisioning local volume: options: %v", options)

	if options.SelectedNode == nil {
		err := fmt.Errorf("Not provide selectedNode in provisionOptions")
		logrus.Error(err)
		return nil, controller.ProvisioningFinished, err
	}
	volPathOnHost, err := volumeRealPath(&options, options.PVName)
	if err != nil {
		return nil, controller.ProvisioningFinished, err
	}
	volPath, err := volumePath(&options, options.PVName)
	if err != nil {
		return nil, controller.ProvisioningFinished, err
	}
	nodeSelector := fmt.Sprintf("kubernetes.io/hostname=%s", options.SelectedNode.Name)
	if err := p.cmdExecutor.OnNodesPods(fmt.Sprintf("mkdir -p %s", volPath),
		metav1.ListOptions{
			LabelSelector: nodeSelector,
		}, metav1.ListOptions{
			LabelSelector: "app=volume-provisioner",
		}); err != nil {
		return nil, controller.ProvisioningFinished, err
	}

	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				Local: &v1.LocalVolumeSource{
					Path: volPathOnHost,
				},
			},
			NodeAffinity: &v1.VolumeNodeAffinity{
				Required: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{options.SelectedNode.Name},
								},
							},
						},
					},
				},
			},
		},
	}, controller.ProvisioningFinished, nil
}

var (
	hostPathOnce                    sync.Once
	hostPathErr                     error
	hostPathVolumePrefixInContainer string // = "/hostfs/data/localvolume"
)

func volumePath(options *controller.ProvisionOptions, pvname string) (string, error) {
	mountPath, err := findLocalVolumeMountedPath(options)
	if err != nil {
		return "", err
	}
	return strutil.JoinPath(mountPath, "localvolume", pvname), nil
}

func volumeRealPath(options *controller.ProvisionOptions, pvname string) (string, error) {
	mountPath, err := findLocalVolumeMountedPath(options)
	if err != nil {
		return "", err
	}
	return strutil.JoinPath("/", strutil.TrimPrefixes(mountPath, "/hostfs"), "localvolume", pvname), nil
}

func findLocalVolumeMountedPath(options *controller.ProvisionOptions) (string, error) {
	if options.StorageClass.Parameters != nil && options.StorageClass.Parameters["hostpath"] != "" {
		return strutil.JoinPath("/hostfs", options.StorageClass.Parameters["hostpath"]), nil
	}
	hostPathOnce.Do(func() {
		mountpoint, err := DiscoverMountPoint()
		if err != nil {
			hostPathErr = err
		}
		hostPathVolumePrefixInContainer = mountpoint
	})
	return hostPathVolumePrefixInContainer, hostPathErr
}