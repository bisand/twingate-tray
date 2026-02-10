package tray

// Menu item IDs
const (
	MenuItemConnect        = 1
	MenuItemDisconnect     = 2
	MenuItemSeparator1     = 3
	MenuItemNetworkInfo    = 4
	MenuItemConnectionTime = 5
	MenuItemStatus         = 6
	MenuItemSeparator2     = 7
	MenuItemRefreshStatus  = 8
	MenuItemConnectionInfo = 9
	MenuItemSeparator3     = 10
	MenuItemExitNode       = 11
	MenuItemResources      = 12
	MenuItemSeparator4     = 13
	MenuItemOpenWebAdmin   = 14
	MenuItemDiagReport     = 15
	MenuItemAutoConnect    = 16
	MenuItemSeparator5     = 17
	MenuItemAbout          = 18
	MenuItemSeparator6     = 19
	MenuItemQuit           = 20

	// Exit node submenu items (100-199)
	MenuItemExitNodeStart  = 101
	MenuItemExitNodeStop   = 102
	MenuItemExitNodeList   = 103
	MenuItemExitNodeSwitch = 104

	// Resources submenu base (200-299 for dynamic resources)
	MenuItemResourcesBase = 200
)

// Icon specifications
const (
	IconSize      = 256
	IconPadding   = 0.08 // 8% padding
	Supersample   = 2    // 2x supersampling for antialiasing
	ViewBoxAspect = 448.0 / 512.0
)
