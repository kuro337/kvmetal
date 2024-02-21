# enabling Traffic interception to Masquerade Traffic to and from VM's

```bash


sudo cp qemuhookintercept /etc/libvirt/hooks
sudo chmod +x /etc/libvirt/hooks/qemuhookintercept
sudo ln -sf /etc/libvirt/hooks/qemuhookintercept /etc/libvirt/hooks/qemu
sudo ln -sf /etc/libvirt/hooks/qemuhookintercept /etc/libvirt/hooks/lxc
sudo service libvirtd restart

Cleanup/Disable

sudo rm -f /etc/libvirt/hooks/qemu
sudo rm -f /etc/libvirt/hooks/lxc
sudo rm -f /etc/libvirt/hooks/qemuhookintercept

go build -o qemuhookintercept hooksinterceptor.go
sudo cp qemuhookintercept /etc/libvirt/hooks/
sudo chmod +x /etc/libvirt/hooks/qemuhookintercept
# confirm File was updated
ls -l /etc/libvirt/hooks/qemuhookintercept

virsh start spark

To compile whole dir

go build -o qemuhookintercept .
# optionally
sudo systemctl stop libvirtd & sudo systemctl stop libvirtd before and after copy

sudo cp qemuhookintercept /etc/libvirt/hooks/
sudo chmod +x /etc/libvirt/hooks/qemuhookintercept

# confirm File was updated
ls -l /etc/libvirt/hooks/qemuhookintercept

virsh start spark

go clean -i ./...
go build -o qemuhookintercept


1. example log
LIBVIRT_HOOK: 2024/02/16 19:38:44 Event received - Domain: spark, Action: prepare, Time: 2024-02-16T19:38:44-05:00
LIBVIRT_HOOK: 2024/02/16 19:38:45 Event received - Domain: spark, Action: start, Time: 2024-02-16T19:38:45-05:00
LIBVIRT_HOOK: 2024/02/16 19:38:45 Event received - Domain: spark, Action: started, Time: 2024-02-16T19:38:45-05:00

os.Args[1] is "spark"
os.Args[2] : Action = "prepare" , "start" , "started" etc.


```
