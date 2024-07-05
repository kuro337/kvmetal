package api

import (
	"fmt"
	"strings"

	"libvirt.org/go/libvirt"
)

type PoolInfo struct {
	Name string
	Path string
}

func ListAllStoragePools(conn *libvirt.Connect) ([]PoolInfo, error) {
	pools, err := conn.ListAllStoragePools(0)
	if err != nil {
		return nil, fmt.Errorf("failed to list storage pools: %v", err)
	}

	var poolInfos []PoolInfo
	for _, pool := range pools {
		defer pool.Free()

		name, err := pool.GetName()
		if err != nil {
			return nil, fmt.Errorf("failed to get pool name: %v", err)
		}

		xmlDesc, err := pool.GetXMLDesc(0)
		if err != nil {
			return nil, fmt.Errorf("failed to get pool XML description: %v", err)
		}

		// Parse the XML to get the path
		// Note: This is a simple string search. For more robust XML parsing,
		// consider using encoding/xml package
		pathStart := strings.Index(xmlDesc, "<path>")
		pathEnd := strings.Index(xmlDesc, "</path>")
		var path string
		if pathStart != -1 && pathEnd != -1 {
			path = xmlDesc[pathStart+6 : pathEnd]
		} else {
			path = "Path not found in XML"
		}

		poolInfos = append(poolInfos, PoolInfo{
			Name: name,
			Path: path,
		})
	}

	return poolInfos, nil
}
