package api

import (
	"fmt"
	"strings"

	"libvirt.org/go/libvirt"
)

func CheckPoolExists(conn *libvirt.Connect, poolName string) bool {
	if _, err := conn.LookupStoragePoolByName(poolName); err != nil {
		return false
	}
	return true
}

func GetPoolPath(conn *libvirt.Connect, poolName string) (string, error) {
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return "", err
	}
	defer pool.Free()

	xmlDesc, err := pool.GetXMLDesc(0)
	if err != nil {
		return "", err
	}

	// Parse the XML to get the path
	// This is a simple string search, you might want to use proper XML parsing for more complex scenarios
	pathStart := strings.Index(xmlDesc, "<path>")
	pathEnd := strings.Index(xmlDesc, "</path>")
	if pathStart == -1 || pathEnd == -1 {
		return "", fmt.Errorf("path not found in pool XML")
	}

	path := xmlDesc[pathStart+6 : pathEnd]
	return path, nil
}

func ListPoolVolumes(conn *libvirt.Connect, poolName string) ([]string, error) {
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup pool: %v", err)
	}
	defer pool.Free()

	// Refresh the pool to get the latest state
	if err := pool.Refresh(0); err != nil {
		return nil, fmt.Errorf("failed to refresh pool: %v", err)
	}

	volumes, err := pool.ListAllStorageVolumes(0)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %v", err)
	}

	var volumePaths []string
	for _, vol := range volumes {
		path, err := vol.GetPath()
		if err != nil {
			vol.Free()
			return nil, fmt.Errorf("failed to get volume path: %v", err)
		}
		volumePaths = append(volumePaths, path)
		vol.Free()
	}

	return volumePaths, nil
}
