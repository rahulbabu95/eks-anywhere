package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
)

type Netbox struct {
	Host    string
	User    string
	Pass    string
	Records []*Machine
	logger  logr.Logger
}

type IpError struct {
	act string
}

func (i *IpError) Error() string {
	return fmt.Sprintf("Error Parsing IP: expected: CIDR Address, got: %v", i.act)
}

func (i *IpError) Is(target error) bool {
	t, ok := target.(*IpError)
	if !ok {
		return false
	}
	return (i.act == t.act || t.act == "")
}

type TypeAssertError struct {
	field string
	exp   string
	act   string
}

func (t *TypeAssertError) Error() string {
	return fmt.Sprintf("Error in Type Assertion: field: %v, expected: %v, got: %v", t.field, t.exp, t.act)
}

func (t *TypeAssertError) Is(target error) bool {
	tar, ok := target.(*TypeAssertError)
	if !ok {
		return false
	}
	return (t.field == tar.field || t.field == "") && (t.exp == tar.exp || t.exp == "") && (t.act == tar.act || t.act == "")
}

type NetboxError struct {
	msg    string
	errMsg string
}

func (n *NetboxError) Error() string {
	return fmt.Sprintf(n.msg + " : " + n.errMsg)
}

func (n *NetboxError) Is(target error) bool {
	tar, ok := target.(*NetboxError)
	if !ok {
		return false
	}
	return (n.msg == tar.msg || n.msg == "") && (n.errMsg == tar.errMsg || n.errMsg == "")
}

