package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

// SystemTray manages the system tray icon using D-Bus StatusNotifierItem
type SystemTray struct {
	conn             *dbus.Conn
	connected        bool
	onConnect        func()
	onDisconnect     func()
	onQuit           func()
	serviceName      string
	objectPath       dbus.ObjectPath
	menuPath         dbus.ObjectPath
	mu               sync.RWMutex
	menuRevision     uint32
	registeredString string

	// Cached icon pixmap data (ARGB, network byte order)
	iconData   []byte
	iconWidth  int32
	iconHeight int32
}

// NewSystemTray creates a new system tray instance
func NewSystemTray(onConnect, onDisconnect, onQuit func()) (*SystemTray, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to D-Bus: %w", err)
	}

	st := &SystemTray{
		conn:         conn,
		connected:    false,
		onConnect:    onConnect,
		onDisconnect: onDisconnect,
		onQuit:       onQuit,
		serviceName:  "org.twingate.StatusNotifierItem",
		objectPath:   "/StatusNotifierItem",
		menuPath:     "/MenuBar",
		menuRevision: 1,
	}

	// Generate initial icon
	st.iconData, st.iconWidth, st.iconHeight = generateIconARGBAntialiased(false)

	return st, nil
}

// Start initializes the system tray and registers with D-Bus
func (st *SystemTray) Start() error {
	// Request a well-known bus name (required by SNI protocol)
	pid := os.Getpid()
	busName := fmt.Sprintf("org.kde.StatusNotifierItem-%d-1", pid)
	reply, err := st.conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return fmt.Errorf("failed to request bus name %s: %w", busName, err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Printf("Warning: Could not become primary owner of %s (reply=%d)", busName, reply)
	}
	st.serviceName = busName

	// Export the StatusNotifierItem interface
	err = st.conn.Export(st, st.objectPath, "org.kde.StatusNotifierItem")
	if err != nil {
		return fmt.Errorf("failed to export StatusNotifierItem: %w", err)
	}

	// Export Properties interface
	propsSpec := map[string]map[string]*prop{
		"org.kde.StatusNotifierItem": {
			"Category":            {Value: "Communications"},
			"Id":                  {Value: "twingate-indicator"},
			"Title":               {Value: "Twingate"},
			"Status":              {Value: "Active"},
			"IconName":            {Value: ""},
			"IconPixmap":          {Getter: st.getIconPixmapProp},
			"OverlayIconName":     {Value: ""},
			"OverlayIconPixmap":   {Value: []iconPixmap{}},
			"AttentionIconName":   {Value: ""},
			"AttentionIconPixmap": {Value: []iconPixmap{}},
			"AttentionMovieName":  {Value: ""},
			"ToolTip":             {Getter: st.getToolTipProp},
			"Menu":                {Value: st.menuPath},
			"ItemIsMenu":          {Value: true},
			"WindowId":            {Value: int32(0)},
		},
	}
	err = st.conn.Export(newPropsHandler(propsSpec), st.objectPath, "org.freedesktop.DBus.Properties")
	if err != nil {
		return fmt.Errorf("failed to export Properties: %w", err)
	}

	// Export introspection for the StatusNotifierItem path
	sniIntrospect := introspect.Node{
		Name: string(st.objectPath),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name: "org.kde.StatusNotifierItem",
				Methods: []introspect.Method{
					{Name: "Activate", Args: []introspect.Arg{
						{Name: "x", Type: "i", Direction: "in"},
						{Name: "y", Type: "i", Direction: "in"},
					}},
					{Name: "SecondaryActivate", Args: []introspect.Arg{
						{Name: "x", Type: "i", Direction: "in"},
						{Name: "y", Type: "i", Direction: "in"},
					}},
					{Name: "ContextMenu", Args: []introspect.Arg{
						{Name: "x", Type: "i", Direction: "in"},
						{Name: "y", Type: "i", Direction: "in"},
					}},
					{Name: "Scroll", Args: []introspect.Arg{
						{Name: "delta", Type: "i", Direction: "in"},
						{Name: "orientation", Type: "s", Direction: "in"},
					}},
				},
				Signals: []introspect.Signal{
					{Name: "NewTitle"},
					{Name: "NewIcon"},
					{Name: "NewAttentionIcon"},
					{Name: "NewOverlayIcon"},
					{Name: "NewToolTip"},
					{Name: "NewStatus", Args: []introspect.Arg{
						{Name: "status", Type: "s"},
					}},
				},
				Properties: []introspect.Property{
					{Name: "Category", Type: "s", Access: "read"},
					{Name: "Id", Type: "s", Access: "read"},
					{Name: "Title", Type: "s", Access: "read"},
					{Name: "Status", Type: "s", Access: "read"},
					{Name: "WindowId", Type: "i", Access: "read"},
					{Name: "IconName", Type: "s", Access: "read"},
					{Name: "IconPixmap", Type: "a(iiay)", Access: "read"},
					{Name: "OverlayIconName", Type: "s", Access: "read"},
					{Name: "OverlayIconPixmap", Type: "a(iiay)", Access: "read"},
					{Name: "AttentionIconName", Type: "s", Access: "read"},
					{Name: "AttentionIconPixmap", Type: "a(iiay)", Access: "read"},
					{Name: "AttentionMovieName", Type: "s", Access: "read"},
					{Name: "ToolTip", Type: "(sa(iiay)ss)", Access: "read"},
					{Name: "ItemIsMenu", Type: "b", Access: "read"},
					{Name: "Menu", Type: "o", Access: "read"},
				},
			},
			{
				Name: "org.freedesktop.DBus.Properties",
				Methods: []introspect.Method{
					{Name: "Get", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "property", Type: "s", Direction: "in"},
						{Name: "value", Type: "v", Direction: "out"},
					}},
					{Name: "GetAll", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "properties", Type: "a{sv}", Direction: "out"},
					}},
					{Name: "Set", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "property", Type: "s", Direction: "in"},
						{Name: "value", Type: "v", Direction: "in"},
					}},
				},
			},
		},
	}
	st.conn.Export(introspect.NewIntrospectable(&sniIntrospect), st.objectPath, "org.freedesktop.DBus.Introspectable")

	// Export DBusMenu interface on the menu path
	err = st.conn.Export(st, st.menuPath, "com.canonical.dbusmenu")
	if err != nil {
		return fmt.Errorf("failed to export DBusMenu: %w", err)
	}

	// Export Properties for the menu path (Version property)
	menuPropsSpec := map[string]map[string]*prop{
		"com.canonical.dbusmenu": {
			"Version":       {Value: uint32(3)},
			"TextDirection": {Value: "ltr"},
			"Status":        {Value: "normal"},
			"IconThemePath": {Value: []string{}},
		},
	}
	err = st.conn.Export(newPropsHandler(menuPropsSpec), st.menuPath, "org.freedesktop.DBus.Properties")
	if err != nil {
		return fmt.Errorf("failed to export menu Properties: %w", err)
	}

	// Export introspection data for the menu path
	menuIntrospect := introspect.Node{
		Name: string(st.menuPath),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name: "com.canonical.dbusmenu",
				Methods: []introspect.Method{
					{Name: "GetLayout", Args: []introspect.Arg{
						{Name: "parentId", Type: "i", Direction: "in"},
						{Name: "recursionDepth", Type: "i", Direction: "in"},
						{Name: "propertyNames", Type: "as", Direction: "in"},
						{Name: "revision", Type: "u", Direction: "out"},
						{Name: "layout", Type: "(ia{sv}av)", Direction: "out"},
					}},
					{Name: "GetGroupProperties", Args: []introspect.Arg{
						{Name: "ids", Type: "ai", Direction: "in"},
						{Name: "propertyNames", Type: "as", Direction: "in"},
						{Name: "properties", Type: "a(ia{sv})", Direction: "out"},
					}},
					{Name: "GetProperty", Args: []introspect.Arg{
						{Name: "id", Type: "i", Direction: "in"},
						{Name: "property", Type: "s", Direction: "in"},
						{Name: "value", Type: "v", Direction: "out"},
					}},
					{Name: "Event", Args: []introspect.Arg{
						{Name: "id", Type: "i", Direction: "in"},
						{Name: "eventId", Type: "s", Direction: "in"},
						{Name: "data", Type: "v", Direction: "in"},
						{Name: "timestamp", Type: "u", Direction: "in"},
					}},
					{Name: "EventGroup", Args: []introspect.Arg{
						{Name: "events", Type: "a(isvu)", Direction: "in"},
						{Name: "idErrors", Type: "ai", Direction: "out"},
					}},
					{Name: "AboutToShow", Args: []introspect.Arg{
						{Name: "id", Type: "i", Direction: "in"},
						{Name: "needUpdate", Type: "b", Direction: "out"},
					}},
					{Name: "AboutToShowGroup", Args: []introspect.Arg{
						{Name: "ids", Type: "ai", Direction: "in"},
						{Name: "updatesNeeded", Type: "ai", Direction: "out"},
						{Name: "idErrors", Type: "ai", Direction: "out"},
					}},
				},
				Signals: []introspect.Signal{
					{Name: "ItemsPropertiesUpdated", Args: []introspect.Arg{
						{Name: "updatedProps", Type: "a(ia{sv})"},
						{Name: "removedProps", Type: "a(ias)"},
					}},
					{Name: "LayoutUpdated", Args: []introspect.Arg{
						{Name: "revision", Type: "u"},
						{Name: "parent", Type: "i"},
					}},
					{Name: "ItemActivationRequested", Args: []introspect.Arg{
						{Name: "id", Type: "i"},
						{Name: "timestamp", Type: "u"},
					}},
				},
				Properties: []introspect.Property{
					{Name: "Version", Type: "u", Access: "read"},
					{Name: "TextDirection", Type: "s", Access: "read"},
					{Name: "Status", Type: "s", Access: "read"},
					{Name: "IconThemePath", Type: "as", Access: "read"},
				},
			},
			{
				Name: "org.freedesktop.DBus.Properties",
				Methods: []introspect.Method{
					{Name: "Get", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "property", Type: "s", Direction: "in"},
						{Name: "value", Type: "v", Direction: "out"},
					}},
					{Name: "GetAll", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "properties", Type: "a{sv}", Direction: "out"},
					}},
					{Name: "Set", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "property", Type: "s", Direction: "in"},
						{Name: "value", Type: "v", Direction: "in"},
					}},
				},
			},
		},
	}
	st.conn.Export(introspect.NewIntrospectable(&menuIntrospect), st.menuPath, "org.freedesktop.DBus.Introspectable")

	// Register with StatusNotifierWatcher
	// The SNI spec says to pass the bus name for registration
	st.registeredString = st.serviceName + string(st.objectPath)
	watcher := st.conn.Object("org.kde.StatusNotifierWatcher", "/StatusNotifierWatcher")
	call := watcher.Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, st.serviceName)
	if call.Err != nil {
		log.Printf("Warning: Could not register with StatusNotifierWatcher: %v", call.Err)
	} else {
		log.Printf("Registered with StatusNotifierWatcher at path: %s", st.registeredString)
	}

	log.Println("System tray initialized")
	return nil
}

