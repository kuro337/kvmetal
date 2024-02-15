package vm

import (
	"fmt"
	"log"
	"log/slog"
	"os/exec"

	"kvmgo/utils"
)

/*
1. while running - detach raw disk
virsh detach-disk spark --target vdb

2. take snapshot
virsh snapshot-create-as --domain spark spark_hadoop --description "Machine with Spark,Hadoop,Java,Scala configured"

3. Reattach user-data.img raw disk
virsh attach-disk spark /home/kuro/Documents/Code/Go/kvmgo/data/artifacts/spark/userdata/user-data.img vdb --cache none

4. To restore the VM to the snapshot
virsh snapshot-revert --domain spark spark_hadoop

5. Permanently Deleting
virsh snapshot-delete --domain spark --snapshotname <snapshot-name>
*/
func SnapshotBegin(vmName, snapshotName, userdataAbsPath, desc string) error {
	is_running, err := utils.IsVMRunning(vmName)
	if err != nil {
		log.Printf("Failed to check if VM is running - could not cleanup ERROR:%s", err)
		return err
	}

	if is_running == false {
		utils.LogError("VM must be in the Running State to take a Snapshot")
		return fmt.Errorf("VM must be in the Running State to take a Snapshot")
	}

	if utils.PathResolvable(userdataAbsPath) == false {
		utils.LogError("User Data Raw Disk Path cannot be Resolved - make sure this is valid before initiating snapshots. Otherwise the disk will not be able to be Reattached to the VM - potentially making it inaccessible normally.")
		return fmt.Errorf("User Data Raw Disk Path cannot be Resolved - make sure this is valid before initiating snapshots. Otherwise the disk will not be able to be Reattached to the VM - potentially making it inaccessible normally.")
	}

	if err := DetachDisk(vmName); err != nil {
		slog.Error("Failed Detaching Disk", "error", err)
		return err
	}

	if err := SnapshotVM(vmName, snapshotName, desc); err != nil {

		slog.Error("Failed Taking Snapshot", "error", err)
		return err
	}

	if err := ReAttachDisk(vmName, userdataAbsPath); err != nil {
		slog.Error("Failed Reattach Disk", "error", err)
		return err
	}

	return nil
}

// DetachDisk temporarily detaches the raw Disk as Point In Time snapshots can only be taken for qCow2 Disks
func DetachDisk(vmName string) error {
	detachCmd := exec.Command("virsh", "detach-disk", vmName, "--target", "vdb")
	if err := detachCmd.Run(); err != nil {
		log.Printf("Failed to Detach user-data.img raw disk for VM %s: %v", vmName, err)
		return err
	}
	log.Printf("User Data Detached Successfully for VM %s", vmName)
	return nil
}

// SnapshotVM takes the snapshot of an Image once the Raw disk has been detached: virsh snapshot-create-as --domain vm_name snap_name --description "My Snapshot"
func SnapshotVM(vmName, snapshotName, desc string) error {
	snapshotCmd := exec.Command("virsh", "snapshot-create-as", vmName, snapshotName, "vdb", "--description", desc)
	if err := snapshotCmd.Run(); err != nil {
		log.Printf("Failed Taking Snapshot of VM %s: %v", vmName, err)
		return err
	}
	log.Printf("Snapshot Successfully Generated for VM %s", vmName)
	log.Printf("To restore the VM Point In Time use the command virsh snapshot-revert --domain %s %s", vmName, snapshotName)
	return nil
}

// ReAttachDisk attaches the userdata raw disk back to the VM once the Snapshot has been completed
func ReAttachDisk(vmName, userdataimgAbsPath string) error {
	reAttachCmd := exec.Command("virsh", "attach-disk", vmName, userdataimgAbsPath, "vdb", "--cache", "none")
	if err := reAttachCmd.Run(); err != nil {
		log.Printf("Failed to Reattach user-data.img raw disk for VM %s: %v", vmName, err)
		return err
	}

	log.Printf("User Data Attached Successfully for VM %s", vmName)
	return nil
}
