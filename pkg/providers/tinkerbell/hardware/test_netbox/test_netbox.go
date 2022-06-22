package main

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/netbox"
)

// var (
// 	err error
// )

func main() {
	n := new(netbox.Netbox)
	n1 := new(netbox.Netbox)
	err := n.ReadFromNetbox()
	err1 := n1.ReadFromNetboxFiltered("eks-a")
	ret, err2 := n1.SerializeMachines(n1.records)
	fmt.Println(err1)
	fmt.Println(err)
}
