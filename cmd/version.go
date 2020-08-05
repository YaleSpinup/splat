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

	"github.com/spf13/cobra"
)

type CmdVersion struct {
	AppVersion string
	BuildTime  string
	GitCommit  string
	GitRef     string
}

var (
	Version *CmdVersion
	long    bool

	versionCmd = &cobra.Command{
		Use:     "version",
		Aliases: []string{"vers"},
		Short:   "Display splat version information",
		RunE: func(_ *cobra.Command, args []string) error {
			if long {
				fmt.Printf("Splat version: %s\nBuildtime: %s\nGitCommit: %s\nGitRef: %s\n", Version.AppVersion, Version.BuildTime, Version.GitCommit, Version.GitRef)
				return nil
			}

			fmt.Printf("%s\n", Version.AppVersion)
			return nil
		},
	}
)

func init() {
	versionCmd.Flags().BoolVarP(&long, "long", "l", false, "get more verbose version information")
}
