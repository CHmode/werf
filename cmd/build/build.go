package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/dapp/pkg/ssh_agent"
	"github.com/flant/dapp/pkg/true_git"
)

type buildRubyCliOptions struct {
	Name             string   `json:"name"`
	Dir              string   `json:"dir"`
	BuildDir         string   `json:"build_dir"`
	TmpDirPrefix     string   `json:"tmp_dir_prefix"`
	SSHKey           []string `json:"ssh_key"`
	RegistryUsername string   `json:"registry_username"`
	RegistryPassword string   `json:"registry_password"`
	Registry         string   `json:"repo"`
}

func runBuild(rubyCliOptions buildRubyCliOptions) error {
	if err := lock.Init(); err != nil {
		return err
	}

	if err := ssh_agent.Init(rubyCliOptions.SSHKey); err != nil {
		return fmt.Errorf("cannot initialize ssh-agent: %s", err)
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	projectDir, err := getProjectDir(rubyCliOptions)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	projectName, err := getProjectName(projectDir, rubyCliOptions)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	buildDir, err := getProjectBuildDir(projectName, rubyCliOptions)
	if err != nil {
		return fmt.Errorf("getting project build dir failed: %s", err)
	}

	tmpDir, err := getProjectTmpDir(rubyCliOptions)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}

	hostDockerConfigDir, err := hostDockerConfigDir(tmpDir, rubyCliOptions)
	if err != nil {
		return fmt.Errorf("getting host docker config dir failed: %s", err)
	}

	if err := docker.Init(hostDockerConfigDir); err != nil {
		return err
	}

	dappfile, err := parseDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("parsing dappfile failed: %s", err)
	}

	c := build.NewConveyor(dappfile, projectDir, projectName, buildDir, tmpDir, ssh_agent.SSHAuthSock)
	if err = c.Build(); err != nil {
		return err
	}

	return nil
}

func parseDappfile(projectDir string) ([]*config.Dimg, error) {
	for _, dappfileName := range []string{"dappfile.yml", "dappfile.yaml"} {
		dappfilePath := path.Join(projectDir, dappfileName)
		if exist, err := file.FileExists(dappfilePath); err != nil {
			return nil, err
		} else if exist {
			return config.ParseDimgs(dappfilePath)
		}
	}

	return nil, errors.New("dappfile.y[a]ml not found")
}

func getProjectDir(rubyCliOptions buildRubyCliOptions) (string, error) {
	if rubyCliOptions.Dir != "" {
		return rubyCliOptions.Dir, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return currentDir, nil
}

func getProjectBuildDir(projectName string, options buildRubyCliOptions) (string, error) {
	if options.BuildDir != "" {
		return options.BuildDir, nil
	} else {
		projectBuildDir := path.Join(dapp.GetHomeDir(), "build", projectName)

		if err := os.MkdirAll(projectBuildDir, os.ModePerm); err != nil {
			return "", err
		}

		return projectBuildDir, nil
	}
}

func getProjectTmpDir(options buildRubyCliOptions) (string, error) {
	var tmpDirPrefix string
	if options.TmpDirPrefix != "" {
		tmpDirPrefix = options.TmpDirPrefix
	} else {
		tmpDirPrefix = "dapp-"
	}

	return ioutil.TempDir("", tmpDirPrefix)
}

func getProjectName(projectDir string, rubyCliOptions buildRubyCliOptions) (string, error) {
	name := path.Base(projectDir)

	if rubyCliOptions.Name != "" {
		name = rubyCliOptions.Name
	} else {
		exist, err := isGitOwnRepoExists(projectDir)
		if err != nil {
			return "", err
		}

		if exist {
			remoteOriginUrl, err := gitOwnRepoOriginUrl(projectDir)
			if err != nil {
				return "", err
			}

			if remoteOriginUrl != "" {
				parts := strings.Split(remoteOriginUrl, "/")
				repoName := parts[len(parts)-1]

				gitEnding := ".git"
				if strings.HasSuffix(repoName, gitEnding) {
					repoName = repoName[0 : len(repoName)-len(gitEnding)]
				}

				name = repoName
			}
		}
	}

	return slug.Slug(name), nil
}

func isGitOwnRepoExists(projectDir string) (bool, error) {
	fileInfo, err := os.Stat(path.Join(projectDir, ".git"))
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	return fileInfo.IsDir(), nil
}

func gitOwnRepoOriginUrl(projectDir string) (string, error) {
	localGitRepo := &git_repo.Local{
		Path:   projectDir,
		GitDir: path.Join(projectDir, ".git"),
	}

	remoteOriginUrl, err := localGitRepo.RemoteOriginUrl()
	if err != nil {
		return "", nil
	}

	return remoteOriginUrl, nil
}

func hostDockerConfigDir(projectTmpDir string, rubyCliOptions buildRubyCliOptions) (string, error) {
	dappDockerConfigEnv := os.Getenv("DAPP_DOCKER_CONFIG")

	username, password, err := dockerCredentials(rubyCliOptions)
	if err != nil {
		return "", err
	}
	areDockerCredentialsNotEmpty := username != "" && password != ""

	if areDockerCredentialsNotEmpty && rubyCliOptions.Registry != "" {
		tmpDockerConfigDir := path.Join(projectTmpDir, "docker")

		if err := os.Mkdir(tmpDockerConfigDir, os.ModePerm); err != nil {
			return "", err
		}

		return tmpDockerConfigDir, nil
	} else if dappDockerConfigEnv != "" {
		return dappDockerConfigEnv, nil
	} else {
		return path.Join(os.Getenv("HOME"), ".docker"), nil
	}
}

func dockerCredentials(rubyCliOptions buildRubyCliOptions) (string, string, error) {
	if rubyCliOptions.RegistryUsername != "" && rubyCliOptions.RegistryPassword != "" {
		return rubyCliOptions.RegistryUsername, rubyCliOptions.RegistryPassword, nil
	} else if os.Getenv("DAPP_DOCKER_CONFIG") != "" {
		return "", "", nil
	} else {
		isGCR, err := isGCR(rubyCliOptions)
		if err != nil {
			return "", "", err
		}

		dappIgnoreCIDockerAutologinEnv := os.Getenv("DAPP_IGNORE_CI_DOCKER_AUTOLOGIN")
		if isGCR || dappIgnoreCIDockerAutologinEnv != "" {
			return "", "", nil
		}

		ciRegistryEnv := os.Getenv("CI_REGISTRY")
		ciJobTokenEnv := os.Getenv("CI_JOB_TOKEN")
		if ciRegistryEnv != "" && ciJobTokenEnv != "" {
			return "gitlab-ci-token", ciJobTokenEnv, nil
		}
	}

	return "", "", nil
}

func isGCR(rubyCliOptions buildRubyCliOptions) (bool, error) {
	registryOption := rubyCliOptions.Registry
	if registryOption != "" {
		if registryOption == ":minikube" {
			return false, nil
		}

		return docker_registry.IsGCR(registryOption)
	}

	return false, nil
}
