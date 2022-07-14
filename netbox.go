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

	//c.Dcim = 

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

type mock struct {}
func (m *mock) DcimCablesBulkDelete(params *DcimCablesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimCablesBulkPartialUpdate(params *DcimCablesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesBulkPartialUpdateOK, error)

func (m *mock)  DcimCablesBulkUpdate(params *DcimCablesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesBulkUpdateOK, error)

func (m *mock) DcimCablesCreate(params *DcimCablesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesCreateCreated, error)

DcimCablesDelete(params *DcimCablesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesDeleteNoContent, error)

DcimCablesList(params *DcimCablesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesListOK, error)

DcimCablesPartialUpdate(params *DcimCablesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesPartialUpdateOK, error)

DcimCablesRead(params *DcimCablesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesReadOK, error)

DcimCablesUpdate(params *DcimCablesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesUpdateOK, error)

DcimConnectedDeviceList(params *DcimConnectedDeviceListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConnectedDeviceListOK, error)

DcimConsolePortTemplatesBulkDelete(params *DcimConsolePortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesBulkDeleteNoContent, error)

DcimConsolePortTemplatesBulkPartialUpdate(params *DcimConsolePortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesBulkPartialUpdateOK, error)

DcimConsolePortTemplatesBulkUpdate(params *DcimConsolePortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesBulkUpdateOK, error)

DcimConsolePortTemplatesCreate(params *DcimConsolePortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesCreateCreated, error)

DcimConsolePortTemplatesDelete(params *DcimConsolePortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesDeleteNoContent, error)

DcimConsolePortTemplatesList(params *DcimConsolePortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesListOK, error)

DcimConsolePortTemplatesPartialUpdate(params *DcimConsolePortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesPartialUpdateOK, error)

DcimConsolePortTemplatesRead(params *DcimConsolePortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesReadOK, error)

DcimConsolePortTemplatesUpdate(params *DcimConsolePortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesUpdateOK, error)

DcimConsolePortsBulkDelete(params *DcimConsolePortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsBulkDeleteNoContent, error)

DcimConsolePortsBulkPartialUpdate(params *DcimConsolePortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsBulkPartialUpdateOK, error)

DcimConsolePortsBulkUpdate(params *DcimConsolePortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsBulkUpdateOK, error)

DcimConsolePortsCreate(params *DcimConsolePortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsCreateCreated, error)

DcimConsolePortsDelete(params *DcimConsolePortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsDeleteNoContent, error)

DcimConsolePortsList(params *DcimConsolePortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsListOK, error)

DcimConsolePortsPartialUpdate(params *DcimConsolePortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsPartialUpdateOK, error)

DcimConsolePortsRead(params *DcimConsolePortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsReadOK, error)

DcimConsolePortsTrace(params *DcimConsolePortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsTraceOK, error)

DcimConsolePortsUpdate(params *DcimConsolePortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsUpdateOK, error)

DcimConsoleServerPortTemplatesBulkDelete(params *DcimConsoleServerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesBulkDeleteNoContent, error)

DcimConsoleServerPortTemplatesBulkPartialUpdate(params *DcimConsoleServerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesBulkPartialUpdateOK, error)

DcimConsoleServerPortTemplatesBulkUpdate(params *DcimConsoleServerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesBulkUpdateOK, error)

DcimConsoleServerPortTemplatesCreate(params *DcimConsoleServerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesCreateCreated, error)

DcimConsoleServerPortTemplatesDelete(params *DcimConsoleServerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesDeleteNoContent, error)

DcimConsoleServerPortTemplatesList(params *DcimConsoleServerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesListOK, error)

DcimConsoleServerPortTemplatesPartialUpdate(params *DcimConsoleServerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesPartialUpdateOK, error)

DcimConsoleServerPortTemplatesRead(params *DcimConsoleServerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesReadOK, error)

DcimConsoleServerPortTemplatesUpdate(params *DcimConsoleServerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesUpdateOK, error)

DcimConsoleServerPortsBulkDelete(params *DcimConsoleServerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsBulkDeleteNoContent, error)

DcimConsoleServerPortsBulkPartialUpdate(params *DcimConsoleServerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsBulkPartialUpdateOK, error)

DcimConsoleServerPortsBulkUpdate(params *DcimConsoleServerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsBulkUpdateOK, error)

DcimConsoleServerPortsCreate(params *DcimConsoleServerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsCreateCreated, error)

DcimConsoleServerPortsDelete(params *DcimConsoleServerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsDeleteNoContent, error)

DcimConsoleServerPortsList(params *DcimConsoleServerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsListOK, error)

DcimConsoleServerPortsPartialUpdate(params *DcimConsoleServerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsPartialUpdateOK, error)

DcimConsoleServerPortsRead(params *DcimConsoleServerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsReadOK, error)

DcimConsoleServerPortsTrace(params *DcimConsoleServerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsTraceOK, error)

DcimConsoleServerPortsUpdate(params *DcimConsoleServerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsUpdateOK, error)

DcimDeviceBayTemplatesBulkDelete(params *DcimDeviceBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesBulkDeleteNoContent, error)

DcimDeviceBayTemplatesBulkPartialUpdate(params *DcimDeviceBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesBulkPartialUpdateOK, error)

DcimDeviceBayTemplatesBulkUpdate(params *DcimDeviceBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesBulkUpdateOK, error)

DcimDeviceBayTemplatesCreate(params *DcimDeviceBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesCreateCreated, error)

DcimDeviceBayTemplatesDelete(params *DcimDeviceBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesDeleteNoContent, error)

DcimDeviceBayTemplatesList(params *DcimDeviceBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesListOK, error)

DcimDeviceBayTemplatesPartialUpdate(params *DcimDeviceBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesPartialUpdateOK, error)

DcimDeviceBayTemplatesRead(params *DcimDeviceBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesReadOK, error)

DcimDeviceBayTemplatesUpdate(params *DcimDeviceBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesUpdateOK, error)

DcimDeviceBaysBulkDelete(params *DcimDeviceBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysBulkDeleteNoContent, error)

DcimDeviceBaysBulkPartialUpdate(params *DcimDeviceBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysBulkPartialUpdateOK, error)

DcimDeviceBaysBulkUpdate(params *DcimDeviceBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysBulkUpdateOK, error)

DcimDeviceBaysCreate(params *DcimDeviceBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysCreateCreated, error)

DcimDeviceBaysDelete(params *DcimDeviceBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysDeleteNoContent, error)

DcimDeviceBaysList(params *DcimDeviceBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysListOK, error)

DcimDeviceBaysPartialUpdate(params *DcimDeviceBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysPartialUpdateOK, error)

DcimDeviceBaysRead(params *DcimDeviceBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysReadOK, error)

DcimDeviceBaysUpdate(params *DcimDeviceBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysUpdateOK, error)

DcimDeviceRolesBulkDelete(params *DcimDeviceRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesBulkDeleteNoContent, error)

DcimDeviceRolesBulkPartialUpdate(params *DcimDeviceRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesBulkPartialUpdateOK, error)

DcimDeviceRolesBulkUpdate(params *DcimDeviceRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesBulkUpdateOK, error)

DcimDeviceRolesCreate(params *DcimDeviceRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesCreateCreated, error)

DcimDeviceRolesDelete(params *DcimDeviceRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesDeleteNoContent, error)

DcimDeviceRolesList(params *DcimDeviceRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesListOK, error)

DcimDeviceRolesPartialUpdate(params *DcimDeviceRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesPartialUpdateOK, error)

DcimDeviceRolesRead(params *DcimDeviceRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesReadOK, error)

DcimDeviceRolesUpdate(params *DcimDeviceRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesUpdateOK, error)

DcimDeviceTypesBulkDelete(params *DcimDeviceTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesBulkDeleteNoContent, error)

DcimDeviceTypesBulkPartialUpdate(params *DcimDeviceTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesBulkPartialUpdateOK, error)

DcimDeviceTypesBulkUpdate(params *DcimDeviceTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesBulkUpdateOK, error)

DcimDeviceTypesCreate(params *DcimDeviceTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesCreateCreated, error)

DcimDeviceTypesDelete(params *DcimDeviceTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesDeleteNoContent, error)

DcimDeviceTypesList(params *DcimDeviceTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesListOK, error)

DcimDeviceTypesPartialUpdate(params *DcimDeviceTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesPartialUpdateOK, error)

DcimDeviceTypesRead(params *DcimDeviceTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesReadOK, error)

DcimDeviceTypesUpdate(params *DcimDeviceTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesUpdateOK, error)

DcimDevicesBulkDelete(params *DcimDevicesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesBulkDeleteNoContent, error)

DcimDevicesBulkPartialUpdate(params *DcimDevicesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesBulkPartialUpdateOK, error)

DcimDevicesBulkUpdate(params *DcimDevicesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesBulkUpdateOK, error)

DcimDevicesCreate(params *DcimDevicesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesCreateCreated, error)

DcimDevicesDelete(params *DcimDevicesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesDeleteNoContent, error)

DcimDevicesList(params *DcimDevicesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesListOK, error)

DcimDevicesNapalm(params *DcimDevicesNapalmParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesNapalmOK, error)

DcimDevicesPartialUpdate(params *DcimDevicesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesPartialUpdateOK, error)

DcimDevicesRead(params *DcimDevicesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesReadOK, error)

DcimDevicesUpdate(params *DcimDevicesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesUpdateOK, error)

DcimFrontPortTemplatesBulkDelete(params *DcimFrontPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesBulkDeleteNoContent, error)

DcimFrontPortTemplatesBulkPartialUpdate(params *DcimFrontPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesBulkPartialUpdateOK, error)

DcimFrontPortTemplatesBulkUpdate(params *DcimFrontPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesBulkUpdateOK, error)

DcimFrontPortTemplatesCreate(params *DcimFrontPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesCreateCreated, error)

DcimFrontPortTemplatesDelete(params *DcimFrontPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesDeleteNoContent, error)

DcimFrontPortTemplatesList(params *DcimFrontPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesListOK, error)

DcimFrontPortTemplatesPartialUpdate(params *DcimFrontPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesPartialUpdateOK, error)

DcimFrontPortTemplatesRead(params *DcimFrontPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesReadOK, error)

DcimFrontPortTemplatesUpdate(params *DcimFrontPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesUpdateOK, error)

DcimFrontPortsBulkDelete(params *DcimFrontPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsBulkDeleteNoContent, error)

DcimFrontPortsBulkPartialUpdate(params *DcimFrontPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsBulkPartialUpdateOK, error)

DcimFrontPortsBulkUpdate(params *DcimFrontPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsBulkUpdateOK, error)

DcimFrontPortsCreate(params *DcimFrontPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsCreateCreated, error)

DcimFrontPortsDelete(params *DcimFrontPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsDeleteNoContent, error)

DcimFrontPortsList(params *DcimFrontPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsListOK, error)

DcimFrontPortsPartialUpdate(params *DcimFrontPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsPartialUpdateOK, error)

DcimFrontPortsPaths(params *DcimFrontPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsPathsOK, error)

DcimFrontPortsRead(params *DcimFrontPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsReadOK, error)

DcimFrontPortsUpdate(params *DcimFrontPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsUpdateOK, error)

DcimInterfaceTemplatesBulkDelete(params *DcimInterfaceTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesBulkDeleteNoContent, error)

DcimInterfaceTemplatesBulkPartialUpdate(params *DcimInterfaceTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesBulkPartialUpdateOK, error)

DcimInterfaceTemplatesBulkUpdate(params *DcimInterfaceTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesBulkUpdateOK, error)

DcimInterfaceTemplatesCreate(params *DcimInterfaceTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesCreateCreated, error)

DcimInterfaceTemplatesDelete(params *DcimInterfaceTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesDeleteNoContent, error)

DcimInterfaceTemplatesList(params *DcimInterfaceTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesListOK, error)

DcimInterfaceTemplatesPartialUpdate(params *DcimInterfaceTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesPartialUpdateOK, error)

DcimInterfaceTemplatesRead(params *DcimInterfaceTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesReadOK, error)

DcimInterfaceTemplatesUpdate(params *DcimInterfaceTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesUpdateOK, error)

DcimInterfacesBulkDelete(params *DcimInterfacesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesBulkDeleteNoContent, error)

DcimInterfacesBulkPartialUpdate(params *DcimInterfacesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesBulkPartialUpdateOK, error)

DcimInterfacesBulkUpdate(params *DcimInterfacesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesBulkUpdateOK, error)

DcimInterfacesCreate(params *DcimInterfacesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesCreateCreated, error)

DcimInterfacesDelete(params *DcimInterfacesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesDeleteNoContent, error)

DcimInterfacesList(params *DcimInterfacesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesListOK, error)

DcimInterfacesPartialUpdate(params *DcimInterfacesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesPartialUpdateOK, error)

DcimInterfacesRead(params *DcimInterfacesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesReadOK, error)

DcimInterfacesTrace(params *DcimInterfacesTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesTraceOK, error)

DcimInterfacesUpdate(params *DcimInterfacesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesUpdateOK, error)

DcimInventoryItemRolesBulkDelete(params *DcimInventoryItemRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesBulkDeleteNoContent, error)

DcimInventoryItemRolesBulkPartialUpdate(params *DcimInventoryItemRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesBulkPartialUpdateOK, error)

DcimInventoryItemRolesBulkUpdate(params *DcimInventoryItemRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesBulkUpdateOK, error)

DcimInventoryItemRolesCreate(params *DcimInventoryItemRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesCreateCreated, error)

DcimInventoryItemRolesDelete(params *DcimInventoryItemRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesDeleteNoContent, error)

DcimInventoryItemRolesList(params *DcimInventoryItemRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesListOK, error)

DcimInventoryItemRolesPartialUpdate(params *DcimInventoryItemRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesPartialUpdateOK, error)

DcimInventoryItemRolesRead(params *DcimInventoryItemRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesReadOK, error)

DcimInventoryItemRolesUpdate(params *DcimInventoryItemRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesUpdateOK, error)

DcimInventoryItemTemplatesBulkDelete(params *DcimInventoryItemTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesBulkDeleteNoContent, error)

DcimInventoryItemTemplatesBulkPartialUpdate(params *DcimInventoryItemTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesBulkPartialUpdateOK, error)

DcimInventoryItemTemplatesBulkUpdate(params *DcimInventoryItemTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesBulkUpdateOK, error)

DcimInventoryItemTemplatesCreate(params *DcimInventoryItemTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesCreateCreated, error)

DcimInventoryItemTemplatesDelete(params *DcimInventoryItemTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesDeleteNoContent, error)

DcimInventoryItemTemplatesList(params *DcimInventoryItemTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesListOK, error)

DcimInventoryItemTemplatesPartialUpdate(params *DcimInventoryItemTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesPartialUpdateOK, error)

DcimInventoryItemTemplatesRead(params *DcimInventoryItemTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesReadOK, error)

DcimInventoryItemTemplatesUpdate(params *DcimInventoryItemTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesUpdateOK, error)

DcimInventoryItemsBulkDelete(params *DcimInventoryItemsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsBulkDeleteNoContent, error)

DcimInventoryItemsBulkPartialUpdate(params *DcimInventoryItemsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsBulkPartialUpdateOK, error)

DcimInventoryItemsBulkUpdate(params *DcimInventoryItemsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsBulkUpdateOK, error)

DcimInventoryItemsCreate(params *DcimInventoryItemsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsCreateCreated, error)

DcimInventoryItemsDelete(params *DcimInventoryItemsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsDeleteNoContent, error)

DcimInventoryItemsList(params *DcimInventoryItemsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsListOK, error)

DcimInventoryItemsPartialUpdate(params *DcimInventoryItemsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsPartialUpdateOK, error)

DcimInventoryItemsRead(params *DcimInventoryItemsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsReadOK, error)

DcimInventoryItemsUpdate(params *DcimInventoryItemsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsUpdateOK, error)

DcimLocationsBulkDelete(params *DcimLocationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsBulkDeleteNoContent, error)

DcimLocationsBulkPartialUpdate(params *DcimLocationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsBulkPartialUpdateOK, error)

DcimLocationsBulkUpdate(params *DcimLocationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsBulkUpdateOK, error)

DcimLocationsCreate(params *DcimLocationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsCreateCreated, error)

DcimLocationsDelete(params *DcimLocationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsDeleteNoContent, error)

DcimLocationsList(params *DcimLocationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsListOK, error)

DcimLocationsPartialUpdate(params *DcimLocationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsPartialUpdateOK, error)

DcimLocationsRead(params *DcimLocationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsReadOK, error)

DcimLocationsUpdate(params *DcimLocationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsUpdateOK, error)

DcimManufacturersBulkDelete(params *DcimManufacturersBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersBulkDeleteNoContent, error)

DcimManufacturersBulkPartialUpdate(params *DcimManufacturersBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersBulkPartialUpdateOK, error)

DcimManufacturersBulkUpdate(params *DcimManufacturersBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersBulkUpdateOK, error)

DcimManufacturersCreate(params *DcimManufacturersCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersCreateCreated, error)

DcimManufacturersDelete(params *DcimManufacturersDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersDeleteNoContent, error)

DcimManufacturersList(params *DcimManufacturersListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersListOK, error)

DcimManufacturersPartialUpdate(params *DcimManufacturersPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersPartialUpdateOK, error)

DcimManufacturersRead(params *DcimManufacturersReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersReadOK, error)

DcimManufacturersUpdate(params *DcimManufacturersUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersUpdateOK, error)

DcimModuleBayTemplatesBulkDelete(params *DcimModuleBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesBulkDeleteNoContent, error)

DcimModuleBayTemplatesBulkPartialUpdate(params *DcimModuleBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesBulkPartialUpdateOK, error)

DcimModuleBayTemplatesBulkUpdate(params *DcimModuleBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesBulkUpdateOK, error)

DcimModuleBayTemplatesCreate(params *DcimModuleBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesCreateCreated, error)

DcimModuleBayTemplatesDelete(params *DcimModuleBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesDeleteNoContent, error)

DcimModuleBayTemplatesList(params *DcimModuleBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesListOK, error)

DcimModuleBayTemplatesPartialUpdate(params *DcimModuleBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesPartialUpdateOK, error)

DcimModuleBayTemplatesRead(params *DcimModuleBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesReadOK, error)

DcimModuleBayTemplatesUpdate(params *DcimModuleBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesUpdateOK, error)

DcimModuleBaysBulkDelete(params *DcimModuleBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysBulkDeleteNoContent, error)

DcimModuleBaysBulkPartialUpdate(params *DcimModuleBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysBulkPartialUpdateOK, error)

DcimModuleBaysBulkUpdate(params *DcimModuleBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysBulkUpdateOK, error)

DcimModuleBaysCreate(params *DcimModuleBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysCreateCreated, error)

DcimModuleBaysDelete(params *DcimModuleBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysDeleteNoContent, error)

DcimModuleBaysList(params *DcimModuleBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysListOK, error)

DcimModuleBaysPartialUpdate(params *DcimModuleBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysPartialUpdateOK, error)

DcimModuleBaysRead(params *DcimModuleBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysReadOK, error)

DcimModuleBaysUpdate(params *DcimModuleBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysUpdateOK, error)

DcimModuleTypesBulkDelete(params *DcimModuleTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesBulkDeleteNoContent, error)

DcimModuleTypesBulkPartialUpdate(params *DcimModuleTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesBulkPartialUpdateOK, error)

DcimModuleTypesBulkUpdate(params *DcimModuleTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesBulkUpdateOK, error)

DcimModuleTypesCreate(params *DcimModuleTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesCreateCreated, error)

DcimModuleTypesDelete(params *DcimModuleTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesDeleteNoContent, error)

DcimModuleTypesList(params *DcimModuleTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesListOK, error)

DcimModuleTypesPartialUpdate(params *DcimModuleTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesPartialUpdateOK, error)

DcimModuleTypesRead(params *DcimModuleTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesReadOK, error)

DcimModuleTypesUpdate(params *DcimModuleTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesUpdateOK, error)

DcimModulesBulkDelete(params *DcimModulesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesBulkDeleteNoContent, error)

DcimModulesBulkPartialUpdate(params *DcimModulesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesBulkPartialUpdateOK, error)

DcimModulesBulkUpdate(params *DcimModulesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesBulkUpdateOK, error)

DcimModulesCreate(params *DcimModulesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesCreateCreated, error)

DcimModulesDelete(params *DcimModulesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesDeleteNoContent, error)

DcimModulesList(params *DcimModulesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesListOK, error)

DcimModulesPartialUpdate(params *DcimModulesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesPartialUpdateOK, error)

DcimModulesRead(params *DcimModulesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesReadOK, error)

DcimModulesUpdate(params *DcimModulesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesUpdateOK, error)

DcimPlatformsBulkDelete(params *DcimPlatformsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsBulkDeleteNoContent, error)

DcimPlatformsBulkPartialUpdate(params *DcimPlatformsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsBulkPartialUpdateOK, error)

DcimPlatformsBulkUpdate(params *DcimPlatformsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsBulkUpdateOK, error)

DcimPlatformsCreate(params *DcimPlatformsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsCreateCreated, error)

DcimPlatformsDelete(params *DcimPlatformsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsDeleteNoContent, error)

DcimPlatformsList(params *DcimPlatformsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsListOK, error)

DcimPlatformsPartialUpdate(params *DcimPlatformsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsPartialUpdateOK, error)

DcimPlatformsRead(params *DcimPlatformsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsReadOK, error)

DcimPlatformsUpdate(params *DcimPlatformsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsUpdateOK, error)

DcimPowerFeedsBulkDelete(params *DcimPowerFeedsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsBulkDeleteNoContent, error)

DcimPowerFeedsBulkPartialUpdate(params *DcimPowerFeedsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsBulkPartialUpdateOK, error)

DcimPowerFeedsBulkUpdate(params *DcimPowerFeedsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsBulkUpdateOK, error)

DcimPowerFeedsCreate(params *DcimPowerFeedsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsCreateCreated, error)

DcimPowerFeedsDelete(params *DcimPowerFeedsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsDeleteNoContent, error)

DcimPowerFeedsList(params *DcimPowerFeedsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsListOK, error)

DcimPowerFeedsPartialUpdate(params *DcimPowerFeedsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsPartialUpdateOK, error)

DcimPowerFeedsRead(params *DcimPowerFeedsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsReadOK, error)

DcimPowerFeedsTrace(params *DcimPowerFeedsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsTraceOK, error)

DcimPowerFeedsUpdate(params *DcimPowerFeedsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsUpdateOK, error)

DcimPowerOutletTemplatesBulkDelete(params *DcimPowerOutletTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesBulkDeleteNoContent, error)

DcimPowerOutletTemplatesBulkPartialUpdate(params *DcimPowerOutletTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesBulkPartialUpdateOK, error)

DcimPowerOutletTemplatesBulkUpdate(params *DcimPowerOutletTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesBulkUpdateOK, error)

DcimPowerOutletTemplatesCreate(params *DcimPowerOutletTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesCreateCreated, error)

DcimPowerOutletTemplatesDelete(params *DcimPowerOutletTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesDeleteNoContent, error)

DcimPowerOutletTemplatesList(params *DcimPowerOutletTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesListOK, error)

DcimPowerOutletTemplatesPartialUpdate(params *DcimPowerOutletTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesPartialUpdateOK, error)

DcimPowerOutletTemplatesRead(params *DcimPowerOutletTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesReadOK, error)

DcimPowerOutletTemplatesUpdate(params *DcimPowerOutletTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesUpdateOK, error)

DcimPowerOutletsBulkDelete(params *DcimPowerOutletsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsBulkDeleteNoContent, error)

DcimPowerOutletsBulkPartialUpdate(params *DcimPowerOutletsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsBulkPartialUpdateOK, error)

DcimPowerOutletsBulkUpdate(params *DcimPowerOutletsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsBulkUpdateOK, error)

DcimPowerOutletsCreate(params *DcimPowerOutletsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsCreateCreated, error)

DcimPowerOutletsDelete(params *DcimPowerOutletsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsDeleteNoContent, error)

DcimPowerOutletsList(params *DcimPowerOutletsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsListOK, error)

DcimPowerOutletsPartialUpdate(params *DcimPowerOutletsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsPartialUpdateOK, error)

DcimPowerOutletsRead(params *DcimPowerOutletsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsReadOK, error)

DcimPowerOutletsTrace(params *DcimPowerOutletsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsTraceOK, error)

DcimPowerOutletsUpdate(params *DcimPowerOutletsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsUpdateOK, error)

DcimPowerPanelsBulkDelete(params *DcimPowerPanelsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsBulkDeleteNoContent, error)

DcimPowerPanelsBulkPartialUpdate(params *DcimPowerPanelsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsBulkPartialUpdateOK, error)

DcimPowerPanelsBulkUpdate(params *DcimPowerPanelsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsBulkUpdateOK, error)

DcimPowerPanelsCreate(params *DcimPowerPanelsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsCreateCreated, error)

DcimPowerPanelsDelete(params *DcimPowerPanelsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsDeleteNoContent, error)

DcimPowerPanelsList(params *DcimPowerPanelsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsListOK, error)

DcimPowerPanelsPartialUpdate(params *DcimPowerPanelsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsPartialUpdateOK, error)

DcimPowerPanelsRead(params *DcimPowerPanelsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsReadOK, error)

DcimPowerPanelsUpdate(params *DcimPowerPanelsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsUpdateOK, error)

DcimPowerPortTemplatesBulkDelete(params *DcimPowerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesBulkDeleteNoContent, error)

DcimPowerPortTemplatesBulkPartialUpdate(params *DcimPowerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesBulkPartialUpdateOK, error)

DcimPowerPortTemplatesBulkUpdate(params *DcimPowerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesBulkUpdateOK, error)

DcimPowerPortTemplatesCreate(params *DcimPowerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesCreateCreated, error)

DcimPowerPortTemplatesDelete(params *DcimPowerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesDeleteNoContent, error)

DcimPowerPortTemplatesList(params *DcimPowerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesListOK, error)

DcimPowerPortTemplatesPartialUpdate(params *DcimPowerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesPartialUpdateOK, error)

DcimPowerPortTemplatesRead(params *DcimPowerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesReadOK, error)

DcimPowerPortTemplatesUpdate(params *DcimPowerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesUpdateOK, error)

DcimPowerPortsBulkDelete(params *DcimPowerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsBulkDeleteNoContent, error)

DcimPowerPortsBulkPartialUpdate(params *DcimPowerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsBulkPartialUpdateOK, error)

DcimPowerPortsBulkUpdate(params *DcimPowerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsBulkUpdateOK, error)

DcimPowerPortsCreate(params *DcimPowerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsCreateCreated, error)

DcimPowerPortsDelete(params *DcimPowerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsDeleteNoContent, error)

DcimPowerPortsList(params *DcimPowerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsListOK, error)

DcimPowerPortsPartialUpdate(params *DcimPowerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsPartialUpdateOK, error)

DcimPowerPortsRead(params *DcimPowerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsReadOK, error)

DcimPowerPortsTrace(params *DcimPowerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsTraceOK, error)

DcimPowerPortsUpdate(params *DcimPowerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsUpdateOK, error)

DcimRackReservationsBulkDelete(params *DcimRackReservationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsBulkDeleteNoContent, error)

DcimRackReservationsBulkPartialUpdate(params *DcimRackReservationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsBulkPartialUpdateOK, error)

DcimRackReservationsBulkUpdate(params *DcimRackReservationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsBulkUpdateOK, error)

DcimRackReservationsCreate(params *DcimRackReservationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsCreateCreated, error)

DcimRackReservationsDelete(params *DcimRackReservationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsDeleteNoContent, error)

DcimRackReservationsList(params *DcimRackReservationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsListOK, error)

DcimRackReservationsPartialUpdate(params *DcimRackReservationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsPartialUpdateOK, error)

DcimRackReservationsRead(params *DcimRackReservationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsReadOK, error)

DcimRackReservationsUpdate(params *DcimRackReservationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsUpdateOK, error)

DcimRackRolesBulkDelete(params *DcimRackRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesBulkDeleteNoContent, error)

DcimRackRolesBulkPartialUpdate(params *DcimRackRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesBulkPartialUpdateOK, error)

DcimRackRolesBulkUpdate(params *DcimRackRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesBulkUpdateOK, error)

DcimRackRolesCreate(params *DcimRackRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesCreateCreated, error)

DcimRackRolesDelete(params *DcimRackRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesDeleteNoContent, error)

DcimRackRolesList(params *DcimRackRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesListOK, error)

DcimRackRolesPartialUpdate(params *DcimRackRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesPartialUpdateOK, error)

DcimRackRolesRead(params *DcimRackRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesReadOK, error)

DcimRackRolesUpdate(params *DcimRackRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesUpdateOK, error)

DcimRacksBulkDelete(params *DcimRacksBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksBulkDeleteNoContent, error)

DcimRacksBulkPartialUpdate(params *DcimRacksBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksBulkPartialUpdateOK, error)

DcimRacksBulkUpdate(params *DcimRacksBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksBulkUpdateOK, error)

DcimRacksCreate(params *DcimRacksCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksCreateCreated, error)

DcimRacksDelete(params *DcimRacksDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksDeleteNoContent, error)

DcimRacksElevation(params *DcimRacksElevationParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksElevationOK, error)

DcimRacksList(params *DcimRacksListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksListOK, error)

DcimRacksPartialUpdate(params *DcimRacksPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksPartialUpdateOK, error)

DcimRacksRead(params *DcimRacksReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksReadOK, error)

DcimRacksUpdate(params *DcimRacksUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksUpdateOK, error)

DcimRearPortTemplatesBulkDelete(params *DcimRearPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesBulkDeleteNoContent, error)

DcimRearPortTemplatesBulkPartialUpdate(params *DcimRearPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesBulkPartialUpdateOK, error)

DcimRearPortTemplatesBulkUpdate(params *DcimRearPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesBulkUpdateOK, error)

DcimRearPortTemplatesCreate(params *DcimRearPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesCreateCreated, error)

DcimRearPortTemplatesDelete(params *DcimRearPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesDeleteNoContent, error)

DcimRearPortTemplatesList(params *DcimRearPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesListOK, error)

DcimRearPortTemplatesPartialUpdate(params *DcimRearPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesPartialUpdateOK, error)

DcimRearPortTemplatesRead(params *DcimRearPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesReadOK, error)

DcimRearPortTemplatesUpdate(params *DcimRearPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesUpdateOK, error)

DcimRearPortsBulkDelete(params *DcimRearPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsBulkDeleteNoContent, error)

DcimRearPortsBulkPartialUpdate(params *DcimRearPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsBulkPartialUpdateOK, error)

DcimRearPortsBulkUpdate(params *DcimRearPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsBulkUpdateOK, error)

DcimRearPortsCreate(params *DcimRearPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsCreateCreated, error)

DcimRearPortsDelete(params *DcimRearPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsDeleteNoContent, error)

DcimRearPortsList(params *DcimRearPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsListOK, error)

DcimRearPortsPartialUpdate(params *DcimRearPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsPartialUpdateOK, error)

DcimRearPortsPaths(params *DcimRearPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsPathsOK, error)

DcimRearPortsRead(params *DcimRearPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsReadOK, error)

DcimRearPortsUpdate(params *DcimRearPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsUpdateOK, error)

DcimRegionsBulkDelete(params *DcimRegionsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsBulkDeleteNoContent, error)

DcimRegionsBulkPartialUpdate(params *DcimRegionsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsBulkPartialUpdateOK, error)

DcimRegionsBulkUpdate(params *DcimRegionsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsBulkUpdateOK, error)

DcimRegionsCreate(params *DcimRegionsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsCreateCreated, error)

DcimRegionsDelete(params *DcimRegionsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsDeleteNoContent, error)

DcimRegionsList(params *DcimRegionsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsListOK, error)

DcimRegionsPartialUpdate(params *DcimRegionsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsPartialUpdateOK, error)

DcimRegionsRead(params *DcimRegionsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsReadOK, error)

DcimRegionsUpdate(params *DcimRegionsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsUpdateOK, error)

DcimSiteGroupsBulkDelete(params *DcimSiteGroupsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsBulkDeleteNoContent, error)

DcimSiteGroupsBulkPartialUpdate(params *DcimSiteGroupsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsBulkPartialUpdateOK, error)

DcimSiteGroupsBulkUpdate(params *DcimSiteGroupsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsBulkUpdateOK, error)

DcimSiteGroupsCreate(params *DcimSiteGroupsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsCreateCreated, error)

DcimSiteGroupsDelete(params *DcimSiteGroupsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsDeleteNoContent, error)

DcimSiteGroupsList(params *DcimSiteGroupsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsListOK, error)

DcimSiteGroupsPartialUpdate(params *DcimSiteGroupsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsPartialUpdateOK, error)

DcimSiteGroupsRead(params *DcimSiteGroupsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsReadOK, error)

DcimSiteGroupsUpdate(params *DcimSiteGroupsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsUpdateOK, error)

DcimSitesBulkDelete(params *DcimSitesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesBulkDeleteNoContent, error)

DcimSitesBulkPartialUpdate(params *DcimSitesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesBulkPartialUpdateOK, error)

DcimSitesBulkUpdate(params *DcimSitesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesBulkUpdateOK, error)

DcimSitesCreate(params *DcimSitesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesCreateCreated, error)

DcimSitesDelete(params *DcimSitesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesDeleteNoContent, error)

DcimSitesList(params *DcimSitesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesListOK, error)

DcimSitesPartialUpdate(params *DcimSitesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesPartialUpdateOK, error)

DcimSitesRead(params *DcimSitesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesReadOK, error)

DcimSitesUpdate(params *DcimSitesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesUpdateOK, error)

DcimVirtualChassisBulkDelete(params *DcimVirtualChassisBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisBulkDeleteNoContent, error)

DcimVirtualChassisBulkPartialUpdate(params *DcimVirtualChassisBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisBulkPartialUpdateOK, error)

DcimVirtualChassisBulkUpdate(params *DcimVirtualChassisBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisBulkUpdateOK, error)

DcimVirtualChassisCreate(params *DcimVirtualChassisCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisCreateCreated, error)

DcimVirtualChassisDelete(params *DcimVirtualChassisDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisDeleteNoContent, error)

DcimVirtualChassisList(params *DcimVirtualChassisListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisListOK, error)

DcimVirtualChassisPartialUpdate(params *DcimVirtualChassisPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisPartialUpdateOK, error)

DcimVirtualChassisRead(params *DcimVirtualChassisReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisReadOK, error)

DcimVirtualChassisUpdate(params *DcimVirtualChassisUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisUpdateOK, error)

SetTransport(transport runtime.ClientTransport)

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
		n.logger.V(1).Error(err, "error parsing IP in start range")
	}

	endIp, _, err := net.ParseCIDR(endIpRange)
	if err != nil {
		n.logger.V(1).Error(err, "error parsing IP in end range")
	}

	trial := net.ParseIP(ip)
	if trial.To4() == nil {

		n.logger.V(0).Error(err, "error parsing IP to IP4 address")

		return false
	}

	if bytes.Compare(trial, startIp) >= 0 && bytes.Compare(trial, endIp) <= 0 {
		return true
	}

	return false
}

