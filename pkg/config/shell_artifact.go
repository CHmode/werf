package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type ShellArtifact struct {
	*ShellDimg
	BuildArtifact             []string
	BuildArtifactCacheVersion string
}

func (c *ShellArtifact) validate() error {
	return nil
}

func (c *ShellArtifact) toRuby() ruby_marshal_config.ShellArtifact {
	shellArtifact := ruby_marshal_config.ShellArtifact{}
	shellArtifact.ShellDimg = c.ShellDimg.toRuby()
	shellArtifact.BuildArtifact.Version = c.BuildArtifactCacheVersion
	shellArtifact.BuildArtifact.Run = c.BuildArtifact
	return shellArtifact
}
