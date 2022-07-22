package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-openapi/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
	"github.com/stretchr/testify/assert"
)

func TestCheckIP(t *testing.T) {
	type checkIpTest struct {
		ctx                     context.Context
		toCheck, ipStart, ipEnd string
		want                    bool
	}

	var checkIpTests = []checkIpTest{
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

	// type expected struct{
	// 	machines []*Machine
	// 	err 	error
	// }

	type inputs struct {
		v    outputs
		err  error
		want []*Machine
	}

	var tests = []inputs{
		// Checking happy flow with control-plane
		{v: outputs{
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
					BMCPassword:  "root"},
			}},
		// Checking happy flow with worker-plane
		{v: outputs{
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
					BMCPassword:  "root"},
			}},

		// Checking unhappy flow with bmcIp without Mask
		{v: outputs{
			bmcIp:       "192.168.2.5",
			bmcUsername: "root",
			bmcPassword: "root",
			disk:        "/dev/sda",
			name:        "dev",
			primIp:      "192.18.2.5/22",
			ifError:     errors.New("cannot parse BMC IP, invalid CIDR address: 192.168.2.5"),
		},
			err: nil, want: []*Machine{
				{},
			}},
		// Checking unhappy flow with IPV6 address for prim IP
		{v: outputs{
			bmcIp:       "192.168.2.5/22",
			bmcUsername: "root",
			bmcPassword: "root",
			disk:        "/dev/sda",
			label:       "control-plane",
			name:        "dev",
			primIp:      "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			ifError:     errors.New("cannot parse Machine IP Address, invalid CIDR address: 2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
		},
			err: nil, want: []*Machine{
				{},
			}},
		// Checking unhappy flow with invalid IPv4 address with mask
		{v: outputs{
			bmcIp:       "192.460.634.516/22",
			bmcUsername: "root",
			bmcPassword: "root",
			disk:        "/dev/sda",
			label:       "",
			name:        "dev",
			primIp:      "192.18.2.5/22",
			ifError:     errors.New("cannot parse BMC IP, invalid CIDR address: 192.460.634.516/22"),
		},
			err: nil, want: []*Machine{
				{},
			}},
		{v: outputs{
			ifError: errors.New("cannot get Devices list: error code 500-Internal Server Error"),
		},
			err: errors.New("error code 500-Internal Server Error"), want: []*Machine{
				{},
			}},
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
			expErr := tt.v.ifError
			//skip assert, and compare the error strings directly.
			assert.EqualErrorf(t, err, expErr.Error(), "Error should be: %v, got: %v", tt.v.ifError, err)
		} else {
			// fmt.Println(n.Records)
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

	var tests = []inputs{
		// Checking happy flow with 1 interface mapped to device
		{v: outputs{
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
			}},
		// Checking happy flow with 3 interfaces mapped to device and primary interface being 1st interface (0-based indexing)
		{v: outputs{
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
			}},
		// Checking Unhappy flow by generating error from API
		{v: outputs{
			device:  "errorDev",
			ifError: errors.New("cannot get Interfaces list: error code 500-Internal Server Error for hostname errorDev"),
		},
			err: errors.New("error code 500-Internal Server Error"), want: []*Machine{}},
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
			// i.Device = &models.NestedDevice{Name: toPointer(tt.v.device)}
			i.MacAddress = &tt.v.MacAddress[idx]
			if idx == tt.v.Tag {
				i.Tags = []*models.NestedTag{{Name: toPointer("eks-a")}}
			}
			dummyInterfaceList[idx] = i
		}

		dummyIntListOK := new(dcim.DcimInterfacesListOK)
		dummyIntListOKBody := new(dcim.DcimInterfacesListOKBody)

		// dummyDevListOK.Payload = new(models.Device)
		dummyIntListOKBody.Results = dummyInterfaceList
		dummyIntListOK.Payload = dummyIntListOKBody
		i := &mock{i: dummyIntListOK, err: tt.err}
		c := &client.NetBoxAPI{Dcim: i}
		// interfaceReq := dcim.NewDcimInterfacesListParams()
		err := n.ReadInterfacesFromNetbox(context.TODO(), c)

		if err != nil {
			expErr := tt.v.ifError
			assert.EqualErrorf(t, err, expErr.Error(), "Error should be: %v, got: %v", tt.v.ifError, err)
		} else {
			fmt.Println(n.Records)
			if diff := cmp.Diff(n.Records, tt.want); diff != "" {
				t.Fatal(diff)
			}
		}
	}
}

