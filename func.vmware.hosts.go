package main

import (
	"context"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

func getHosts(ctx context.Context, vcenterClient *vim25.Client) ([]mo.HostSystem, error) {
	m := view.NewManager(vcenterClient)

	v, err := m.CreateContainerView(ctx, vcenterClient.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all hosts
	// Reference: https://dp-downloads.broadcom.com/api-content/apis/API_VWSA_001/7.0/html/vim.HostSystem.html
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		return nil, err
	}
	return hss, nil
}
