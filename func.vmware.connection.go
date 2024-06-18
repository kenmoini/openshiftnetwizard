// This file is likely obsolete and can be safely removed.
package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"net"
	"net/url"
	neturl "net/url"
	"sync"

	klog "k8s.io/klog/v2"

	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/sts"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

var vcenterUsername string
var vcenterPassword string
var vcenterEndpoint string
var vcenterInsecureFlag bool

var vCenterClient *vim25.Client

const (
	userAgentName            = "openshift-network-wizard"
	RoundTripperDefaultCount = 3
)

var (
	clientLock sync.Mutex
)

// Connect makes connection to vCenter and sets VSphereConnection.Client.
// If connection.Client is already set, it obtains the existing user session.
// if user session is not valid, connection.Client will be set to the new client.
func (connection *VSphereConnection) Connect(ctx context.Context) error {
	var err error
	clientLock.Lock()
	defer clientLock.Unlock()

	if connection.Client == nil {
		connection.Client, err = connection.NewClient(ctx)
		if err != nil {
			klog.Errorf("Failed to create govmomi client. err: %+v", err)
			return err
		}
		return nil
	}
	m := session.NewManager(connection.Client)
	userSession, err := m.UserSession(ctx)
	if err != nil {
		klog.Errorf("Error while obtaining user session. err: %+v", err)
		return err
	}
	if userSession != nil {
		return nil
	}
	klog.Warning("Creating new client session since the existing session is not valid or not authenticated")

	connection.Client, err = connection.NewClient(ctx)
	if err != nil {
		klog.Errorf("Failed to create govmomi client. err: %+v", err)
		return err
	}
	return nil
}

// Signer returns an sts.Signer for use with SAML token auth if connection is configured for such.
// Returns nil if username/password auth is configured for the connection.
func (connection *VSphereConnection) Signer(ctx context.Context, client *vim25.Client) (*sts.Signer, error) {
	// TODO: Add separate fields for certificate and private-key.
	// For now we can leave the config structs and validation as-is and
	// decide to use LoginByToken if the username value is PEM encoded.
	b, _ := pem.Decode([]byte(connection.Username))
	if b == nil {
		return nil, nil
	}

	cert, err := tls.X509KeyPair([]byte(connection.Username), []byte(connection.Password))
	if err != nil {
		klog.Errorf("Failed to load X509 key pair. err: %+v", err)
		return nil, err
	}

	tokens, err := sts.NewClient(ctx, client)
	if err != nil {
		klog.Errorf("Failed to create STS client. err: %+v", err)
		return nil, err
	}

	req := sts.TokenRequest{
		Certificate: &cert,
		Delegatable: true,
	}

	signer, err := tokens.Issue(ctx, req)
	if err != nil {
		klog.Errorf("Failed to issue SAML token. err: %+v", err)
		return nil, err
	}

	return signer, nil
}

// login calls SessionManager.LoginByToken if certificate and private key are configured,
// otherwise calls SessionManager.Login with user and password.
func (connection *VSphereConnection) login(ctx context.Context, client *vim25.Client) error {
	m := session.NewManager(client)
	connection.credentialsLock.Lock()
	defer connection.credentialsLock.Unlock()

	signer, err := connection.Signer(ctx, client)
	if err != nil {
		return err
	}

	if signer == nil {
		klog.V(3).Infof("SessionManager.Login with username %q", connection.Username)
		return m.Login(ctx, neturl.UserPassword(connection.Username, connection.Password))
	}

	klog.V(3).Infof("SessionManager.LoginByToken with certificate %q", connection.Username)

	header := soap.Header{Security: signer}

	return m.LoginByToken(client.WithHeader(ctx, header))
}

// Logout calls SessionManager.Logout for the given connection.
func (connection *VSphereConnection) Logout(ctx context.Context) {
	m := session.NewManager(connection.Client)
	if err := m.Logout(ctx); err != nil {
		klog.Errorf("Logout failed: %s", err)
	}
}

// NewClient creates a new govmomi client for the VSphereConnection obj
func (connection *VSphereConnection) NewClient(ctx context.Context) (*vim25.Client, error) {
	url, err := soap.ParseURL(net.JoinHostPort(connection.Hostname, connection.Port))
	if err != nil {
		klog.Errorf("Failed to parse URL: %s. err: %+v", url, err)
		return nil, err
	}

	sc := soap.NewClient(url, connection.Insecure)

	if ca := connection.CACert; ca != "" {
		if err := sc.SetRootCAs(ca); err != nil {
			return nil, err
		}
	}

	tpHost := connection.Hostname + ":" + connection.Port
	sc.SetThumbprint(tpHost, connection.Thumbprint)

	client, err := vim25.NewClient(ctx, sc)
	if err != nil {
		klog.Errorf("Failed to create new client. err: %+v", err)
		return nil, err
	}
	client.UserAgent = userAgentName
	err = connection.login(ctx, client)
	if err != nil {
		return nil, err
	}

	if connection.RoundTripperCount == 0 {
		connection.RoundTripperCount = RoundTripperDefaultCount
	}
	client.RoundTripper = vim25.Retry(client.RoundTripper, vim25.TemporaryNetworkError(int(connection.RoundTripperCount)))
	return client, nil
}

