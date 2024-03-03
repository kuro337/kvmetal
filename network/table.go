package network

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

// Helper function to convert PortMapping slices to a string representation.
func portMapToString(portMap []PortMapping) string {
	var strSlice []string
	for _, pm := range portMap {
		strSlice = append(strSlice, fmt.Sprintf("%s %d->%d", pm.Protocol, pm.VMPort, pm.HostPort))
	}
	return strings.Join(strSlice, ", ")
}

// CreateTableFromConfig generates a concise table from a ForwardingConfig and returns it as a string.
func CreateTableFromConfig(config ForwardingConfig) string {
	var stringBuilder strings.Builder
	t := table.NewWriter()
	t.SetOutputMirror(&stringBuilder)
	t.SetStyle(table.StyleLight)

	// Define the header for the table.
	t.AppendHeader(table.Row{"VM Name", "Port Mapping", "Host IP", "External IP", "Interface"})

	// Prepare the data for the table.
	portMapStr := portMapToString(config.PortMap)
	//	portRangeStr := portRangeToString(config.PortRange)
	hostIP := config.HostIP.String()
	externalIP := config.ExternalIP.String()

	// Append the configuration data as a row.
	t.AppendRow(table.Row{
		config.VMName, portMapStr, hostIP, externalIP, config.Interface,
	})

	t.Render()

	return stringBuilder.String()
}

/*
// Helper function to convert PortRange slices to a string representation.
func portRangeToString(portRange []PortRange) string {
	var strSlice []string
	for _, pr := range portRange {
		strSlice = append(strSlice, fmt.Sprintf("%s %d-%d->%d-%d", pr.Protocol, pr.VMStartPort, pr.VMEndPortNum, pr.HostStartPortNum, pr.HostEndPortNum))
	}
	return strings.Join(strSlice, ", ")
}
*/
