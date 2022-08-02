package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-openapi/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
	"github.com/netbox-community/go-netbox/netbox/models"
)

func TestCheckIP(t *testing.T) {
	type checkIpTest struct {
		ctx                     context.Context
		toCheck, ipStart, ipEnd string
		want                    bool
	}

	checkIpTests := []checkIpTest{
		{context.TODO(), "10.80.21.32", "10.80.21.31/21", "10.80.21.51/21", true},
		{context.TODO(), "10.80.21.35", "10.80.21.31/21", "10.80.21.51/21", true},
		{context.TODO(), "25.82.21.32", "10.80.21.31/21", "10.80.21.51/21", false},
		{context.TODO(), "100.100.100.1000", "10.80.21.31/21", "10.80.21.51/21", false},
		{context.TODO(), "25.82.21.32", "10.800.21.31/21", "10.80.21.51/21", false},
		{context.TODO(), "25.82.21.32", "10.80.21.31/21", "10.800.21.51/21", false},
	}

	n := new(Netbox)
	n.logger = logr.Discard()
	for _, test := range checkIpTests {
		if output := n.CheckIp(test.ctx, test.toCheck, test.ipStart, test.ipEnd); output != test.want {
			t.Errorf("output %v not equal to expected %v", test.toCheck, test.want)
		}
	}
}

func toPointer(v string) *string { return &v }

func TestReadDevicesFromNetbox(t *testing.T) {
	type outputs struct {
		bmcIp       string
		bmcUsername string
		bmcPassword string
		disk        string
		label       string
		name        string
		primIp      string
		ifError     error
	}

	type inputs struct {
		v    outputs
		err  error
		want []*Machine
	}

	tests := []inputs{
		// Checking happy flow with control-plane
		{
			v: outputs{
				bmcIp:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				label:       "control-plane",
				name:        "dev",
				primIp:      "192.18.2.5/22",
				ifError:     nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:  "dev",
					IPAddress: "192.18.2.5",
					Netmask:   "255.255.252.0",
					Disk:      "/dev/sda",
					Labels: map[string]string{
						"type": "control-plane",
					},
					BMCIPAddress: "192.168.2.5",
					BMCUsername:  "root",
					BMCPassword:  "root",
				},
			},
		},
		// Checking happy flow with worker-plane
		{
			v: outputs{
				bmcIp:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIp:      "192.18.2.5/22",
				ifError:     nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:  "dev",
					IPAddress: "192.18.2.5",
					Netmask:   "255.255.252.0",
					Disk:      "/dev/sda",
					Labels: map[string]string{
						"type": "worker-plane",
					},
					BMCIPAddress: "192.168.2.5",
					BMCUsername:  "root",
					BMCPassword:  "root",
				},
			},
		},

		// Checking unhappy flow with bmcIp without Mask
		{
			v: outputs{
				bmcIp:       "192.168.2.5",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIp:      "192.18.2.5/22",
				ifError:     &IpError{"192.168.2.5"},
			},
			err: nil, want: []*Machine{
				{},
			},
		},
		// Checking unhappy flow with IPV6 address for prim IP
		{
			v: outputs{
				bmcIp:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				label:       "control-plane",
				name:        "dev",
				primIp:      "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				ifError:     &IpError{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
			},
			err: nil, want: []*Machine{
				{},
			},
		},
		// Checking unhappy flow with invalid IPv4 address with mask
		{
			v: outputs{
				bmcIp:       "192.460.634.516/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				label:       "",
				name:        "dev",
				primIp:      "192.18.2.5/22",
				ifError:     &IpError{"192.460.634.516/22"},
			},
			err: nil, want: []*Machine{
				{},
			},
		},
		{
			v: outputs{
				ifError: &NetboxError{"cannot get Devices list", "error code 500-Internal Server Error"},
			},
			err: errors.New("error code 500-Internal Server Error"), want: []*Machine{},
		},
	}

	for _, tt := range tests {
		n := new(Netbox)
		n.logger = logr.Discard()
		d := new(models.DeviceWithConfigContext)
		d.Tags = []*models.NestedTag{{Name: &tt.v.label}}
		d.Name = toPointer(tt.v.name)
		d.CustomFields = map[string]interface{}{
			"bmc_ip":       map[string]interface{}{"address": tt.v.bmcIp},
			"bmc_username": tt.v.bmcUsername,
			"bmc_password": tt.v.bmcPassword,
			"disk":         tt.v.disk,
		}
		d.PrimaryIp4 = &models.NestedIPAddress{Address: toPointer(tt.v.primIp)}
		dummyDevListOK := new(dcim.DcimDevicesListOK)
		dummyDevListOKBody := new(dcim.DcimDevicesListOKBody)

		// dummyDevListOK.Payload = new(models.Device)
		dummyDevListOKBody.Results = []*models.DeviceWithConfigContext{d}
		dummyDevListOK.Payload = dummyDevListOKBody
		v := &mock{v: dummyDevListOK, err: tt.err}
		c := &client.NetBoxAPI{Dcim: v}
		deviceReq := dcim.NewDcimDevicesListParams()
		err := n.ReadDevicesFromNetbox(context.TODO(), c, deviceReq)

		if err != nil {
			if !errors.Is(err, tt.v.ifError) {
				t.Fatal("Got: ", err.Error(), "want: ", tt.v.ifError)
			}

		} else {
			if diff := cmp.Diff(n.Records, tt.want); diff != "" {
				t.Fatal(diff)
			}
		}
	}
}