// iconPixmap represents a single icon pixmap for the SNI protocol: (width, height, argb_data)
type iconPixmap struct {
	Width  int32
	Height int32
	Data   []byte
}

// UpdateStatus updates the tray icon and emits signals when connection state changes
func (st *SystemTray) UpdateStatus(connected bool) {
	st.mu.Lock()
	prevConnected := st.connected
	if prevConnected == connected {
		st.mu.Unlock()
		return
	}
	st.connected = connected
	st.iconData, st.iconWidth, st.iconHeight = generateIconARGBAntialiased(connected)
	st.menuRevision++
	revision := st.menuRevision
	st.mu.Unlock()

	// Emit D-Bus signals for icon change
	st.conn.Emit(st.objectPath, "org.kde.StatusNotifierItem.NewIcon")
	st.conn.Emit(st.objectPath, "org.kde.StatusNotifierItem.NewToolTip")

	// Emit menu layout changed
	st.conn.Emit(st.menuPath, "com.canonical.dbusmenu.LayoutUpdated", revision, int32(0))

	tooltip := "Disconnected from Twingate"
	if connected {
		tooltip = "Connected to Twingate"
	}
	log.Printf("Tray status updated: %s", tooltip)
}

// Stop removes the system tray item
func (st *SystemTray) Stop() {
	if st.registeredString != "" {
		log.Println("Unregistering from StatusNotifierWatcher")
	}
	if st.conn != nil {
		st.conn.Close()
	}
}

