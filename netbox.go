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
	debug   bool
}

// Need to return io.EOF when no more Records are available.
// This method need to be a generator.
func (n *Netbox) Read() (Machine, error) {
	return Machine{}, nil
}

// 1. call Netbox, and get VM devices, maybe match on some filter of a VM device?
// return value in some kind slice of VM devices
// 2. translate from netbox data type to Machine // for testability we might want a single function here.
// do we translate them all at once or one by one when Read() is called?
// 3. Read() walks through the list of n.Records and returns them one by one

func (n *Netbox) ReadFromNetbox(ctx context.Context, Host string, ValidationToken string) error {
	// call netbox
	// get the Records
	// put them in n.Records

	//Hardcoded as there were issues setting this as env variable in my dev desk. Shouldn't be a problem as would have different implementation for prod
	//as customers are not going to share this with us
	token := ValidationToken
	netboxHost := Host

	transport := httptransport.New(netboxHost, client.DefaultBasePath, []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	c := client.New(transport, nil)

	//Get the devices list from netbox to populate the Machine values
	deviceReq := dcim.NewDcimDevicesListParams()
	err := n.ReadDevicesFromNetbox(ctx, c, deviceReq)

	// deviceRes, err := c.Dcim.DcimDevicesList(deviceReq, nil)
	if err != nil {
		return fmt.Errorf("cannot get Devices list: %v ", err)
	}

	err = n.ReadInterfacesFromNetbox(ctx, c)
	// interfacesRes, err := c.Dcim.DcimInterfacesList(interfacesReq, nil)
	if err != nil {
		return fmt.Errorf("error reading Interfaces list: %v ", err)

	}

	//Get the Interfaces list from netbox to populate the Machine gateway and nameserver value
	ipamReq := ipam.NewIpamIPRangesListParams()
	n.ReadIpRangeFromNetbox(ctx, c, ipamReq)

	n.logger.V(1).Info("ALL DEVICES")

	for _, machine := range n.Records {
		n.logger.V(1).Info(machine.Hostname, machine.IPAddress, machine.MACAddress, machine.BMCIPAddress)

	}

	return nil
}

// Field used for filtering
func (n *Netbox) ReadFromNetboxFiltered(ctx context.Context, Host string, ValidationToken string, filterTag string) error {
	//Hardcoded as there were issues setting this as env variable in my dev desk. Shouldn't be a problem as would have different implementation for prod
	//as customers are not going to share this with us

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
		n.logger.V(1).Info(machine.Hostname, machine.IPAddress, machine.MACAddress, machine.BMCIPAddress)
	}
	return nil

}

