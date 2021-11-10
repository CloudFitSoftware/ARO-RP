package network

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
)

// PublicIPAddressesClientAddons contains addons for PublicIPAddressesClient
type PublicIPAddressesClientAddons interface {
	DeleteAndWait(ctx context.Context, resourceGroupName string, publicIPAddressName string) (err error)
}

func (c *publicIPAddressesClient) DeleteAndWait(ctx context.Context, resourceGroupName string, publicIPAddressName string) error {
	future, err := c.Delete(ctx, resourceGroupName, publicIPAddressName)
	if err != nil {
		return err
	}

	return future.WaitForCompletionRef(ctx, c.Client)
}
