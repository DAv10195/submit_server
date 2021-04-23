// +build !windows

package path

func GetDefaultConfigDirPath() string {
	return "/etc/submit-server/"
}

func GetDefaultDbDirPath() string {
	return "var/cache/submit-server/db"
}
