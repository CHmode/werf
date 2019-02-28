package cleanup

import (
	"fmt"
	"path"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"

	"github.com/spf13/cobra"
)

var CmdData struct {
	WithoutKube bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Cleanup project images from images repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ApplyLogOptions(&CommonCmdData); err != nil {
				cmd.Help()
				fmt.Println()
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runCleanup()
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to delete images from the specified images repo.")
	common.SetupInsecureRepo(&CommonCmdData, cmd)
	common.SetupImagesCleanupPolicies(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)

	cmd.Flags().BoolVarP(&CmdData.WithoutKube, "without-kube", "", false, "Do not skip deployed kubernetes images")

	return cmd
}

func runCleanup() error {
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

	if err := kube.Init(kube.InitOptions{KubeContext: *CommonCmdData.KubeContext, KubeConfig: *CommonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}
	common.LogProjectDir(projectDir)

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	var imagesNames []string
	for _, image := range werfConfig.Images {
		imagesNames = append(imagesNames, image.Name)
	}

	commonRepoOptions := cleaning.CommonRepoOptions{
		ImagesRepo:  imagesRepo,
		ImagesNames: imagesNames,
		DryRun:      *CommonCmdData.DryRun,
	}

	var localRepo *git_repo.Local
	gitDir := path.Join(projectDir, ".git")
	if exist, err := util.DirExists(gitDir); err != nil {
		return err
	} else if exist {
		localRepo = &git_repo.Local{
			Path:   projectDir,
			GitDir: gitDir,
		}
	}

	policies, err := common.GetImagesCleanupPolicies(&CommonCmdData)
	if err != nil {
		return err
	}

	kubernetesClients, err := kube.GetAllClients(kube.GetClientsOptions{KubeConfig: *CommonCmdData.KubeConfig})
	if err != nil {
		return fmt.Errorf("unable to get kubernetes clusters connections: %s", err)
	}

	imagesCleanupOptions := cleaning.ImagesCleanupOptions{
		CommonRepoOptions: commonRepoOptions,
		LocalGit:          localRepo,
		KubernetesClients: kubernetesClients,
		WithoutKube:       CmdData.WithoutKube,
		Policies:          policies,
	}

	if err := cleaning.ImagesCleanup(imagesCleanupOptions); err != nil {
		return err
	}

	return nil
}