type mock struct {
	v   *dcim.DcimDevicesListOK
	i   *dcim.DcimInterfacesListOK
	err error
}

func (m *mock) DcimCablesBulkDelete(params *dcim.DcimCablesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimCablesBulkPartialUpdate(params *dcim.DcimCablesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimCablesBulkUpdate(params *dcim.DcimCablesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimCablesCreate(params *dcim.DcimCablesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimCablesDelete(params *dcim.DcimCablesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimCablesList(params *dcim.DcimCablesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesListOK, error) {
	return nil, nil
}
func (m *mock) DcimCablesPartialUpdate(params *dcim.DcimCablesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimCablesRead(params *dcim.DcimCablesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimCablesUpdate(params *dcim.DcimCablesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimCablesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConnectedDeviceList(params *dcim.DcimConnectedDeviceListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConnectedDeviceListOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesBulkDelete(params *dcim.DcimConsolePortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesBulkPartialUpdate(params *dcim.DcimConsolePortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesBulkUpdate(params *dcim.DcimConsolePortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesCreate(params *dcim.DcimConsolePortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesDelete(params *dcim.DcimConsolePortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesList(params *dcim.DcimConsolePortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesPartialUpdate(params *dcim.DcimConsolePortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesRead(params *dcim.DcimConsolePortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortTemplatesUpdate(params *dcim.DcimConsolePortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsBulkDelete(params *dcim.DcimConsolePortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsBulkPartialUpdate(params *dcim.DcimConsolePortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsBulkUpdate(params *dcim.DcimConsolePortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsCreate(params *dcim.DcimConsolePortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsDelete(params *dcim.DcimConsolePortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsList(params *dcim.DcimConsolePortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsListOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsPartialUpdate(params *dcim.DcimConsolePortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsRead(params *dcim.DcimConsolePortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsTrace(params *dcim.DcimConsolePortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsTraceOK, error) {
	return nil, nil
}
func (m *mock) DcimConsolePortsUpdate(params *dcim.DcimConsolePortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsolePortsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesBulkDelete(params *dcim.DcimConsoleServerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesBulkPartialUpdate(params *dcim.DcimConsoleServerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesBulkUpdate(params *dcim.DcimConsoleServerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesCreate(params *dcim.DcimConsoleServerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesDelete(params *dcim.DcimConsoleServerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesList(params *dcim.DcimConsoleServerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesPartialUpdate(params *dcim.DcimConsoleServerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesRead(params *dcim.DcimConsoleServerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortTemplatesUpdate(params *dcim.DcimConsoleServerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsBulkDelete(params *dcim.DcimConsoleServerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsBulkPartialUpdate(params *dcim.DcimConsoleServerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsBulkUpdate(params *dcim.DcimConsoleServerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsCreate(params *dcim.DcimConsoleServerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsDelete(params *dcim.DcimConsoleServerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsList(params *dcim.DcimConsoleServerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsListOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsPartialUpdate(params *dcim.DcimConsoleServerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsRead(params *dcim.DcimConsoleServerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsTrace(params *dcim.DcimConsoleServerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsTraceOK, error) {
	return nil, nil
}
func (m *mock) DcimConsoleServerPortsUpdate(params *dcim.DcimConsoleServerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesBulkDelete(params *dcim.DcimDeviceBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesBulkPartialUpdate(params *dcim.DcimDeviceBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesBulkUpdate(params *dcim.DcimDeviceBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesCreate(params *dcim.DcimDeviceBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesDelete(params *dcim.DcimDeviceBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesList(params *dcim.DcimDeviceBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesPartialUpdate(params *dcim.DcimDeviceBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesRead(params *dcim.DcimDeviceBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBayTemplatesUpdate(params *dcim.DcimDeviceBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysBulkDelete(params *dcim.DcimDeviceBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysBulkPartialUpdate(params *dcim.DcimDeviceBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysBulkUpdate(params *dcim.DcimDeviceBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysCreate(params *dcim.DcimDeviceBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysDelete(params *dcim.DcimDeviceBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysList(params *dcim.DcimDeviceBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysListOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysPartialUpdate(params *dcim.DcimDeviceBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysRead(params *dcim.DcimDeviceBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysReadOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceBaysUpdate(params *dcim.DcimDeviceBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceBaysUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesBulkDelete(params *dcim.DcimDeviceRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesBulkPartialUpdate(params *dcim.DcimDeviceRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesBulkUpdate(params *dcim.DcimDeviceRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesCreate(params *dcim.DcimDeviceRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesDelete(params *dcim.DcimDeviceRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesList(params *dcim.DcimDeviceRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesListOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesPartialUpdate(params *dcim.DcimDeviceRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesRead(params *dcim.DcimDeviceRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceRolesUpdate(params *dcim.DcimDeviceRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceRolesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesBulkDelete(params *dcim.DcimDeviceTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesBulkPartialUpdate(params *dcim.DcimDeviceTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesBulkUpdate(params *dcim.DcimDeviceTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesCreate(params *dcim.DcimDeviceTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesDelete(params *dcim.DcimDeviceTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesList(params *dcim.DcimDeviceTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesListOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesPartialUpdate(params *dcim.DcimDeviceTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesRead(params *dcim.DcimDeviceTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimDeviceTypesUpdate(params *dcim.DcimDeviceTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDeviceTypesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDevicesBulkDelete(params *dcim.DcimDevicesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDevicesBulkPartialUpdate(params *dcim.DcimDevicesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDevicesBulkUpdate(params *dcim.DcimDevicesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDevicesCreate(params *dcim.DcimDevicesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimDevicesDelete(params *dcim.DcimDevicesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimDevicesList(params *dcim.DcimDevicesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesListOK, error) {
	return m.v, m.err
}
func (m *mock) DcimDevicesNapalm(params *dcim.DcimDevicesNapalmParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesNapalmOK, error) {
	return nil, nil
}
func (m *mock) DcimDevicesPartialUpdate(params *dcim.DcimDevicesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimDevicesRead(params *dcim.DcimDevicesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimDevicesUpdate(params *dcim.DcimDevicesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimDevicesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesBulkDelete(params *dcim.DcimFrontPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesBulkPartialUpdate(params *dcim.DcimFrontPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesBulkUpdate(params *dcim.DcimFrontPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesCreate(params *dcim.DcimFrontPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesDelete(params *dcim.DcimFrontPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesList(params *dcim.DcimFrontPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesPartialUpdate(params *dcim.DcimFrontPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesRead(params *dcim.DcimFrontPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortTemplatesUpdate(params *dcim.DcimFrontPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsBulkDelete(params *dcim.DcimFrontPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsBulkPartialUpdate(params *dcim.DcimFrontPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsBulkUpdate(params *dcim.DcimFrontPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsCreate(params *dcim.DcimFrontPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsDelete(params *dcim.DcimFrontPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsList(params *dcim.DcimFrontPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsListOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsPartialUpdate(params *dcim.DcimFrontPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsPaths(params *dcim.DcimFrontPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsPathsOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsRead(params *dcim.DcimFrontPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimFrontPortsUpdate(params *dcim.DcimFrontPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimFrontPortsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesBulkDelete(params *dcim.DcimInterfaceTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesBulkPartialUpdate(params *dcim.DcimInterfaceTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesBulkUpdate(params *dcim.DcimInterfaceTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesCreate(params *dcim.DcimInterfaceTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesDelete(params *dcim.DcimInterfaceTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesList(params *dcim.DcimInterfaceTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesPartialUpdate(params *dcim.DcimInterfaceTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesRead(params *dcim.DcimInterfaceTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfaceTemplatesUpdate(params *dcim.DcimInterfaceTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesBulkDelete(params *dcim.DcimInterfacesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesBulkPartialUpdate(params *dcim.DcimInterfacesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesBulkUpdate(params *dcim.DcimInterfacesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesCreate(params *dcim.DcimInterfacesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesDelete(params *dcim.DcimInterfacesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesList(params *dcim.DcimInterfacesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesListOK, error) {
	return m.i, m.err
}
func (m *mock) DcimInterfacesPartialUpdate(params *dcim.DcimInterfacesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesRead(params *dcim.DcimInterfacesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesTrace(params *dcim.DcimInterfacesTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesTraceOK, error) {
	return nil, nil
}
func (m *mock) DcimInterfacesUpdate(params *dcim.DcimInterfacesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInterfacesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesBulkDelete(params *dcim.DcimInventoryItemRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesBulkPartialUpdate(params *dcim.DcimInventoryItemRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesBulkUpdate(params *dcim.DcimInventoryItemRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesCreate(params *dcim.DcimInventoryItemRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesDelete(params *dcim.DcimInventoryItemRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesList(params *dcim.DcimInventoryItemRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesListOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesPartialUpdate(params *dcim.DcimInventoryItemRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesRead(params *dcim.DcimInventoryItemRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemRolesUpdate(params *dcim.DcimInventoryItemRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesBulkDelete(params *dcim.DcimInventoryItemTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesBulkPartialUpdate(params *dcim.DcimInventoryItemTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesBulkUpdate(params *dcim.DcimInventoryItemTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesCreate(params *dcim.DcimInventoryItemTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesDelete(params *dcim.DcimInventoryItemTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesList(params *dcim.DcimInventoryItemTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesPartialUpdate(params *dcim.DcimInventoryItemTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesRead(params *dcim.DcimInventoryItemTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemTemplatesUpdate(params *dcim.DcimInventoryItemTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsBulkDelete(params *dcim.DcimInventoryItemsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsBulkPartialUpdate(params *dcim.DcimInventoryItemsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsBulkUpdate(params *dcim.DcimInventoryItemsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsCreate(params *dcim.DcimInventoryItemsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsDelete(params *dcim.DcimInventoryItemsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsList(params *dcim.DcimInventoryItemsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsListOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsPartialUpdate(params *dcim.DcimInventoryItemsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsRead(params *dcim.DcimInventoryItemsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimInventoryItemsUpdate(params *dcim.DcimInventoryItemsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimInventoryItemsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimLocationsBulkDelete(params *dcim.DcimLocationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimLocationsBulkPartialUpdate(params *dcim.DcimLocationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimLocationsBulkUpdate(params *dcim.DcimLocationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimLocationsCreate(params *dcim.DcimLocationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimLocationsDelete(params *dcim.DcimLocationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimLocationsList(params *dcim.DcimLocationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsListOK, error) {
	return nil, nil
}
func (m *mock) DcimLocationsPartialUpdate(params *dcim.DcimLocationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimLocationsRead(params *dcim.DcimLocationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimLocationsUpdate(params *dcim.DcimLocationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimLocationsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersBulkDelete(params *dcim.DcimManufacturersBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersBulkPartialUpdate(params *dcim.DcimManufacturersBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersBulkUpdate(params *dcim.DcimManufacturersBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersCreate(params *dcim.DcimManufacturersCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersDelete(params *dcim.DcimManufacturersDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersList(params *dcim.DcimManufacturersListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersListOK, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersPartialUpdate(params *dcim.DcimManufacturersPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersRead(params *dcim.DcimManufacturersReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersReadOK, error) {
	return nil, nil
}
func (m *mock) DcimManufacturersUpdate(params *dcim.DcimManufacturersUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimManufacturersUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesBulkDelete(params *dcim.DcimModuleBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesBulkPartialUpdate(params *dcim.DcimModuleBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesBulkUpdate(params *dcim.DcimModuleBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesCreate(params *dcim.DcimModuleBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesDelete(params *dcim.DcimModuleBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesList(params *dcim.DcimModuleBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesPartialUpdate(params *dcim.DcimModuleBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesRead(params *dcim.DcimModuleBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBayTemplatesUpdate(params *dcim.DcimModuleBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysBulkDelete(params *dcim.DcimModuleBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysBulkPartialUpdate(params *dcim.DcimModuleBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysBulkUpdate(params *dcim.DcimModuleBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysCreate(params *dcim.DcimModuleBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysDelete(params *dcim.DcimModuleBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysList(params *dcim.DcimModuleBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysListOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysPartialUpdate(params *dcim.DcimModuleBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysRead(params *dcim.DcimModuleBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysReadOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleBaysUpdate(params *dcim.DcimModuleBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleBaysUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesBulkDelete(params *dcim.DcimModuleTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesBulkPartialUpdate(params *dcim.DcimModuleTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesBulkUpdate(params *dcim.DcimModuleTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesCreate(params *dcim.DcimModuleTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesDelete(params *dcim.DcimModuleTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesList(params *dcim.DcimModuleTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesListOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesPartialUpdate(params *dcim.DcimModuleTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesRead(params *dcim.DcimModuleTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimModuleTypesUpdate(params *dcim.DcimModuleTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModuleTypesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModulesBulkDelete(params *dcim.DcimModulesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModulesBulkPartialUpdate(params *dcim.DcimModulesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModulesBulkUpdate(params *dcim.DcimModulesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModulesCreate(params *dcim.DcimModulesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimModulesDelete(params *dcim.DcimModulesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimModulesList(params *dcim.DcimModulesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesListOK, error) {
	return nil, nil
}
func (m *mock) DcimModulesPartialUpdate(params *dcim.DcimModulesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimModulesRead(params *dcim.DcimModulesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimModulesUpdate(params *dcim.DcimModulesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimModulesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsBulkDelete(params *dcim.DcimPlatformsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsBulkPartialUpdate(params *dcim.DcimPlatformsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsBulkUpdate(params *dcim.DcimPlatformsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsCreate(params *dcim.DcimPlatformsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsDelete(params *dcim.DcimPlatformsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsList(params *dcim.DcimPlatformsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsListOK, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsPartialUpdate(params *dcim.DcimPlatformsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsRead(params *dcim.DcimPlatformsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimPlatformsUpdate(params *dcim.DcimPlatformsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPlatformsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsBulkDelete(params *dcim.DcimPowerFeedsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsBulkPartialUpdate(params *dcim.DcimPowerFeedsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsBulkUpdate(params *dcim.DcimPowerFeedsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsCreate(params *dcim.DcimPowerFeedsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsDelete(params *dcim.DcimPowerFeedsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsList(params *dcim.DcimPowerFeedsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsListOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsPartialUpdate(params *dcim.DcimPowerFeedsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsRead(params *dcim.DcimPowerFeedsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsTrace(params *dcim.DcimPowerFeedsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsTraceOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerFeedsUpdate(params *dcim.DcimPowerFeedsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerFeedsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesBulkDelete(params *dcim.DcimPowerOutletTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesBulkPartialUpdate(params *dcim.DcimPowerOutletTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesBulkUpdate(params *dcim.DcimPowerOutletTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesCreate(params *dcim.DcimPowerOutletTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesDelete(params *dcim.DcimPowerOutletTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesList(params *dcim.DcimPowerOutletTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesPartialUpdate(params *dcim.DcimPowerOutletTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesRead(params *dcim.DcimPowerOutletTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletTemplatesUpdate(params *dcim.DcimPowerOutletTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsBulkDelete(params *dcim.DcimPowerOutletsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsBulkPartialUpdate(params *dcim.DcimPowerOutletsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsBulkUpdate(params *dcim.DcimPowerOutletsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsCreate(params *dcim.DcimPowerOutletsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsDelete(params *dcim.DcimPowerOutletsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsList(params *dcim.DcimPowerOutletsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsListOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsPartialUpdate(params *dcim.DcimPowerOutletsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsRead(params *dcim.DcimPowerOutletsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsTrace(params *dcim.DcimPowerOutletsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsTraceOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerOutletsUpdate(params *dcim.DcimPowerOutletsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerOutletsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsBulkDelete(params *dcim.DcimPowerPanelsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsBulkPartialUpdate(params *dcim.DcimPowerPanelsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsBulkUpdate(params *dcim.DcimPowerPanelsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsCreate(params *dcim.DcimPowerPanelsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsDelete(params *dcim.DcimPowerPanelsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsList(params *dcim.DcimPowerPanelsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsListOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsPartialUpdate(params *dcim.DcimPowerPanelsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsRead(params *dcim.DcimPowerPanelsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPanelsUpdate(params *dcim.DcimPowerPanelsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPanelsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesBulkDelete(params *dcim.DcimPowerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesBulkPartialUpdate(params *dcim.DcimPowerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesBulkUpdate(params *dcim.DcimPowerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesCreate(params *dcim.DcimPowerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesDelete(params *dcim.DcimPowerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesList(params *dcim.DcimPowerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesPartialUpdate(params *dcim.DcimPowerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesRead(params *dcim.DcimPowerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortTemplatesUpdate(params *dcim.DcimPowerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsBulkDelete(params *dcim.DcimPowerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsBulkPartialUpdate(params *dcim.DcimPowerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsBulkUpdate(params *dcim.DcimPowerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsCreate(params *dcim.DcimPowerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsDelete(params *dcim.DcimPowerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsList(params *dcim.DcimPowerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsListOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsPartialUpdate(params *dcim.DcimPowerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsRead(params *dcim.DcimPowerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsTrace(params *dcim.DcimPowerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsTraceOK, error) {
	return nil, nil
}
func (m *mock) DcimPowerPortsUpdate(params *dcim.DcimPowerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimPowerPortsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsBulkDelete(params *dcim.DcimRackReservationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsBulkPartialUpdate(params *dcim.DcimRackReservationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsBulkUpdate(params *dcim.DcimRackReservationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsCreate(params *dcim.DcimRackReservationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsDelete(params *dcim.DcimRackReservationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsList(params *dcim.DcimRackReservationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsListOK, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsPartialUpdate(params *dcim.DcimRackReservationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsRead(params *dcim.DcimRackReservationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimRackReservationsUpdate(params *dcim.DcimRackReservationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackReservationsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesBulkDelete(params *dcim.DcimRackRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesBulkPartialUpdate(params *dcim.DcimRackRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesBulkUpdate(params *dcim.DcimRackRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesCreate(params *dcim.DcimRackRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesDelete(params *dcim.DcimRackRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesList(params *dcim.DcimRackRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesListOK, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesPartialUpdate(params *dcim.DcimRackRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesRead(params *dcim.DcimRackRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimRackRolesUpdate(params *dcim.DcimRackRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRackRolesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRacksBulkDelete(params *dcim.DcimRacksBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRacksBulkPartialUpdate(params *dcim.DcimRacksBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRacksBulkUpdate(params *dcim.DcimRacksBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRacksCreate(params *dcim.DcimRacksCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimRacksDelete(params *dcim.DcimRacksDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRacksElevation(params *dcim.DcimRacksElevationParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksElevationOK, error) {
	return nil, nil
}
func (m *mock) DcimRacksList(params *dcim.DcimRacksListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksListOK, error) {
	return nil, nil
}
func (m *mock) DcimRacksPartialUpdate(params *dcim.DcimRacksPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRacksRead(params *dcim.DcimRacksReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksReadOK, error) {
	return nil, nil
}
func (m *mock) DcimRacksUpdate(params *dcim.DcimRacksUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRacksUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesBulkDelete(params *dcim.DcimRearPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesBulkPartialUpdate(params *dcim.DcimRearPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesBulkUpdate(params *dcim.DcimRearPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesCreate(params *dcim.DcimRearPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesDelete(params *dcim.DcimRearPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesList(params *dcim.DcimRearPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesListOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesPartialUpdate(params *dcim.DcimRearPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesRead(params *dcim.DcimRearPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortTemplatesUpdate(params *dcim.DcimRearPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsBulkDelete(params *dcim.DcimRearPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsBulkPartialUpdate(params *dcim.DcimRearPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsBulkUpdate(params *dcim.DcimRearPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsCreate(params *dcim.DcimRearPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsDelete(params *dcim.DcimRearPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsList(params *dcim.DcimRearPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsListOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsPartialUpdate(params *dcim.DcimRearPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsPaths(params *dcim.DcimRearPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsPathsOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsRead(params *dcim.DcimRearPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimRearPortsUpdate(params *dcim.DcimRearPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRearPortsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRegionsBulkDelete(params *dcim.DcimRegionsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRegionsBulkPartialUpdate(params *dcim.DcimRegionsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRegionsBulkUpdate(params *dcim.DcimRegionsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRegionsCreate(params *dcim.DcimRegionsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimRegionsDelete(params *dcim.DcimRegionsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimRegionsList(params *dcim.DcimRegionsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsListOK, error) {
	return nil, nil
}
func (m *mock) DcimRegionsPartialUpdate(params *dcim.DcimRegionsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimRegionsRead(params *dcim.DcimRegionsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimRegionsUpdate(params *dcim.DcimRegionsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimRegionsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsBulkDelete(params *dcim.DcimSiteGroupsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsBulkPartialUpdate(params *dcim.DcimSiteGroupsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsBulkUpdate(params *dcim.DcimSiteGroupsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsCreate(params *dcim.DcimSiteGroupsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsDelete(params *dcim.DcimSiteGroupsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsList(params *dcim.DcimSiteGroupsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsListOK, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsPartialUpdate(params *dcim.DcimSiteGroupsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsRead(params *dcim.DcimSiteGroupsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsReadOK, error) {
	return nil, nil
}
func (m *mock) DcimSiteGroupsUpdate(params *dcim.DcimSiteGroupsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSiteGroupsUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSitesBulkDelete(params *dcim.DcimSitesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimSitesBulkPartialUpdate(params *dcim.DcimSitesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSitesBulkUpdate(params *dcim.DcimSitesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSitesCreate(params *dcim.DcimSitesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimSitesDelete(params *dcim.DcimSitesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimSitesList(params *dcim.DcimSitesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesListOK, error) {
	return nil, nil
}
func (m *mock) DcimSitesPartialUpdate(params *dcim.DcimSitesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimSitesRead(params *dcim.DcimSitesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesReadOK, error) {
	return nil, nil
}
func (m *mock) DcimSitesUpdate(params *dcim.DcimSitesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimSitesUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisBulkDelete(params *dcim.DcimVirtualChassisBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisBulkPartialUpdate(params *dcim.DcimVirtualChassisBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisBulkUpdate(params *dcim.DcimVirtualChassisBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisCreate(params *dcim.DcimVirtualChassisCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisCreateCreated, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisDelete(params *dcim.DcimVirtualChassisDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisDeleteNoContent, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisList(params *dcim.DcimVirtualChassisListParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisListOK, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisPartialUpdate(params *dcim.DcimVirtualChassisPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisPartialUpdateOK, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisRead(params *dcim.DcimVirtualChassisReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisReadOK, error) {
	return nil, nil
}
func (m *mock) DcimVirtualChassisUpdate(params *dcim.DcimVirtualChassisUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...dcim.ClientOption) (*dcim.DcimVirtualChassisUpdateOK, error) {
	return nil, nil
}
func (m *mock) SetTransport(transport runtime.ClientTransport) {
}