// --- StatusNotifierItem D-Bus methods ---

func (st *SystemTray) Title() (string, *dbus.Error) {
	return "Twingate", nil
}

func (st *SystemTray) Id() (string, *dbus.Error) {
	return "twingate-indicator", nil
}

func (st *SystemTray) Status() (string, *dbus.Error) {
	return "Active", nil
}

func (st *SystemTray) Category() (string, *dbus.Error) {
	return "Communications", nil
}

// IconName returns empty string to force the use of IconPixmap
func (st *SystemTray) IconName() (string, *dbus.Error) {
	return "", nil
}

// IconPixmap returns the icon as pixmap data in the SNI format: a(iiay)
func (st *SystemTray) IconPixmap() ([]iconPixmap, *dbus.Error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return []iconPixmap{
		{
			Width:  st.iconWidth,
			Height: st.iconHeight,
			Data:   st.iconData,
		},
	}, nil
}

func (st *SystemTray) ToolTip() (interface{}, *dbus.Error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	tooltip := "Twingate - Disconnected"
	if st.connected {
		tooltip = "Twingate - Connected"
	}

	// ToolTip type: (sa(iiay)ss) = (icon_name, icon_pixmap[], title, description)
	return struct {
		IconName string
		Pixmaps  []iconPixmap
		Title    string
		Desc     string
	}{
		IconName: "",
		Pixmaps:  []iconPixmap{},
		Title:    "Twingate",
		Desc:     tooltip,
	}, nil
}

