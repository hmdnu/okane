package template

func sidebarLinkClass(active string, item string) string {
	base := "whitespace-nowrap rounded px-3 py-2 text-sm font-medium"
	if active == item {
		return base + " bg-gray-900 text-white"
	}

	return base + " text-gray-600 hover:bg-gray-100 hover:text-gray-900"
}
