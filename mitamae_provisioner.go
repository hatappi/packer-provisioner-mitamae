package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/hashicorp/packer/template/interpolate"
	// "log"
	"os"
	"regexp"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Mitamae_version string

	Bin_dir string

	Recipe_path string

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

	if mp.config.Recipe_path == "" {
		return errors.New("recipe_path is required")
	}

	if mp.config.Bin_dir == "" {
		mp.config.Bin_dir = "/usr/local/bin"
	}

	if mp.config.Mitamae_version == "" {
		mp.config.Mitamae_version = "v1.4.5"
	}

	return nil
}

func (mp *MitamaeProvisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	filename, err := mp.mitamae_name(comm)
	if err != nil {
		return err
	}

	err = mp.mitamae_download(ui, comm, filename)
	if err != nil {
		return err
	}

	err = mp.exec_recipe(comm, filename)
	if err != nil {
		return err
	}

	return nil
}

func (mp *MitamaeProvisioner) exec_recipe(comm packer.Communicator, filename string) error {
	bin_dir := fmt.Sprintf("%s/%s", mp.config.Bin_dir, filename)

	var cmd packer.RemoteCmd
	cmd.Command = fmt.Sprintf("%s local %s", bin_dir, mp.config.Recipe_path)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := comm.Start(&cmd); err != nil {
		return err
	}

	cmd.Wait()

	if stderr.String() != "" {
		return errors.New(stderr.String())
	}

	return nil
}

func (mp *MitamaeProvisioner) mitamae_name(comm packer.Communicator) (string, error) {
	var cmd packer.RemoteCmd
	cmd.Command = "uname -s -m | awk '{printf(\"%s-%s\",$2,tolower($1))}'"
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := comm.Start(&cmd); err != nil {
		return "", err
	}
	cmd.Wait()
	os_arch := stdout.String()

	os := regexp.MustCompile(`^[^-]+-`).ReplaceAllString(os_arch, "")
	switch os {
	case "linux", "darwin":
		return fmt.Sprintf("mitamae-%s", os_arch), nil
	default:
		return "", errors.New(fmt.Sprintf("%s is not support.", os_arch))
	}
}

func (mp *MitamaeProvisioner) mitamae_download(ui packer.Ui, comm packer.Communicator, filename string) error {
	download_url := fmt.Sprintf("https://github.com/itamae-kitchen/mitamae/releases/download/%s/%s", mp.config.Mitamae_version, filename)
	bin_path := fmt.Sprintf("%s/%s", mp.config.Bin_dir, filename)

	ui.Say(fmt.Sprintf("Download MItamae from %s to %s", download_url, bin_path))

	var cmd packer.RemoteCmd
	cmd.Command = fmt.Sprintf("wget %s -q -P %s -O %s", download_url, mp.config.Bin_dir, filename)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := comm.Start(&cmd); err != nil {
		return err
	}
	cmd.Wait()
	if stderr.String() != "" {
		return errors.New(stderr.String())
	}

	var chmod_cmd packer.RemoteCmd
	chmod_cmd.Command = fmt.Sprintf("chmod +x %s", bin_path)
	var chmod_stderr bytes.Buffer
	chmod_cmd.Stderr = &chmod_stderr
	if err := comm.Start(&chmod_cmd); err != nil {
		return err
	}
	chmod_cmd.Wait()
	if chmod_stderr.String() != "" {
		return errors.New(chmod_stderr.String())
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
