package models

import (
	"container/list"
)

// Location represents a location with users and devices
type Location struct {
	userList   list.List
	deviceList list.List
}