//Function to check if a given ip address (ip parameter) falls in between a start (startIpRange parameter) and end (endIpRange parameter) IP address
func (n *Netbox) CheckIp(ctx context.Context, ip string, startIpRange string, endIpRange string) bool {
	startIp, _, err := net.ParseCIDR(startIpRange)
	if err != nil {
		n.logger.Error(err, "error parsing IP in start range")
	}

	endIp, _, err := net.ParseCIDR(endIpRange)
	if err != nil {
		n.logger.Error(err, "error parsing IP in end range")
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

func (n *Netbox) ReadDevicesFromNetbox(ctx context.Context, client *client.NetBoxAPI, deviceReq *dcim.DcimDevicesListParams) error {

	option := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}

	deviceRes, err := client.Dcim.DcimDevicesList(deviceReq, nil, option)
	if err != nil {
		return fmt.Errorf("cannot get Devices list: %v ", err)
	}

	device_payload := deviceRes.GetPayload()

	for _, device := range device_payload.Results {
		machine := new(Machine)
		machine.Hostname = *device.Name

		//Custom fields are returned as an interface by the API, type assertion to check for validity of the response
		customFields, Ok := device.CustomFields.(map[string]interface{})
		if !Ok {
			return fmt.Errorf("cannot get Device Custom fields from Netbox, %v", Ok)
		}

		bmcIPMap, Ok := customFields["bmc_ip"].(map[string]interface{})
		if !Ok {
			return fmt.Errorf("cannot get BMC IP from  Netbox, %v", Ok)
		}

		bmcIPVal, Ok := bmcIPMap["address"].(string)
		if !Ok {
			return fmt.Errorf("cannot get BMC IP from  Netbox, %v", Ok)
		}

		//Check if the string returned in for bmc_ip is a valid IP.
		bmcIPValAdd, bmcIPValMask, err := net.ParseCIDR(bmcIPVal)
		if err != nil {
			return fmt.Errorf("cannot parse BMC IP, %v", err)
		}

		machine.BMCIPAddress = bmcIPValAdd.String()
		//Get the netmask for the machine using bmc_ip as the value also contains mask.
		machine.Netmask = net.IP(bmcIPValMask.Mask).String()
		bmcUserVal, Ok := customFields["bmc_username"].(string)
		if !Ok {
			return fmt.Errorf("incompatibile datatype for bmc_Username returned from netbox, %v", Ok)
		}
		machine.BMCUsername = bmcUserVal

		bmcPassVal, Ok := customFields["bmc_password"].(string)
		if !Ok {
			return fmt.Errorf("incompatibile datatype for bmc_password returned from netbox, %v", Ok)
		}
		machine.BMCPassword = bmcPassVal

		diskVal, Ok := customFields["disk"].(string)
		if !Ok {
			return fmt.Errorf("incompatibile datatype for disk returned from netbox, %v", Ok)
		}
		machine.Disk = diskVal

		//Obtain the machine IP from primary IP which contains IP/mask value
		machineIpAdd, _, err := net.ParseCIDR(*device.PrimaryIp4.Address)
		if err != nil {
			return fmt.Errorf("cannot parse Machine IP Address, %v", err)
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
			return fmt.Errorf("cannot get Interfaces list: %v for hostname %v ", err, interfacesReq.Device)

		}
		interfacesResults := interfacesRes.GetPayload().Results

		// Check if we get 1 or more interfaces and handle accordingly
		// No need for length checking
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
					return fmt.Errorf("cannot get ipRange Custom fields from Netbox: %v", Ok)
				}

				gatewayIpMap, Ok := customFields["gateway"].(map[string]interface{})
				if !Ok {
					return fmt.Errorf("cannot get gateway IP from Netbox: %v", Ok)
				}

				gatewayIpVal, Ok := gatewayIpMap["address"].(string)
				if !Ok {
					return fmt.Errorf("cannot get gateway IP from Netbox: %v", Ok)
				}

				//Check if the string returned in for gatewayIpVal is a valid IP.
				gatewayIpAdd, _, err := net.ParseCIDR(gatewayIpVal)
				if err != nil {
					return fmt.Errorf("cannot parse Gateway IP: %v", err)
				}

				nameserversIps, Ok := customFields["nameservers"].([]interface{})
				if !Ok {
					return fmt.Errorf("cannot get nameservers IP from Netbox: %v", Ok)
				}

				var nsIp Nameservers

				for _, nameserverIp := range nameserversIps {
					nameserversIpsMap, Ok := nameserverIp.(map[string]interface{})
					if !Ok {
						return fmt.Errorf("here cannot get nameservers IP from Netbox: %v", Ok)
					}

					nameserverIpVal, Ok := nameserversIpsMap["address"].(string)
					if !Ok {
						return fmt.Errorf("here cannot get nameservers IP from Netbox: %v", Ok)
					}

					//Parse CIDR reasoning and explanation about the type returned by netbox
					//Check if string returned by nameserverIpVal is a valid IP.
					nameserverIpAdd, _, err := net.ParseCIDR(nameserverIpVal)
					if err != nil {
						return fmt.Errorf("cannot parse Nameservers IP: %v", err)
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

func (n *Netbox) SerializeMachines(machines []*Machine) ([]byte, error) {
	ret, err := json.MarshalIndent(machines, "", " ")
	if err != nil {
		return nil, fmt.Errorf("error in encoding Machines to byte Array: %v", err)
	}
	return ret, nil
}
