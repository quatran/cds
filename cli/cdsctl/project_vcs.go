package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ovh/cds/cli"
	"github.com/ovh/cds/sdk/cdsclient"
	"github.com/ovh/cds/sdk/exportentities"
)

var projectVCSCmd = cli.Command{
	Name:    "vcs",
	Aliases: []string{"vcs"},
	Short:   "Manage VCS on a CDS project",
}

func projectVCS() *cobra.Command {
	return cli.NewCommand(projectVCSCmd, nil, []*cobra.Command{
		cli.NewListCommand(projectVCSListCmd, projectVCSListFunc, nil, withAllCommandModifiers()...),
		cli.NewDeleteCommand(projectVCSDeleteCmd, projectVCSDeleteFunc, nil, withAllCommandModifiers()...),
		cli.NewCommand(projectVCSImportCmd, projectVCSImportFunc, nil, withAllCommandModifiers()...),
		cli.NewCommand(projectVCSExportCmd, projectVCSExportFunc, nil, withAllCommandModifiers()...),
	})
}

var projectVCSListCmd = cli.Command{
	Name:  "list",
	Short: "List VCS available on a project",
	Ctx: []cli.Arg{
		{Name: _ProjectKey},
	},
}

func projectVCSListFunc(v cli.Values) (cli.ListResult, error) {
	pfs, err := client.ProjectVCSList(v.GetString(_ProjectKey))
	return cli.AsListResult(pfs), err
}

var projectVCSDeleteCmd = cli.Command{
	Name:  "delete",
	Short: "Delete a VCS configuration on a project",
	Ctx: []cli.Arg{
		{Name: _ProjectKey},
	},
	Args: []cli.Arg{
		{Name: "name"},
	},
}

func projectVCSDeleteFunc(v cli.Values) error {
	return client.ProjectVCSDelete(v.GetString(_ProjectKey), v.GetString("name"))
}

var projectVCSImportCmd = cli.Command{
	Name:    "import",
	Short:   "Import a VCS configuration on a project from a yaml file",
	Example: "cdsctl project vcs import MY-PROJECT file.yml",
	Ctx: []cli.Arg{
		{Name: _ProjectKey},
	},
	Args: []cli.Arg{
		{Name: "filename"},
	},
	Flags: []cli.Flag{
		{Name: "force", Type: cli.FlagBool},
	},
}

func projectVCSImportFunc(v cli.Values) error {
	f, err := os.Open(v.GetString("filename"))
	if err != nil {
		return cli.WrapError(err, "unable to open file %s", v.GetString("filename"))
	}
	defer f.Close()

	var mods []cdsclient.RequestModifier
	if v.GetBool("force") {
		mods = append(mods, cdsclient.Force())
	}

	_, err = client.ProjectVCSImport(v.GetString(_ProjectKey), f, mods...)
	return err
}

var projectVCSExportCmd = cli.Command{
	Name:    "export",
	Short:   "Export a VCS configuration from a project to stdout",
	Example: "cdsctl vcs export MY-PROJECT MY-VCS-SERVER-NAME > file.yaml",
	Ctx: []cli.Arg{
		{Name: _ProjectKey},
	},
	Args: []cli.Arg{
		{Name: "name"},
	},
}

func projectVCSExportFunc(v cli.Values) error {
	pf, err := client.ProjectVCSGet(v.GetString(_ProjectKey), v.GetString("name"))
	if err != nil {
		return err
	}

	btes, err := exportentities.Marshal(pf, exportentities.FormatYAML)
	if err != nil {
		return err
	}

	fmt.Println(string(btes))
	return nil
}
