package cmd

import (
	"errors"
	"os/exec"

	"github.com/DaanV2/projects-tool/config"
	"github.com/DaanV2/projects-tool/ide"
	"github.com/DaanV2/projects-tool/projects"
	"github.com/DaanV2/projects-tool/regex"
	"github.com/DaanV2/projects-tool/slicesx"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

func SelectWorkload(cmd *cobra.Command, args []string) {
	var selectProject *projects.Project
	c := config.GetConfig()
	projs := projects.GetProjects(c.ProjectFolders)

	// If filter pattern specified apply
	if len(args) > 0 {
		filter := args[0]
		projs = slicesx.Filter(projs, func(item *projects.Project) bool {
			return regex.IsMatch(item.Folder, filter)
		})
	}

	projOptions := make([]huh.Option[projects.Project], 0, len(projs))
	for _, proj := range projs {
		projOptions = append(projOptions, huh.NewOption(proj.Name, *proj))
	}

	// If there are projects left to select, make a form
	if len(projs) > 1 {
		selectProject = projs[0]
		selectForm := huh.NewSelect[projects.Project]().
			Title("Select project").
			Description("The project to open").
			Options(projOptions...).
			Value(selectProject)

		FatalIfNotAbort(Display("", "", selectForm))
		// Else if there is only 1 select that one
	} else if len(projs) == 1 {
		selectProject = projs[0]
	}

	if selectProject == nil {
		log.Fatal("no project selected")
		return
	}

	userIde, _ := cmd.Flags().GetString("ide")
	err := invokeIDE(ide.GetIDE(userIde), selectProject, c)
	if err != nil {
		log.Fatal("troubling launching the ide", "error", err)
		return
	}

}

func invokeIDE(ideC ide.IDE, project *projects.Project, config *config.Config) error {
	if ideC == nil {
		ideC = findIDE(project, config)
	}
	if ideC == nil {
		return errors.New("couldn't find a IDE to launch this project for")
	}
	ocom := ideC.OpenCommand(project.Folder)
	log.Debug("Attempting...",
		"command", ocom,
		"project", project.Name,
		"folder", project.Folder,
	)

	log.Info("Launching...", "command", ocom, "project", project.Name)
	cmd := exec.Command(ocom)
	cmd.Dir = project.Folder
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

func findIDE(project *projects.Project, config *config.Config) ide.IDE {
	for _, i := range config.IDE {
		if regex.IsMatch(project.Folder, i.PathFilter) {
			return i.IDE.Get()
		}
	}

	return config.DefaultIDE.Get()
}