// UpdateCredentials updates username and password.
// Note: Updated username and password will be used when there is no session active
func (connection *VSphereConnection) UpdateCredentials(username string, password string) {
	connection.credentialsLock.Lock()
	defer connection.credentialsLock.Unlock()
	connection.Username = username
	connection.Password = password
}

// Logout closes existing connections to remote vCenter endpoints.
func (connMgr *ConnectionManager) Logout() {
	for _, vsphereIns := range connMgr.VsphereInstanceMap {
		connMgr.Lock()
		c := vsphereIns.Conn.Client
		connMgr.Unlock()
		if c != nil {
			vsphereIns.Conn.Logout(context.TODO())
		}
	}
}

// Verify validates the configuration by attempting to connect to the
// configured, remote vCenter endpoints.
func (connMgr *ConnectionManager) Verify() error {
	for _, vcInstance := range connMgr.VsphereInstanceMap {
		err := connMgr.Connect(context.Background(), vcInstance)
		if err == nil {
			klog.V(3).Infof("vCenter connect %s succeeded.", vcInstance.Cfg.VCenterIP)
		} else {
			klog.Errorf("vCenter %s failed. Err: %q", vcInstance.Cfg.VCenterIP, err)
			return err
		}
	}
	return nil
}

// VerifyWithContext is the same as Verify but allows a Go Context
// to control the lifecycle of the connection event.
func (connMgr *ConnectionManager) VerifyWithContext(ctx context.Context) error {
	for _, vcInstance := range connMgr.VsphereInstanceMap {
		err := connMgr.Connect(ctx, vcInstance)
		if err == nil {
			klog.V(3).Infof("vCenter connect %s succeeded.", vcInstance.Cfg.VCenterIP)
		} else {
			klog.Errorf("vCenter %s failed. Err: %q", vcInstance.Cfg.VCenterIP, err)
			return err
		}
	}
	return nil
}

// APIVersion returns the version of the vCenter API
func (connMgr *ConnectionManager) APIVersion(vcInstance *VSphereInstance) (string, error) {
	if err := connMgr.Connect(context.Background(), vcInstance); err != nil {
		return "", err
	}

	return vcInstance.Conn.Client.ServiceContent.About.ApiVersion, nil
}

// Connect connects to vCenter with existing credentials
// If credentials are invalid:
//  1. It will fetch credentials from credentialManager
//  2. Update the credentials
//  3. Connects again to vCenter with fetched credentials
func (connMgr *ConnectionManager) Connect(ctx context.Context, vcInstance *VSphereInstance) error {
	connMgr.Lock()
	defer connMgr.Unlock()

	err := vcInstance.Conn.Connect(ctx)
	if err == nil {
		return nil
	}

	// if !vclib.IsInvalidCredentialsError(err) || connMgr.credentialManagers == nil {
	// 	klog.Errorf("Cannot connect to vCenter with err: %v", err)
	// 	return err
	// }

	// klog.V(2).Infof("Invalid credentials. Fetching credentials from secrets. vcServer=%s credentialHolder=%s",
	// 	vcInstance.Cfg.VCenterIP, vcInstance.Cfg.SecretRef)

	// credMgr := connMgr.credentialManagers[vcInstance.Cfg.SecretRef]
	// if credMgr == nil {
	// 	klog.Errorf("Unable to find credential manager for vcServer=%s credentialHolder=%s", vcInstance.Cfg.VCenterIP, vcInstance.Cfg.SecretRef)
	// 	return ErrUnableToFindCredentialManager
	// }
	// credentials, err := credMgr.GetCredential(vcInstance.Cfg.VCenterIP)
	// if err != nil {
	// 	klog.Error("Failed to get credentials from Secret Credential Manager with err:", err)
	// 	return err
	// }

	vcInstance.Conn.UpdateCredentials(vcenterUsername, vcenterPassword)
	return vcInstance.Conn.Connect(ctx)
}

// =====================================================================================================================

// NewVCenterClient creates a vim25.Client for use to connect to a vCenter server
func NewVCenterClient(ctx context.Context) (*vim25.Client, error) {
	// Parse URL from string
	u, err := soap.ParseURL(vcenterEndpoint)
	if err != nil {
		return nil, err
	}

	// Override username and/or password as required
	u.User = url.UserPassword(vcenterUsername, vcenterPassword)

	// Share govc's session cache
	s := &cache.Session{
		URL:      u,
		Insecure: vcenterInsecureFlag,
	}

	c := new(vim25.Client)
	err = s.Login(ctx, c, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}
