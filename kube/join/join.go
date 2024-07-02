package join

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"kvmgo/kube"
	"kvmgo/network"
)

// JoinNodes joins the worker nodes with the Control Node for kubernetes
func JoinNodesNew(nodes []string) error {
	controlDomain := nodes[0]
	n := len(nodes)

	if n < 2 {
		return errors.New("Not enough nodes to form a cluster")
	}

	fmt.Println(controlDomain)

	control, err := kube.NewControl(controlDomain)
	if err != nil {
		return fmt.Errorf("Error:%s", err)
	}

	joinCmd, serr, err := control.GetJoinCmd()
	if err != nil {
		return fmt.Errorf("Error:%s %s", serr, err)
	}

	log.Println(joinCmd)

	ch := make(chan error, n-1)

	for i := 1; i < n; i++ {
		go func(node string) {
			_, err := RunJoinCmdNew(node, joinCmd)
			ch <- err
		}(nodes[i])
	}

	var errs strings.Builder

	for i := 1; i < n; i++ {
		err := <-ch
		if err != nil {
			errs.WriteString(fmt.Sprintf("Failed to join: %s. ", err))
		}
	}

	close(ch)

	str := errs.String()
	if str != "" {
		return fmt.Errorf("All workers were not joined successfully. %s", str)
	}

	log.Println("Pausing 5 Seconds before checking kubectl get nodes")
	time.Sleep(5 * time.Second)

	ctlnodes, serr, err := control.CheckNodes()

	log.Printf("kubectl resp: %s , serr: %s\n", ctlnodes, serr)

	return nil
}

// JoinNodes joins the worker nodes with the Control Node for kubernetes
func JoinNodes(nodes []string) error {
	control := nodes[0]
	n := len(nodes)

	if n < 2 {
		return errors.New("Not enough nodes to form a cluster")
	}

	// mclient, err := network.GetSSHClient(control)

	joinCmd, err := GetJoinCmd(control)
	if err != nil {
		return err
	}

	ch := make(chan error, n-1)

	for i := 1; i < n; i++ {
		go func(node string) {
			_, err := RunJoinCmd(node, joinCmd)
			ch <- err
		}(nodes[i])
	}

	var errs strings.Builder

	for i := 1; i < n; i++ {
		err := <-ch
		if err != nil {
			errs.WriteString(fmt.Sprintf("Failed to join: %s. ", err))
		}
	}

	close(ch)

	str := errs.String()
	if str != "" {
		return fmt.Errorf("All workers were not joined successfully. %s", str)
	}

	log.Println("Pausing 5 Seconds before checking kubectl get nodes")
	time.Sleep(5 * time.Second)

	// nodes , err := CheckNodes()

	return nil
}

// CheckNodes returns the current nodes active on the Cluster
func CheckNodes(control *network.VMClient) (string, error) {
	// kubectl get nodes
	out, _, err := control.RunCmd("kubectl get nodes")
	//	out, err = kssh.RunCmd(mclient, "ls")
	if err != nil {
		return "", fmt.Errorf("failed cmd Error:%s", err)
	}

	return out, nil
}

// GetJoinCmd gets the kubeadm Cluster join command for workers
func GetJoinCmd(control string) (string, error) {
	// mclient, err := kssh.EstablishSsh(control)
	mclient, err := network.GetSSHClient(control)
	if err != nil {
		return "", fmt.Errorf("Failed to conn control Error:%s", err)
	}
	defer mclient.Close()

	// kubectl get nodes
	out, _, err := mclient.RunCmd("tail -10 kubeadm-init.log")
	//	out, err = kssh.RunCmd(mclient, "ls")
	if err != nil {
		return "", fmt.Errorf("failed cmd Error:%s", err)
	}

	lines := strings.Split(out, "\n")

	var join strings.Builder
	join.WriteString("sudo ")
	f := false
	for _, line := range lines {

		l := strings.TrimSpace(line)
		if f == true {
			join.WriteString(l)
			break
		}
		if strings.Contains(l, "kubeadm") {
			f = true
			join.WriteString(strings.TrimSuffix(l, "\\"))
			join.WriteRune(' ')
		}

	}

	return join.String(), nil
}

// RunJoinCmd runs the kubeadm join on the worker and returns sout & err if any
func RunJoinCmdNew(worker, joinCmd string) (string, error) {
	wclient, err := kube.NewWorker(worker)
	if err != nil {
		return "", fmt.Errorf("Failed to conn worker Error:%s", err)
	}
	out, err := wclient.RunJoinCmd(joinCmd)
	if err != nil {
		return "", fmt.Errorf("failed cmd Error:%s", err)
	}
	return out, nil
}

// RunJoinCmd runs the kubeadm join on the worker and returns sout & err if any
func RunJoinCmd(worker, joinCmd string) (string, error) {
	wclient, err := network.GetSSHClient(worker)
	if err != nil {
		return "", fmt.Errorf("Failed to conn worker Error:%s", err)
	}

	defer wclient.Close()

	out, _, err := wclient.RunCmd(joinCmd)
	if err != nil {
		return "", fmt.Errorf("failed cmd Error:%s", err)
	}

	return out, nil
}
