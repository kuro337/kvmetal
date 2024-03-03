package constants

type Distro int

const (
	Ubuntu Distro = iota
	Debian
)

type Dependency string

const (
	Zsh                     Dependency = "zsh"
	Hadoop                  Dependency = "Hadoop"
	KubernetesControlCalico Dependency = "KubernetesControlPlaneCalico"
	KubernetesControlCilium Dependency = "KubernetesControlPlaneCilium"
	KubeWorker              Dependency = "KubernetesWorkerNode"
	Calico                  Dependency = "Calico"
	Cilium                  Dependency = "Cilium"
	Spark                   Dependency = "Spark"
	Java11                  Dependency = "Java11"
	Scala                   Dependency = "Scala"
	Sbt                     Dependency = "Sbt"
	Helm                    Dependency = "Helm"
)
