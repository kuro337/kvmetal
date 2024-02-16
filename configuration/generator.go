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
}

type ConfigBuilder struct {
	distro   Distro
	deps     []constants.Dependency
	pkgs     []constants.CloudInitPkg
	username string
	password string
	hostname string
}

func NewConfigBuilder(distro constants.Distro, deps []constants.Dependency, pkgs []constants.CloudInitPkg, username, password, hostname string) (*ConfigBuilder, error) {
	var osdistro Distro

	switch distro {
	case constants.Ubuntu:
		osdistro = &ubuntu.UbuntuConfig{}
	default:
		log.Printf("Unknown Distro Passed")
		return nil, fmt.Errorf("Unknown Distribution")
	}

	return &ConfigBuilder{
		distro:   osdistro,
		deps:     deps,
		pkgs:     pkgs,
		username: username,
		password: password,
		hostname: hostname,
	}, nil
}

func (c *ConfigBuilder) CreateCloudInitData() string {
	var userDataBuilder strings.Builder
	baseUserData := SubstituteHostNameAndFqdnUserdata(c.distro.DefaultCloudInit(), c.hostname)

	userDataBuilder.WriteString(baseUserData + "\n")

	userDataBuilder.WriteString(c.BuildPackages())
	userDataBuilder.WriteString(c.BuildRunCmds())

	return userDataBuilder.String()
}

func (c *ConfigBuilder) BuildPackages() string {
	var pkgBuilder strings.Builder

	pkgBuilder.WriteString("\npackages:\n")

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

func (c *ConfigBuilder) BuildRunCmds() string {
	var runCmdBuilder strings.Builder
	runCmdBuilder.WriteString("\nruncmd:\n")

	for _, dependency := range c.deps {

		runCmd := c.distro.GetRunCmd(dependency)
		if runCmd == "" {
			log.Printf("No Run Command Found for Dependency")
			continue
		}
		runCmdBuilder.WriteString("\n  ## Run Command for Dependency: " + string(dependency) + runCmd)
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
