/*
MIT License

Copyright (c) Nhost

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package cmd

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-getter"
	"github.com/mrinalwahal/cli/nhost"
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{"up"},
	Short:   "Upgrade this utility to latest version",
	Long: `Automatically check for the latest available version of this
utility and upgrade to it.`,
	Run: func(cmd *cobra.Command, args []string) {

		release, err := nhost.LatestRelease()
		if err != nil {
			log.Debug(err)
			log.Fatal("Failed to fetch latest release")
		}

		if release.TagName == Version {
			log.WithField("component", release.TagName).Info("You already have the latest version. Hurray!")
		} else {
			log.WithField("component", release.TagName).Info("New version available")

			asset := release.Asset()

			// initialize hashicorp go-getter client
			client := &getter.Client{
				Ctx:  context.Background(),
				Dir:  false,
				Src:  asset.BrowserDownloadURL,
				Mode: getter.ClientModeDir,
			}

			// get installed binary destination
			DESTINATION, err := exec.LookPath("nhost")
			if err != nil {
				log.WithField("compnent", Version).Debug(err)
				log.WithField("compnent", Version).Fatal("Failed to fetch installed binary path")
			}
			client.Dst = filepath.Dir(DESTINATION)
			log.WithField("compnent", release.TagName).Debugf("Installing in %s", client.Dst)

			// remove the installed binary
			if err := deletePath(DESTINATION); err != nil {
				log.WithField("compnent", Version).Debug(err)
				log.WithField("compnent", Version).Fatal("Failed to remove already installed binary")
			}

			// download the new one
			if err := client.Get(); err != nil {
				log.WithField("compnent", release.TagName).Debug(err)
				log.WithField("compnent", release.TagName).Fatal("Failed to download release")
			}

			log.WithField("compnent", release.TagName).Info("New release downloaded")
			log.Infof("Check version with: %vnhost version%v", Bold, Reset)
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upgradeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upgradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
