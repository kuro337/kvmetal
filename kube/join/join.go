package join

import (
	"errors"
	"fmt"
	"strings"

	kssh "kvmgo/network/ssh"
)

// JoinNodes joins the worker nodes with the Control Node for kubernetes
func JoinNodes(nodes []string) error {
	control := nodes[0]
	n := len(nodes)

	if n < 2 {
		return errors.New("Not enough nodes to form a cluster")
	}

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

	return nil
}

// GetJoinCmd gets the kubeadm Cluster join command for workers
func GetJoinCmd(control string) (string, error) {
	mclient, err := kssh.EstablishSsh(control)
	if err != nil {
		return "", fmt.Errorf("Failed to conn control Error:%s", err)
	}
	defer mclient.Close()

	// kubectl get nodes
	out, err := kssh.RunCmd(mclient, "tail -10 kubeadm-init.log")
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
func RunJoinCmd(worker, joinCmd string) (string, error) {
	wclient, err := kssh.EstablishSsh(worker)
	if err != nil {
		return "", fmt.Errorf("Failed to conn worker Error:%s", err)
	}

	defer wclient.Close()

	out, err := kssh.RunCmd(wclient, joinCmd)
	if err != nil {
		return "", fmt.Errorf("failed cmd Error:%s", err)
	}

	return out, nil
}
