// Package engine provides the VR UI component for VR support.

package engine

import (
	"encoding/json"
	"math"
	"sync"
)

// VRUIComponent tracks VR UI panel state for an entity.
// It manages panel position in 3D space, gaze interaction, and comfort settings.
type VRUIComponent struct {
	mu sync.RWMutex

	// Enabled controls whether VR UI is active
	Enabled bool `json:"enabled"`

	// Panels is the collection of UI panels
	Panels map[string]*VRUIPanel `json:"panels"`

	// GazeTarget is the currently gazed-at panel (empty if none)
	GazeTarget string `json:"gaze_target"`

	// GazeHoverTime tracks how long the user has been looking at the target
	GazeHoverTime float64 `json:"gaze_hover_time"`

	// GazeActivationTime is the time required to activate by gaze
	GazeActivationTime float64 `json:"gaze_activation_time"`

	// ComfortVignetteEnabled enables vignette during movement
	ComfortVignetteEnabled bool `json:"comfort_vignette_enabled"`

	// ComfortVignetteIntensity is the current vignette intensity (0-1)
	ComfortVignetteIntensity float64 `json:"comfort_vignette_intensity"`

	// ComfortVignetteSpeed controls how fast vignette appears/fades
	ComfortVignetteSpeed float64 `json:"comfort_vignette_speed"`

	// UIScale is the global scale factor for all panels
	UIScale float64 `json:"ui_scale"`
}

// VRUIPanel represents a floating UI panel in 3D space.
type VRUIPanel struct {
	// ID is the unique panel identifier
	ID string `json:"id"`

	// PanelType categorizes the panel (health, inventory, minimap, menu, etc.)
	PanelType string `json:"panel_type"`

	// WorldX is the X position in world space
	WorldX float64 `json:"world_x"`

	// WorldY is the Y position in world space
	WorldY float64 `json:"world_y"`

	// WorldZ is the Z position (distance from camera)
	WorldZ float64 `json:"world_z"`

	// Width is the panel width in world units
	Width float64 `json:"width"`

	// Height is the panel height in world units
	Height float64 `json:"height"`

	// FollowHead controls if panel follows head movement
	FollowHead bool `json:"follow_head"`

	// FollowDistance maintains distance when following
	FollowDistance float64 `json:"follow_distance"`

	// FollowOffsetX is the horizontal offset from head when FollowHead is true
	FollowOffsetX float64 `json:"follow_offset_x"`

	// FollowOffsetY is the vertical offset from head when FollowHead is true
	FollowOffsetY float64 `json:"follow_offset_y"`

	// Opacity is the panel opacity (0-1)
	Opacity float64 `json:"opacity"`

	// Visible controls if panel is rendered
	Visible bool `json:"visible"`

	// GazeHighlighted is true when being looked at
	GazeHighlighted bool `json:"gaze_highlighted"`

	// Interactive controls if panel responds to gaze
	Interactive bool `json:"interactive"`

	// LockedToHand anchors panel to a controller hand
	LockedToHand string `json:"locked_to_hand"`
}

// Panel type constants
const (
	PanelTypeHealth       = "health"
	PanelTypeInventory    = "inventory"
	PanelTypeMinimap      = "minimap"
	PanelTypeMenu         = "menu"
	PanelTypeQuest        = "quest"
	PanelTypeChat         = "chat"
	PanelTypeNotification = "notification"
)

// DefaultGazeActivationTime is the default time to activate by gaze
const DefaultGazeActivationTime = 1.5

// DefaultUIScale is the default UI scale factor
const DefaultUIScale = 1.0

// DefaultPanelDistance is the default distance from camera
const DefaultPanelDistance = 2.0

// NewVRUIComponent creates a new VR UI component with defaults.
func NewVRUIComponent() *VRUIComponent {
	return &VRUIComponent{
		Enabled:                false,
		Panels:                 make(map[string]*VRUIPanel),
		GazeActivationTime:     DefaultGazeActivationTime,
		ComfortVignetteEnabled: true,
		ComfortVignetteSpeed:   3.0,
		UIScale:                DefaultUIScale,
	}
}

// Type returns the component type identifier.
func (c *VRUIComponent) Type() string {
	return "vr_ui"
}

// SetEnabled enables or disables VR UI.
func (c *VRUIComponent) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Enabled = enabled
}

// IsEnabled returns whether VR UI is enabled.
func (c *VRUIComponent) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Enabled
}

