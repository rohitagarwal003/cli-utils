// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cli-utils/pkg/inventory"
)

// InventoryPolicyFilter implements ValidationFilter interface to determine
// if an object should be pruned (deleted) because of the InventoryPolicy
// and if the objects owning inventory identifier matchs the inventory id.
type InventoryPolicyFilter struct {
	Inv       inventory.InventoryInfo
	InvPolicy inventory.InventoryPolicy
}

// Name returns a filter identifier for logging.
func (ipf InventoryPolicyFilter) Name() string {
	return "InventoryPolictyFilter"
}

// Filter returns true if the passed object should NOT be pruned (deleted)
// because the "prevent remove" annotation is present; otherwise returns
// false. Never returns an error.
func (ipf InventoryPolicyFilter) Filter(obj *unstructured.Unstructured) (bool, error) {
	// Check the inventory id "match" and the adopt policy to determine
	// if an object should be pruned (deleted).
	if !inventory.CanPrune(ipf.Inv, obj, ipf.InvPolicy) {
		return true, nil
	}
	return false, nil
}
