package configuration

import (
	"fmt"
	"log"
	"strings"

	"kvmgo/configuration/ubuntu"
	"kvmgo/constants"
)

type Distro interface {
	GetImageUrl() string
	GetVersion() string
	DefaultCloudInit() string
	GetRunCmd(constants.Dependency) string
	GetPackage(dep constants.CloudInitPkg) string
	GetInitSvc(svc constants.InitSvc) string
}

type ConfigBuilder struct {
	Component Preset
	distro    Distro
	deps      []constants.Dependency
	pkgs      []constants.CloudInitPkg
	initsvc   []constants.InitSvc
	discovery string
	username  string
	password  string
	hostname  string
	sshpubkey string
}

/*
Build the Configuration with Run Commands, Packages, and Metadata for the VM
Pass a Dependency List and a Package List to configure VM at Boot

Check for samples of using the ConfigBuilder
  - cli.configuration.presets.CreateKubeControlPlaneUserData
  - cli.configuration.presets.CreateHadoopUserData
*/
func NewConfigBuilder(
	component Preset,
	distro constants.Distro,
	deps []constants.Dependency,
	pkgs []constants.CloudInitPkg,
	initSvc []constants.InitSvc,
	username,
	password,
	hostname, sshkey string,
) (*ConfigBuilder, error) {
	var osdistro Distro

	switch distro {
	case constants.Ubuntu:
		osdistro = &ubuntu.UbuntuConfig{}
	default:
		log.Printf("Unknown Distro Passed")
		return nil, fmt.Errorf("Unknown Distribution")
	}

	return &ConfigBuilder{
		Component: component,
		distro:    osdistro,
		deps:      deps,
		pkgs:      pkgs,
		username:  username,
		password:  password,
		hostname:  hostname,
		sshpubkey: sshkey,
	}, nil
}

func (c *ConfigBuilder) CreateCloudInitData() string {
	var userDataBuilder strings.Builder
	baseUserData := SubstituteHostNameAndFqdnUserdataSSHPublicKey(
		c.distro.DefaultCloudInit(),
		c.hostname,
		c.sshpubkey)

	userDataBuilder.WriteString(baseUserData + "\n")
	userDataBuilder.WriteString(c.BuildInitSvc())

	userDataBuilder.WriteString(c.BuildPackages())
	userDataBuilder.WriteString(c.BuildRunCmds())

	return c.Component.Substitutions(userDataBuilder.String())

	// return userDataBuilder.String()
}

func (c *ConfigBuilder) BuildInitSvc() string {
	var initSvcBuilder strings.Builder

	initSvcBuilder.WriteString("\npackages:\n")

	for _, svc := range c.initsvc {
		initService := c.distro.GetInitSvc(svc)
		if initService == "" {
			log.Printf("No Run Command Found for Dependency")
			continue
		}
		initSvcBuilder.WriteString("\n" + initService + "\n")
	}

	return initSvcBuilder.String()
}

func (c *ConfigBuilder) BuildPackages() string {
	var pkgBuilder strings.Builder

	//	pkgBuilder.WriteString("\npackages:\n") Already do this in BuildInitSvc

	for _, pkg := range c.pkgs {
		pkgCode := c.distro.GetPackage(pkg)
		if pkgCode == "" {
			log.Printf("No Run Command Found for Dependency")
			continue
		}
		pkgBuilder.WriteString("  - " + pkgCode + "\n")
	}

	return pkgBuilder.String()
}

/*
Generate the runCmd that specifies Boot Instructions to install deps.
Multiple Dependencies can be passed.
*/
func (c *ConfigBuilder) BuildRunCmds() string {
	var runCmdBuilder strings.Builder
	runCmdBuilder.WriteString("\nruncmd:\n")

	for _, dependency := range c.deps {

		runCmd := c.distro.GetRunCmd(dependency)
		if runCmd == "" {
			log.Printf("No Run Command Found for Dependency")
			continue
		}
		runCmdBuilder.WriteString("\n  # Run Command for Dependency: " + string(dependency) + runCmd)
	}

	return runCmdBuilder.String()
}

/*

./main --config spark hadoop zsh

1. ./main --config hadoop zsh

// specifies user/pass and kernel settings - common
cloud_data := Generator(ubuntuBase) // constants.CloudInitUbuntu

cloud_data.AddRunCmds(Zsh,Hadoop) // adds run cmds for Zsh Hadoop

done.

_____


*/
