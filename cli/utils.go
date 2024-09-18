package cli

import (
	"fmt"
	"strings"

	"kvmgo/types/fpath"
)

// ResolveArtifactsPath gets the Artifacts Path for the VM - i.e Resolves data/images
// and appends the VM name & gets the images path where all Images are cached
// ex. /abs/path/data/images  , /abs/data/artifacts/<vmanme>
func ResolveArtifactsPath(vmName string) (imagesPath, artifactsPath *fpath.FilePath, ferr error) {
	// Resolve Images path - i.e VM/OS images are stored/cached here
	imgsPath, err := fpath.NewPath("data/images", false)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to Resolve VM Images Path. Error:%s", err)
	}

	// Artifacts Path & Validation for the File Path being resolvable
	artifactPath, err := fpath.NewPath(fmt.Sprintf("data/artifacts/%s", vmName), false)
	if err != nil {
		return imgsPath, nil, fmt.Errorf("Artifacts path could not be resolved from cwd. Error :%s", err)
	}

	return imgsPath, artifactPath, nil
}

// SplitKubeJoinNodes splits the input string such as --join="control,worker1,worker2" and returns the nodes
// with the Control Domain being the head and Workers being the Tail
func SplitKubeJoinNodes(nodes string) ([]string, error) {
	var domains []string

	split := strings.Split(strings.TrimSpace(nodes), ",")

	if len(split) <= 1 {
		return nil, fmt.Errorf("Must pass at least a Control Plane domain and a Worker Domain to join nodes")
	}

	for _, domain := range split {
		domains = append(domains, domain)
	}

	return domains, nil
}