func (n *Netbox) ReadDevicesFromNetbox(ctx context.Context, client *client.NetBoxAPI, deviceReq *dcim.DcimDevicesListParams) error {
	opt := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}
	deviceRes, err := client.Dcim.DcimDevicesList(deviceReq, nil, opt)
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

	n.logger.V(1).Info("step 1 - Reading devices successul", "num_machines", len(n.Records))

	return nil
}

func (n *Netbox) ReadInterfacesFromNetbox(ctx context.Context, client *client.NetBoxAPI) error {
	//Get the Interfaces list from netbox to populate the Machine mac value
	interfacesReq := dcim.NewDcimInterfacesListParams()

	for _, record := range n.Records {
		interfacesReq.Device = &record.Hostname
		interfacesRes, err := client.Dcim.DcimInterfacesList(interfacesReq, nil)

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

	n.logger.V(1).Info("step 2 - Reading intefaces successful, MAC addresses set")

	return nil
}

func (n *Netbox) ReadIpRangeFromNetbox(ctx context.Context, client *client.NetBoxAPI, ipamReq *ipam.IpamIPRangesListParams) error {
	ipamRes, err := client.Ipam.IpamIPRangesList(ipamReq, nil)

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

	n.logger.V(1).Info("step 3 - Reading IPAM data successful, all DCIM calls are complete")

	return nil
}

func (n *Netbox) SerializeMachines(machines []*Machine) ([]byte, error) {
	ret, err := json.MarshalIndent(machines, "", " ")
	if err != nil {
		return nil, fmt.Errorf("error in encoding Machines to byte Array: %v", err)
	}
	return ret, nil
}
