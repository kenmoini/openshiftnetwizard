package main

import (
	"context"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

func getFolders(ctx context.Context, vcenterClient *vim25.Client) ([]mo.Folder, error) {
	m := view.NewManager(vcenterClient)

	v, err := m.CreateContainerView(ctx, vcenterClient.ServiceContent.RootFolder, []string{"Folder"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all folders
	// Reference: https://dp-downloads.broadcom.com/api-content/apis/API_VWSA_001/7.0/html/vim.Folder.html
	var folders []mo.Folder
	err = v.Retrieve(ctx, []string{"Folder"}, []string{"namespace", "childType", "childEntity", "parent", "name", "namespace", "value"}, &folders)
	if err != nil {
		return nil, err
	}
	return folders, nil
}

type InventoryItem struct {
	Id       string
	Name     string
	Type     string
	Children []InventoryItem
}