// AddPanel adds a UI panel with the given settings.
func (c *VRUIComponent) AddPanel(id, panelType string, x, y, z, width, height float64) *VRUIPanel {
	c.mu.Lock()
	defer c.mu.Unlock()

	panel := &VRUIPanel{
		ID:             id,
		PanelType:      panelType,
		WorldX:         x,
		WorldY:         y,
		WorldZ:         z,
		Width:          width,
		Height:         height,
		FollowHead:     false,
		FollowDistance: DefaultPanelDistance,
		Opacity:        1.0,
		Visible:        true,
		Interactive:    true,
	}

	c.Panels[id] = panel
	return panel
}

// RemovePanel removes a panel by ID.
func (c *VRUIComponent) RemovePanel(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Panels, id)
}

// GetPanel returns a panel by ID.
func (c *VRUIComponent) GetPanel(id string) *VRUIPanel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Panels[id]
}

// GetAllPanels returns all panels.
func (c *VRUIComponent) GetAllPanels() []*VRUIPanel {
	c.mu.RLock()
	defer c.mu.RUnlock()

	panels := make([]*VRUIPanel, 0, len(c.Panels))
	for _, p := range c.Panels {
		panels = append(panels, p)
	}
	return panels
}

// SetPanelPosition sets the 3D position of a panel.
func (c *VRUIComponent) SetPanelPosition(id string, x, y, z float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if panel, ok := c.Panels[id]; ok {
		panel.WorldX = x
		panel.WorldY = y
		panel.WorldZ = z
	}
}

// SetPanelVisible sets the visibility of a panel.
func (c *VRUIComponent) SetPanelVisible(id string, visible bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if panel, ok := c.Panels[id]; ok {
		panel.Visible = visible
	}
}

// SetPanelFollowHead sets whether a panel follows head movement.
func (c *VRUIComponent) SetPanelFollowHead(id string, follow bool, distance float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if panel, ok := c.Panels[id]; ok {
		panel.FollowHead = follow
		if distance > 0 {
			panel.FollowDistance = distance
		}
	}
}

// SetPanelLockedToHand locks a panel to a controller hand.
func (c *VRUIComponent) SetPanelLockedToHand(id, hand string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if panel, ok := c.Panels[id]; ok {
		panel.LockedToHand = hand
	}
}

// SetGazeTarget sets the currently gazed-at panel.
func (c *VRUIComponent) SetGazeTarget(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset previous target
	if c.GazeTarget != "" && c.GazeTarget != id {
		if prev, ok := c.Panels[c.GazeTarget]; ok {
			prev.GazeHighlighted = false
		}
		c.GazeHoverTime = 0
	}

	c.GazeTarget = id

	// Highlight new target
	if id != "" {
		if panel, ok := c.Panels[id]; ok {
			panel.GazeHighlighted = true
		}
	}
}

// GetGazeTarget returns the currently gazed-at panel.
func (c *VRUIComponent) GetGazeTarget() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GazeTarget
}

// UpdateGazeHover updates the gaze hover time.
func (c *VRUIComponent) UpdateGazeHover(deltaTime float64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.GazeTarget == "" {
		return false
	}

	c.GazeHoverTime += deltaTime

	// Check if activation time reached
	if c.GazeHoverTime >= c.GazeActivationTime {
		c.GazeHoverTime = 0
		return true
	}

	return false
}

// GetGazeProgress returns the progress toward gaze activation (0-1).
func (c *VRUIComponent) GetGazeProgress() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.GazeActivationTime <= 0 {
		return 1.0
	}
	progress := c.GazeHoverTime / c.GazeActivationTime
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// SetGazeActivationTime sets the time required to activate by gaze.
func (c *VRUIComponent) SetGazeActivationTime(time float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time < 0.1 {
		time = 0.1
	}
	c.GazeActivationTime = time
}

// SetComfortVignetteEnabled enables or disables comfort vignette.
func (c *VRUIComponent) SetComfortVignetteEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ComfortVignetteEnabled = enabled
}

// IsComfortVignetteEnabled returns whether comfort vignette is enabled.
func (c *VRUIComponent) IsComfortVignetteEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ComfortVignetteEnabled
}

// SetComfortVignetteIntensity sets the vignette intensity (0-1).
func (c *VRUIComponent) SetComfortVignetteIntensity(intensity float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1 {
		intensity = 1
	}
	c.ComfortVignetteIntensity = intensity
}