func TestReadInterfacesFromNetbox(t *testing.T) {
	type outputs struct {
		MacAddress []string
		Name       []string
		device     string
		Tag        int
		ifError    error
	}

	type inputs struct {
		v    outputs
		err  error
		want []*Machine
	}

	tests := []inputs{
		// Checking happy flow with 1 interface mapped to device
		{
			v: outputs{
				MacAddress: []string{"CC:48:3A:11:F4:C1"},
				Name:       []string{"GigabitEthernet1"},
				device:     "eksa-dev01",
				ifError:    nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:   "eksa-dev01",
					MACAddress: "CC:48:3A:11:F4:C1",
				},
			},
		},
		// Checking happy flow with 3 interfaces mapped to device and primary interface being 1st interface (0-based indexing)
		{
			v: outputs{
				MacAddress: []string{"CC:48:3A:11:F4:C1", "CC:48:3A:11:EA:11", "CC:48:3A:11:EA:61"},
				Name:       []string{"GigabitEthernet1", "GigabitEthernet1-a", "GigabitEthernet1-b"},
				device:     "eksa-dev01",
				Tag:        1,
				ifError:    nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:   "eksa-dev01",
					MACAddress: "CC:48:3A:11:EA:11",
				},
			},
		},
		// Checking Unhappy flow by generating error from API
		{
			v: outputs{
				device:  "errorDev",
				ifError: &NetboxError{"cannot get Interfaces list", "error code 500-Internal Server Error"},
			},
			err: errors.New("error code 500-Internal Server Error"), want: []*Machine{},
		},
	}
	for _, tt := range tests {
		n := new(Netbox)
		dummyMachine := &Machine{
			Hostname: tt.v.device,
		}

		n.Records = append(n.Records, dummyMachine)
		n.logger = logr.Discard()

		dummyInterfaceList := make([]*models.Interface, len(tt.v.MacAddress))
		for idx := range tt.v.MacAddress {
			i := new(models.Interface)
			i.Name = &tt.v.Name[idx]

			i.MacAddress = &tt.v.MacAddress[idx]
			if idx == tt.v.Tag {
				i.Tags = []*models.NestedTag{{Name: toPointer("eks-a")}}
			}
			dummyInterfaceList[idx] = i
		}

		dummyIntListOK := new(dcim.DcimInterfacesListOK)
		dummyIntListOKBody := new(dcim.DcimInterfacesListOKBody)
		dummyIntListOKBody.Results = dummyInterfaceList
		dummyIntListOK.Payload = dummyIntListOKBody
		i := &mock{i: dummyIntListOK, err: tt.err}
		c := &client.NetBoxAPI{Dcim: i}

		err := n.ReadInterfacesFromNetbox(context.TODO(), c)

		if err != nil {
			if !errors.Is(err, tt.v.ifError) {
				t.Fatal("Got: ", err.Error(), "want: ", tt.v.ifError)
			}
		} else {
			fmt.Println(n.Records)
			if diff := cmp.Diff(n.Records, tt.want); diff != "" {
				t.Fatal(diff)
			}
		}
	}
}

