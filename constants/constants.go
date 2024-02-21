package constants

type Distro int

const (
	Ubuntu Distro = iota
	Debian
)

type Dependency string

const (
	Zsh               Dependency = "zsh"
	Hadoop            Dependency = "Hadoop"
	KubernetesControl Dependency = "KubernetesControlPlane"
	KubeWorker        Dependency = "KubernetesWorkerNode"
	Spark             Dependency = "Spark"
	Java11            Dependency = "Java11"
	Scala             Dependency = "Scala"
	Sbt               Dependency = "Sbt"
	Helm              Dependency = "Helm"
)
