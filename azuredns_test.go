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
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	//"strconv"

	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"github.com/Azure/go-autorest/autorest/to"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/tests"

)

func newTestInterface() (dnsprovider.Interface, error) {
	p := dnsprovider.RegisteredDnsProviders()
	for _, s := range p {
		fmt.Printf("Found registered provider %s\n", s)
	}

	// Use this to test the real cloud service.
	i, err := newAzureDNSProvider()

	// Use this to stub out the entire cloud service
	//i, err :=  newFakeInterface() 
	return i, err
}

func newAzureDNSProvider() (dnsprovider.Interface, error) {
	i, err := dnsprovider.GetDnsProvider(ProviderName, strings.NewReader("\n[Global]\nsubscription-id = ffa90503-6fe8-4ab5-9bf1-059f81a6d8bb\ntenant-id = 66841164-1e0e-4ffc-a0d2-ce36f95ec41d\nclient-id = eebec4ca-c175-45dc-b763-943607ce4838\nsecret = 9f4ba4e0-be35-4821-9aa6-4caadfaba4fa\nresourceGroup = delete-dns"))
	if i == nil || err != nil {
		fmt.Printf("DNS provider %s not registered", ProviderName)
		os.Exit(1)
	}

	return i, nil
}

func newFakeInterface() (dnsprovider.Interface, error) {
	zoneName := "example.com"
	api := NewAzureDNSAPIStub()
	var svc AzureDNSAPI
	svc = api

	iface := New(&svc)

	// Add a fake zone to test against.
	var params dns.Zone = dns.Zone{
		Name:     &zoneName,
		Location: to.StringPtr("global"),
	}

	if iface.service == nil {
		fmt.Printf("Service interface should not be null")
		os.Exit(1)
	}

	_, err := svc.CreateOrUpdateZone( zoneName, params, "", "")
	if err != nil {
		return nil, err
	}
	return iface, nil
}

var interface_ dnsprovider.Interface