func (st *SystemTray) Menu() (dbus.ObjectPath, *dbus.Error) {
	return st.menuPath, nil
}

func (st *SystemTray) ItemIsMenu() (bool, *dbus.Error) {
	return true, nil
}

// Activate handles left-click - with ItemIsMenu=true, the panel should show
// the DBusMenu directly, but some panels still call Activate
func (st *SystemTray) Activate(x int32, y int32) *dbus.Error {
	log.Println("Tray icon activated (left-click)")
	return nil
}

func (st *SystemTray) SecondaryActivate(x int32, y int32) *dbus.Error {
	log.Println("Tray icon secondary activated (middle-click)")
	return nil
}

func (st *SystemTray) ContextMenu(x int32, y int32) *dbus.Error {
	log.Println("Tray icon context menu (right-click)")
	return nil
}

func (st *SystemTray) Scroll(delta int32, orientation string) *dbus.Error {
	return nil
}

func (st *SystemTray) XAyatanaLabel() (string, *dbus.Error) {
	return "", nil
}

func (st *SystemTray) XAyatanaLabelGuide() (string, *dbus.Error) {
	return "", nil
}

func (st *SystemTray) XAyatanaOrderingIndex() (uint32, *dbus.Error) {
	return 0, nil
}

func (st *SystemTray) AttentionIconName() (string, *dbus.Error) {
	return "", nil
}

func (st *SystemTray) AttentionIconPixmap() ([]iconPixmap, *dbus.Error) {
	return []iconPixmap{}, nil
}

func (st *SystemTray) AttentionMovieName() (string, *dbus.Error) {
	return "", nil
}

func (st *SystemTray) WindowsMenu() (dbus.ObjectPath, *dbus.Error) {
	return dbus.ObjectPath("/"), nil
}

// --- DBusMenu Interface Implementation ---

// menuLayoutItem represents a DBusMenu layout item: (ia{sv}av)
type menuLayoutItem struct {
	ID         int32
	Properties map[string]dbus.Variant
	Children   []dbus.Variant
}

// makeMenuItem creates a menuLayoutItem wrapped in a dbus.Variant
func makeMenuItem(id int32, props map[string]dbus.Variant) dbus.Variant {
	return dbus.MakeVariant(menuLayoutItem{
		ID:         id,
		Properties: props,
		Children:   []dbus.Variant{},
	})
}

