package guild_housing

// permissions.go defines access control levels for guild housing.
// This includes permission types and validation logic for rank-based access.
//
// Code relocated from: types.go

import "github.com/opd-ai/venture/pkg/network/federation/guild"

// Permission represents access levels for guild houses.
type Permission int

const (
	PermissionNone   Permission = iota // No access
	PermissionView                     // Can enter and view
	PermissionUse                      // Can use crafting stations
	PermissionManage                   // Can place/remove furniture
	PermissionAdmin                    // Full control
)

// String returns human-readable permission name.
func (p Permission) String() string {
	switch p {
	case PermissionNone:
		return "None"
	case PermissionView:
		return "View"
	case PermissionUse:
		return "Use"
	case PermissionManage:
		return "Manage"
	case PermissionAdmin:
		return "Admin"
	default:
		return "Unknown"
	}
}

// DefaultPermissions returns standard rank-based permissions for guild houses.
func DefaultPermissions() map[guild.Rank]Permission {
	return map[guild.Rank]Permission{
		guild.RankLeader:  PermissionAdmin,
		guild.RankOfficer: PermissionManage,
		guild.RankMember:  PermissionUse,
		guild.RankRecruit: PermissionView,
	}
}
