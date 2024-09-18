package utils

import (
	"fmt"
	"log"
	"strings"
)

// RunningVMs Parses the names from space seperated output lines - and gets the Middle Element
func RunningVMs(output string) ([]string, error) {
	var names []string

	lines := strings.Split(output, "\n") // Split into lines

	for _, line := range lines[2:] { // iterate over lines, skip 1st 2 (headers)
		if line == "" {
			continue // Skip empty lines
		}

		fields := strings.Fields(line) // Split line into fields by whitespace

		// Check if the fields slice has at least 2 elements (id and name)
		if len(fields) < 2 {
			return nil, fmt.Errorf("unexpected format: %s", line)
		}

		// Second field (index 1) is the name
		names = append(names, fields[1])
	}

	return names, nil
}

/*
ExtractField extracts a specific field from each line of a structured, space-separated output.

skipLines specifies how many lines to skip at the beginning (e.g., for headers).

fieldIndex is the index of the field to extract, starting from 0.

Usage:

	stdout := `Id   Name          State

			-----------------------------

	 	1    kubecontrol   running
	 	2    kubeworker    running`

			names, _ := ExtractField(output, 2, 1)

			fmt.Println("Fields:", names) // Fields:[kubecontrol kubeworker]
*/
func ExtractField(output string, skipLines int, fieldIndex int) ([]string, error) {
	var fieldsExtracted []string

	// Split the output into lines.
	lines := strings.Split(output, "\n")

	// Iterate over lines, skipping the specified number of lines at the start.
	for _, line := range lines[skipLines:] {
		if line == "" {
			continue // Skip empty lines.
		}

		// Split line into fields, assuming whitespace separation.
		fields := strings.Fields(line)

		// Check if the fields slice has enough elements.
		if len(fields) <= fieldIndex {
			return nil, fmt.Errorf("unexpected format or missing field in line: %s", line)
		}

		// Extract the specified field.
		fieldsExtracted = append(fieldsExtracted, fields[fieldIndex])
	}

	return fieldsExtracted, nil
}

type VM struct {
	Name  string
	State string
}

// var allVMs []VM // Global State for Virtual Machines

// ListVMs parses the output string and returns a slice of VM structs.
// If print is true, it also prints the VMs in a formatted table.
func ListVMs(skipLines int, print bool) ([]VM, error) {
	var vms []VM
	output, _ := ExecCmd("virsh list --all", false)
	lines := strings.Split(output, "\n")
	for _, line := range lines[skipLines:] {
		if line == "" {
			continue // Skip empty lines.
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("unexpected format or missing field in line: %s", line)
		}

		vm := VM{
			Name:  fields[1],
			State: fields[2],
		}

		vms = append(vms, vm)
	}

	if print {
		if len(vms) == 0 {
			log.Printf("No Virtual Machines Running")
		} else {
			PrintVMs(vms)
		}
	}

	return vms, nil
}

// PrintVMs prints the list of VMs in a formatted table.
func PrintVMs(vms []VM) {
	fmt.Println("Id   Name          State")
	fmt.Println("------------------------------")
	for i, vm := range vms {
		fmt.Printf("%2d    %-12s %s\n", i+1, vm.Name, vm.State)
	}
}