// getMenuItems returns the current menu item definitions
func (st *SystemTray) getMenuItems() map[int32]map[string]dbus.Variant {
	st.mu.RLock()
	connected := st.connected
	st.mu.RUnlock()

	items := make(map[int32]map[string]dbus.Variant)

	// Root item
	items[0] = map[string]dbus.Variant{
		"children-display": dbus.MakeVariant("submenu"),
	}

	// Item 1: Connect or Disconnect
	if connected {
		items[1] = map[string]dbus.Variant{
			"label":   dbus.MakeVariant("Disconnect"),
			"enabled": dbus.MakeVariant(true),
			"visible": dbus.MakeVariant(true),
		}
	} else {
		items[1] = map[string]dbus.Variant{
			"label":   dbus.MakeVariant("Connect"),
			"enabled": dbus.MakeVariant(true),
			"visible": dbus.MakeVariant(true),
		}
	}

	// Item 2: Separator
	items[2] = map[string]dbus.Variant{
		"type":    dbus.MakeVariant("separator"),
		"visible": dbus.MakeVariant(true),
	}

	// Item 3: Connection Info
	items[3] = map[string]dbus.Variant{
		"label":   dbus.MakeVariant("Connection Info..."),
		"enabled": dbus.MakeVariant(true),
		"visible": dbus.MakeVariant(true),
	}

	// Item 4: Status (informational, disabled)
	statusText := "Status: Disconnected"
	if connected {
		statusText = "Status: Connected"
	}
	items[4] = map[string]dbus.Variant{
		"label":   dbus.MakeVariant(statusText),
		"enabled": dbus.MakeVariant(false),
		"visible": dbus.MakeVariant(true),
	}

	// Item 5: Separator
	items[5] = map[string]dbus.Variant{
		"type":    dbus.MakeVariant("separator"),
		"visible": dbus.MakeVariant(true),
	}

	// Item 6: Quit
	items[6] = map[string]dbus.Variant{
		"label":   dbus.MakeVariant("Quit"),
		"enabled": dbus.MakeVariant(true),
		"visible": dbus.MakeVariant(true),
	}

	return items
}

// GetLayout returns the menu layout tree
func (st *SystemTray) GetLayout(parentId int32, recursionDepth int32, propertyNames []string) (uint32, menuLayoutItem, *dbus.Error) {
	st.mu.RLock()
	connected := st.connected
	revision := st.menuRevision
	st.mu.RUnlock()

	log.Printf("GetLayout called: parentId=%d, depth=%d, props=%v", parentId, recursionDepth, propertyNames)

	// Build menu items as children of root
	var children []dbus.Variant

	// Item 1: Connect or Disconnect
	connectLabel := "Connect"
	if connected {
		connectLabel = "Disconnect"
	}
	children = append(children, makeMenuItem(1, map[string]dbus.Variant{
		"label":   dbus.MakeVariant(connectLabel),
		"enabled": dbus.MakeVariant(true),
		"visible": dbus.MakeVariant(true),
	}))

	// Item 2: Separator
	children = append(children, makeMenuItem(2, map[string]dbus.Variant{
		"type":    dbus.MakeVariant("separator"),
		"visible": dbus.MakeVariant(true),
	}))

	// Item 3: Connection Info
	children = append(children, makeMenuItem(3, map[string]dbus.Variant{
		"label":   dbus.MakeVariant("Connection Info..."),
		"enabled": dbus.MakeVariant(true),
		"visible": dbus.MakeVariant(true),
	}))

	// Item 4: Status (disabled, informational)
	statusText := "Status: Disconnected"
	if connected {
		statusText = "Status: Connected"
	}
	children = append(children, makeMenuItem(4, map[string]dbus.Variant{
		"label":   dbus.MakeVariant(statusText),
		"enabled": dbus.MakeVariant(false),
		"visible": dbus.MakeVariant(true),
	}))

	// Item 5: Separator
	children = append(children, makeMenuItem(5, map[string]dbus.Variant{
		"type":    dbus.MakeVariant("separator"),
		"visible": dbus.MakeVariant(true),
	}))

	// Item 6: Quit
	children = append(children, makeMenuItem(6, map[string]dbus.Variant{
		"label":   dbus.MakeVariant("Quit"),
		"enabled": dbus.MakeVariant(true),
		"visible": dbus.MakeVariant(true),
	}))

	// Root menu item
	root := menuLayoutItem{
		ID: 0,
		Properties: map[string]dbus.Variant{
			"children-display": dbus.MakeVariant("submenu"),
		},
		Children: children,
	}

	return revision, root, nil
}

// Event handles menu item clicks
func (st *SystemTray) Event(id int32, eventId string, data dbus.Variant, timestamp uint32) *dbus.Error {
	log.Printf("Menu event: id=%d, event=%s", id, eventId)

	if eventId != "clicked" {
		return nil
	}

	switch id {
	case 1: // Connect/Disconnect
		st.mu.RLock()
		connected := st.connected
		st.mu.RUnlock()

		if connected {
			log.Println("Menu: Disconnect clicked")
			go st.onDisconnect()
		} else {
			log.Println("Menu: Connect clicked")
			go st.onConnect()
		}
	case 3: // Connection Info
		log.Println("Menu: Connection Info clicked")
		go func() {
			info := gatherConnectionInfo()
			showStatusDialog(info)
		}()
	case 6: // Quit
		log.Println("Menu: Quit clicked")
		go st.onQuit()
	}

	return nil
}

