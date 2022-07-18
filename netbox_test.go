package main

import (
	"context"
	"testing"
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
	}

	n := new(Netbox)
	for _, test := range checkIpTests {
		if output := n.CheckIp(test.ctx, test.toCheck, test.ipStart, test.ipEnd); output != test.want {
			t.Errorf("output %v not equal to expected %v", test.toCheck, test.want)
		}
	}
}

func TestReadDevicesFromNetbox(t *testing.T) {

	type DeviceClientMock struct{}

	// func (m *DeviceClientMock) DcimCablesBulkDelete(params *DcimCablesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesBulkDeleteNoContent, error) {
	// 	return nil, nil
	// }

	// func (m *DeviceClientMock) DcimCablesBulkPartialUpdate(params *DcimCablesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesBulkPartialUpdateOK, error){
	// 	return nil,nil
	// }

	// func (m *DeviceClientMock)  DcimCablesBulkUpdate(params *DcimCablesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimCablesCreate(params *DcimCablesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesCreateCreated, error)

	// func (m *DeviceClientMock) DcimCablesDelete(params *DcimCablesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimCablesList(params *DcimCablesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesListOK, error)

	// func (m *DeviceClientMock) DcimCablesPartialUpdate(params *DcimCablesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimCablesRead(params *DcimCablesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesReadOK, error)

	// func (m *DeviceClientMock) DcimCablesUpdate(params *DcimCablesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimCablesUpdateOK, error)

	// func (m *DeviceClientMock) DcimConnectedDeviceList(params *DcimConnectedDeviceListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConnectedDeviceListOK, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesBulkDelete(params *DcimConsolePortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesBulkPartialUpdate(params *DcimConsolePortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesBulkUpdate(params *DcimConsolePortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesCreate(params *DcimConsolePortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesCreateCreated, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesDelete(params *DcimConsolePortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesList(params *DcimConsolePortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesListOK, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesPartialUpdate(params *DcimConsolePortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesRead(params *DcimConsolePortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesReadOK, error)

	// func (m *DeviceClientMock) DcimConsolePortTemplatesUpdate(params *DcimConsolePortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortTemplatesUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsolePortsBulkDelete(params *DcimConsolePortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsolePortsBulkPartialUpdate(params *DcimConsolePortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsolePortsBulkUpdate(params *DcimConsolePortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsolePortsCreate(params *DcimConsolePortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsCreateCreated, error)

	// func (m *DeviceClientMock) DcimConsolePortsDelete(params *DcimConsolePortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsolePortsList(params *DcimConsolePortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsListOK, error)

	// func (m *DeviceClientMock) DcimConsolePortsPartialUpdate(params *DcimConsolePortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsolePortsRead(params *DcimConsolePortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsReadOK, error)

	// func (m *DeviceClientMock) DcimConsolePortsTrace(params *DcimConsolePortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsTraceOK, error)

	// func (m *DeviceClientMock) DcimConsolePortsUpdate(params *DcimConsolePortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsolePortsUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesBulkDelete(params *DcimConsoleServerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesBulkPartialUpdate(params *DcimConsoleServerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesBulkUpdate(params *DcimConsoleServerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesCreate(params *DcimConsoleServerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesCreateCreated, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesDelete(params *DcimConsoleServerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesList(params *DcimConsoleServerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesListOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesPartialUpdate(params *DcimConsoleServerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesRead(params *DcimConsoleServerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesReadOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortTemplatesUpdate(params *DcimConsoleServerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortTemplatesUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsBulkDelete(params *DcimConsoleServerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsBulkPartialUpdate(params *DcimConsoleServerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsBulkUpdate(params *DcimConsoleServerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsCreate(params *DcimConsoleServerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsCreateCreated, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsDelete(params *DcimConsoleServerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsList(params *DcimConsoleServerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsListOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsPartialUpdate(params *DcimConsoleServerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsRead(params *DcimConsoleServerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsReadOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsTrace(params *DcimConsoleServerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsTraceOK, error)

	// func (m *DeviceClientMock) DcimConsoleServerPortsUpdate(params *DcimConsoleServerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimConsoleServerPortsUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesBulkDelete(params *DcimDeviceBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesBulkPartialUpdate(params *DcimDeviceBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesBulkUpdate(params *DcimDeviceBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesCreate(params *DcimDeviceBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesCreateCreated, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesDelete(params *DcimDeviceBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesList(params *DcimDeviceBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesListOK, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesPartialUpdate(params *DcimDeviceBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesRead(params *DcimDeviceBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesReadOK, error)

	// func (m *DeviceClientMock) DcimDeviceBayTemplatesUpdate(params *DcimDeviceBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBayTemplatesUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBaysBulkDelete(params *DcimDeviceBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceBaysBulkPartialUpdate(params *DcimDeviceBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBaysBulkUpdate(params *DcimDeviceBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBaysCreate(params *DcimDeviceBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysCreateCreated, error)

	// func (m *DeviceClientMock) DcimDeviceBaysDelete(params *DcimDeviceBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceBaysList(params *DcimDeviceBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysListOK, error)

	// func (m *DeviceClientMock) DcimDeviceBaysPartialUpdate(params *DcimDeviceBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceBaysRead(params *DcimDeviceBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysReadOK, error)

	// func (m *DeviceClientMock) DcimDeviceBaysUpdate(params *DcimDeviceBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceBaysUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceRolesBulkDelete(params *DcimDeviceRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceRolesBulkPartialUpdate(params *DcimDeviceRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceRolesBulkUpdate(params *DcimDeviceRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceRolesCreate(params *DcimDeviceRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesCreateCreated, error)

	// func (m *DeviceClientMock) DcimDeviceRolesDelete(params *DcimDeviceRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceRolesList(params *DcimDeviceRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesListOK, error)

	// func (m *DeviceClientMock) DcimDeviceRolesPartialUpdate(params *DcimDeviceRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceRolesRead(params *DcimDeviceRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesReadOK, error)

	// func (m *DeviceClientMock) DcimDeviceRolesUpdate(params *DcimDeviceRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceRolesUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceTypesBulkDelete(params *DcimDeviceTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceTypesBulkPartialUpdate(params *DcimDeviceTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceTypesBulkUpdate(params *DcimDeviceTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceTypesCreate(params *DcimDeviceTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesCreateCreated, error)

	// func (m *DeviceClientMock) DcimDeviceTypesDelete(params *DcimDeviceTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDeviceTypesList(params *DcimDeviceTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesListOK, error)

	// func (m *DeviceClientMock) DcimDeviceTypesPartialUpdate(params *DcimDeviceTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDeviceTypesRead(params *DcimDeviceTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesReadOK, error)

	// func (m *DeviceClientMock) DcimDeviceTypesUpdate(params *DcimDeviceTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDeviceTypesUpdateOK, error)

	// func (m *DeviceClientMock) DcimDevicesBulkDelete(params *DcimDevicesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDevicesBulkPartialUpdate(params *DcimDevicesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDevicesBulkUpdate(params *DcimDevicesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimDevicesCreate(params *DcimDevicesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesCreateCreated, error)

	// func (m *DeviceClientMock) DcimDevicesDelete(params *DcimDevicesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimDevicesList(params *DcimDevicesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesListOK, error)

	// func (m *DeviceClientMock) DcimDevicesNapalm(params *DcimDevicesNapalmParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesNapalmOK, error)

	// func (m *DeviceClientMock) DcimDevicesPartialUpdate(params *DcimDevicesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimDevicesRead(params *DcimDevicesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesReadOK, error)

	// func (m *DeviceClientMock) DcimDevicesUpdate(params *DcimDevicesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimDevicesUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesBulkDelete(params *DcimFrontPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesBulkPartialUpdate(params *DcimFrontPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesBulkUpdate(params *DcimFrontPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesCreate(params *DcimFrontPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesCreateCreated, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesDelete(params *DcimFrontPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesList(params *DcimFrontPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesListOK, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesPartialUpdate(params *DcimFrontPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesRead(params *DcimFrontPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesReadOK, error)

	// func (m *DeviceClientMock) DcimFrontPortTemplatesUpdate(params *DcimFrontPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortTemplatesUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortsBulkDelete(params *DcimFrontPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimFrontPortsBulkPartialUpdate(params *DcimFrontPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortsBulkUpdate(params *DcimFrontPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortsCreate(params *DcimFrontPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsCreateCreated, error)

	// func (m *DeviceClientMock) DcimFrontPortsDelete(params *DcimFrontPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimFrontPortsList(params *DcimFrontPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsListOK, error)

	// func (m *DeviceClientMock) DcimFrontPortsPartialUpdate(params *DcimFrontPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimFrontPortsPaths(params *DcimFrontPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsPathsOK, error)

	// func (m *DeviceClientMock) DcimFrontPortsRead(params *DcimFrontPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsReadOK, error)

	// func (m *DeviceClientMock) DcimFrontPortsUpdate(params *DcimFrontPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimFrontPortsUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesBulkDelete(params *DcimInterfaceTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesBulkPartialUpdate(params *DcimInterfaceTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesBulkUpdate(params *DcimInterfaceTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesCreate(params *DcimInterfaceTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesCreateCreated, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesDelete(params *DcimInterfaceTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesList(params *DcimInterfaceTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesListOK, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesPartialUpdate(params *DcimInterfaceTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesRead(params *DcimInterfaceTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesReadOK, error)

	// func (m *DeviceClientMock) DcimInterfaceTemplatesUpdate(params *DcimInterfaceTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfaceTemplatesUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfacesBulkDelete(params *DcimInterfacesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInterfacesBulkPartialUpdate(params *DcimInterfacesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfacesBulkUpdate(params *DcimInterfacesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfacesCreate(params *DcimInterfacesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesCreateCreated, error)

	// func (m *DeviceClientMock) DcimInterfacesDelete(params *DcimInterfacesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInterfacesList(params *DcimInterfacesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesListOK, error)

	// func (m *DeviceClientMock) DcimInterfacesPartialUpdate(params *DcimInterfacesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInterfacesRead(params *DcimInterfacesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesReadOK, error)

	// func (m *DeviceClientMock) DcimInterfacesTrace(params *DcimInterfacesTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesTraceOK, error)

	// func (m *DeviceClientMock) DcimInterfacesUpdate(params *DcimInterfacesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInterfacesUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesBulkDelete(params *DcimInventoryItemRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesBulkPartialUpdate(params *DcimInventoryItemRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesBulkUpdate(params *DcimInventoryItemRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesCreate(params *DcimInventoryItemRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesCreateCreated, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesDelete(params *DcimInventoryItemRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesList(params *DcimInventoryItemRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesListOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesPartialUpdate(params *DcimInventoryItemRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesRead(params *DcimInventoryItemRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesReadOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemRolesUpdate(params *DcimInventoryItemRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemRolesUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesBulkDelete(params *DcimInventoryItemTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesBulkPartialUpdate(params *DcimInventoryItemTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesBulkUpdate(params *DcimInventoryItemTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesCreate(params *DcimInventoryItemTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesCreateCreated, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesDelete(params *DcimInventoryItemTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesList(params *DcimInventoryItemTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesListOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesPartialUpdate(params *DcimInventoryItemTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesRead(params *DcimInventoryItemTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesReadOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemTemplatesUpdate(params *DcimInventoryItemTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemTemplatesUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemsBulkDelete(params *DcimInventoryItemsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInventoryItemsBulkPartialUpdate(params *DcimInventoryItemsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemsBulkUpdate(params *DcimInventoryItemsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemsCreate(params *DcimInventoryItemsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsCreateCreated, error)

	// func (m *DeviceClientMock) DcimInventoryItemsDelete(params *DcimInventoryItemsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimInventoryItemsList(params *DcimInventoryItemsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsListOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemsPartialUpdate(params *DcimInventoryItemsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemsRead(params *DcimInventoryItemsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsReadOK, error)

	// func (m *DeviceClientMock) DcimInventoryItemsUpdate(params *DcimInventoryItemsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimInventoryItemsUpdateOK, error)

	// func (m *DeviceClientMock) DcimLocationsBulkDelete(params *DcimLocationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimLocationsBulkPartialUpdate(params *DcimLocationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimLocationsBulkUpdate(params *DcimLocationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimLocationsCreate(params *DcimLocationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsCreateCreated, error)

	// func (m *DeviceClientMock) DcimLocationsDelete(params *DcimLocationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimLocationsList(params *DcimLocationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsListOK, error)

	// func (m *DeviceClientMock) DcimLocationsPartialUpdate(params *DcimLocationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimLocationsRead(params *DcimLocationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsReadOK, error)

	// func (m *DeviceClientMock) DcimLocationsUpdate(params *DcimLocationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimLocationsUpdateOK, error)

	// func (m *DeviceClientMock) DcimManufacturersBulkDelete(params *DcimManufacturersBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimManufacturersBulkPartialUpdate(params *DcimManufacturersBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimManufacturersBulkUpdate(params *DcimManufacturersBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimManufacturersCreate(params *DcimManufacturersCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersCreateCreated, error)

	// func (m *DeviceClientMock) DcimManufacturersDelete(params *DcimManufacturersDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimManufacturersList(params *DcimManufacturersListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersListOK, error)

	// func (m *DeviceClientMock) DcimManufacturersPartialUpdate(params *DcimManufacturersPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimManufacturersRead(params *DcimManufacturersReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersReadOK, error)

	// func (m *DeviceClientMock) DcimManufacturersUpdate(params *DcimManufacturersUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimManufacturersUpdateOK, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesBulkDelete(params *DcimModuleBayTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesBulkPartialUpdate(params *DcimModuleBayTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesBulkUpdate(params *DcimModuleBayTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesCreate(params *DcimModuleBayTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesCreateCreated, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesDelete(params *DcimModuleBayTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesList(params *DcimModuleBayTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesListOK, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesPartialUpdate(params *DcimModuleBayTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesRead(params *DcimModuleBayTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesReadOK, error)

	// func (m *DeviceClientMock) DcimModuleBayTemplatesUpdate(params *DcimModuleBayTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBayTemplatesUpdateOK, error)

	// func (m *DeviceClientMock) DcimModuleBaysBulkDelete(params *DcimModuleBaysBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysBulkDeleteNoContent, error)

	// func (m *DeviceClientMock) DcimModuleBaysBulkPartialUpdate(params *DcimModuleBaysBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysBulkPartialUpdateOK, error)

	// func (m *DeviceClientMock) DcimModuleBaysBulkUpdate(params *DcimModuleBaysBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysBulkUpdateOK, error)

	// func (m *DeviceClientMock) DcimModuleBaysCreate(params *DcimModuleBaysCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysCreateCreated, error)

	// func (m *DeviceClientMock) DcimModuleBaysDelete(params *DcimModuleBaysDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysDeleteNoContent, error)

	// DcimModuleBaysList(params *DcimModuleBaysListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysListOK, error)

	// DcimModuleBaysPartialUpdate(params *DcimModuleBaysPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysPartialUpdateOK, error)

	// DcimModuleBaysRead(params *DcimModuleBaysReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysReadOK, error)

	// DcimModuleBaysUpdate(params *DcimModuleBaysUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleBaysUpdateOK, error)

	// DcimModuleTypesBulkDelete(params *DcimModuleTypesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesBulkDeleteNoContent, error)

	// DcimModuleTypesBulkPartialUpdate(params *DcimModuleTypesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesBulkPartialUpdateOK, error)

	// DcimModuleTypesBulkUpdate(params *DcimModuleTypesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesBulkUpdateOK, error)

	// DcimModuleTypesCreate(params *DcimModuleTypesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesCreateCreated, error)

	// DcimModuleTypesDelete(params *DcimModuleTypesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesDeleteNoContent, error)

	// DcimModuleTypesList(params *DcimModuleTypesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesListOK, error)

	// DcimModuleTypesPartialUpdate(params *DcimModuleTypesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesPartialUpdateOK, error)

	// DcimModuleTypesRead(params *DcimModuleTypesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesReadOK, error)

	// DcimModuleTypesUpdate(params *DcimModuleTypesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModuleTypesUpdateOK, error)

	// DcimModulesBulkDelete(params *DcimModulesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesBulkDeleteNoContent, error)

	// DcimModulesBulkPartialUpdate(params *DcimModulesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesBulkPartialUpdateOK, error)

	// DcimModulesBulkUpdate(params *DcimModulesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesBulkUpdateOK, error)

	// DcimModulesCreate(params *DcimModulesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesCreateCreated, error)

	// DcimModulesDelete(params *DcimModulesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesDeleteNoContent, error)

	// DcimModulesList(params *DcimModulesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesListOK, error)

	// DcimModulesPartialUpdate(params *DcimModulesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesPartialUpdateOK, error)

	// DcimModulesRead(params *DcimModulesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesReadOK, error)

	// DcimModulesUpdate(params *DcimModulesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimModulesUpdateOK, error)

	// DcimPlatformsBulkDelete(params *DcimPlatformsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsBulkDeleteNoContent, error)

	// DcimPlatformsBulkPartialUpdate(params *DcimPlatformsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsBulkPartialUpdateOK, error)

	// DcimPlatformsBulkUpdate(params *DcimPlatformsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsBulkUpdateOK, error)

	// DcimPlatformsCreate(params *DcimPlatformsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsCreateCreated, error)

	// DcimPlatformsDelete(params *DcimPlatformsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsDeleteNoContent, error)

	// DcimPlatformsList(params *DcimPlatformsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsListOK, error)

	// DcimPlatformsPartialUpdate(params *DcimPlatformsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsPartialUpdateOK, error)

	// DcimPlatformsRead(params *DcimPlatformsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsReadOK, error)

	// DcimPlatformsUpdate(params *DcimPlatformsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPlatformsUpdateOK, error)

	// DcimPowerFeedsBulkDelete(params *DcimPowerFeedsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsBulkDeleteNoContent, error)

	// DcimPowerFeedsBulkPartialUpdate(params *DcimPowerFeedsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsBulkPartialUpdateOK, error)

	// DcimPowerFeedsBulkUpdate(params *DcimPowerFeedsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsBulkUpdateOK, error)

	// DcimPowerFeedsCreate(params *DcimPowerFeedsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsCreateCreated, error)

	// DcimPowerFeedsDelete(params *DcimPowerFeedsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsDeleteNoContent, error)

	// DcimPowerFeedsList(params *DcimPowerFeedsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsListOK, error)

	// DcimPowerFeedsPartialUpdate(params *DcimPowerFeedsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsPartialUpdateOK, error)

	// DcimPowerFeedsRead(params *DcimPowerFeedsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsReadOK, error)

	// DcimPowerFeedsTrace(params *DcimPowerFeedsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsTraceOK, error)

	// DcimPowerFeedsUpdate(params *DcimPowerFeedsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerFeedsUpdateOK, error)

	// DcimPowerOutletTemplatesBulkDelete(params *DcimPowerOutletTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesBulkDeleteNoContent, error)

	// DcimPowerOutletTemplatesBulkPartialUpdate(params *DcimPowerOutletTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesBulkPartialUpdateOK, error)

	// DcimPowerOutletTemplatesBulkUpdate(params *DcimPowerOutletTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesBulkUpdateOK, error)

	// DcimPowerOutletTemplatesCreate(params *DcimPowerOutletTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesCreateCreated, error)

	// DcimPowerOutletTemplatesDelete(params *DcimPowerOutletTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesDeleteNoContent, error)

	// DcimPowerOutletTemplatesList(params *DcimPowerOutletTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesListOK, error)

	// DcimPowerOutletTemplatesPartialUpdate(params *DcimPowerOutletTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesPartialUpdateOK, error)

	// DcimPowerOutletTemplatesRead(params *DcimPowerOutletTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesReadOK, error)

	// DcimPowerOutletTemplatesUpdate(params *DcimPowerOutletTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletTemplatesUpdateOK, error)

	// DcimPowerOutletsBulkDelete(params *DcimPowerOutletsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsBulkDeleteNoContent, error)

	// DcimPowerOutletsBulkPartialUpdate(params *DcimPowerOutletsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsBulkPartialUpdateOK, error)

	// DcimPowerOutletsBulkUpdate(params *DcimPowerOutletsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsBulkUpdateOK, error)

	// DcimPowerOutletsCreate(params *DcimPowerOutletsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsCreateCreated, error)

	// DcimPowerOutletsDelete(params *DcimPowerOutletsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsDeleteNoContent, error)

	// DcimPowerOutletsList(params *DcimPowerOutletsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsListOK, error)

	// DcimPowerOutletsPartialUpdate(params *DcimPowerOutletsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsPartialUpdateOK, error)

	// DcimPowerOutletsRead(params *DcimPowerOutletsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsReadOK, error)

	// DcimPowerOutletsTrace(params *DcimPowerOutletsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsTraceOK, error)

	// DcimPowerOutletsUpdate(params *DcimPowerOutletsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerOutletsUpdateOK, error)

	// DcimPowerPanelsBulkDelete(params *DcimPowerPanelsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsBulkDeleteNoContent, error)

	// DcimPowerPanelsBulkPartialUpdate(params *DcimPowerPanelsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsBulkPartialUpdateOK, error)

	// DcimPowerPanelsBulkUpdate(params *DcimPowerPanelsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsBulkUpdateOK, error)

	// DcimPowerPanelsCreate(params *DcimPowerPanelsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsCreateCreated, error)

	// DcimPowerPanelsDelete(params *DcimPowerPanelsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsDeleteNoContent, error)

	// DcimPowerPanelsList(params *DcimPowerPanelsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsListOK, error)

	// DcimPowerPanelsPartialUpdate(params *DcimPowerPanelsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsPartialUpdateOK, error)

	// DcimPowerPanelsRead(params *DcimPowerPanelsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsReadOK, error)

	// DcimPowerPanelsUpdate(params *DcimPowerPanelsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPanelsUpdateOK, error)

	// DcimPowerPortTemplatesBulkDelete(params *DcimPowerPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesBulkDeleteNoContent, error)

	// DcimPowerPortTemplatesBulkPartialUpdate(params *DcimPowerPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesBulkPartialUpdateOK, error)

	// DcimPowerPortTemplatesBulkUpdate(params *DcimPowerPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesBulkUpdateOK, error)

	// DcimPowerPortTemplatesCreate(params *DcimPowerPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesCreateCreated, error)

	// DcimPowerPortTemplatesDelete(params *DcimPowerPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesDeleteNoContent, error)

	// DcimPowerPortTemplatesList(params *DcimPowerPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesListOK, error)

	// DcimPowerPortTemplatesPartialUpdate(params *DcimPowerPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesPartialUpdateOK, error)

	// DcimPowerPortTemplatesRead(params *DcimPowerPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesReadOK, error)

	// DcimPowerPortTemplatesUpdate(params *DcimPowerPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortTemplatesUpdateOK, error)

	// DcimPowerPortsBulkDelete(params *DcimPowerPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsBulkDeleteNoContent, error)

	// DcimPowerPortsBulkPartialUpdate(params *DcimPowerPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsBulkPartialUpdateOK, error)

	// DcimPowerPortsBulkUpdate(params *DcimPowerPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsBulkUpdateOK, error)

	// DcimPowerPortsCreate(params *DcimPowerPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsCreateCreated, error)

	// DcimPowerPortsDelete(params *DcimPowerPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsDeleteNoContent, error)

	// DcimPowerPortsList(params *DcimPowerPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsListOK, error)

	// DcimPowerPortsPartialUpdate(params *DcimPowerPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsPartialUpdateOK, error)

	// DcimPowerPortsRead(params *DcimPowerPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsReadOK, error)

	// DcimPowerPortsTrace(params *DcimPowerPortsTraceParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsTraceOK, error)

	// DcimPowerPortsUpdate(params *DcimPowerPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimPowerPortsUpdateOK, error)

	// DcimRackReservationsBulkDelete(params *DcimRackReservationsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsBulkDeleteNoContent, error)

	// DcimRackReservationsBulkPartialUpdate(params *DcimRackReservationsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsBulkPartialUpdateOK, error)

	// DcimRackReservationsBulkUpdate(params *DcimRackReservationsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsBulkUpdateOK, error)

	// DcimRackReservationsCreate(params *DcimRackReservationsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsCreateCreated, error)

	// DcimRackReservationsDelete(params *DcimRackReservationsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsDeleteNoContent, error)

	// DcimRackReservationsList(params *DcimRackReservationsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsListOK, error)

	// DcimRackReservationsPartialUpdate(params *DcimRackReservationsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsPartialUpdateOK, error)

	// DcimRackReservationsRead(params *DcimRackReservationsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsReadOK, error)

	// DcimRackReservationsUpdate(params *DcimRackReservationsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackReservationsUpdateOK, error)

	// DcimRackRolesBulkDelete(params *DcimRackRolesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesBulkDeleteNoContent, error)

	// DcimRackRolesBulkPartialUpdate(params *DcimRackRolesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesBulkPartialUpdateOK, error)

	// DcimRackRolesBulkUpdate(params *DcimRackRolesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesBulkUpdateOK, error)

	// DcimRackRolesCreate(params *DcimRackRolesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesCreateCreated, error)

	// DcimRackRolesDelete(params *DcimRackRolesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesDeleteNoContent, error)

	// DcimRackRolesList(params *DcimRackRolesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesListOK, error)

	// DcimRackRolesPartialUpdate(params *DcimRackRolesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesPartialUpdateOK, error)

	// DcimRackRolesRead(params *DcimRackRolesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesReadOK, error)

	// DcimRackRolesUpdate(params *DcimRackRolesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRackRolesUpdateOK, error)

	// DcimRacksBulkDelete(params *DcimRacksBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksBulkDeleteNoContent, error)

	// DcimRacksBulkPartialUpdate(params *DcimRacksBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksBulkPartialUpdateOK, error)

	// DcimRacksBulkUpdate(params *DcimRacksBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksBulkUpdateOK, error)

	// DcimRacksCreate(params *DcimRacksCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksCreateCreated, error)

	// DcimRacksDelete(params *DcimRacksDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksDeleteNoContent, error)

	// DcimRacksElevation(params *DcimRacksElevationParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksElevationOK, error)

	// DcimRacksList(params *DcimRacksListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksListOK, error)

	// DcimRacksPartialUpdate(params *DcimRacksPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksPartialUpdateOK, error)

	// DcimRacksRead(params *DcimRacksReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksReadOK, error)

	// DcimRacksUpdate(params *DcimRacksUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRacksUpdateOK, error)

	// DcimRearPortTemplatesBulkDelete(params *DcimRearPortTemplatesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesBulkDeleteNoContent, error)

	// DcimRearPortTemplatesBulkPartialUpdate(params *DcimRearPortTemplatesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesBulkPartialUpdateOK, error)

	// DcimRearPortTemplatesBulkUpdate(params *DcimRearPortTemplatesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesBulkUpdateOK, error)

	// DcimRearPortTemplatesCreate(params *DcimRearPortTemplatesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesCreateCreated, error)

	// DcimRearPortTemplatesDelete(params *DcimRearPortTemplatesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesDeleteNoContent, error)

	// DcimRearPortTemplatesList(params *DcimRearPortTemplatesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesListOK, error)

	// DcimRearPortTemplatesPartialUpdate(params *DcimRearPortTemplatesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesPartialUpdateOK, error)

	// DcimRearPortTemplatesRead(params *DcimRearPortTemplatesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesReadOK, error)

	// DcimRearPortTemplatesUpdate(params *DcimRearPortTemplatesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortTemplatesUpdateOK, error)

	// DcimRearPortsBulkDelete(params *DcimRearPortsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsBulkDeleteNoContent, error)

	// DcimRearPortsBulkPartialUpdate(params *DcimRearPortsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsBulkPartialUpdateOK, error)

	// DcimRearPortsBulkUpdate(params *DcimRearPortsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsBulkUpdateOK, error)

	// DcimRearPortsCreate(params *DcimRearPortsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsCreateCreated, error)

	// DcimRearPortsDelete(params *DcimRearPortsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsDeleteNoContent, error)

	// DcimRearPortsList(params *DcimRearPortsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsListOK, error)

	// DcimRearPortsPartialUpdate(params *DcimRearPortsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsPartialUpdateOK, error)

	// DcimRearPortsPaths(params *DcimRearPortsPathsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsPathsOK, error)

	// DcimRearPortsRead(params *DcimRearPortsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsReadOK, error)

	// DcimRearPortsUpdate(params *DcimRearPortsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRearPortsUpdateOK, error)

	// DcimRegionsBulkDelete(params *DcimRegionsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsBulkDeleteNoContent, error)

	// DcimRegionsBulkPartialUpdate(params *DcimRegionsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsBulkPartialUpdateOK, error)

	// DcimRegionsBulkUpdate(params *DcimRegionsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsBulkUpdateOK, error)

	// DcimRegionsCreate(params *DcimRegionsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsCreateCreated, error)

	// DcimRegionsDelete(params *DcimRegionsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsDeleteNoContent, error)

	// DcimRegionsList(params *DcimRegionsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsListOK, error)

	// DcimRegionsPartialUpdate(params *DcimRegionsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsPartialUpdateOK, error)

	// DcimRegionsRead(params *DcimRegionsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsReadOK, error)

	// DcimRegionsUpdate(params *DcimRegionsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimRegionsUpdateOK, error)

	// DcimSiteGroupsBulkDelete(params *DcimSiteGroupsBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsBulkDeleteNoContent, error)

	// DcimSiteGroupsBulkPartialUpdate(params *DcimSiteGroupsBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsBulkPartialUpdateOK, error)

	// DcimSiteGroupsBulkUpdate(params *DcimSiteGroupsBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsBulkUpdateOK, error)

	// DcimSiteGroupsCreate(params *DcimSiteGroupsCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsCreateCreated, error)

	// DcimSiteGroupsDelete(params *DcimSiteGroupsDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsDeleteNoContent, error)

	// DcimSiteGroupsList(params *DcimSiteGroupsListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsListOK, error)

	// DcimSiteGroupsPartialUpdate(params *DcimSiteGroupsPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsPartialUpdateOK, error)

	// DcimSiteGroupsRead(params *DcimSiteGroupsReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsReadOK, error)

	// DcimSiteGroupsUpdate(params *DcimSiteGroupsUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSiteGroupsUpdateOK, error)

	// DcimSitesBulkDelete(params *DcimSitesBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesBulkDeleteNoContent, error)

	// DcimSitesBulkPartialUpdate(params *DcimSitesBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesBulkPartialUpdateOK, error)

	// DcimSitesBulkUpdate(params *DcimSitesBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesBulkUpdateOK, error)

	// DcimSitesCreate(params *DcimSitesCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesCreateCreated, error)

	// DcimSitesDelete(params *DcimSitesDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesDeleteNoContent, error)

	// DcimSitesList(params *DcimSitesListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesListOK, error)

	// DcimSitesPartialUpdate(params *DcimSitesPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesPartialUpdateOK, error)

	// DcimSitesRead(params *DcimSitesReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesReadOK, error)

	// DcimSitesUpdate(params *DcimSitesUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimSitesUpdateOK, error)

	// DcimVirtualChassisBulkDelete(params *DcimVirtualChassisBulkDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisBulkDeleteNoContent, error)

	// DcimVirtualChassisBulkPartialUpdate(params *DcimVirtualChassisBulkPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisBulkPartialUpdateOK, error)

	// DcimVirtualChassisBulkUpdate(params *DcimVirtualChassisBulkUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisBulkUpdateOK, error)

	// DcimVirtualChassisCreate(params *DcimVirtualChassisCreateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisCreateCreated, error)

	// DcimVirtualChassisDelete(params *DcimVirtualChassisDeleteParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisDeleteNoContent, error)

	// DcimVirtualChassisList(params *DcimVirtualChassisListParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisListOK, error)

	// DcimVirtualChassisPartialUpdate(params *DcimVirtualChassisPartialUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisPartialUpdateOK, error)

	// DcimVirtualChassisRead(params *DcimVirtualChassisReadParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisReadOK, error)

	// DcimVirtualChassisUpdate(params *DcimVirtualChassisUpdateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DcimVirtualChassisUpdateOK, error)

	// SetTransport(transport runtime.ClientTransport)

}
