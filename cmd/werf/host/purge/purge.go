package reset

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge werf images, stages, cache and other data of all projects on host machine",
		Long: common.GetLongCommandDescription(`Purge werf images, stages, cache and other data of all projects on host machine.

The data include:
* Old service tmp dirs, which werf creates during every build, publish, deploy and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.
* Shared context:
  * Mounts which persists between several builds (mounts from build_dir).

WARNING: Do not run this command during any other werf command is working on the host machine. This command is supposed to be run manually.`),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ApplyLogOptions(&CommonCmdData); err != nil {
				cmd.Help()
				fmt.Println()
				return err
			}
			common.LogVersion()

			return runReset()
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "")
	common.SetupInsecureRepo(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)

	return cmd
}

func runReset() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{AllowInsecureRepo: *CommonCmdData.InsecureRepo}); err != nil {
		return err
	}

	if err := docker.Init(common.ApplyAndGetDockerConfig(&CommonCmdData)); err != nil {
		return err
	}

	commonOptions := cleaning.CommonOptions{
		DryRun:         *CommonCmdData.DryRun,
		SkipUsedImages: false,
		RmiForce:       true,
		RmForce:        true,
	}
	if err := cleaning.HostPurge(commonOptions); err != nil {
		return err
	}

	return nil
}