// EventGroup handles batch menu events
func (st *SystemTray) EventGroup(events []struct {
	ID        int32
	EventID   string
	Data      dbus.Variant
	Timestamp uint32
}) ([]int32, *dbus.Error) {
	for _, ev := range events {
		st.Event(ev.ID, ev.EventID, ev.Data, ev.Timestamp)
	}
	return []int32{}, nil
}

func (st *SystemTray) Version() (uint32, *dbus.Error) {
	return 3, nil
}

func (st *SystemTray) AboutToShow(id int32) (bool, *dbus.Error) {
	return true, nil
}

func (st *SystemTray) AboutToShowGroup(ids []int32) ([]int32, []int32, *dbus.Error) {
	return ids, []int32{}, nil
}

func (st *SystemTray) GetProperty(id int32, property string) (dbus.Variant, *dbus.Error) {
	items := st.getMenuItems()
	if props, ok := items[id]; ok {
		if val, ok := props[property]; ok {
			return val, nil
		}
	}
	return dbus.MakeVariant(""), nil
}

// groupPropertyItem represents a menu item with its properties for GetGroupProperties
type groupPropertyItem struct {
	ID         int32
	Properties map[string]dbus.Variant
}

func (st *SystemTray) GetGroupProperties(ids []int32, propertyNames []string) ([]groupPropertyItem, *dbus.Error) {
	log.Printf("GetGroupProperties called: ids=%v, properties=%v", ids, propertyNames)
	items := st.getMenuItems()
	var result []groupPropertyItem

	for _, id := range ids {
		if props, ok := items[id]; ok {
			// Filter properties if propertyNames is specified
			if len(propertyNames) > 0 {
				filtered := make(map[string]dbus.Variant)
				for _, name := range propertyNames {
					if val, ok := props[name]; ok {
						filtered[name] = val
					}
				}
				result = append(result, groupPropertyItem{ID: id, Properties: filtered})
			} else {
				result = append(result, groupPropertyItem{ID: id, Properties: props})
			}
		}
	}

	if result == nil {
		result = []groupPropertyItem{}
	}
	return result, nil
}

// --- Properties handler ---

func (st *SystemTray) getIconPixmapProp() interface{} {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return []iconPixmap{
		{
			Width:  st.iconWidth,
			Height: st.iconHeight,
			Data:   st.iconData,
		},
	}
}

func (st *SystemTray) getToolTipProp() interface{} {
	tooltip, _ := st.ToolTip()
	return tooltip
}

// prop defines a D-Bus property with optional dynamic getter
type prop struct {
	Value  interface{}
	Getter func() interface{}
}

type propsHandler struct {
	props map[string]map[string]*prop
}

func newPropsHandler(props map[string]map[string]*prop) *propsHandler {
	return &propsHandler{props: props}
}

func (p *propsHandler) Get(iface string, property string) (dbus.Variant, *dbus.Error) {
	ifaceProps, ok := p.props[iface]
	if !ok {
		return dbus.Variant{}, dbus.MakeFailedError(fmt.Errorf("interface %s not found", iface))
	}

	prop, ok := ifaceProps[property]
	if !ok {
		return dbus.Variant{}, dbus.MakeFailedError(fmt.Errorf("property %s not found", property))
	}

	var value interface{}
	if prop.Getter != nil {
		value = prop.Getter()
	} else {
		value = prop.Value
	}

	return dbus.MakeVariant(value), nil
}

func (p *propsHandler) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	ifaceProps, ok := p.props[iface]
	if !ok {
		return nil, dbus.MakeFailedError(fmt.Errorf("interface %s not found", iface))
	}

	result := make(map[string]dbus.Variant)
	for name, prop := range ifaceProps {
		var value interface{}
		if prop.Getter != nil {
			value = prop.Getter()
		} else {
			value = prop.Value
		}
		result[name] = dbus.MakeVariant(value)
	}

	return result, nil
}

func (p *propsHandler) Set(iface string, property string, value dbus.Variant) *dbus.Error {
	return dbus.MakeFailedError(fmt.Errorf("property %s is not writable", property))
}
