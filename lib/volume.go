package lib

import (
	"fmt"

	"libvirt.org/go/libvirt"
)

type Volume struct {
	volume *libvirt.StorageVol
	Name   string
	Path   string
}

func (v *Volume) Free() {
	v.volume.Free()
}

func (v *Volume) String() string {
	return fmt.Sprintf("Name:%s Path:%s", v.Name, v.Path)
}

func (v *Volume) Delete() error {
	//	STORAGE_VOL_DELETE_NORMAL         = StorageVolDeleteFlags(C.VIR_STORAGE_VOL_DELETE_NORMAL)         // Delete metadata only (fast)
	//	STORAGE_VOL_DELETE_ZEROED         = StorageVolDeleteFlags(C.VIR_STORAGE_VOL_DELETE_ZEROED)         // Clear all data to zeros (slow)
	//	STORAGE_VOL_DELETE_WITH_SNAPSHOTS = StorageVolDeleteFlags(C.VIR_STORAGE_VOL_DELETE_WITH_SNAPSHOTS) // Force removal of volume, even if in use
	return v.volume.Delete(libvirt.STORAGE_VOL_DELETE_NORMAL)
}

func NewVolume(vol *libvirt.StorageVol) (*Volume, error) {
	errStr := ""
	volume := &Volume{volume: vol}

	path, err1 := vol.GetPath()
	name, err2 := vol.GetName()

	if err1 != nil {
		errStr = fmt.Sprintf("path err:%s ", err1.Error())
	} else {
		volume.Path = path
	}
	if err2 != nil {
		errStr = errStr + fmt.Sprintf("name err:%s", err2.Error())
	} else {
		volume.Name = name
	}

	if errStr != "" {
		return volume, fmt.Errorf("Error retrieving Volume Info:%s", errStr)
	}
	return volume, nil
}

// ListPoolVolumes lists the volumes/images associated with the Storage Pool such as ubuntu for Images
func GetVolumes(pool *libvirt.StoragePool) ([]*Volume, error) {
	volumes, err := pool.ListAllStorageVolumes(0)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %v", err)
	}
	var vols []*Volume
	for _, vol := range volumes {
		path, err1 := vol.GetPath()
		name, err2 := vol.GetName()
		if err1 != nil || err2 != nil {
			vol.Free()
			return nil, fmt.Errorf("failed to get volume Path and Name: %s %s", err1.Error(), err2.Error())
		}
		vols = append(vols, &Volume{Name: name, Path: path, volume: &vol})
		vol.Free()
	}

	return vols, nil
}
