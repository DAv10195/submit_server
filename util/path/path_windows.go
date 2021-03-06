// +build windows

package path

func GetDefaultConfigFilePath() string {
	return "C:\\ProgramData\\submit-server"
}

func GetDefaultDbDirPath() string {
	return "C:\\ProgramData\\submit-server\\db"
}
