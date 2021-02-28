// +build !windows

package cmd

import (
	"fmt"
)

func getDefaultConfigFilePath() string {
	return fmt.Sprintf("/etc/submit-server/%s", defaultConfigFileName)
}

func getDefaultDbDirPath() string {
	return "var/cache/submit-server/db"
}
