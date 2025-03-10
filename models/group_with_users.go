package models

// GroupWithUsers represents a group along with the list of users that belong to it.
type GroupWithUsers struct {
	Group Group  `json:"group"`
	Users []User `json:"users"`
}