// ReadFromNetbox Function calls 3 helper functions which makes API calls to Netbox and sets Records field with required Hardware value
func (n *Netbox) ReadFromNetbox(ctx context.Context, Host string, ValidationToken string) error {

	token := ValidationToken
	netboxHost := Host

	transport := httptransport.New(netboxHost, client.DefaultBasePath, []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	c := client.New(transport, nil)

	//Get the devices list from netbox to populate the Machine values
	deviceReq := dcim.NewDcimDevicesListParams()
	err := n.ReadDevicesFromNetbox(ctx, c, deviceReq)

	if err != nil {
		return fmt.Errorf("cannot get Devices list: %v ", err)
	}

	err = n.ReadInterfacesFromNetbox(ctx, c)
	if err != nil {
		return fmt.Errorf("error reading Interfaces list: %v ", err)

	}

	// Get the Interfaces list from netbox to populate the Machine gateway and nameserver value
	ipamReq := ipam.NewIpamIPRangesListParams()
	n.ReadIpRangeFromNetbox(ctx, c, ipamReq)

	n.logger.V(1).Info("ALL DEVICES")

	for _, machine := range n.Records {
		n.logger.V(1).Info("Device Read: ", "Host", machine.Hostname, "IP", machine.IPAddress, "MAC", machine.MACAddress, "BMC-IP", machine.BMCIPAddress)

	}

	return nil
}

// ReadFromNetboxFiltered Function calls 3 helper functions with a filter tag which makes API calls to Netbox and sets Records field with required Hardware value
func (n *Netbox) ReadFromNetboxFiltered(ctx context.Context, Host string, ValidationToken string, filterTag string) error {

	token := ValidationToken
	netboxHost := Host

	transport := httptransport.New(netboxHost, client.DefaultBasePath, []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	c := client.New(transport, nil)

	//Get the devices list from netbox to populate the Machine values
	deviceReq := dcim.NewDcimDevicesListParams()
	deviceReq.Tag = &filterTag

	err := n.ReadDevicesFromNetbox(ctx, c, deviceReq)
	if err != nil {
		return fmt.Errorf("could not get Devices list: %v", err)
	}
	//Get the Interfaces list from netbox to populate the Machine mac value
	err = n.ReadInterfacesFromNetbox(ctx, c)

	if err != nil {
		return fmt.Errorf("error reading Interfaces list: %v ", err)
	}

	//Get the Interfaces list from netbox to populate the Machine gateway and nameserver value
	ipamReq := ipam.NewIpamIPRangesListParams()
	n.ReadIpRangeFromNetbox(ctx, c, ipamReq)

	n.logger.V(1).Info("FILTERED DEVICES")
	for _, machine := range n.Records {
		n.logger.V(1).Info("Device Read: ", "Host", machine.Hostname, "IP", machine.IPAddress, "MAC", machine.MACAddress, "BMC-IP", machine.BMCIPAddress)
	}
	return nil

}

// CheckIp Function to check if a given ip address falls in between a start and end IP address
func (n *Netbox) CheckIp(ctx context.Context, ip string, startIpRange string, endIpRange string) bool {
	startIp, _, err := net.ParseCIDR(startIpRange)
	if err != nil {
		n.logger.Error(err, "error parsing IP in start range")
		return false
	}

	endIp, _, err := net.ParseCIDR(endIpRange)
	if err != nil {
		n.logger.Error(err, "error parsing IP in end range")
		return false
	}

	trial := net.ParseIP(ip)
	if trial.To4() == nil {

		n.logger.Error(err, "error parsing IP to IP4 address")
		return false
	}

	if bytes.Compare(trial, startIp) >= 0 && bytes.Compare(trial, endIp) <= 0 {
		return true
	}

	return false
}

// ReadDevicesFromNetbox Function fetches the devices list from Netbox and sets HostName, BMC info, Ip addr, Disk and Labels
func (n *Netbox) ReadDevicesFromNetbox(ctx context.Context, client *client.NetBoxAPI, deviceReq *dcim.DcimDevicesListParams) error {

	option := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}

	deviceRes, err := client.Dcim.DcimDevicesList(deviceReq, nil, option)
	if err != nil {
		return &NetboxError{"cannot get Devices list", err.Error()}
	}

	device_payload := deviceRes.GetPayload()

	for _, device := range device_payload.Results {
		machine := new(Machine)
		machine.Hostname = *device.Name

		//Custom fields are returned as an interface by the API, type assertion to check for validity of the response
		customFields, Ok := device.CustomFields.(map[string]interface{})
		if !Ok {
			return &TypeAssertError{"CustomFields", "map[string]interface{}", fmt.Sprintf("%T", device.CustomFields)}
		}

		bmcIPMap, Ok := customFields["bmc_ip"].(map[string]interface{})
		if !Ok {

			return &TypeAssertError{"bmc_ip", "map[string]interface{}", fmt.Sprintf("%T", customFields["bmc_ip"])}
			//return fmt.Errorf("type Assertion error for BMC IP, %v", Ok)
		}

		bmcIPVal, Ok := bmcIPMap["address"].(string)
		if !Ok {
			return &TypeAssertError{"bmc_ip_address", "string", fmt.Sprintf("%T", bmcIPMap["address"])}
		}

		//Check if the string returned in for bmc_ip is a valid IP.
		bmcIPValAdd, bmcIPValMask, err := net.ParseCIDR(bmcIPVal)
		if err != nil {
			return &IpError{bmcIPVal}
		}

		machine.BMCIPAddress = bmcIPValAdd.String()
		//Get the netmask for the machine using bmc_ip as the value also contains mask.
		machine.Netmask = net.IP(bmcIPValMask.Mask).String()
		bmcUserVal, Ok := customFields["bmc_username"].(string)
		if !Ok {
			return &TypeAssertError{"bmc_username", "string", fmt.Sprintf("%T", customFields["bmc_username"])}
		}
		machine.BMCUsername = bmcUserVal

		bmcPassVal, Ok := customFields["bmc_password"].(string)
		if !Ok {
			return &TypeAssertError{"bmc_password", "string", fmt.Sprintf("%T", customFields["bmc_password"])}
		}
		machine.BMCPassword = bmcPassVal

		diskVal, Ok := customFields["disk"].(string)
		if !Ok {
			return &TypeAssertError{"disk", "string", fmt.Sprintf("%T", customFields["disk"])}
		}
		machine.Disk = diskVal

		//Obtain the machine IP from primary IP which contains IP/mask value
		machineIpAdd, _, err := net.ParseCIDR(*device.PrimaryIp4.Address)
		if err != nil {

			return &IpError{*device.PrimaryIp4.Address}
			// return fmt.Errorf("cannot parse Machine IP Address, %v", err)
		}
		machine.IPAddress = machineIpAdd.String()

		labelMap := make(map[string]string)
		controlFlag := false
		for _, tag := range device.Tags {

			if *tag.Name == "control-plane" {

				labelMap["type"] = "control-plane"
				controlFlag = !controlFlag
				break
			}
		}
		if !controlFlag {
			labelMap["type"] = "worker-plane"
		}
		machine.Labels = labelMap
		n.Records = append(n.Records, machine)
	}

	n.logger.Info("step 1 - Reading devices successul", "num_machines", len(n.Records))
	return nil
}

// ReadInterfacesFromNetbox Function fetches the interfaces list from Netbox and sets the MAC address for each record
func (n *Netbox) ReadInterfacesFromNetbox(ctx context.Context, client *client.NetBoxAPI) error {
	//Get the Interfaces list from netbox to populate the Machine mac value
	interfacesReq := dcim.NewDcimInterfacesListParams()

	option := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}
	for _, record := range n.Records {
		interfacesReq.Device = &record.Hostname
		interfacesRes, err := client.Dcim.DcimInterfacesList(interfacesReq, nil, option)

		if err != nil {
			return &NetboxError{"cannot get Interfaces list", err.Error()}
		}
		interfacesResults := interfacesRes.GetPayload().Results
		if len(interfacesResults) == 1 {
			record.MACAddress = *interfacesResults[0].MacAddress
		} else {
			for _, interfaces := range interfacesResults {
				for _, tagName := range interfaces.Tags {
					if *tagName.Name == "eks-a" {
						record.MACAddress = *interfaces.MacAddress
					}
				}
			}
		}
	}

	n.logger.Info("step 2 - Reading intefaces successful, MAC addresses set")

	return nil
}