// GetComfortVignetteIntensity returns the current vignette intensity.
func (c *VRUIComponent) GetComfortVignetteIntensity() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ComfortVignetteIntensity
}

// UpdateComfortVignette updates vignette based on movement.
func (c *VRUIComponent) UpdateComfortVignette(isMoving bool, deltaTime float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.ComfortVignetteEnabled {
		c.ComfortVignetteIntensity = 0
		return
	}

	targetIntensity := 0.0
	if isMoving {
		targetIntensity = 0.3 // 30% vignette when moving
	}

	// Smoothly interpolate toward target
	diff := targetIntensity - c.ComfortVignetteIntensity
	c.ComfortVignetteIntensity += diff * c.ComfortVignetteSpeed * deltaTime

	// Clamp
	if c.ComfortVignetteIntensity < 0 {
		c.ComfortVignetteIntensity = 0
	}
	if c.ComfortVignetteIntensity > 1 {
		c.ComfortVignetteIntensity = 1
	}
}

// SetUIScale sets the global UI scale.
func (c *VRUIComponent) SetUIScale(scale float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if scale < 0.5 {
		scale = 0.5
	}
	if scale > 2.0 {
		scale = 2.0
	}
	c.UIScale = scale
}

// GetUIScale returns the global UI scale.
func (c *VRUIComponent) GetUIScale() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.UIScale
}

// CalculateGazeRayIntersection checks if a gaze ray hits a panel.
// gazeDir is the normalized gaze direction, gazeOrigin is the eye position.
// Returns the panel ID if hit, empty string otherwise.
func (c *VRUIComponent) CalculateGazeRayIntersection(gazeOriginX, gazeOriginY, gazeOriginZ, gazeDirX, gazeDirY, gazeDirZ float64) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	closestDist := math.MaxFloat64
	closestPanel := ""

	for id, panel := range c.Panels {
		if !panel.Visible || !panel.Interactive {
			continue
		}

		// Simple ray-plane intersection for panels facing camera (Z-aligned)
		// Panel is assumed to be perpendicular to Z axis at panel.WorldZ
		if gazeDirZ == 0 {
			continue
		}

		t := (panel.WorldZ - gazeOriginZ) / gazeDirZ
		if t < 0 {
			continue // Behind camera
		}

		// Calculate intersection point
		hitX := gazeOriginX + t*gazeDirX
		hitY := gazeOriginY + t*gazeDirY

		// Check if within panel bounds
		halfW := panel.Width * c.UIScale / 2
		halfH := panel.Height * c.UIScale / 2

		if hitX >= panel.WorldX-halfW && hitX <= panel.WorldX+halfW &&
			hitY >= panel.WorldY-halfH && hitY <= panel.WorldY+halfH {
			if t < closestDist {
				closestDist = t
				closestPanel = id
			}
		}
	}

	return closestPanel
}

// Serialize converts the component to JSON bytes.
func (c *VRUIComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize loads component state from JSON bytes.
func (c *VRUIComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}

// CreateDefaultPanels sets up the standard VR UI panel layout.
func (c *VRUIComponent) CreateDefaultPanels() {
	// Health bar - upper left, follows head
	health := c.AddPanel("health", PanelTypeHealth, -0.3, 0.2, DefaultPanelDistance, 0.2, 0.05)
	health.FollowHead = true

	// Minimap - upper right, follows head
	minimap := c.AddPanel("minimap", PanelTypeMinimap, 0.3, 0.2, DefaultPanelDistance, 0.15, 0.15)
	minimap.FollowHead = true

	// Inventory - locked to left wrist
	inv := c.AddPanel("inventory", PanelTypeInventory, 0, 0, 0.3, 0.3, 0.4)
	inv.LockedToHand = ControllerLeft
	inv.Visible = false // Hidden until activated

	// Menu - center, fixed
	menu := c.AddPanel("menu", PanelTypeMenu, 0, 0, DefaultPanelDistance, 0.8, 0.6)
	menu.Visible = false // Hidden until activated

	// Quest tracker - right side
	quest := c.AddPanel("quest", PanelTypeQuest, 0.4, 0, DefaultPanelDistance, 0.2, 0.3)
	quest.FollowHead = true

	// Notifications - top center
	notif := c.AddPanel("notifications", PanelTypeNotification, 0, 0.3, DefaultPanelDistance, 0.4, 0.1)
	notif.FollowHead = true
	notif.Interactive = false
}
