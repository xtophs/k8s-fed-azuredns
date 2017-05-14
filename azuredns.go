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
	//"bytes"
	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	gcfg "gopkg.in/gcfg.v1"

)

const (
	ProviderName = "azure-azuredns"
)

type Config struct {
	Global struct {
		SubscriptionID string `gcfg:"subscription-id"`
		ClientID       string `gcfg:"client-id"`
		Secret         string `gcfg:"secret"`
		TenantID       string `gcfg:"tenant-id"`
		ResourceGroup  string `gcfg:"resourceGroup"`
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

	var azConfig Config
	if err := gcfg.ReadInto(&azConfig, config); err != nil {
		glog.Errorf("Couldn't read config: %v", err)
		return nil, err
		}
		
	// TODO
	// check config
	glog.V(4).Infof("Azure DNS config: %v", azConfig)

	return NewClients(azConfig), nil
}
