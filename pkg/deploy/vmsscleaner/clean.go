package vmsscleaner

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/Azure/ARO-RP/pkg/util/azureclient/mgmt/compute"
)

type Interface interface {
	RemoveFailedNewScaleset(ctx context.Context, rgName, vmssToDelete string) (retry bool)
	UpdateVMSSProbes(ctx context.Context, rgName string) (retry bool)
}

type cleaner struct {
	log  *logrus.Entry
	vmss compute.VirtualMachineScaleSetsClient
}

func New(log *logrus.Entry, vmss compute.VirtualMachineScaleSetsClient) Interface {
	return &cleaner{
		log:  log,
		vmss: vmss,
	}
}

// RemoveFailedNewScaleset attempts to delete the new VMSS from the current deployment if necessary and returns
// with whether or not deployment should be retried
func (c *cleaner) RemoveFailedNewScaleset(ctx context.Context, rgName, vmssToDelete string) (retry bool) {
	scalesets, err := c.vmss.List(ctx, rgName)
	if err != nil {
		c.log.Warn(err)
		return false
	}

	switch len(scalesets) {
	case 0:
		// If there are no scalesets, can retry again without worrying about deletion
		return true
	case 1:
		// If there is a single scaleset, can retry iff the name differs from vmssToDelete
		return *scalesets[0].Name != vmssToDelete
	}

	for _, vmss := range scalesets {
		if *vmss.Name != vmssToDelete {
			// If it's not the newly deployed VMSS, skip it.
			continue
		}

		c.log.Printf("deleting failed or unhealthy scaleset %s", vmssToDelete)
		err = c.vmss.DeleteAndWait(ctx, rgName, vmssToDelete)
		if err != nil {
			c.log.Warn(err)
			return false // If deletion failed, vmssToDelete still exists. Don't retry.
		}
	}
	// If vmssToDelete was found and deleted successfully, deployment can be retried
	// If it was not returned from List, assume it does not exist and that deployment can be retried.
	return true
}

// UpdateVMSSProbes attempts to remove the probes references so we can update the load balancer rules
func (c *cleaner) UpdateVMSSProbes(ctx context.Context, rgName string) (retry bool) {
	scalesets, err := c.vmss.List(ctx, rgName)
	if err != nil {
		c.log.Warn(err)
		return false
	}

	for _, vmss := range scalesets {
		name := *vmss.Name
		if !strings.HasPrefix(name, "gateway-vmss-") {
			continue
		}

		//removes the network profile healthprobe
		vmss.VirtualMachineProfile.NetworkProfile.HealthProbe = nil
		c.log.Printf("removing the probe reference from the vmss %s", name)
		err = c.vmss.CreateOrUpdateAndWait(ctx, rgName, name, vmss)
		if err != nil {
			c.log.Warn(err)
			return false // If update failed, gateway vmss still has a reference to the probe
		}
	}
	return true
}
