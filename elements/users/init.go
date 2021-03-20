package users

import "github.com/DAv10195/submit_server/util/containers"

var Roles = containers.NewStringSet()

func init() {
	Roles.Add(Admin, Secretary, StandardUser, Agent)
}
