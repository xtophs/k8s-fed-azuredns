/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azuredns

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"github.com/xtophs/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

// Compile time check for interface adherence
var _ dnsprovider.Interface = Interface{}

var _ AzureDNSAPI = Clients{}

type Interface struct {
	service *AzureDNSAPI
}

type AzureDNSAPI interface {
	GetZonesClient() *dns.ZonesClient
	GetRecordSetsClient() *dns.RecordSetsClient
	GetResourceGroupName() string
}

type Clients struct {
	rc      dns.RecordSetsClient
	conf    Config
	confMap map[string]string
}

// func NewAzureDNSConfig(subId string, resGroup string, tenantId string, clientId string, secret string) *azureDnsConfig {
// 	return &azureDnsConfig{
// 		subId,
// 		resGroup,
// 		tenantId,
// 		clientId,
// 		secret,
// 	}
// }

func (a Clients) GetZonesClient() *dns.ZonesClient {
	zc := dns.NewZonesClient(a.conf.Global.SubscriptionID)
	spt, err := helpers.NewServicePrincipalTokenFromCredentials(a.confMap, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		glog.Fatalf("Error authenticating to Azure DNS: %v", err)
		return nil
	}

	zc.Authorizer = autorest.NewBearerAuthorizer(spt)
	return &zc
}

func (a Clients) GetRecordSetsClient() *dns.RecordSetsClient {
	rc := dns.NewRecordSetsClient(a.conf.Global.SubscriptionID)
	spt, err := helpers.NewServicePrincipalTokenFromCredentials(a.confMap, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		glog.Fatalf("Error authenticating to Azure DNS: %v", err)
		return nil
	}

	rc.Authorizer = autorest.NewBearerAuthorizer(spt)
	return &rc
}

func (a Clients) GetResourceGroupName() string {
	return a.conf.Global.ResourceGroup
}

// New builds an Interface, with a specified azurednsAPI implementation.
// This is useful for testing purposes, but also if we want an instance with with custom AWS options.
func New(service *AzureDNSAPI) *Interface {
	return &Interface{service}
}

func NewClients(config Config) *Interface {

	// TODO
	c := map[string]string{
		"AZURE_CLIENT_ID":       config.Global.ClientID,
		"AZURE_CLIENT_SECRET":   config.Global.Secret,
		"AZURE_SUBSCRIPTION_ID": config.Global.SubscriptionID,
		"AZURE_TENANT_ID":       config.Global.TenantID}

	if err := checkEnvVar(&c); err != nil {
		glog.Fatalf("Error in config: %v", err)
		return nil
	}

	var clients *Clients
	clients = &Clients{}

	glog.V(4).Infof("Created Azure DNS clients for subscription: %s", config.Global.SubscriptionID)

	clients.conf = config
	clients.confMap = c
	var api AzureDNSAPI = clients
	return &Interface{&api}
}

func (i Interface) Zones() (zones dnsprovider.Zones, supported bool) {
	return Zones{&i}, true
}

func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for varName, value := range *envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}
