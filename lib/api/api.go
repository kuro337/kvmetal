package api

import (
	"fmt"
	"log"
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
	log.Printf("Checking if Pool %s exists\n", poolName)

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

func DeletePool(conn *libvirt.Connect, poolName string, deleteContents bool) error {
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("failed to lookup pool: %v", err)
	}
	defer pool.Free()

	// Stop the pool
	err = pool.Destroy()
	if err != nil {
		return fmt.Errorf("failed to stop pool: %v", err)
	}

	if deleteContents {
		// Delete all volumes in the pool
		volumes, err := pool.ListAllStorageVolumes(0)
		if err != nil {
			return fmt.Errorf("failed to list volumes: %v", err)
		}
		for _, vol := range volumes {
			err = vol.Delete(0)
			if err != nil {
				vol.Free()
				return fmt.Errorf("failed to delete volume: %v", err)
			}
			vol.Free()
		}

		// Delete the pool itself (including on-disk data)
		err = pool.Delete(libvirt.STORAGE_POOL_DELETE_NORMAL)
		if err != nil {
			return fmt.Errorf("failed to delete pool data: %v", err)
		}
	}

	// Undefine the pool
	err = pool.Undefine()
	if err != nil {
		return fmt.Errorf("failed to undefine pool: %v", err)
	}

	return nil
}
