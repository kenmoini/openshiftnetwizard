package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
)

func vmwareScreen(_ fyne.Window) fyne.CanvasObject {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Ensure your vCenter Root Certificate Authority is added to your system's trust stores", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)
	return container.NewCenter(content)
}

func makeVMWareConnectionTab(win fyne.Window) fyne.CanvasObject {
	server := widget.NewEntry()
	server.SetPlaceHolder("https://vcenter.example.com")
	server.Validator = validation.NewRegexp(`https:\/\/[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#()?&//=]*)`, "not a valid url")

	username := widget.NewEntry()
	username.SetPlaceHolder("administrator@vsphere.local")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	validateSSLCheck := widget.NewCheck("Skip TLS Validation?", func(bool) {})
	validateSSLCheck.SetChecked(true)

	// Check for environment variables
	envVCenterEndpoint, envVCenterEndpointExists := os.LookupEnv("VCENTER_SERVER")
	if envVCenterEndpointExists {
		server.SetText(envVCenterEndpoint)
	}
	envVCenterUsername, envVCenterUsernameExists := os.LookupEnv("VCENTER_USERNAME")
	if envVCenterUsernameExists {
		username.SetText(envVCenterUsername)
	}
	envVCenterPassword, envVCenterPasswordExists := os.LookupEnv("VCENTER_PASSWORD")
	if envVCenterPasswordExists {
		password.SetText(envVCenterPassword)
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Server", Widget: server, HintText: ""},
			{Text: "Username", Widget: username, HintText: ""},
			{Text: "Password", Widget: password, HintText: ""},
		},
		OnCancel: func() {
			server.SetText("")
			username.SetText("")
			password.SetText("")
			vCenterClient = nil
		},
		OnSubmit: func() {
			ctx := context.Background()

			// Parse the url and add the username and password
			serverURL, err := soap.ParseURL(server.Text)
			if err != nil {
				fmt.Println(err)
			}
			serverURL.User = url.UserPassword(username.Text, password.Text)

			// Share govc's session cache
			s := &cache.Session{
				URL:      serverURL,
				Insecure: validateSSLCheck.Checked,
			}

			// Create a new client
			c := new(vim25.Client)
			err = s.Login(ctx, c, nil)
			if err != nil {
				dialog.ShowError(err, win)
			} else {
				rc := new(rest.Client)
				err = s.Login(ctx, rc, nil)
				if err != nil {
					dialog.ShowError(err, win)
				} else {
					dialog.ShowInformation("Success", "Authenticated with vCenter REST API successfully", win)
				}
				vCenterClient = c
			}
		},
	}
	form.Append("Security", validateSSLCheck)
	return form
}

func makeVMWareNetworksTab(win fyne.Window) fyne.CanvasObject {
	networkList := widget.NewMultiLineEntry()

	return container.NewVScroll(container.NewVBox(
		widget.NewButton("List Networks", func() {
			if vCenterClient == nil {
				dialog.ShowInformation("Information", "Please connect to a vCenter server first", win)
			} else {
				ctx := context.Background()
				networks, err := listNetworks(ctx, vCenterClient)
				if err != nil {
					dialog.ShowError(err, win)
				} else {
					networkList.SetText(fmt.Sprintf("%v", networks))
				}
			}
		}),
		layout.NewSpacer(),
		networkList,
	))
}

func makeVMWareInventoryTab(win fyne.Window) fyne.CanvasObject {
	if vCenterClient == nil {
		dialog.ShowInformation("Information", "Please connect to a vCenter server first", win)
		return container.NewVBox()
	} else {
		inventoryList := widget.NewMultiLineEntry()
		vcenterServer := vCenterClient.URL().Host
		//data := make(map[string][]string)

		//tree := widget.NewTreeWithStrings(data)
		tree := widget.NewTree(
			func(id widget.TreeNodeID) []widget.TreeNodeID {
				switch id {
				case "":
					return []widget.TreeNodeID{vcenterServer, "a", "b", "c"}
				case "a":
					return []widget.TreeNodeID{"a1", "a2"}
				}
				return []string{}
			},
			func(id widget.TreeNodeID) bool {
				return id == "" || id == "a"
			},
			func(branch bool) fyne.CanvasObject {
				if branch {
					return widget.NewLabel("Branch template")
				}
				return widget.NewLabel("Leaf template")
			},
			func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
				text := id
				if branch {
					text += " (branch)"
				}
				o.(*widget.Label).SetText(text)
			})

		return container.NewBorder(container.NewVBox(
			widget.NewButton("List Inventory", func() {
				ctx := context.Background()
				inventory, err := getFolders(ctx, vCenterClient)
				if err != nil {
					dialog.ShowError(err, win)
				} else {
					i, err := json.Marshal(inventory)
					//inventoryList.SetText(fmt.Sprintf("%+v", inventory))
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(string(i))
						selectedInventory := makeInventoryTree(inventory)
						fmt.Printf("%+v", selectedInventory)
					}
					inventoryList.SetText(string(i))
					//fmt.Printf("%+v", inventory)
				}
			})), nil, nil, nil, tree)

	}
}

func makeInventoryTree(inventory []mo.Folder) map[string][]InventoryItem {
	data := make(map[string][]InventoryItem)
	for _, item := range inventory {
		data[item.Name] = append(data[item.Name], InventoryItem{
			Id:   item.Self.Value,
			Name: item.Name,
			Type: item.Self.Type,
		})
	}
	return data
}
