package join

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"kvmgo/kube"
	"kvmgo/network"
)

// JoinNodes joins the worker nodes with the Control Node for kubernetes
func JoinNodes(nodes []string) ([]string, error) {
	slices.Sort(nodes)

	controlDomain := nodes[0]
	n := len(nodes)

	if n < 2 {
		return nil, errors.New("Not enough nodes to form a cluster")
	}

	fmt.Println(controlDomain)

	control, err := kube.NewControl(controlDomain)
	if err != nil {
		return nil, fmt.Errorf("Error:%s", err)
	}

	joinCmd, serr, err := control.GetJoinCmd()
	if err != nil {
		return nil, fmt.Errorf("Error:%s %s", serr, err)
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
		return nil, fmt.Errorf("All workers were not joined successfully. %s", str)
	}

	log.Println("Pausing 5 Seconds before checking kubectl get nodes")

	var joinedNodes []string

	backoffs := 3

	for i := 0; i < backoffs; i++ {
		secs := 5 * (i + 1)
		time.Sleep(time.Duration(secs) * time.Second)
		joined, err := VerifyNodes(control, nodes)
		joinedNodes = joined
		if err != nil {
			log.Printf("error from verify nodes: %s\n", err)
		}
		if len(joined) >= n {
			return joined, nil
		}
	}

	return joinedNodes, fmt.Errorf("Not able to confirm all nodes are Actively running")
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

func VerifyNodes(control *kube.KubeClient, nodes []string) ([]string, error) {
	_, err := control.CheckNodesN()
	if err != nil {
		return nil, err
	}

	var joinedNodes []string
	for _, node := range nodes {
		if jn, ok := control.Nodes[node]; ok {
			log.Printf("Node joined cluster: %s, Status: %s\n", node, jn.Status)
			joinedNodes = append(joinedNodes, node)
		}
	}

	x, y := len(joinedNodes), len(nodes)
	if x < y {
		log.Printf("Joined nodes: %d vs. expected %d - backoff and retry again\n", x, y)

		return joinedNodes, fmt.Errorf("Joined nodes: %d vs. expected %d - backoff and retry again\n", x, y)
	}

	return joinedNodes, nil
}
