package main

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/netbox"
)

func main() {
	n := new(netbox.Netbox)
	n1 := new(netbox.Netbox)
	_ = n.ReadFromNetbox()
	_ = n1.ReadFromNetboxFiltered("eks-a")
	ret, err2 := n1.SerializeMachines(n1.Records)
	machines, err := netbox.ReadMachinesBytes(ret)
	fmt.Println(err)
	fmt.Println(ret, err2)
	// for _, machine := range machines {
	// 	fmt.Println(*machine)
	// }
	file, err := netbox.WriteToCsv(machines)
	fmt.Println("error witing to csv: ", err)
	fmt.Println(file)

	err := n.ReadFromNetbox()
	err1 := n1.ReadFromNetboxFiltered("eks-a")
	fmt.Println(err1)
	fmt.Println(err)

}
