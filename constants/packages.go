package constants

/*
In CloudInit user-data - we can define packages directly to get initialized on the Machine

# If they are not available - we need to manually provide a Shell Script to initialize the dependency

Example of using packages in our Cloud Init Data

		packages:
	  	- zsh
	  	- openjdk-11-jdk
	  	- git
	  	- curl
*/
// Define a custom type for your package names
type CloudInitPkg string

// Define constants of type PackageName
const (
	ZSH        CloudInitPkg = "zsh"
	OpenJDK11  CloudInitPkg = "openjdk-11-jdk"
	Git        CloudInitPkg = "git"
	Curl       CloudInitPkg = "curl"
	NetTools   CloudInitPkg = "net-tools"
	BuildTools CloudInitPkg = "build-tools"
)
