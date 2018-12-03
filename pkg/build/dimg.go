package build

import (
	"fmt"
	"os"
	"strings"

	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/image"
)

type Dimg struct {
	baseImageName     string
	baseImageDimgName string

	stages    []stage.Interface
	baseImage *image.Stage
}

func (d *Dimg) SetStages(stages []stage.Interface) {
	d.stages = stages
}

func (d *Dimg) GetStages() []stage.Interface {
	return d.stages
}

func (d *Dimg) GetStage(name stage.StageName) stage.Interface {
	for _, stage := range d.stages {
		if stage.Name() == name {
			return stage
		}
	}

	return nil
}

func (d *Dimg) LatestStage() stage.Interface {
	return d.stages[len(d.stages)-1]
}

func (d *Dimg) GetName() string {
	return ""
}

func (d *Dimg) SetupBaseImage(c *Conveyor) {
	baseImageName := d.baseImageName
	if d.baseImageDimgName != "" {
		baseImageName = c.GetDimg(d.baseImageDimgName).LatestStage().GetImage().Name()
	}

	d.baseImage = c.GetOrCreateImage(nil, baseImageName)
}

func (d *Dimg) GetBaseImage() *image.Stage {
	return d.baseImage
}

func (d *Dimg) PrepareBaseImage(c *Conveyor) error {
	fromImage := d.stages[0].GetImage()

	if fromImage.IsImageExists() {
		return nil
	}

	if d.baseImageDimgName != "" {
		return nil
	}

	ciRegistry := os.Getenv("CI_REGISTRY")
	if ciRegistry != "" && strings.HasPrefix(fromImage.GetName(), ciRegistry) {
		err := c.GetDockerAuthorizer().LoginBaseImage(ciRegistry)
		if err != nil {
			return fmt.Errorf("login into repo %s for base image %s failed: %s", ciRegistry, fromImage.GetName(), err)
		}
	}

	if d.baseImage.IsImageExists() {
		err := d.baseImage.Pull()
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: cannot pull base image %s: %s\n", d.baseImage.GetName(), err)
			fmt.Fprintf(os.Stderr, "WARNING: using existing image %s without pull\n", d.baseImage.GetName())
		}
		return nil
	}

	err := d.baseImage.Pull()
	if err != nil {
		return fmt.Errorf("image %s pull failed: %s", d.baseImage.GetName(), err)
	}

	return nil
}
