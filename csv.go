package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

func ReadMachinesBytes(machines []byte) ([]*Machine, error) {
	var hardwareMachines []*Machine
	err := json.Unmarshal(machines, &hardwareMachines)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling the input byte stream: %v", err)
	}
	return hardwareMachines, nil
}

func WriteToCsv(machines []*Machine) (*os.File, error) {

	//Create a csv file usign Os operations
	file, err := os.Create("csv")
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	headers := [11]string{"hostname", "bmc_ip", "bmc_username", "bmc_password", "mac", "ip_address", "netmask", "gateway", "nameservers", "labels", "disk"}
	err = writer.Write(headers[:])
	if err != nil {
		return nil, fmt.Errorf("error Writing Column names into file: %v", err)
	}
	var machinesString [][]string
	for _, machine := range machines {
		nsCombined := extractNameServers(machine.Nameservers)
		row := []string{machine.Hostname, machine.BMCIPAddress, machine.BMCUsername, machine.BMCPassword, machine.MACAddress, machine.Netmask, machine.Gateway, nsCombined, "type=" + machine.Labels["type"], machine.Disk}
		machinesString = append(machinesString, row)
	}
	writer.WriteAll(machinesString)
	return file, nil
}

func extractNameServers(nameservers []string) string {
	nsCombined := ""
	for idx, ns := range nameservers {
		if idx == 0 {
			nsCombined += ns
		} else {
			nsCombined = nsCombined + "|" + ns
		}
	}
	return nsCombined
}
