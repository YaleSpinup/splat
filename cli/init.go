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
package cli

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	githubRepository, githubReleaseTag, localPath, urlBase, outDir string

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
				tempDir, err := ioutil.TempDir("", "splat")
				if err != nil {
					return fmt.Errorf("failed to create temporary directory: %s", err)
				}

				defer func() {
					if err := os.RemoveAll(tempDir); err != nil {
						fmt.Printf("failed to cleanup after myself... there might be temporary files left in %s", tempDir)
					}
				}()

				templatePath, err = downloadGithubRelease(tempDir)
				if err != nil {
					return fmt.Errorf("failed to download template for '%s:%s': %s", githubRepository, githubReleaseTag, err)
				}
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
	initCmd.Flags().StringVarP(&githubRepository, "github", "g", "YaleSpinup/api-tmpl", "Pull template from a Github repository")
	initCmd.Flags().StringVarP(&githubReleaseTag, "tag", "t", "", "Use this release tag instead of latest when pulling from Github")
	initCmd.Flags().StringVarP(&localPath, "local", "l", "", "path to a local template")
	initCmd.Flags().StringVarP(&urlBase, "url", "u", "/v1/test", "base path for url routes")
	initCmd.Flags().StringVarP(&outDir, "out", "o", "", "path to write output")
}

func initializeProject(pkgName, templatePath string) (string, error) {
	appName := path.Base(pkgName)

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	absPath := wd + "/" + appName
	if outDir != "" {
		if filepath.IsAbs(outDir) {
			absPath = outDir
		} else {
			absPath = filepath.Clean(wd + "/" + outDir)
		}
	}

	project := &Project{
		AbsolutePath: absPath,
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

func downloadGithubRelease(path string) (string, error) {
	client := github.NewClient(nil)

	splitRep := strings.Split(githubRepository, "/")
	if len(splitRep) < 2 {
		return "", fmt.Errorf("badly formated repository '%s'", githubRepository)
	}

	owner := splitRep[len(splitRep)-2]
	repo := splitRep[len(splitRep)-1]

	var url string
	if githubReleaseTag != "" {
		release, _, err := client.Repositories.GetReleaseByTag(context.TODO(), owner, repo, githubReleaseTag)
		if err != nil {
			return "", err
		}
		url = *release.ZipballURL
	} else {
		release, _, err := client.Repositories.GetLatestRelease(context.TODO(), owner, repo)
		if err != nil {
			return "", err
		}
		url = *release.ZipballURL
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to start download of template: %s", err)
	}
	defer resp.Body.Close()

	outPath := fmt.Sprintf("%s/template.zip", path)
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary download file %s: %s", outPath, err)
	}
	defer outFile.Close()

	cb, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to temporary download file %s: %s", outPath, err)
	}

	log.Debugf("downloaded %d bytes to %s", cb, outPath)

	extractDir := fmt.Sprintf("%s/template", path)
	if err := unzip(outPath, extractDir); err != nil {
		return "", fmt.Errorf("failed to extract archive %s: %s", outPath, err)
	}

	return extractDir, nil
}

// unzip unzips a zip, modified from: https://golangcode.com/unzip-files-in-go/
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	topdir := ""
	for _, f := range r.File {
		fname := strings.TrimPrefix(f.Name, topdir)
		fpath := filepath.Join(dest, fname)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// create directories unllesss it's the toplevel directory that looks like organization-reponame-commit
			if topdir != "" {
				os.MkdirAll(fpath, os.ModePerm)
			} else {
				topdir = f.Name
			}
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
