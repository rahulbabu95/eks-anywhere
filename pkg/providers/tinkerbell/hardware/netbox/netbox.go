package netbox

import (
	"bytes"
	"errors"
	"fmt"
	"json"
	"log"
	"net"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
)

type Netbox struct {
	Host    string
	User    string
	Pass    string
	records []*hardware.Machine
}

// Need to return io.EOF when no more records are available.
// This method need to be a generator.
func (n *Netbox) Read() (hardware.Machine, error) {
	return hardware.Machine{}, nil
}

// 1. call Netbox, and get VM devices, maybe match on some filter of a VM device?
// return value in some kind slice of VM devices
// 2. translate from netbox data type to hardware.Machine // for testability we might want a single function here.
// do we translate them all at once or one by one when Read() is called?
// 3. Read() walks through the list of n.records and returns them one by one

func (n *Netbox) ReadFromNetbox() error {
	// call netbox
	// get the records
	// put them in n.records

	//Hardcoded as there were issues setting this as env variable in my dev desk. Shouldn't be a problem as would have different implementation for prod
	//as customers are not going to share this with us
	// token := os.Getenv("NETBOX_TOKEN")
	token := "0123456789abcdef0123456789abcdef01234567"

	if token == "" {
		return fmt.Errorf("NETBOX_TOKEN is not set")
	}

	// netboxHost := os.Getenv("NETBOX_HOST")
	netboxHost := "localhost:8000"
	if netboxHost == "" {
		return fmt.Errorf("NETBOX_HOST is not set")
	}

	transport := httptransport.New(netboxHost, client.DefaultBasePath, []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	c := client.New(transport, nil)

	//Get the devices list from netbox to populate the hardware.Machine values
	deviceReq := dcim.NewDcimDevicesListParams()
	err := n.ReadDevicesFromNetbox(c, deviceReq)

	// deviceRes, err := c.Dcim.DcimDevicesList(deviceReq, nil)
	if err != nil {
		return fmt.Errorf("cannot get Devices list: %v ", err)

	}

	err = n.ReadInterfacesFromNetbox(c)
	// interfacesRes, err := c.Dcim.DcimInterfacesList(interfacesReq, nil)
	if err != nil {
		return fmt.Errorf("error reading Interfaces list: %v ", err)

	}

	//Get the Interfaces list from netbox to populate the hardware.Machine gateway and nameserver value
	ipamReq := ipam.NewIpamIPRangesListParams()
	n.ReadIpRangeFromNetbox(c, ipamReq)
	fmt.Println("----------------------------------------ALL DEVICES---------------------------------------------------")
	for _, machine := range n.records {
		fmt.Println(machine)
	}

	return nil
}

// Field used for filtering
func (n *Netbox) ReadFromNetboxFiltered(filterTag string) error {
	//Hardcoded as there were issues setting this as env variable in my dev desk. Shouldn't be a problem as would have different implementation for prod
	//as customers are not going to share this with us
	// token := os.Getenv("NETBOX_TOKEN")
	token := "0123456789abcdef0123456789abcdef01234567"

	if token == "" {
		return fmt.Errorf("NETBOX_TOKEN is not set")
	}

	// netboxHost := os.Getenv("NETBOX_HOST")
	netboxHost := "localhost:8000"
	if netboxHost == "" {
		return fmt.Errorf("NETBOX_HOST is not set")
	}

	transport := httptransport.New(netboxHost, client.DefaultBasePath, []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	c := client.New(transport, nil)

	//Get the devices list from netbox to populate the hardware.Machine values
	deviceReq := dcim.NewDcimDevicesListParams()

	// filterTag := "eks-a"
	deviceReq.Tag = &filterTag

	err := n.ReadDevicesFromNetbox(c, deviceReq)
	if err != nil {
		return fmt.Errorf("Could not get Devices list: %v", err)
	}
	//Get the Interfaces list from netbox to populate the hardware.Machine mac value
	err = n.ReadInterfacesFromNetbox(c)

	if err != nil {
		return fmt.Errorf("error reading Interfaces list: %v ", err)
	}

	//Get the Interfaces list from netbox to populate the hardware.Machine gateway and nameserver value
	ipamReq := ipam.NewIpamIPRangesListParams()
	n.ReadIpRangeFromNetbox(c, ipamReq)

	fmt.Println("----------------------------------------FILTERED DEVICES---------------------------------------------------")
	for _, machine := range n.records {
		fmt.Println(machine)
	}
	return nil

}

//Function to check if a given ip address (ip parameter) falls in between a start (startIpRange parameter) and end (endIpRange parameter) IP address
func (n *Netbox) check(ip string, startIpRange string, endIpRange string) bool {
	startIp, _, err := net.ParseCIDR(startIpRange)
	if err != nil {
		log.Fatal(err)
	}

	endIp, _, err := net.ParseCIDR(endIpRange)
	if err != nil {
		log.Fatal(err)
	}

	trial := net.ParseIP(ip)
	if trial.To4() == nil {
		fmt.Printf("%v is not an IPv4 address\n", trial)
		return false
	}

	if bytes.Compare(trial, startIp) >= 0 && bytes.Compare(trial, endIp) <= 0 {
		// fmt.Printf("%v is between %v and %v\n", trial, startIp, endIp)
		return true
	}

	fmt.Printf("%v is NOT between %v and %v\n", trial, startIp, endIp)
	return false
}

func (n *Netbox) ReadDevicesFromNetbox(client *client.NetBoxAPI, deviceReq *dcim.DcimDevicesListParams) error {

	deviceRes, err := client.Dcim.DcimDevicesList(deviceReq, nil)
	if err != nil {
		fmt.Errorf("cannot get Devices list: %v ", err)

	}

	device_payload := deviceRes.GetPayload()
	// var n.records []hardware.Machine

	for _, device := range device_payload.Results {
		machine := new(hardware.Machine)
		machine.Hostname = *device.Name

		//Custom fields are returned as an interface by the API, type assertion to check for validity of the response
		customFields, Ok := device.CustomFields.(map[string]interface{})
		if !Ok {
			fmt.Errorf("cannot get Device Custom fields from Netbox, %v", Ok)
		}

		bmcIPMap, Ok := customFields["bmc_ip"].(map[string]interface{})
		if !Ok {
			fmt.Errorf("cannot get BMC IP from  Netbox, %v", Ok)
		}

		bmcIPVal, Ok := bmcIPMap["address"].(string)
		if !Ok {
			fmt.Errorf("cannot get BMC IP from  Netbox, %v", Ok)
		}

		//Check if the string returned in for bmc_ip is a valid IP.
		bmcIPValAdd, bmcIPValMask, err := net.ParseCIDR(bmcIPVal)
		if err != nil {
			fmt.Errorf("cannot parse BMC IP, %v", err)
		}

		machine.BMCIPAddress = bmcIPValAdd.String()
		//Get the netmask for the machine using bmc_ip as the value also contains mask.
		machine.Netmask = net.IP(bmcIPValMask.Mask).String()
		bmcUserVal, Ok := customFields["bmc_username"].(string)
		if !Ok {
			fmt.Errorf("incompatibile datatype for bmc_Username returned from netbox, %v", Ok)
		}
		machine.BMCUsername = bmcUserVal

		bmcPassVal, Ok := customFields["bmc_password"].(string)
		if !Ok {
			fmt.Errorf("incompatibile datatype for bmc_password returned from netbox, %v", Ok)
		}
		machine.BMCPassword = bmcPassVal

		diskVal, Ok := customFields["disk"].(string)
		if !Ok {
			fmt.Errorf("incompatibile datatype for disk returned from netbox, %v", Ok)
		}
		machine.Disk = diskVal

		//Obtain the machine IP from primary IP which contains IP/mask value
		machineIpAdd, _, err := net.ParseCIDR(*device.PrimaryIp4.Address)
		if err != nil {
			fmt.Errorf("Cannot parse Machine IP Address, %v", err)
		}
		machine.IPAddress = machineIpAdd.String()

		labelMap := make(map[string]string)
		controlFlag := false
		for _, tag := range device.Tags {
			// fmt.Println(*device.Name, *tag.Name)

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
		n.records = append(n.records, machine)
	}
	return nil
}

func (n *Netbox) ReadInterfacesFromNetbox(client *client.NetBoxAPI) error {
	//Get the Interfaces list from netbox to populate the hardware.Machine mac value
	interfacesReq := dcim.NewDcimInterfacesListParams()
	for idx, _ := range n.records {
		interfacesReq.Device = &n.records[idx].Hostname
		interfacesRes, err := client.Dcim.DcimInterfacesList(interfacesReq, nil)

		if err != nil {
			return fmt.Errorf("cannot get Interfaces list: %v for hostname %v ", err, interfacesReq.Device)

		}
		interfacesResults := interfacesRes.GetPayload().Results

		// Check if we get 1 or more interfaces and handle accordingly
		// No need for length checking
		if len(interfacesResults) > 1 {
			for _, interfaces := range interfacesResults {
				if len(interfaces.Tags) != 0 {
					for _, tagName := range interfaces.Tags {
						if *tagName.Name == "eks-a" {
							n.records[idx].MACAddress = *interfaces.MacAddress
						}
					}
				}
			}
		} else if len(interfacesResults) == 1 {
			n.records[idx].MACAddress = *interfacesResults[0].MacAddress
		} else {
			fmt.Errorf(("Received empty interfaces response from Netbox"))
		}
		// fmt.Println(machine.MACAddress)
	}
	return nil
}

func (n *Netbox) ReadIpRangeFromNetbox(client *client.NetBoxAPI, ipamReq *ipam.IpamIPRangesListParams) error {
	ipamRes, err := client.Ipam.IpamIPRangesList(ipamReq, nil)

	if err != nil {
		return fmt.Errorf("cannot get IP ranges list: %v ", err)
	}
	ipam_payload := ipamRes.GetPayload()
	// change the loop for optimization
	for _, ipRange := range ipam_payload.Results {
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
			return errors.New("cannot get nameservers IP from Netbox")
		}

		var nsIp hardware.Nameservers

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

		for idx, _ := range n.records {

			//Check if the IP of machine lies between the start and end address in the IP range. If so, update the nameserver and gateway value of the machine
			if n.check(n.records[idx].IPAddress, *ipRange.StartAddress, *ipRange.EndAddress) {
				n.records[idx].Nameservers = nsIp
				n.records[idx].Gateway = gatewayIpAdd.String()
			}
			// fmt.Println(machine, "mac addr: ", machine.MACAddress)
		}
	}
	return nil
}

func (n *Netbox) SerializeMachines(machines []*hardware.Machine) [] byte, error {
	ret, err := json.MarshalIndent(machines, "", " ")
	if err != nil {
		return nil, fmt.Errorf("Error in encoding Machines to byte Array: %v", Ok)
	}
	fmt.Println(string(ret))
	return ret, nil
}
