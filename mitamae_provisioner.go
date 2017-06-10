package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	MitamaeVersion string `mapstructure:"mitamae_version"`

	BinDir string `mapstructure:"bin_dir"`

	Option string

	RecipePath string `mapstructure:"recipe_path"`

	ctx interpolate.Context
}

type MitamaeProvisioner struct {
	config Config
}

func (mp *MitamaeProvisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&mp.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &mp.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if mp.config.RecipePath == "" {
		return errors.New("recipe_path is required")
	}

	if mp.config.BinDir == "" {
		mp.config.BinDir = "/usr/local/bin"
	}

	if mp.config.MitamaeVersion == "" {
		mp.config.MitamaeVersion = "v1.4.5"
	}

	return nil
}

func (mp *MitamaeProvisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("MItamae provisioning start")

	filename, err := mp.getMItamaeFileName(comm)
	if err != nil {
		return err
	}

	err = mp.downloadMItamae(ui, comm, filename)
	if err != nil {
		return err
	}

	err = mp.execRecipe(ui, comm, filename)
	if err != nil {
		return err
	}

	ui.Say("MItamae provisioning end")

	return nil
}

func (mp *MitamaeProvisioner) getMItamaeFileName(comm packer.Communicator) (string, error) {
	var cmd packer.RemoteCmd
	cmd.Command = "uname -s -m | awk '{printf(\"%s-%s\",$2,tolower($1))}'"

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := comm.Start(&cmd); err != nil {
		return "", err
	}
	cmd.Wait()
	if stderr.String() != "" {
		return "", errors.New(stderr.String())
	}
	osArch := stdout.String()

	os := regexp.MustCompile(`^[^-]+-`).ReplaceAllString(osArch, "")
	switch os {
	case "linux", "darwin":
		return fmt.Sprintf("mitamae-%s", osArch), nil
	default:
		return "", fmt.Errorf("%s is not support.", osArch)
	}
}

func (mp *MitamaeProvisioner) downloadMItamae(ui packer.Ui, comm packer.Communicator, filename string) error {
	downloadUrl := fmt.Sprintf("https://github.com/itamae-kitchen/mitamae/releases/download/%s/%s", mp.config.MitamaeVersion, filename)
	binPath := fmt.Sprintf("%s/%s", mp.config.BinDir, filename)

	log.Printf("Start MItamae Download from %s to %s", downloadUrl, binPath)

	var cmd packer.RemoteCmd
	cmd.Command = fmt.Sprintf("wget %s -q -P %s -O %s && chmod +x %s", downloadUrl, mp.config.BinDir, filename, binPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := comm.Start(&cmd); err != nil {
		return err
	}
	cmd.Wait()
	if stderr.String() != "" {
		return errors.New(stderr.String())
	}

	log.Printf("Success MItamae Download")

	return nil
}

func (mp *MitamaeProvisioner) execRecipe(ui packer.Ui, comm packer.Communicator, filename string) error {
	binPath := fmt.Sprintf("%s/%s", mp.config.BinDir, filename)

	var cmd packer.RemoteCmd
	cmd.Command = fmt.Sprintf("%s local %s %s", binPath, mp.config.Option, mp.config.RecipePath)
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("MItamae execution failed")
	}

	return nil
}

func (mp *MitamaeProvisioner) Cancel() {
	os.Exit(0)
}

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterProvisioner(new(MitamaeProvisioner))
	server.Serve()
}
