package build

import (
	"fmt"

	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"

	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/util"
)

func NewPushPhase(repo string, opts PushOptions) *PushPhase {
	tagsByScheme := map[TagScheme][]string{
		CustomScheme:    opts.Tags,
		CIScheme:        opts.TagsByCI,
		GitBranchScheme: opts.TagsByGitBranch,
		GitTagScheme:    opts.TagsByGitTag,
		GitCommitScheme: opts.TagsByGitCommit,
	}
	return &PushPhase{Repo: repo, TagsByScheme: tagsByScheme, WithStages: opts.WithStages}
}

const (
	CustomScheme    TagScheme = "custom"
	GitTagScheme    TagScheme = "git_tag"
	GitBranchScheme TagScheme = "git_branch"
	GitCommitScheme TagScheme = "git_commit"
	CIScheme        TagScheme = "ci"

	RepoDimgstageTagFormat = "dimgstage-%s"
)

type TagScheme string

type PushPhase struct {
	WithStages   bool
	Repo         string
	TagsByScheme map[TagScheme][]string
}

func (p *PushPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("PushPhase.Run\n")
	}

	err := c.GetDockerAuthorizer().LoginForPush(p.Repo)
	if err != nil {
		return fmt.Errorf("login into '%s' for push failed: %s", p.Repo, err)
	}

	for _, dimg := range c.DimgsInOrder {
		if p.WithStages {
			if dimg.GetName() == "" {
				fmt.Printf("# Pushing dimg stages cache\n")
			} else {
				fmt.Printf("# Pushing dimg/%s stages cache\n", dimg.GetName())
			}

			err := p.pushDimgStages(c, dimg)
			if err != nil {
				return fmt.Errorf("unable to push dimg %s stages: %s", dimg.GetName(), err)
			}
		}

		if !dimg.isArtifact {
			if dimg.GetName() == "" {
				fmt.Printf("# Pushing dimg\n")
			} else {
				fmt.Printf("# Pushing dimg/%s\n", dimg.GetName())
			}

			err := p.pushDimg(c, dimg)
			if err != nil {
				return fmt.Errorf("unable to push dimg %s: %s", dimg.GetName(), err)
			}
		}
	}

	return nil
}

func (p *PushPhase) pushDimgStages(c *Conveyor, dimg *Dimg) error {
	stages := dimg.GetStages()

	existingStagesTags, err := docker_registry.DimgstageTags(p.Repo)
	if err != nil {
		return fmt.Errorf("error fetching existing stages cache list %s: %s", p.Repo, err)
	}

	for _, stage := range stages {
		stageTagName := fmt.Sprintf(RepoDimgstageTagFormat, stage.GetSignature())
		stageImageName := fmt.Sprintf("%s:%s", p.Repo, stageTagName)

		if util.IsStringsContainValue(existingStagesTags, stageTagName) {
			if dimg.GetName() == "" {
				fmt.Printf("# Ignore existing in repo image %s for dimg stage/%s\n", stageImageName, stage.Name())
			} else {
				fmt.Printf("# Ignore existing in repo image %s for dimg/%s stage/%s\n", stageImageName, dimg.GetName(), stage.Name())
			}

			continue
		}

		err := func() error {
			var err error

			imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(stageImageName))
			err = lock.Lock(imageLockName, lock.LockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}
			defer lock.Unlock(imageLockName)

			if dimg.GetName() == "" {
				fmt.Printf("# Pushing image %s for dimg stage/%s\n", stageImageName, stage.Name())
			} else {
				fmt.Printf("# Pushing image %s for dimg/%s stage/%s\n", stageImageName, dimg.GetName(), stage.Name())
			}

			stageImage := c.GetImage(stage.GetImage().Name())

			err = stageImage.Export(stageImageName)
			if err != nil {
				return fmt.Errorf("error pushing %s: %s", stageImageName, err)
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PushPhase) pushDimg(c *Conveyor, dimg *Dimg) error {
	var dimgRepository string
	if dimg.GetName() != "" {
		dimgRepository = fmt.Sprintf("%s/%s", p.Repo, dimg.GetName())
	} else {
		dimgRepository = p.Repo
	}

	existingTags, err := docker_registry.DimgTags(dimgRepository)
	if err != nil {
		return fmt.Errorf("error fetch existing tags of dimg %s: %s", dimgRepository, err)
	}

	stages := dimg.GetStages()
	lastStageImage := stages[len(stages)-1].GetImage()

	for scheme, tags := range p.TagsByScheme {
	ProcessingTags:
		for _, tag := range tags {
			dimgImageName := fmt.Sprintf("%s:%s", dimgRepository, tag)

			if util.IsStringsContainValue(existingTags, tag) {
				parentID, err := docker_registry.ImageParentId(dimgImageName)
				if err != nil {
					return fmt.Errorf("unable to get image %s parent id: %s", dimgImageName, err)
				}

				if lastStageImage.ID() == parentID {
					if dimg.GetName() == "" {
						fmt.Printf("# Ignore existing in repo image %s for dimg\n", dimgImageName)
					} else {
						fmt.Printf("# Ignore existing in repo image %s for dimg/%s\n", dimgImageName, dimg.GetName())
					}
					continue ProcessingTags
				}
			}

			err := func() error {
				var err error

				imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(dimgImageName))
				err = lock.Lock(imageLockName, lock.LockOptions{})
				if err != nil {
					return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
				}
				defer lock.Unlock(imageLockName)

				fmt.Printf("# Build %s layer with tag scheme '%s'\n", dimgImageName, scheme)

				pushImage := image.NewDimgImage(c.GetImage(lastStageImage.Name()), dimgImageName)

				pushImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
					"dapp-tag-scheme": string(scheme),
					"dapp-dimg":       "true",
				})

				err = pushImage.Build(image.BuildOptions{})
				if err != nil {
					return fmt.Errorf("error building %s with tag scheme '%s': %s", dimgImageName, scheme, err)
				}

				if dimg.GetName() == "" {
					fmt.Printf("# Pushing image %s for dimg\n", dimgImageName)
				} else {
					fmt.Printf("# Pushing image %s for dimg/%s\n", dimgImageName, dimg.GetName())
				}

				err = pushImage.Export()
				if err != nil {
					return fmt.Errorf("error pushing %s: %s", dimgImageName, err)
				}

				return nil
			}()

			if err != nil {
				return err
			}
		}
	}

	return nil
}
