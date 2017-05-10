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
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
)

// Compile time check for interface adherence
var _ dnsprovider.ResourceRecordSet = ResourceRecordSet{}

type ResourceRecordSet struct {
	impl   *dns.RecordSet
	rrsets *ResourceRecordSets
}

func (rrset ResourceRecordSet) Name() string {
	return *rrset.impl.Name
}

func (rrset ResourceRecordSet) Rrdatas() []string {
	// Sigh - need to unpack the strings out of the azuredns ResourceRecords
	result := recordSetPropertiesToRrDatas(rrset.impl)

	return result
}

func (rrset ResourceRecordSet) Ttl() int64 {
	// same behavior as the route 53 provider
	if rrset.impl.TTL != nil {
		return *rrset.impl.TTL
	}
	return 0
}

func (rrset ResourceRecordSet) Type() rrstype.RrsType {
	return rrstype.RrsType(strings.TrimPrefix(*rrset.impl.Type, "Microsoft.Network/dnszones/"))
}

// azurednsResourceRecordSet returns the azuredns ResourceRecordSet object for the ResourceRecordSet
// This is a "back door" that allows for limited access to the ResourceRecordSet,
// without having to requery it.
// Using this method should be avoided where possible; instead prefer to add functionality
// to the cross-provider ResourceRecordSet interface.
func (rrset ResourceRecordSet) azurednsResourceRecordSet() *dns.RecordSet {
	return rrset.impl
}
