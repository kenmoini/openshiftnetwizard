package main

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

func listNetworks(ctx context.Context, vcenterClient *vim25.Client) ([]mo.Network, error) {
	m := view.NewManager(vcenterClient)

	v, err := m.CreateContainerView(ctx, vcenterClient.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		fmt.Println(err)
	}

	defer v.Destroy(ctx)

	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Network.html
	var networks []mo.Network
	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	if err != nil {
		return nil, err
	}

	for _, net := range networks {
		fmt.Printf("%s: %s\n", net.Name, net.Reference())
		fmt.Printf("%s: %v\n", net.Name, net)
	}
	return networks, nil
}
