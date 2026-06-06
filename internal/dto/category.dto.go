package dto

import "time"

type CategoryItem struct {
	ID        uint
	Name      string
	IsGlobal  bool
	CreatedAt time.Time
}

type CategoryManagementData struct {
	Profile    SidebarProfile
	Categories []CategoryItem
}
