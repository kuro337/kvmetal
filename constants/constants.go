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
	Kafka                   Dependency = "Kafka"
	Calico                  Dependency = "Calico"
	Cilium                  Dependency = "Cilium"
	Spark                   Dependency = "Spark"
	Java11                  Dependency = "Java11"
	Scala                   Dependency = "Scala"
	Sbt                     Dependency = "Sbt"
	Helm                    Dependency = "Helm"
	JDK_SCALA               Dependency = "Jdk_Scala"
)