func TestTypeAssertions(t *testing.T) {
	type outputs struct {
		bmcIp       interface{}
		bmcUsername interface{}
		bmcPassword interface{}
		disk        interface{}
		name        string
		primIp      string
	}

	type inputs struct {
		v    outputs
		err  error
		want error
	}

	tests := []inputs{
		{
			v: outputs{
				bmcIp:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIp:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_ip", "map[string]interface{}", "string"},
		},
		{
			v: outputs{
				bmcIp:       map[string]interface{}{"address": 192.431},
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIp:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_ip_address", "string", "float64"},
		},
		{
			v: outputs{
				bmcIp:       map[string]interface{}{"address": "192.168.2.5/22"},
				bmcUsername: []string{"root1", "root2"},
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIp:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_username", "string", "[]string"},
		},
		{
			v: outputs{
				bmcIp:       map[string]interface{}{"address": "192.168.2.5/22"},
				bmcUsername: "root1",
				bmcPassword: []string{"root1", "root2"},
				disk:        "/dev/sda",
				name:        "dev",
				primIp:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_password", "string", "[]string"},
		},
		{
			v: outputs{
				bmcIp:       map[string]interface{}{"address": "192.168.2.5/22"},
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        123,
				name:        "dev",
				primIp:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"disk", "string", "int"},
		}}

	for _, tt := range tests {
		n := new(Netbox)
		n.logger = logr.Discard()
		d := new(models.DeviceWithConfigContext)
		d.Name = toPointer(tt.v.name)

		d.CustomFields = map[string]interface{}{
			"bmc_ip":       tt.v.bmcIp,
			"bmc_username": tt.v.bmcUsername,
			"bmc_password": tt.v.bmcPassword,
			"disk":         tt.v.disk,
		}
		d.PrimaryIp4 = &models.NestedIPAddress{Address: toPointer(tt.v.primIp)}
		dummyDevListOK := new(dcim.DcimDevicesListOK)
		dummyDevListOKBody := new(dcim.DcimDevicesListOKBody)

		dummyDevListOKBody.Results = []*models.DeviceWithConfigContext{d}
		dummyDevListOK.Payload = dummyDevListOKBody
		v := &mock{v: dummyDevListOK, err: tt.err}
		c := &client.NetBoxAPI{Dcim: v}
		deviceReq := dcim.NewDcimDevicesListParams()
		err := n.ReadDevicesFromNetbox(context.TODO(), c, deviceReq)

		if err != nil {
			if !errors.Is(err, tt.want) {
				t.Fatal("Got: ", err.Error(), "want: ", tt.want)
			}
		} else {
			if diff := cmp.Diff(n.Records, tt.want); diff != "" {
				t.Fatal(diff)
			}
		}
	}
}

func TestReadIpRangeFromNetbox(t *testing.T) {
	type outputs struct {
		gatewayIp    interface{}
		nameserverIp []interface{}
		startIp      string
		endIp        string
		ifError      error
	}

	type inputs struct {
		v    outputs
		err  error
		want []*Machine
	}

	tests := []inputs{
		{
			v: outputs{
				gatewayIp:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIp: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIp:      "10.80.12.20/22",
				endIp:        "10.80.12.30/22",
			},
			err: nil, want: []*Machine{
				{
					IPAddress:   "10.80.12.25",
					Gateway:     "10.80.8.1",
					Nameservers: Nameservers{"208.91.112.53"},
				},
			},
		},
		{
			v: outputs{
				gatewayIp:    map[string]interface{}{"address": "10.800.8.1/22"},
				nameserverIp: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIp:      "10.80.12.20/22",
				endIp:        "10.80.12.30/22",
				ifError:      &IpError{"10.800.8.1/22"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIp:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIp: []interface{}{map[string]interface{}{"address": "208.910.112.53/22"}},
				startIp:      "10.80.12.20/22",
				endIp:        "10.80.12.30/22",
				ifError:      &IpError{"208.910.112.53/22"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIp:    map[string]string{"address": "10.80.8.1/22"},
				nameserverIp: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIp:      "10.80.12.20/22",
				endIp:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"gatewayIP", "map[string]interface{}", "map[string]string"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIp:    map[string]interface{}{"address": 102.45},
				nameserverIp: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIp:      "10.80.12.20/22",
				endIp:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"gatewayAddr", "string", "float64"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIp:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIp: []interface{}{"208.91.112.53/22", "208.91.112.53/22"},
				startIp:      "10.80.12.20/22",
				endIp:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"nameserversIPMap", "map[string]interface{}", "string"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIp:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIp: []interface{}{map[string]interface{}{"address": 208.91}},
				startIp:      "10.80.12.20/22",
				endIp:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"nameserversIPMap", "string", "float64"},
			},
			err: nil, want: []*Machine{},
		},
	}

	for _, tt := range tests {
		n := new(Netbox)
		dummyMachine := &Machine{
			IPAddress: "10.80.12.25",
		}

		n.Records = append(n.Records, dummyMachine)
		n.logger = logr.Discard()

		d := new(models.IPRange)
		d.StartAddress = &tt.v.startIp
		d.EndAddress = &tt.v.endIp
		d.CustomFields = map[string]interface{}{
			"gateway":     tt.v.gatewayIp,
			"nameservers": tt.v.nameserverIp,
		}
		dummyIprangeListOk := new(ipam.IpamIPRangesListOK)
		dummyIprangeListOkBody := new(ipam.IpamIPRangesListOKBody)
		dummyIprangeListOkBody.Results = []*models.IPRange{d}
		dummyIprangeListOk.Payload = dummyIprangeListOkBody
		i := &mock{ip: dummyIprangeListOk, err: tt.err}
		c := &client.NetBoxAPI{Ipam: i}

		ipRangeReq := ipam.NewIpamIPRangesListParams()
		err := n.ReadIpRangeFromNetbox(context.TODO(), c, ipRangeReq)

		if err != nil {
			if !errors.Is(err, tt.v.ifError) {
				t.Fatal("Got: ", err.Error(), "want: ", tt.v.ifError)
			}
		} else {
			fmt.Println(n.Records)
			if diff := cmp.Diff(n.Records, tt.want); diff != "" {
				t.Fatal(diff)
			}
		}
	}
}

func TestSerializeMachines(t *testing.T) {
	var test = []*Machine{{Hostname: "Dev1", IPAddress: "10.80.8.21", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:F4:C1", Disk: "/dev/sda", Labels: map[string]string{"type": "worker-plane"}, BMCIPAddress: "10.80.12.20", BMCUsername: "root", BMCPassword: "pPyU6mAO"},
		{Hostname: "Dev2", IPAddress: "10.80.8.22", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:EA:11", Disk: "/dev/sda", Labels: map[string]string{"type": "control-plane"}, BMCIPAddress: "10.80.12.21", BMCUsername: "root", BMCPassword: "pPyU6mAO"},
	}

	want := createMachineString(test)
	n := new(Netbox)
	n.logger = logr.Discard()

	got, err := n.SerializeMachines(test)
	if err != nil {
		t.Fatal("Error: ", err)
	}

	if !bytes.EqualFold(got, []byte(want)) {
		t.Fatal(cmp.Diff(got, []byte(want)))
	}
}

type mock struct {
	v   *dcim.DcimDevicesListOK
	i   *dcim.DcimInterfacesListOK
	ip  *ipam.IpamIPRangesListOK
	err error
}

func (m *mock) DcimCablesBulkDelete(_ *dcim.DcimCablesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimCablesBulkPartialUpdate(_ *dcim.DcimCablesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesBulkUpdate(_ *dcim.DcimCablesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesCreate(_ *dcim.DcimCablesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimCablesDelete(_ *dcim.DcimCablesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimCablesList(_ *dcim.DcimCablesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesListOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesPartialUpdate(_ *dcim.DcimCablesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesRead(_ *dcim.DcimCablesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesUpdate(_ *dcim.DcimCablesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConnectedDeviceList(_ *dcim.DcimConnectedDeviceListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConnectedDeviceListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesBulkDelete(_ *dcim.DcimConsolePortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesBulkPartialUpdate(_ *dcim.DcimConsolePortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesBulkUpdate(_ *dcim.DcimConsolePortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesCreate(_ *dcim.DcimConsolePortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesDelete(_ *dcim.DcimConsolePortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesList(_ *dcim.DcimConsolePortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesPartialUpdate(_ *dcim.DcimConsolePortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesRead(_ *dcim.DcimConsolePortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesUpdate(_ *dcim.DcimConsolePortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsBulkDelete(_ *dcim.DcimConsolePortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsBulkPartialUpdate(_ *dcim.DcimConsolePortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsBulkUpdate(_ *dcim.DcimConsolePortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsCreate(_ *dcim.DcimConsolePortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsDelete(_ *dcim.DcimConsolePortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsList(_ *dcim.DcimConsolePortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsPartialUpdate(_ *dcim.DcimConsolePortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsRead(_ *dcim.DcimConsolePortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsTrace(_ *dcim.DcimConsolePortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsUpdate(_ *dcim.DcimConsolePortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesBulkDelete(_ *dcim.DcimConsoleServerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesBulkPartialUpdate(_ *dcim.DcimConsoleServerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesBulkUpdate(_ *dcim.DcimConsoleServerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesCreate(_ *dcim.DcimConsoleServerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesDelete(_ *dcim.DcimConsoleServerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesList(_ *dcim.DcimConsoleServerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesPartialUpdate(_ *dcim.DcimConsoleServerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesRead(_ *dcim.DcimConsoleServerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesUpdate(_ *dcim.DcimConsoleServerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsBulkDelete(_ *dcim.DcimConsoleServerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsBulkPartialUpdate(_ *dcim.DcimConsoleServerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsBulkUpdate(_ *dcim.DcimConsoleServerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsCreate(_ *dcim.DcimConsoleServerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsDelete(_ *dcim.DcimConsoleServerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsList(_ *dcim.DcimConsoleServerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsPartialUpdate(_ *dcim.DcimConsoleServerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsRead(_ *dcim.DcimConsoleServerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsTrace(_ *dcim.DcimConsoleServerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsUpdate(_ *dcim.DcimConsoleServerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesBulkDelete(_ *dcim.DcimDeviceBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesBulkPartialUpdate(_ *dcim.DcimDeviceBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesBulkUpdate(_ *dcim.DcimDeviceBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesCreate(_ *dcim.DcimDeviceBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesDelete(_ *dcim.DcimDeviceBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesList(_ *dcim.DcimDeviceBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesPartialUpdate(_ *dcim.DcimDeviceBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesRead(_ *dcim.DcimDeviceBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesUpdate(_ *dcim.DcimDeviceBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysBulkDelete(_ *dcim.DcimDeviceBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysBulkPartialUpdate(_ *dcim.DcimDeviceBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysBulkUpdate(_ *dcim.DcimDeviceBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysCreate(_ *dcim.DcimDeviceBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysDelete(_ *dcim.DcimDeviceBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysList(_ *dcim.DcimDeviceBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysPartialUpdate(_ *dcim.DcimDeviceBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysRead(_ *dcim.DcimDeviceBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysUpdate(_ *dcim.DcimDeviceBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesBulkDelete(_ *dcim.DcimDeviceRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesBulkPartialUpdate(_ *dcim.DcimDeviceRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesBulkUpdate(_ *dcim.DcimDeviceRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesCreate(_ *dcim.DcimDeviceRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesDelete(_ *dcim.DcimDeviceRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesList(_ *dcim.DcimDeviceRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesPartialUpdate(_ *dcim.DcimDeviceRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesRead(_ *dcim.DcimDeviceRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesUpdate(_ *dcim.DcimDeviceRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesBulkDelete(_ *dcim.DcimDeviceTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesBulkPartialUpdate(_ *dcim.DcimDeviceTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesBulkUpdate(_ *dcim.DcimDeviceTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesCreate(_ *dcim.DcimDeviceTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesDelete(_ *dcim.DcimDeviceTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesList(_ *dcim.DcimDeviceTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesPartialUpdate(_ *dcim.DcimDeviceTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesRead(_ *dcim.DcimDeviceTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesUpdate(_ *dcim.DcimDeviceTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesBulkDelete(_ *dcim.DcimDevicesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDevicesBulkPartialUpdate(_ *dcim.DcimDevicesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesBulkUpdate(_ *dcim.DcimDevicesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesCreate(_ *dcim.DcimDevicesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDevicesDelete(_ *dcim.DcimDevicesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDevicesList(_ *dcim.DcimDevicesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesListOK, error) {
	return m.v, m.err
}

func (m *mock) DcimDevicesNapalm(_ *dcim.DcimDevicesNapalmParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesNapalmOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesPartialUpdate(_ *dcim.DcimDevicesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesRead(_ *dcim.DcimDevicesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesUpdate(_ *dcim.DcimDevicesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesBulkDelete(_ *dcim.DcimFrontPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesBulkPartialUpdate(_ *dcim.DcimFrontPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesBulkUpdate(_ *dcim.DcimFrontPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesCreate(_ *dcim.DcimFrontPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesDelete(_ *dcim.DcimFrontPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesList(_ *dcim.DcimFrontPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesPartialUpdate(_ *dcim.DcimFrontPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesRead(_ *dcim.DcimFrontPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesUpdate(_ *dcim.DcimFrontPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsBulkDelete(_ *dcim.DcimFrontPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsBulkPartialUpdate(_ *dcim.DcimFrontPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsBulkUpdate(_ *dcim.DcimFrontPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsCreate(_ *dcim.DcimFrontPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsDelete(_ *dcim.DcimFrontPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsList(_ *dcim.DcimFrontPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsPartialUpdate(_ *dcim.DcimFrontPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsPaths(_ *dcim.DcimFrontPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsPathsOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsRead(_ *dcim.DcimFrontPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsUpdate(_ *dcim.DcimFrontPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesBulkDelete(_ *dcim.DcimInterfaceTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesBulkPartialUpdate(_ *dcim.DcimInterfaceTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesBulkUpdate(_ *dcim.DcimInterfaceTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesCreate(_ *dcim.DcimInterfaceTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesDelete(_ *dcim.DcimInterfaceTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesList(_ *dcim.DcimInterfaceTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesPartialUpdate(_ *dcim.DcimInterfaceTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesRead(_ *dcim.DcimInterfaceTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesUpdate(_ *dcim.DcimInterfaceTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesBulkDelete(_ *dcim.DcimInterfacesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesBulkPartialUpdate(_ *dcim.DcimInterfacesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesBulkUpdate(_ *dcim.DcimInterfacesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesCreate(_ *dcim.DcimInterfacesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesDelete(_ *dcim.DcimInterfacesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesList(_ *dcim.DcimInterfacesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesListOK, error) {
	return m.i, m.err
}

func (m *mock) DcimInterfacesPartialUpdate(_ *dcim.DcimInterfacesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesRead(_ *dcim.DcimInterfacesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesTrace(_ *dcim.DcimInterfacesTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesUpdate(_ *dcim.DcimInterfacesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesBulkDelete(_ *dcim.DcimInventoryItemRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesBulkPartialUpdate(_ *dcim.DcimInventoryItemRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesBulkUpdate(_ *dcim.DcimInventoryItemRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesCreate(_ *dcim.DcimInventoryItemRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesDelete(_ *dcim.DcimInventoryItemRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesList(_ *dcim.DcimInventoryItemRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesListOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesPartialUpdate(_ *dcim.DcimInventoryItemRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesRead(_ *dcim.DcimInventoryItemRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesUpdate(_ *dcim.DcimInventoryItemRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesBulkDelete(_ *dcim.DcimInventoryItemTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesBulkPartialUpdate(_ *dcim.DcimInventoryItemTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesBulkUpdate(_ *dcim.DcimInventoryItemTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesCreate(_ *dcim.DcimInventoryItemTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesDelete(_ *dcim.DcimInventoryItemTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesList(_ *dcim.DcimInventoryItemTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesPartialUpdate(_ *dcim.DcimInventoryItemTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesRead(_ *dcim.DcimInventoryItemTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesUpdate(_ *dcim.DcimInventoryItemTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsBulkDelete(_ *dcim.DcimInventoryItemsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsBulkPartialUpdate(_ *dcim.DcimInventoryItemsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsBulkUpdate(_ *dcim.DcimInventoryItemsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsCreate(_ *dcim.DcimInventoryItemsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsDelete(_ *dcim.DcimInventoryItemsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsList(_ *dcim.DcimInventoryItemsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsListOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsPartialUpdate(_ *dcim.DcimInventoryItemsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsRead(_ *dcim.DcimInventoryItemsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsUpdate(_ *dcim.DcimInventoryItemsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsBulkDelete(_ *dcim.DcimLocationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimLocationsBulkPartialUpdate(_ *dcim.DcimLocationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsBulkUpdate(_ *dcim.DcimLocationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsCreate(_ *dcim.DcimLocationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimLocationsDelete(_ *dcim.DcimLocationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimLocationsList(_ *dcim.DcimLocationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsListOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsPartialUpdate(_ *dcim.DcimLocationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsRead(_ *dcim.DcimLocationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsUpdate(_ *dcim.DcimLocationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersBulkDelete(_ *dcim.DcimManufacturersBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersBulkPartialUpdate(_ *dcim.DcimManufacturersBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersBulkUpdate(_ *dcim.DcimManufacturersBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersCreate(_ *dcim.DcimManufacturersCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersDelete(_ *dcim.DcimManufacturersDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersList(_ *dcim.DcimManufacturersListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersListOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersPartialUpdate(_ *dcim.DcimManufacturersPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersRead(_ *dcim.DcimManufacturersReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersReadOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersUpdate(_ *dcim.DcimManufacturersUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesBulkDelete(_ *dcim.DcimModuleBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesBulkPartialUpdate(_ *dcim.DcimModuleBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesBulkUpdate(_ *dcim.DcimModuleBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesCreate(_ *dcim.DcimModuleBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesDelete(_ *dcim.DcimModuleBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesList(_ *dcim.DcimModuleBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesPartialUpdate(_ *dcim.DcimModuleBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesRead(_ *dcim.DcimModuleBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesUpdate(_ *dcim.DcimModuleBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysBulkDelete(_ *dcim.DcimModuleBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysBulkPartialUpdate(_ *dcim.DcimModuleBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysBulkUpdate(_ *dcim.DcimModuleBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysCreate(_ *dcim.DcimModuleBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysDelete(_ *dcim.DcimModuleBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysList(_ *dcim.DcimModuleBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysListOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysPartialUpdate(_ *dcim.DcimModuleBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysRead(_ *dcim.DcimModuleBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysUpdate(_ *dcim.DcimModuleBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesBulkDelete(_ *dcim.DcimModuleTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesBulkPartialUpdate(_ *dcim.DcimModuleTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesBulkUpdate(_ *dcim.DcimModuleTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesCreate(_ *dcim.DcimModuleTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesDelete(_ *dcim.DcimModuleTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesList(_ *dcim.DcimModuleTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesListOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesPartialUpdate(_ *dcim.DcimModuleTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesRead(_ *dcim.DcimModuleTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesUpdate(_ *dcim.DcimModuleTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesBulkDelete(_ *dcim.DcimModulesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModulesBulkPartialUpdate(_ *dcim.DcimModulesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesBulkUpdate(_ *dcim.DcimModulesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesCreate(_ *dcim.DcimModulesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModulesDelete(_ *dcim.DcimModulesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModulesList(_ *dcim.DcimModulesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesListOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesPartialUpdate(_ *dcim.DcimModulesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesRead(_ *dcim.DcimModulesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesUpdate(_ *dcim.DcimModulesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsBulkDelete(_ *dcim.DcimPlatformsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsBulkPartialUpdate(_ *dcim.DcimPlatformsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsBulkUpdate(_ *dcim.DcimPlatformsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsCreate(_ *dcim.DcimPlatformsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsDelete(_ *dcim.DcimPlatformsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsList(_ *dcim.DcimPlatformsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsPartialUpdate(_ *dcim.DcimPlatformsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsRead(_ *dcim.DcimPlatformsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsUpdate(_ *dcim.DcimPlatformsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsBulkDelete(_ *dcim.DcimPowerFeedsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsBulkPartialUpdate(_ *dcim.DcimPowerFeedsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsBulkUpdate(_ *dcim.DcimPowerFeedsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsCreate(_ *dcim.DcimPowerFeedsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsDelete(_ *dcim.DcimPowerFeedsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsList(_ *dcim.DcimPowerFeedsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsPartialUpdate(_ *dcim.DcimPowerFeedsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsRead(_ *dcim.DcimPowerFeedsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsTrace(_ *dcim.DcimPowerFeedsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsUpdate(_ *dcim.DcimPowerFeedsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesBulkDelete(_ *dcim.DcimPowerOutletTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesBulkPartialUpdate(_ *dcim.DcimPowerOutletTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesBulkUpdate(_ *dcim.DcimPowerOutletTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesCreate(_ *dcim.DcimPowerOutletTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesDelete(_ *dcim.DcimPowerOutletTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesList(_ *dcim.DcimPowerOutletTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesPartialUpdate(_ *dcim.DcimPowerOutletTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesRead(_ *dcim.DcimPowerOutletTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesUpdate(_ *dcim.DcimPowerOutletTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsBulkDelete(_ *dcim.DcimPowerOutletsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsBulkPartialUpdate(_ *dcim.DcimPowerOutletsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsBulkUpdate(_ *dcim.DcimPowerOutletsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsCreate(_ *dcim.DcimPowerOutletsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsDelete(_ *dcim.DcimPowerOutletsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsList(_ *dcim.DcimPowerOutletsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsPartialUpdate(_ *dcim.DcimPowerOutletsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsRead(_ *dcim.DcimPowerOutletsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsTrace(_ *dcim.DcimPowerOutletsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsUpdate(_ *dcim.DcimPowerOutletsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsBulkDelete(_ *dcim.DcimPowerPanelsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsBulkPartialUpdate(_ *dcim.DcimPowerPanelsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsBulkUpdate(_ *dcim.DcimPowerPanelsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsCreate(_ *dcim.DcimPowerPanelsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsDelete(_ *dcim.DcimPowerPanelsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsList(_ *dcim.DcimPowerPanelsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsPartialUpdate(_ *dcim.DcimPowerPanelsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsRead(_ *dcim.DcimPowerPanelsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsUpdate(_ *dcim.DcimPowerPanelsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesBulkDelete(_ *dcim.DcimPowerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesBulkPartialUpdate(_ *dcim.DcimPowerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesBulkUpdate(_ *dcim.DcimPowerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesCreate(_ *dcim.DcimPowerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesDelete(_ *dcim.DcimPowerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesList(_ *dcim.DcimPowerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesPartialUpdate(_ *dcim.DcimPowerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesRead(_ *dcim.DcimPowerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesUpdate(_ *dcim.DcimPowerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsBulkDelete(_ *dcim.DcimPowerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsBulkPartialUpdate(_ *dcim.DcimPowerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsBulkUpdate(_ *dcim.DcimPowerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsCreate(_ *dcim.DcimPowerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsDelete(_ *dcim.DcimPowerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsList(_ *dcim.DcimPowerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsPartialUpdate(_ *dcim.DcimPowerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsRead(_ *dcim.DcimPowerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsTrace(_ *dcim.DcimPowerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsUpdate(_ *dcim.DcimPowerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsBulkDelete(_ *dcim.DcimRackReservationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsBulkPartialUpdate(_ *dcim.DcimRackReservationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsBulkUpdate(_ *dcim.DcimRackReservationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsCreate(_ *dcim.DcimRackReservationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsDelete(_ *dcim.DcimRackReservationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsList(_ *dcim.DcimRackReservationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsListOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsPartialUpdate(_ *dcim.DcimRackReservationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsRead(_ *dcim.DcimRackReservationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsUpdate(_ *dcim.DcimRackReservationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesBulkDelete(_ *dcim.DcimRackRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesBulkPartialUpdate(_ *dcim.DcimRackRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesBulkUpdate(_ *dcim.DcimRackRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesCreate(_ *dcim.DcimRackRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesDelete(_ *dcim.DcimRackRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesList(_ *dcim.DcimRackRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesListOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesPartialUpdate(_ *dcim.DcimRackRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesRead(_ *dcim.DcimRackRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesUpdate(_ *dcim.DcimRackRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksBulkDelete(_ *dcim.DcimRacksBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRacksBulkPartialUpdate(_ *dcim.DcimRacksBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksBulkUpdate(_ *dcim.DcimRacksBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksCreate(_ *dcim.DcimRacksCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRacksDelete(_ *dcim.DcimRacksDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRacksElevation(_ *dcim.DcimRacksElevationParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksElevationOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksList(_ *dcim.DcimRacksListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksListOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksPartialUpdate(_ *dcim.DcimRacksPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksRead(_ *dcim.DcimRacksReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksUpdate(_ *dcim.DcimRacksUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesBulkDelete(_ *dcim.DcimRearPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesBulkPartialUpdate(_ *dcim.DcimRearPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesBulkUpdate(_ *dcim.DcimRearPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesCreate(_ *dcim.DcimRearPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesDelete(_ *dcim.DcimRearPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesList(_ *dcim.DcimRearPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesPartialUpdate(_ *dcim.DcimRearPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesRead(_ *dcim.DcimRearPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesUpdate(_ *dcim.DcimRearPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsBulkDelete(_ *dcim.DcimRearPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsBulkPartialUpdate(_ *dcim.DcimRearPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsBulkUpdate(_ *dcim.DcimRearPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsCreate(_ *dcim.DcimRearPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsDelete(_ *dcim.DcimRearPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsList(_ *dcim.DcimRearPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsPartialUpdate(_ *dcim.DcimRearPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsPaths(_ *dcim.DcimRearPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsPathsOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsRead(_ *dcim.DcimRearPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsUpdate(_ *dcim.DcimRearPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsBulkDelete(_ *dcim.DcimRegionsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRegionsBulkPartialUpdate(_ *dcim.DcimRegionsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsBulkUpdate(_ *dcim.DcimRegionsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsCreate(_ *dcim.DcimRegionsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRegionsDelete(_ *dcim.DcimRegionsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRegionsList(_ *dcim.DcimRegionsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsListOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsPartialUpdate(_ *dcim.DcimRegionsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsRead(_ *dcim.DcimRegionsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsUpdate(_ *dcim.DcimRegionsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsBulkDelete(_ *dcim.DcimSiteGroupsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsBulkPartialUpdate(_ *dcim.DcimSiteGroupsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsBulkUpdate(_ *dcim.DcimSiteGroupsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsCreate(_ *dcim.DcimSiteGroupsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsDelete(_ *dcim.DcimSiteGroupsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsList(_ *dcim.DcimSiteGroupsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsListOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsPartialUpdate(_ *dcim.DcimSiteGroupsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsRead(_ *dcim.DcimSiteGroupsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsUpdate(_ *dcim.DcimSiteGroupsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesBulkDelete(_ *dcim.DcimSitesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSitesBulkPartialUpdate(_ *dcim.DcimSitesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesBulkUpdate(_ *dcim.DcimSitesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesCreate(_ *dcim.DcimSitesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimSitesDelete(_ *dcim.DcimSitesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSitesList(_ *dcim.DcimSitesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesListOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesPartialUpdate(_ *dcim.DcimSitesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesRead(_ *dcim.DcimSitesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesUpdate(_ *dcim.DcimSitesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisBulkDelete(_ *dcim.DcimVirtualChassisBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisBulkPartialUpdate(_ *dcim.DcimVirtualChassisBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisBulkUpdate(_ *dcim.DcimVirtualChassisBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisCreate(_ *dcim.DcimVirtualChassisCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisDelete(_ *dcim.DcimVirtualChassisDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisList(_ *dcim.DcimVirtualChassisListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisListOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisPartialUpdate(_ *dcim.DcimVirtualChassisPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisRead(_ *dcim.DcimVirtualChassisReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisReadOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisUpdate(_ *dcim.DcimVirtualChassisUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesBulkDelete(_ *ipam.IpamAggregatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesBulkPartialUpdate(_ *ipam.IpamAggregatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesBulkUpdate(_ *ipam.IpamAggregatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesCreate(_ *ipam.IpamAggregatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesDelete(_ *ipam.IpamAggregatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesList(_ *ipam.IpamAggregatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesListOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesPartialUpdate(_ *ipam.IpamAggregatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesRead(_ *ipam.IpamAggregatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesUpdate(_ *ipam.IpamAggregatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAggregatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsBulkDelete(_ *ipam.IpamAsnsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAsnsBulkPartialUpdate(_ *ipam.IpamAsnsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsBulkUpdate(_ *ipam.IpamAsnsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsCreate(_ *ipam.IpamAsnsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamAsnsDelete(_ *ipam.IpamAsnsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAsnsList(_ *ipam.IpamAsnsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsListOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsPartialUpdate(_ *ipam.IpamAsnsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsRead(_ *ipam.IpamAsnsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsUpdate(_ *ipam.IpamAsnsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamAsnsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsBulkDelete(_ *ipam.IpamFhrpGroupAssignmentsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsBulkPartialUpdate(_ *ipam.IpamFhrpGroupAssignmentsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsBulkUpdate(_ *ipam.IpamFhrpGroupAssignmentsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsCreate(_ *ipam.IpamFhrpGroupAssignmentsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsDelete(_ *ipam.IpamFhrpGroupAssignmentsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsList(_ *ipam.IpamFhrpGroupAssignmentsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsListOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsPartialUpdate(_ *ipam.IpamFhrpGroupAssignmentsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsRead(_ *ipam.IpamFhrpGroupAssignmentsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsUpdate(_ *ipam.IpamFhrpGroupAssignmentsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsBulkDelete(_ *ipam.IpamFhrpGroupsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsBulkPartialUpdate(_ *ipam.IpamFhrpGroupsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsBulkUpdate(_ *ipam.IpamFhrpGroupsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsCreate(_ *ipam.IpamFhrpGroupsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsDelete(_ *ipam.IpamFhrpGroupsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsList(_ *ipam.IpamFhrpGroupsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsListOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsPartialUpdate(_ *ipam.IpamFhrpGroupsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsRead(_ *ipam.IpamFhrpGroupsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsUpdate(_ *ipam.IpamFhrpGroupsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamFhrpGroupsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesBulkDelete(_ *ipam.IpamIPAddressesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesBulkPartialUpdate(_ *ipam.IpamIPAddressesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesBulkUpdate(_ *ipam.IpamIPAddressesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesCreate(_ *ipam.IpamIPAddressesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesDelete(_ *ipam.IpamIPAddressesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesList(_ *ipam.IpamIPAddressesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesListOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesPartialUpdate(_ *ipam.IpamIPAddressesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesRead(_ *ipam.IpamIPAddressesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesUpdate(_ *ipam.IpamIPAddressesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPAddressesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesAvailableIpsCreate(_ *ipam.IpamIPRangesAvailableIpsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesAvailableIpsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesAvailableIpsList(_ *ipam.IpamIPRangesAvailableIpsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesAvailableIpsListOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesBulkDelete(_ *ipam.IpamIPRangesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesBulkPartialUpdate(_ *ipam.IpamIPRangesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesBulkUpdate(_ *ipam.IpamIPRangesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesCreate(_ *ipam.IpamIPRangesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesDelete(_ *ipam.IpamIPRangesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesList(_ *ipam.IpamIPRangesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesListOK, error) {
	return m.ip, nil
}

func (m *mock) IpamIPRangesPartialUpdate(_ *ipam.IpamIPRangesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesRead(_ *ipam.IpamIPRangesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesUpdate(_ *ipam.IpamIPRangesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamIPRangesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailableIpsCreate(_ *ipam.IpamPrefixesAvailableIpsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesAvailableIpsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailableIpsList(_ *ipam.IpamPrefixesAvailableIpsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesAvailableIpsListOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailablePrefixesCreate(_ *ipam.IpamPrefixesAvailablePrefixesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesAvailablePrefixesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailablePrefixesList(_ *ipam.IpamPrefixesAvailablePrefixesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesAvailablePrefixesListOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesBulkDelete(_ *ipam.IpamPrefixesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesBulkPartialUpdate(_ *ipam.IpamPrefixesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesBulkUpdate(_ *ipam.IpamPrefixesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesCreate(_ *ipam.IpamPrefixesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesDelete(_ *ipam.IpamPrefixesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesList(_ *ipam.IpamPrefixesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesListOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesPartialUpdate(_ *ipam.IpamPrefixesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesRead(_ *ipam.IpamPrefixesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesUpdate(_ *ipam.IpamPrefixesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamPrefixesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsBulkDelete(_ *ipam.IpamRirsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRirsBulkPartialUpdate(_ *ipam.IpamRirsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsBulkUpdate(_ *ipam.IpamRirsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsCreate(_ *ipam.IpamRirsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamRirsDelete(_ *ipam.IpamRirsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRirsList(_ *ipam.IpamRirsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsListOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsPartialUpdate(_ *ipam.IpamRirsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsRead(_ *ipam.IpamRirsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsUpdate(_ *ipam.IpamRirsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRirsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesBulkDelete(_ *ipam.IpamRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRolesBulkPartialUpdate(_ *ipam.IpamRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesBulkUpdate(_ *ipam.IpamRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesCreate(_ *ipam.IpamRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamRolesDelete(_ *ipam.IpamRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRolesList(_ *ipam.IpamRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesListOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesPartialUpdate(_ *ipam.IpamRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesRead(_ *ipam.IpamRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesUpdate(_ *ipam.IpamRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsBulkDelete(_ *ipam.IpamRouteTargetsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsBulkPartialUpdate(_ *ipam.IpamRouteTargetsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsBulkUpdate(_ *ipam.IpamRouteTargetsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsCreate(_ *ipam.IpamRouteTargetsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsDelete(_ *ipam.IpamRouteTargetsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsList(_ *ipam.IpamRouteTargetsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsListOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsPartialUpdate(_ *ipam.IpamRouteTargetsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsRead(_ *ipam.IpamRouteTargetsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsUpdate(_ *ipam.IpamRouteTargetsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamRouteTargetsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesBulkDelete(_ *ipam.IpamServiceTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesBulkPartialUpdate(_ *ipam.IpamServiceTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesBulkUpdate(_ *ipam.IpamServiceTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesCreate(_ *ipam.IpamServiceTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesDelete(_ *ipam.IpamServiceTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesList(_ *ipam.IpamServiceTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesPartialUpdate(_ *ipam.IpamServiceTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesRead(_ *ipam.IpamServiceTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesUpdate(_ *ipam.IpamServiceTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServiceTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesBulkDelete(_ *ipam.IpamServicesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServicesBulkPartialUpdate(_ *ipam.IpamServicesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesBulkUpdate(_ *ipam.IpamServicesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesCreate(_ *ipam.IpamServicesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamServicesDelete(_ *ipam.IpamServicesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServicesList(_ *ipam.IpamServicesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesListOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesPartialUpdate(_ *ipam.IpamServicesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesRead(_ *ipam.IpamServicesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesUpdate(_ *ipam.IpamServicesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamServicesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsAvailableVlansCreate(_ *ipam.IpamVlanGroupsAvailableVlansCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsAvailableVlansCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsAvailableVlansList(_ *ipam.IpamVlanGroupsAvailableVlansListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsAvailableVlansListOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsBulkDelete(_ *ipam.IpamVlanGroupsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsBulkPartialUpdate(_ *ipam.IpamVlanGroupsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsBulkUpdate(_ *ipam.IpamVlanGroupsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsCreate(_ *ipam.IpamVlanGroupsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsDelete(_ *ipam.IpamVlanGroupsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsList(_ *ipam.IpamVlanGroupsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsListOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsPartialUpdate(_ *ipam.IpamVlanGroupsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsRead(_ *ipam.IpamVlanGroupsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsUpdate(_ *ipam.IpamVlanGroupsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlanGroupsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansBulkDelete(_ *ipam.IpamVlansBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlansBulkPartialUpdate(_ *ipam.IpamVlansBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansBulkUpdate(_ *ipam.IpamVlansBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansCreate(_ *ipam.IpamVlansCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVlansDelete(_ *ipam.IpamVlansDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlansList(_ *ipam.IpamVlansListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansListOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansPartialUpdate(_ *ipam.IpamVlansPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansRead(_ *ipam.IpamVlansReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansReadOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansUpdate(_ *ipam.IpamVlansUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVlansUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsBulkDelete(_ *ipam.IpamVrfsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVrfsBulkPartialUpdate(_ *ipam.IpamVrfsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsBulkUpdate(_ *ipam.IpamVrfsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsCreate(_ *ipam.IpamVrfsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVrfsDelete(_ *ipam.IpamVrfsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVrfsList(_ *ipam.IpamVrfsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsListOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsPartialUpdate(_ *ipam.IpamVrfsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsRead(_ *ipam.IpamVrfsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsUpdate(_ *ipam.IpamVrfsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ipam.ClientOption) (*ipam.IpamVrfsUpdateOK, error) {
	return nil, nil
}

func (m *mock) SetTransport(transport runtime.ClientTransport) {
}
