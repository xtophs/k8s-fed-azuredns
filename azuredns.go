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

// azuredns is the implementation of pkg/dnsprovider interface for azuredns
package azuredns

import (
	"fmt"
	"io"

	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

const (
	ProviderName = "azure-azuredns"
)

type Config struct {
	Global struct {
		subscriptionID string `gcfg:"subscriptionID"`
		clientID       string `gcfg:"clientID"`
		secret         string `gcfg:"secret"`
		tenantID       string `gcfg:"tenantID"`
	}
}

func init() {
	dnsprovider.RegisterDnsProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		glog.V(5).Infof("Registering Azure DNS provider\n")
		fmt.Printf("Registering Azure DNS provider\n")
		return newazuredns(config)
	})
}

// newazuredns creates a new instance of an AWS azuredns DNS Interface.
func newazuredns(config io.Reader) (*Interface, error) {

	// TODO: create config struct
	// This is test data
	azConfig := NewAzureDNSConfig("ffa90503-6fe8-4ab5-9bf1-059f81a6d8bb",
		"delete-dns",
		"66841164-1e0e-4ffc-a0d2-ce36f95ec41d",
		"eebec4ca-c175-45dc-b763-943607ce4838",
		"9f4ba4e0-be35-4821-9aa6-4caadfaba4fa")
	glog.V(4).Infof("Azure DNS config: %v", config)

	return NewClients(*azConfig), nil
}
