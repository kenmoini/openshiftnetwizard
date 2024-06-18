package main

import (
	"sync"

	"github.com/vmware/govmomi/vim25"
	vcfg "k8s.io/cloud-provider-vsphere/pkg/common/config"
	vclib "k8s.io/cloud-provider-vsphere/pkg/common/vclib"
)

// VSphereConnection contains information for connecting to vCenter
type VSphereConnection struct {
	Client            *vim25.Client
	Username          string
	Password          string
	Hostname          string
	Port              string
	CACert            string
	Thumbprint        string
	Insecure          bool
	RoundTripperCount uint
	credentialsLock   sync.Mutex
}

// ConnectionManager encapsulates vCenter connections
type ConnectionManager struct {
	sync.Mutex

	// The k8s client init from the cloud provider service account
	//client clientset.Interface

	// Maps the VC server to VSphereInstance
	VsphereInstanceMap map[string]*VSphereInstance
	// CredentialManager per VC
	// The global CredentialManager will have an entry in this map with the key of "Global"
	//credentialManagers map[string]*cm.CredentialManager
	// InformerManagers per VC
	// The global InformerManager will have an entry in this map with the key of "Global"
	//informerManagers map[string]*k8s.InformerManager
}

// VSphereInstance represents a vSphere instance where one or more kubernetes nodes are running.
type VSphereInstance struct {
	Conn *vclib.VSphereConnection
	Cfg  *vcfg.VirtualCenterConfig
}

// VMDiscoveryInfo contains VM info about a discovered VM
type VMDiscoveryInfo struct {
	TenantRef  string
	DataCenter *vclib.Datacenter
	VM         *vclib.VirtualMachine
	VcServer   string
	UUID       string
	NodeName   string
}

// FcdDiscoveryInfo contains FCD info about a discovered FCD
type FcdDiscoveryInfo struct {
	TenantRef  string
	DataCenter *vclib.Datacenter
	FCDInfo    *vclib.FirstClassDiskInfo
	VcServer   string
}

// ListDiscoveryInfo represents a VC/DC pair
type ListDiscoveryInfo struct {
	TenantRef  string
	VcServer   string
	DataCenter *vclib.Datacenter
}

// ZoneDiscoveryInfo contains VC+DC info based on a given zone
type ZoneDiscoveryInfo struct {
	TenantRef  string
	DataCenter *vclib.Datacenter
	VcServer   string
}
