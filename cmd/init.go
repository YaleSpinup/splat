/*
Copyright © 2020 Yale University

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	gitRepository, localPath, urlBase string

	initCmd = &cobra.Command{
		Use:     "init [package name]",
		Aliases: []string{"initialize", "initialise", "create", "new"},
		Short:   "Initialize a new Spinup API",

		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected one arg (the fully qualified package name)")
			}

			templatePath := ""
			if localPath != "" {
				templatePath = localPath
			} else {
				// git clone
				// templatePath = clonedDir
				return fmt.Errorf("git repositories are currently unsupported")
			}

			if templatePath == "" {
				return fmt.Errorf("templatePath cannot be empty, set a git repo or local template directory")
			}

			templatePath = filepath.Clean(templatePath)

			projectPath, err := initializeProject(args[0], templatePath)
			if err != nil {
				return err
			}

			fmt.Printf("Your Spinup application is ready at\n%s\n", projectPath)

			return nil
		},
	}
)

func init() {
	initCmd.Flags().StringVarP(&gitRepository, "git", "g", "https://github.com/YaleSpinup/api-tmpl", "template repository")
	initCmd.Flags().StringVarP(&localPath, "local", "l", "", "path to a local template")
	initCmd.Flags().StringVarP(&urlBase, "url", "u", "/v1/test", "base path for url routes")
}

func initializeProject(pkgName, templatePath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	appName := path.Base(pkgName)
	project := &Project{
		AbsolutePath: wd + "/" + appName,
		AppName:      appName,
		Copyright:    copyrightLine(),
		DockerName:   dockerName(appName),
		PkgName:      pkgName,
		TemplatePath: templatePath,
		UrlBase:      urlBase,
	}

	if err := project.Create(); err != nil {
		return "", err
	}

	return project.AbsolutePath, nil
}

func copyrightLine() string {
	return fmt.Sprintf("Copyright © %s Yale University", time.Now().Format("2006"))
}

func dockerName(name string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(name)), "-_")
}