func TestMain(m *testing.M) {
	fmt.Printf("Parsing flags.\n")
	flag.Parse()
	var err error
	fmt.Printf("Getting new test interface.\n")
	interface_, err = newTestInterface()
	if err != nil {
		fmt.Printf("Error creating interface: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Running tests...\n")
	os.Exit(m.Run())
}

// zones returns the zones interface for the configured dns provider account/project,
// or fails if it can't be found
func zones(t *testing.T) dnsprovider.Zones {
	zonesInterface, supported := interface_.Zones()
	if !supported {
		t.Fatalf("Zones interface not supported by interface %v", interface_)
	} else {
		t.Logf("Got zones %v\n", zonesInterface)
	}
	return zonesInterface
}

// firstZone returns the first zone for the configured dns provider account/project,
// or fails if it can't be found
func firstZone(t *testing.T) dnsprovider.Zone {
	t.Logf("Getting zones")
	z := zones(t)

	zones, err := z.List()
	if err != nil {
		t.Fatalf("Failed to list zones: %v", err)
	} else {
		t.Logf("Got zone list: %v with %i zones\n", zones, len(zones))
	}
	if len(zones) < 1 {
		t.Fatalf("Zone listing returned %d, expected >= %d", len(zones), 1)
	} else {
		t.Logf("Got at least 1 zone in list:%v\n", zones[0])
		t.Logf("Got at least 1 zone in list:%s\n", zones[0].Name())
	}
	return zones[0]
}

func xtophsZone(t *testing.T) dnsprovider.Zone {
	t.Logf("Getting xtoph zones")
	z := zones(t)

	zones, err := z.List()
	var zone dnsprovider.Zone
	for _, z := range zones {
		if z.Name() == "xtophs.com" {
			zone = z
			break
		} 
	}
	if err != nil {
		t.Fatalf("Failed to get xtoph zone: %v", err)
	} else {
		t.Logf("Got zone list: %v with %i zones\n", zones, len(zones))
	}

	if z == nil {
		t.Fatalf("Failed to get xtoph zone: %v", err)		
	}
	return zone
}

/* rrs returns the ResourceRecordSets interface for a given zone */
func rrs(t *testing.T, zone dnsprovider.Zone) (r dnsprovider.ResourceRecordSets) {
	rrsets, supported := zone.ResourceRecordSets()
	if !supported {
		t.Fatalf("ResourceRecordSets interface not supported by zone %v", zone)
		return r
	}
	return rrsets
}

func listRrsOrFail(t *testing.T, rrsets dnsprovider.ResourceRecordSets) []dnsprovider.ResourceRecordSet {
	rrset, err := rrsets.List()
	if err != nil {
		t.Fatalf("Failed to list recordsets: %v", err)
	} else {
		if len(rrset) < 0 {
			t.Fatalf("Record set length=%d, expected >=0", len(rrset))
		} else {
			t.Logf("Got %d recordsets: %v", len(rrset), rrset)
		}
	}
	return rrset
}

func getExampleRrs(zone dnsprovider.Zone) dnsprovider.ResourceRecordSet {
	rrsets, _ := zone.ResourceRecordSets()
	return rrsets.New("www1."+zone.Name(), []string{"13.88.18.250", "13.88.18.250"}, 180, rrstype.A)
}


func getExampleCNAMERrs(zone dnsprovider.Zone) dnsprovider.ResourceRecordSet {
	rrsets, _ := zone.ResourceRecordSets()
	return rrsets.New("www1.hack.local.", []string{"alias." + zone.Name()}, 180, rrstype.CNAME)
}


func getInvalidRrs(zone dnsprovider.Zone) dnsprovider.ResourceRecordSet {
	rrsets, _ := zone.ResourceRecordSets()
	return rrsets.New("www12."+zone.Name(), []string{"rubbish", "rubbish"}, 180, rrstype.A)
}

func addRrsetOrFail(t *testing.T, rrsets dnsprovider.ResourceRecordSets, rrset dnsprovider.ResourceRecordSet) {
	err := rrsets.StartChangeset().Add(rrset).Apply()
	if err != nil {
		t.Fatalf("Failed to add recordsets: %v", err)
	}
}

/* TestZonesList verifies that listing of zones succeeds */
func TestZonesList(t *testing.T) {
	firstZone(t)
}

func TestAddRemoveCNAme(t *testing.T) {
	zone := firstZone(t)
	sets := rrs(t, zone)
	rrset := getExampleCNAMERrs(zone)
	addRrsetOrFail(t, sets, rrset)
	err := sets.StartChangeset().Remove(rrset).Apply()
	if err != nil {
		//Try again to clean up.
		defer sets.StartChangeset().Remove(rrset).Apply()
		t.Errorf("Failed to remove resource record set %v after adding", rrset)
	} else {
		t.Logf("Successfully removed resource set %v after adding", rrset)
	}
	//Check that it's gone
	list := listRrsOrFail(t, sets)
	found := false
	for _, set := range list {
		if set.Name() == rrset.Name() {
			found = true
			break
		}
	}
	if found {
		t.Errorf("Deleted resource record set %v is still present", rrset)
	}
	
}
/* TestZonesID verifies that the id of the zone is returned with the prefix removed */
func TestZonesID(t *testing.T) {
	zone := firstZone(t)

	// Check /hostedzone/ prefix is removed
	zoneID := zone.ID()
	if zoneID != zone.Name() {
		t.Fatalf("Unexpected zone id: %q", zoneID)
	}
}

/* TestZoneAddSuccess verifies that addition of a valid managed DNS zone succeeds */
func TestZoneAddSuccess(t *testing.T) {
	testZoneName := "ubernetes.testing"
	z := zones(t)
	input, err := z.New(testZoneName)
	if err != nil {
		t.Errorf("Failed to allocate new zone object %s: %v", testZoneName, err)
	}
	zone, err := z.Add(input)
	if err != nil {
		t.Errorf("Failed to create new managed DNS zone %s: %v", testZoneName, err)
	}
	defer func(zone dnsprovider.Zone) {
		if zone != nil {			
			if err := z.Remove(zone); err != nil {
				t.Errorf("Failed to delete zone %v: %v", zone, err)
			}
		}
	}(zone)
	t.Logf("Successfully added managed DNS zone: %v", zone)
}

/* TestResourceRecordSetsList verifies that listing of RRS's succeeds */
func TestResourceRecordSetsList(t *testing.T) {
	listRrsOrFail(t, rrs(t, firstZone(t)))
}

/* TestResourceRecordSetsAddSuccess verifies that addition of a valid RRS succeeds */
func TestResourceRecordSetsAddSuccess(t *testing.T) {
	//t.Logf("XTOPH")
	zone := firstZone(t)
	sets := rrs(t, zone)
	set := getExampleRrs(zone)
	addRrsetOrFail(t, sets, set)
	defer sets.StartChangeset().Remove(set).Apply()
	t.Logf("Successfully added resource record set: %v", set)
}

/* TestResourceRecordSetsAdditionVisible verifies that added RRS is visible after addition */
func TestResourceRecordSetsAdditionVisible(t *testing.T) {
	zone := firstZone(t)
	sets := rrs(t, zone)
	rrset := getExampleRrs(zone)
	addRrsetOrFail(t, sets, rrset)
	t.Logf("Successfully added resource record set: %v", rrset)
	found := false
	for _, record := range listRrsOrFail(t, sets) {
		if record.Name() == rrset.Name() {
			found = true
			break
		}
	}
	defer sets.StartChangeset().Remove(rrset).Apply()

	if !found {
		t.Errorf("Failed to find added resource record set %s", rrset.Name())
	}
}

/* TestResourceRecordSetsAddDuplicateFail verifies that addition of a duplicate RRS fails */
func TestResourceRecordSetsAddDuplicateFail(t *testing.T) {
	zone := firstZone(t)
	sets := rrs(t, zone)
	rrset := getExampleRrs(zone)
	addRrsetOrFail(t, sets, rrset)
	defer sets.StartChangeset().Remove(rrset).Apply()
	t.Logf("Successfully added resource record set: %v", rrset)
	// Try to add it again, and verify that the call fails.
	err := sets.StartChangeset().Add(rrset).Apply()
	if err == nil {
		defer sets.StartChangeset().Remove(rrset).Apply()
		t.Errorf("Should have failed to add duplicate resource record %v, but succeeded instead.", rrset)
	} else {
		t.Logf("Correctly failed to add duplicate resource record %v: %v", rrset, err)
	}
}

/* TestResourceRecordSetsRemove verifies that the removal of an existing RRS succeeds */
func TestResourceRecordSetsRemove(t *testing.T) {
	zone := firstZone(t)
	sets := rrs(t, zone)
	rrset := getExampleRrs(zone)
	addRrsetOrFail(t, sets, rrset)
	err := sets.StartChangeset().Remove(rrset).Apply()
	if err != nil {
		// Try again to clean up.
		defer sets.StartChangeset().Remove(rrset).Apply()
		t.Errorf("Failed to remove resource record set %v after adding", rrset)
	} else {
		t.Logf("Successfully removed resource set %v after adding", rrset)
	}
}

/* TestResourceRecordSetsRemoveGone verifies that a removed RRS no longer exists */
func TestResourceRecordSetsRemoveGone(t *testing.T) {
	zone := firstZone(t)
	sets := rrs(t, zone)
	rrset := getExampleRrs(zone)
	addRrsetOrFail(t, sets, rrset)
	err := sets.StartChangeset().Remove(rrset).Apply()
	if err != nil {
		//Try again to clean up.
		defer sets.StartChangeset().Remove(rrset).Apply()
		t.Errorf("Failed to remove resource record set %v after adding", rrset)
	} else {
		t.Logf("Successfully removed resource set %v after adding", rrset)
	}
	//Check that it's gone
	list := listRrsOrFail(t, sets)
	found := false
	for _, set := range list {
		if set.Name() == rrset.Name() {
			found = true
			break
		}
	}
	if found {
		t.Errorf("Deleted resource record set %v is still present", rrset)
	}
}

// func TestResourceRecordSetPaging(t * testing.T){
// 	// TODO
// 	zone := firstZone(t)
// 	sets := rrs(t, zone)
// 	addchanges := sets.StartChangeset()
// 	deletechanges := sets.StartChangeset()
// 	rrsets, _ := zone.ResourceRecordSets()
// 	for i:=0; i < 50; i++ {
// 		s := strconv.Itoa(i)
// 		r := rrsets.New("www12"+s+"."+zone.Name(), []string{"10.10.10." + s, "169.20.20." + s}, 180, rrstype.A)
// 		addchanges.Add(r)
// 		deletechanges.Add(r)
// 	}
// 	err := addchanges.Apply()
// 	if err != nil {
// 		t.Fatalf("Failed to add %i recordsets: %v", 50, err)
// 	}

// 	rrset, err := rrsets.List()
// 	if err != nil {
// 		t.Fatalf("Failed to list recordsets: %v", err)
// 	} else {
// 		if len(rrset) < 50 {
// 			t.Fatalf("Record set length=%d, expected >=0", len(rrset))
// 		} else {
// 			t.Logf("Got %d recordsets: %v", len(rrset), rrset)
// 		}
// 	}

// 	err = deletechanges.Apply()
// 	if err != nil {
// 		t.Fatalf("Failed to add %i recordsets: %v", 50, err)
// 	}

// 	return 
// }


/* TestResourceRecordSetsReplace verifies that replacing an RRS works */
func TestResourceRecordSetsReplace(t *testing.T) {
	zone := firstZone(t)
	tests.CommonTestResourceRecordSetsReplace(t, zone)
}

/* TestResourceRecordSetsReplaceAll verifies that we can remove an RRS and create one with a different name*/
func TestResourceRecordSetsReplaceAll(t *testing.T) {
	zone := firstZone(t)
	tests.CommonTestResourceRecordSetsReplaceAll(t, zone)
}

/* TestResourceRecordSetsHonorsType verifies that we can add records of the same name but different types */
func TestResourceRecordSetsDifferentTypes(t *testing.T) {
	zone := firstZone(t)
	tests.CommonTestResourceRecordSetsDifferentTypes(t, zone)
}
