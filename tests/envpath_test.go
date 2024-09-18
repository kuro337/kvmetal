package tests

import (
	"testing"

	"kvmgo/types/fpath"
)

func TestEnvPathFns(t *testing.T) {
	env := fpath.NewEnvPath("/home/kuro/.cargo/bin")

	_ = env.GenerateNewPath()

	aliased := env.GetAliased()

	t.Log(aliased)
}

/*


export PATH=/bin:$HOME/.local/bin:$HOME/.sdkman/candidates/java/current/bin:$HOME/.sdkman/candidates/sbt/current/bin:$HOME/.vscode-server/bin/8b3775030ed1a69b13e4f4c628c612102e30a681/bin/remote-cli:$HOME/Documents/k8/istio/demo/istio-1.20.1/bin:$HOME/go/bin:/sbin:/snap/bin:/usr/bin:/usr/games:/usr/local/bin:/usr/local/games:/usr/local/go/bin:/usr/local/go1.22.0.linux-amd64/bin:/usr/local/kafka/bin:/usr/local/sbin:/usr/sbin:$HOME/.cargo/bin

*/
