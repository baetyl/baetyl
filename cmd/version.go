package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/containerd/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/spf13/cobra"
)

// Compile parameter
var (
	Version  string
	Revision string
	Platform specs.Platform
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show the version of baetyl",
	Long:  ``,
	Run:   version,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	Platform = platforms.DefaultSpec()
}

func version(cmd *cobra.Command, args []string) {
	fmt.Printf("Version:      %s\nGit revision: %s\nGo version:   %s\nPlatform:     %s\n\n", Version, Revision, runtime.Version(), format(Platform))
}

func format(platform specs.Platform) string {
	if platform.OS == "" {
		return "unknown"
	}

	return joinNotEmpty(platform.OS, platform.Architecture, platform.Variant)
}

func joinNotEmpty(s ...string) string {
	var ss []string
	for _, s := range s {
		if s == "" {
			continue
		}

		ss = append(ss, s)
	}

	return strings.Join(ss, "/")
}
