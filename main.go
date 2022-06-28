package main

import (
	"fmt"
)

func main() {
	n := new(Netbox)
	n1 := new(Netbox)
	_ = n.ReadFromNetbox()
	_ = n1.ReadFromNetboxFiltered("eks-a")
	ret, err2 := n1.SerializeMachines(n1.Records)
	machines, err := ReadMachinesBytes(ret)
	fmt.Println(err)
	fmt.Println(ret, err2)
	// for _, machine := range machines {
	// 	fmt.Println(*machine)
	// }
	file, err := WriteToCsv(machines)
	fmt.Println("error witing to csv: ", err)
	fmt.Println(file)

	err = n.ReadFromNetbox()
	if err != nil {
		fmt.Printf("n.ReadFromNetbox() = %v\n", err)
	}
	err = n1.ReadFromNetboxFiltered("eks-a")
	if err != nil {
		fmt.Printf("n1.ReadFromNetboxFiltered() = %v\n", err)
	}

}
