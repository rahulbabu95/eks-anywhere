package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
)

func TestReadMachineBytes(t *testing.T) {
	n := new(Netbox)
	machines := []*Machine{
		{Hostname: "Dev1", IPAddress: "10.80.8.21", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:F4:C1", Disk: "/dev/sda", Labels: map[string]string{"type": "worker-plane"}, BMCIPAddress: "10.80.12.20", BMCUsername: "root", BMCPassword: "pPyU6mAO"},
		{Hostname: "Dev2", IPAddress: "10.80.8.22", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:EA:11", Disk: "/dev/sda", Labels: map[string]string{"type": "control-plane"}, BMCIPAddress: "10.80.12.21", BMCUsername: "root", BMCPassword: "pPyU6mAO"},
	}
	n.logger = logr.Discard()
	machinesRawString := createMachineString(machines)

	// check happy flow by serializing machines
	machinesUncorruptBytes := []byte(machinesRawString)
	machinesRead, _ := ReadMachinesBytes(context.TODO(), machinesUncorruptBytes, n)

	if diff := cmp.Diff(machines, machinesRead); diff != "" {
		t.Fatal(diff)
	}

	// check unhappy flow by corrupting bytes i.e. swap first and last byte
	machinesCorruptBytes := machinesUncorruptBytes
	machinesCorruptBytes[0], machinesCorruptBytes[len(machinesCorruptBytes)-1] = machinesCorruptBytes[len(machinesCorruptBytes)-1], machinesCorruptBytes[0]
	_, err := ReadMachinesBytes(context.TODO(), machinesCorruptBytes, n)
	if err == nil {
		t.Fatal()
	}
}

func TestWriteToCSV(t *testing.T) {
	var machines = []*Machine{{Hostname: "eksa-dev01", IPAddress: "10.80.8.21", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:F4:C1", Disk: "/dev/sda", Labels: map[string]string{"type": "control-plane"}, BMCIPAddress: "10.80.12.20", BMCUsername: "root", BMCPassword: "root"},
		{Hostname: "eksa-dev02", IPAddress: "10.80.8.22", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:EA:11", Disk: "/dev/sda", Labels: map[string]string{"type": "worker-plane"}, BMCIPAddress: "10.80.12.21", BMCUsername: "root", BMCPassword: "root"},
	}
	exp_file, err := os.Open("testdata/results.csv")
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(exp_file)
	exp_records, _ := reader.ReadAll()
	// errChan := make(chan error)
	n := new(Netbox)
	n.logger = logr.Discard()
	WriteToCSV(context.TODO(), machines, n)
	act_file, err := os.Open("hardware.csv")
	if err != nil {
		t.Fatal(err)
	}
	reader2 := csv.NewReader(act_file)
	act_records, _ := reader2.ReadAll()
	for i := range act_records {
		for j := range act_records[i] {
			if diff := cmp.Diff(act_records[i][j], exp_records[i][j]); diff != "" {
				t.Fatal("Field: ", act_records[0][j], diff)
			}
		}
	}
}

func createMachineString(machines []*Machine) string {
	var rawMachineString = `[`

	for idx, m := range machines {
		t := fmt.Sprintf(`
 {
  "Hostname": %q,
  "IPAddress": %q,
  "Netmask": %q,
  "Gateway": %q,
  "Nameservers": [
   %q
  ],
  "MACAddress": %q,
  "Disk": %q,
  "Labels": {
   "type": %q
  },
  "BMCIPAddress": %q,
  "BMCUsername": %q,
  "BMCPassword": %q
 }`, m.Hostname, m.IPAddress, m.Netmask, m.Gateway, strings.Join(m.Nameservers, ","), m.MACAddress, m.Disk, m.Labels["type"], m.BMCIPAddress, m.BMCUsername, m.BMCPassword)

		rawMachineString += t

		if idx != len(machines)-1 {
			rawMachineString += `,`
		}
	}
	rawMachineString += `
]`
	return rawMachineString
}

func TestExtractNameServers(t *testing.T) {
	type nsTest struct {
		ns   []string
		want string
	}

	nsTests := []nsTest{
		{[]string{"121.63.48.96", "121.63.58.96"}, "121.63.48.96|121.63.58.96"},
		{[]string{"121.63.48.96", "121.63.58.96", "121.63.68.96"}, "121.63.48.96|121.63.58.96|121.63.68.96"},
		{[]string{"", "121.63.58.96", "121.63.68.96"}, "|121.63.58.96|121.63.68.96"},
	}

	for _, test := range nsTests {
		got := extractNameServers(test.ns)
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Fatal(diff)
		}
	}
}
