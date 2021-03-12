// +build windows

package path

func GetDefaultConfigDirPath() string {
	return "C:\\ProgramData\\submit-server"
}

func GetDefaultDbDirPath() string {
	return "C:\\ProgramData\\submit-server\\db"
}
