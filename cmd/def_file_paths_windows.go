// +build windows

package cmd

import (
	"fmt"
)

func getDefaultConfigFilePath() string {
	return fmt.Sprintf("C:\\ProgramData\\submit-server\\%s", defaultConfigFileName)
}

func getDefaultDbDirPath() string {
	return "C:\\ProgramData\\submit-server\\db"
}