// ReadIpRangeFromNetbox Function fetches IP ranges from Netbox and sets the Gateway and nameserver address for each record
func (n *Netbox) ReadIpRangeFromNetbox(ctx context.Context, client *client.NetBoxAPI, ipamReq *ipam.IpamIPRangesListParams) error {

	option := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}
	ipamRes, err := client.Ipam.IpamIPRangesList(ipamReq, nil, option)
	if err != nil {
		return fmt.Errorf("cannot get IP ranges list: %v ", err)
	}
	ipam_payload := ipamRes.GetPayload()

	for _, record := range n.Records {
		for _, ipRange := range ipam_payload.Results {
			//Check if the IP of machine lies between the start and end address in the IP range. If so, update the nameserver and gateway value of the machine
			if n.CheckIp(ctx, record.IPAddress, *ipRange.StartAddress, *ipRange.EndAddress) {
				customFields, Ok := ipRange.CustomFields.(map[string]interface{})
				if !Ok {
					return &TypeAssertError{"customFields", "map[string]interface{}", fmt.Sprintf("%T", ipRange.CustomFields)}
				}

				gatewayIpMap, Ok := customFields["gateway"].(map[string]interface{})
				if !Ok {
					return &TypeAssertError{"gatewayIP", "map[string]interface{}", fmt.Sprintf("%T", customFields["gateway"])}
				}

				gatewayIpVal, Ok := gatewayIpMap["address"].(string)
				if !Ok {
					return &TypeAssertError{"gatewayAddr", "string", fmt.Sprintf("%T", gatewayIpMap["address"])}
				}

				//Check if the string returned in for gatewayIpVal is a valid IP.
				gatewayIpAdd, _, err := net.ParseCIDR(gatewayIpVal)
				if err != nil {
					return &IpError{gatewayIpVal}
				}

				nameserversIps, Ok := customFields["nameservers"].([]interface{})
				if !Ok {
					return &TypeAssertError{"nameservers", "[]interface{}", fmt.Sprintf("%T", customFields["nameservers"])}
				}

				var nsIp Nameservers

				for _, nameserverIp := range nameserversIps {
					nameserversIpsMap, Ok := nameserverIp.(map[string]interface{})
					if !Ok {
						return &TypeAssertError{"nameserversIPMap", "map[string]interface{}", fmt.Sprintf("%T", nameserverIp)}
					}

					nameserverIpVal, Ok := nameserversIpsMap["address"].(string)
					if !Ok {
						return &TypeAssertError{"nameserversIPMap", "string", fmt.Sprintf("%T", nameserversIpsMap["address"])}
					}

					//Parse CIDR reasoning and explanation about the type returned by netbox
					//Check if string returned by nameserverIpVal is a valid IP.
					nameserverIpAdd, _, err := net.ParseCIDR(nameserverIpVal)
					if err != nil {
						return &IpError{nameserverIpVal}
					}

					nsIp = append(nsIp, nameserverIpAdd.String())
				}
				record.Nameservers = nsIp
				record.Gateway = gatewayIpAdd.String()
			}
		}
	}

	n.logger.Info("step 3 - Reading IPAM data successful, all DCIM calls are complete")

	return nil
}

// SerializeMachines Function takes in a arry of machine slices as input and converts them into byte array.
func (n *Netbox) SerializeMachines(machines []*Machine) ([]byte, error) {
	ret, err := json.MarshalIndent(machines, "", " ")
	if err != nil {
		return nil, fmt.Errorf("error in encoding Machines to byte Array: %v", err)
	}
	return ret, nil
}
