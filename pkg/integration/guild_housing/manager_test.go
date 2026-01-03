package guild_housing

import (
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world/housing"
)

func TestPermissionString(t *testing.T) {
	tests := []struct {
		name string
		perm Permission
		want string
	}{
		{"none", PermissionNone, "None"},
		{"view", PermissionView, "View"},
		{"use", PermissionUse, "Use"},
		{"manage", PermissionManage, "Manage"},
		{"admin", PermissionAdmin, "Admin"},
		{"unknown", Permission(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.perm.String(); got != tt.want {
				t.Errorf("Permission.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpgradeTierString(t *testing.T) {
	tests := []struct {
		name string
		tier UpgradeTier
		want string
	}{
		{"basic", TierBasic, "Basic"},
		{"standard", TierStandard, "Standard"},
		{"advanced", TierAdvanced, "Advanced"},
		{"master", TierMaster, "Master"},
		{"unknown", UpgradeTier(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.want {
				t.Errorf("UpgradeTier.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpgradeTierCost(t *testing.T) {
	tests := []struct {
		name string
		tier UpgradeTier
		want int
	}{
		{"basic", TierBasic, 0},
		{"standard", TierStandard, 10000},
		{"advanced", TierAdvanced, 50000},
		{"master", TierMaster, 100000},
		{"unknown", UpgradeTier(99), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.Cost(); got != tt.want {
				t.Errorf("UpgradeTier.Cost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpgradeTierBonusMultiplier(t *testing.T) {
	tests := []struct {
		name string
		tier UpgradeTier
		want float64
	}{
		{"basic", TierBasic, 1.0},
		{"standard", TierStandard, 1.2},
		{"advanced", TierAdvanced, 1.5},
		{"master", TierMaster, 2.0},
		{"unknown", UpgradeTier(99), 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.BonusMultiplier(); got != tt.want {
				t.Errorf("UpgradeTier.BonusMultiplier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionTypeString(t *testing.T) {
	tests := []struct {
		name string
		txn  TransactionType
		want string
	}{
		{"deposit", TransactionDeposit, "Deposit"},
		{"withdraw", TransactionWithdraw, "Withdraw"},
		{"unknown", TransactionType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.txn.String(); got != tt.want {
				t.Errorf("TransactionType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGuildHousingComponentType(t *testing.T) {
	comp := GuildHousingComponent{
		GuildID:    "guild-001",
		Permission: PermissionUse,
		Tier:       TierStandard,
	}

	if got := comp.Type(); got != "guild_housing" {
		t.Errorf("GuildHousingComponent.Type() = %v, want %v", got, "guild_housing")
	}
}

func TestCreateGuildHouse(t *testing.T) {
	manager := NewManager()
	size := housing.SizeLarge // 24x24

	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	if house.GuildID != "guild-001" {
		t.Errorf("GuildID = %v, want %v", house.GuildID, "guild-001")
	}
	if house.OwnerID != "player-001" {
		t.Errorf("OwnerID = %v, want %v", house.OwnerID, "player-001")
	}
	if house.Size != size {
		t.Errorf("Size = %v, want %v", house.Size, size)
	}
	if house.Tier != TierBasic {
		t.Errorf("Tier = %v, want %v", house.Tier, TierBasic)
	}

	if house.Permissions[guild.RankLeader] != PermissionAdmin {
		t.Errorf("Leader permission = %v, want %v", house.Permissions[guild.RankLeader], PermissionAdmin)
	}
	if house.Permissions[guild.RankOfficer] != PermissionManage {
		t.Errorf("Officer permission = %v, want %v", house.Permissions[guild.RankOfficer], PermissionManage)
	}
	if house.Permissions[guild.RankMember] != PermissionUse {
		t.Errorf("Member permission = %v, want %v", house.Permissions[guild.RankMember], PermissionUse)
	}
	if house.Permissions[guild.RankRecruit] != PermissionView {
		t.Errorf("Recruit permission = %v, want %v", house.Permissions[guild.RankRecruit], PermissionView)
	}
}

func TestGetGuildHouse(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium
	created := manager.CreateGuildHouse("guild-001", "player-001", size)

	retrieved, err := manager.GetGuildHouse(created.HouseID)
	if err != nil {
		t.Fatalf("GetGuildHouse() error = %v", err)
	}
	if retrieved.HouseID != created.HouseID {
		t.Errorf("HouseID = %v, want %v", retrieved.HouseID, created.HouseID)
	}

	_, err = manager.GetGuildHouse("nonexistent")
	if err == nil {
		t.Error("GetGuildHouse() expected error for nonexistent house")
	}
}

func TestGetGuildHouses(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium

	manager.CreateGuildHouse("guild-001", "player-001", size)
	manager.CreateGuildHouse("guild-001", "player-002", size)
	manager.CreateGuildHouse("guild-002", "player-003", size)

	houses := manager.GetGuildHouses("guild-001")
	if len(houses) != 2 {
		t.Errorf("GetGuildHouses() count = %v, want 2", len(houses))
	}

	houses = manager.GetGuildHouses("guild-002")
	if len(houses) != 1 {
		t.Errorf("GetGuildHouses() count = %v, want 1", len(houses))
	}
}

func TestSetPermission(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	err := manager.SetPermission(house.HouseID, guild.RankMember, PermissionManage)
	if err != nil {
		t.Fatalf("SetPermission() error = %v", err)
	}

	retrieved, _ := manager.GetGuildHouse(house.HouseID)
	if retrieved.Permissions[guild.RankMember] != PermissionManage {
		t.Errorf("Permission = %v, want %v", retrieved.Permissions[guild.RankMember], PermissionManage)
	}

	err = manager.SetPermission("nonexistent", guild.RankMember, PermissionManage)
	if err == nil {
		t.Error("SetPermission() expected error for nonexistent house")
	}
}

func TestCheckPermission(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	tests := []struct {
		name     string
		rank     guild.Rank
		required Permission
		want     bool
	}{
		{"leader_admin", guild.RankLeader, PermissionAdmin, true},
		{"leader_manage", guild.RankLeader, PermissionManage, true},
		{"officer_manage", guild.RankOfficer, PermissionManage, true},
		{"officer_admin", guild.RankOfficer, PermissionAdmin, false},
		{"member_use", guild.RankMember, PermissionUse, true},
		{"member_manage", guild.RankMember, PermissionManage, false},
		{"recruit_view", guild.RankRecruit, PermissionView, true},
		{"recruit_use", guild.RankRecruit, PermissionUse, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := manager.CheckPermission(house.HouseID, tt.rank, tt.required); got != tt.want {
				t.Errorf("CheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}

	if manager.CheckPermission("nonexistent", guild.RankLeader, PermissionAdmin) {
		t.Error("CheckPermission() should return false for nonexistent house")
	}
}

func TestAddCraftingStation(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	err := manager.AddCraftingStation(house.HouseID, "station-001")
	if err != nil {
		t.Fatalf("AddCraftingStation() error = %v", err)
	}

	err = manager.AddCraftingStation(house.HouseID, "station-002")
	if err != nil {
		t.Fatalf("AddCraftingStation() error = %v", err)
	}

	stations, err := manager.GetCraftingStations(house.HouseID)
	if err != nil {
		t.Fatalf("GetCraftingStations() error = %v", err)
	}
	if len(stations) != 2 {
		t.Errorf("station count = %v, want 2", len(stations))
	}

	err = manager.AddCraftingStation("nonexistent", "station-003")
	if err == nil {
		t.Error("AddCraftingStation() expected error for nonexistent house")
	}
}

func TestGetCraftingStations(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	stations, err := manager.GetCraftingStations(house.HouseID)
	if err != nil {
		t.Fatalf("GetCraftingStations() error = %v", err)
	}
	if len(stations) != 0 {
		t.Errorf("initial station count = %v, want 0", len(stations))
	}

	_, err = manager.GetCraftingStations("nonexistent")
	if err == nil {
		t.Error("GetCraftingStations() expected error for nonexistent house")
	}
}

func TestCreateGuildStorage(t *testing.T) {
	manager := NewManager()
	capacity := 500

	storage := manager.CreateGuildStorage("guild-001", capacity)

	if storage.GuildID != "guild-001" {
		t.Errorf("GuildID = %v, want %v", storage.GuildID, "guild-001")
	}
	if storage.Capacity != capacity {
		t.Errorf("Capacity = %v, want %v", storage.Capacity, capacity)
	}
	if len(storage.Items) != 0 {
		t.Errorf("initial items count = %v, want 0", len(storage.Items))
	}
}

func TestGetGuildStorage(t *testing.T) {
	manager := NewManager()
	created := manager.CreateGuildStorage("guild-001", 500)

	retrieved, err := manager.GetGuildStorage(created.StorageID)
	if err != nil {
		t.Fatalf("GetGuildStorage() error = %v", err)
	}
	if retrieved.StorageID != created.StorageID {
		t.Errorf("StorageID = %v, want %v", retrieved.StorageID, created.StorageID)
	}

	_, err = manager.GetGuildStorage("nonexistent")
	if err == nil {
		t.Error("GetGuildStorage() expected error for nonexistent storage")
	}
}

func TestDepositItem(t *testing.T) {
	manager := NewManager()
	storage := manager.CreateGuildStorage("guild-001", 500)

	err := manager.DepositItem(storage.StorageID, "player-001", "item-001", 10)
	if err != nil {
		t.Fatalf("DepositItem() error = %v", err)
	}

	items, _ := manager.GetStorageItems(storage.StorageID)
	if len(items) != 1 {
		t.Errorf("item count = %v, want 1", len(items))
	}
	if items["item-001"].Quantity != 10 {
		t.Errorf("item quantity = %v, want 10", items["item-001"].Quantity)
	}

	err = manager.DepositItem(storage.StorageID, "player-002", "item-001", 5)
	if err != nil {
		t.Fatalf("DepositItem() error = %v", err)
	}

	items, _ = manager.GetStorageItems(storage.StorageID)
	if items["item-001"].Quantity != 15 {
		t.Errorf("item quantity = %v, want 15", items["item-001"].Quantity)
	}

	err = manager.DepositItem("nonexistent", "player-001", "item-001", 1)
	if err == nil {
		t.Error("DepositItem() expected error for nonexistent storage")
	}
}

func TestWithdrawItem(t *testing.T) {
	manager := NewManager()
	storage := manager.CreateGuildStorage("guild-001", 500)

	manager.DepositItem(storage.StorageID, "player-001", "item-001", 10)

	withdrawn, err := manager.WithdrawItem(storage.StorageID, "player-002", "item-001", 5)
	if err != nil {
		t.Fatalf("WithdrawItem() error = %v", err)
	}
	if withdrawn != 5 {
		t.Errorf("withdrawn = %v, want 5", withdrawn)
	}

	items, _ := manager.GetStorageItems(storage.StorageID)
	if items["item-001"].Quantity != 5 {
		t.Errorf("remaining quantity = %v, want 5", items["item-001"].Quantity)
	}

	withdrawn, err = manager.WithdrawItem(storage.StorageID, "player-003", "item-001", 10)
	if err != nil {
		t.Fatalf("WithdrawItem() error = %v", err)
	}
	if withdrawn != 5 {
		t.Errorf("withdrawn = %v, want 5 (partial)", withdrawn)
	}

	items, _ = manager.GetStorageItems(storage.StorageID)
	if _, exists := items["item-001"]; exists {
		t.Error("item should be removed when quantity reaches 0")
	}

	_, err = manager.WithdrawItem("nonexistent", "player-001", "item-001", 1)
	if err == nil {
		t.Error("WithdrawItem() expected error for nonexistent storage")
	}

	_, err = manager.WithdrawItem(storage.StorageID, "player-001", "nonexistent", 1)
	if err == nil {
		t.Error("WithdrawItem() expected error for nonexistent item")
	}
}

func TestGetStorageItems(t *testing.T) {
	manager := NewManager()
	storage := manager.CreateGuildStorage("guild-001", 500)

	manager.DepositItem(storage.StorageID, "player-001", "item-001", 10)
	manager.DepositItem(storage.StorageID, "player-002", "item-002", 20)

	items, err := manager.GetStorageItems(storage.StorageID)
	if err != nil {
		t.Fatalf("GetStorageItems() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("item count = %v, want 2", len(items))
	}

	_, err = manager.GetStorageItems("nonexistent")
	if err == nil {
		t.Error("GetStorageItems() expected error for nonexistent storage")
	}
}

func TestGetTransactions(t *testing.T) {
	manager := NewManager()
	storage := manager.CreateGuildStorage("guild-001", 500)

	manager.DepositItem(storage.StorageID, "player-001", "item-001", 10)
	manager.WithdrawItem(storage.StorageID, "player-002", "item-001", 5)

	transactions, err := manager.GetTransactions(storage.StorageID)
	if err != nil {
		t.Fatalf("GetTransactions() error = %v", err)
	}
	if len(transactions) != 2 {
		t.Errorf("transaction count = %v, want 2", len(transactions))
	}
	if transactions[0].Action != TransactionDeposit {
		t.Errorf("transaction[0] action = %v, want %v", transactions[0].Action, TransactionDeposit)
	}
	if transactions[1].Action != TransactionWithdraw {
		t.Errorf("transaction[1] action = %v, want %v", transactions[1].Action, TransactionWithdraw)
	}

	_, err = manager.GetTransactions("nonexistent")
	if err == nil {
		t.Error("GetTransactions() expected error for nonexistent storage")
	}
}

func TestUpgradeHouse(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	err := manager.UpgradeHouse(house.HouseID, 10000)
	if err != nil {
		t.Fatalf("UpgradeHouse() error = %v", err)
	}

	retrieved, _ := manager.GetGuildHouse(house.HouseID)
	if retrieved.Tier != TierStandard {
		t.Errorf("Tier = %v, want %v", retrieved.Tier, TierStandard)
	}

	err = manager.UpgradeHouse(house.HouseID, 50000)
	if err != nil {
		t.Fatalf("UpgradeHouse() error = %v", err)
	}

	retrieved, _ = manager.GetGuildHouse(house.HouseID)
	if retrieved.Tier != TierAdvanced {
		t.Errorf("Tier = %v, want %v", retrieved.Tier, TierAdvanced)
	}

	err = manager.UpgradeHouse(house.HouseID, 5000)
	if err == nil {
		t.Error("UpgradeHouse() expected error for insufficient gold")
	}

	err = manager.UpgradeHouse(house.HouseID, 100000)
	if err != nil {
		t.Fatalf("UpgradeHouse() error = %v", err)
	}

	err = manager.UpgradeHouse(house.HouseID, 100000)
	if err == nil {
		t.Error("UpgradeHouse() expected error for max tier")
	}

	err = manager.UpgradeHouse("nonexistent", 10000)
	if err == nil {
		t.Error("UpgradeHouse() expected error for nonexistent house")
	}
}

func TestGetUpgradeBonus(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	bonus, err := manager.GetUpgradeBonus(house.HouseID)
	if err != nil {
		t.Fatalf("GetUpgradeBonus() error = %v", err)
	}
	if bonus != 1.0 {
		t.Errorf("bonus = %v, want 1.0", bonus)
	}

	manager.UpgradeHouse(house.HouseID, 10000)
	bonus, _ = manager.GetUpgradeBonus(house.HouseID)
	if bonus != 1.2 {
		t.Errorf("bonus = %v, want 1.2", bonus)
	}

	_, err = manager.GetUpgradeBonus("nonexistent")
	if err == nil {
		t.Error("GetUpgradeBonus() expected error for nonexistent house")
	}
}

func TestCreateMeetingHall(t *testing.T) {
	manager := NewManager()
	hall := manager.CreateMeetingHall("guild-001", 50)

	if hall.GuildID != "guild-001" {
		t.Errorf("GuildID = %v, want %v", hall.GuildID, "guild-001")
	}
	if hall.MaxCapacity != 50 {
		t.Errorf("MaxCapacity = %v, want 50", hall.MaxCapacity)
	}
	if hall.ChatRadius != 150.0 {
		t.Errorf("ChatRadius = %v, want 150.0", hall.ChatRadius)
	}
	if len(hall.Members) != 0 {
		t.Errorf("initial members count = %v, want 0", len(hall.Members))
	}
}

func TestAddMemberToHall(t *testing.T) {
	manager := NewManager()
	hall := manager.CreateMeetingHall("guild-001", 2)

	err := manager.AddMemberToHall(hall, "player-001")
	if err != nil {
		t.Fatalf("AddMemberToHall() error = %v", err)
	}
	if len(hall.Members) != 1 {
		t.Errorf("member count = %v, want 1", len(hall.Members))
	}

	err = manager.AddMemberToHall(hall, "player-001")
	if err == nil {
		t.Error("AddMemberToHall() expected error for duplicate member")
	}

	err = manager.AddMemberToHall(hall, "player-002")
	if err != nil {
		t.Fatalf("AddMemberToHall() error = %v", err)
	}

	err = manager.AddMemberToHall(hall, "player-003")
	if err == nil {
		t.Error("AddMemberToHall() expected error for capacity exceeded")
	}
}

func TestRemoveMemberFromHall(t *testing.T) {
	manager := NewManager()
	hall := manager.CreateMeetingHall("guild-001", 10)

	manager.AddMemberToHall(hall, "player-001")
	manager.AddMemberToHall(hall, "player-002")

	manager.RemoveMemberFromHall(hall, "player-001")
	if len(hall.Members) != 1 {
		t.Errorf("member count = %v, want 1", len(hall.Members))
	}
	if hall.Members[0] != "player-002" {
		t.Errorf("remaining member = %v, want player-002", hall.Members[0])
	}

	manager.RemoveMemberFromHall(hall, "nonexistent")
	if len(hall.Members) != 1 {
		t.Error("RemoveMemberFromHall() should not modify hall for nonexistent member")
	}
}

func TestSaveLoad(t *testing.T) {
	manager := NewManager()
	size := housing.SizeMedium

	house1 := manager.CreateGuildHouse("guild-001", "player-001", size)
	house2 := manager.CreateGuildHouse("guild-002", "player-002", size)
	storage := manager.CreateGuildStorage("guild-001", 500)
	manager.DepositItem(storage.StorageID, "player-001", "item-001", 10)

	data, err := manager.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	newManager := NewManager()
	err = newManager.Load(data)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	loadedHouse1, err := newManager.GetGuildHouse(house1.HouseID)
	if err != nil {
		t.Fatalf("GetGuildHouse() error = %v", err)
	}
	if loadedHouse1.GuildID != "guild-001" {
		t.Errorf("loaded GuildID = %v, want %v", loadedHouse1.GuildID, "guild-001")
	}

	loadedHouse2, err := newManager.GetGuildHouse(house2.HouseID)
	if err != nil {
		t.Fatalf("GetGuildHouse() error = %v", err)
	}
	if loadedHouse2.GuildID != "guild-002" {
		t.Errorf("loaded GuildID = %v, want %v", loadedHouse2.GuildID, "guild-002")
	}

	loadedStorage, err := newManager.GetGuildStorage(storage.StorageID)
	if err != nil {
		t.Fatalf("GetGuildStorage() error = %v", err)
	}
	if loadedStorage.GuildID != "guild-001" {
		t.Errorf("loaded storage GuildID = %v, want %v", loadedStorage.GuildID, "guild-001")
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "invalid JSON",
			data:    []byte(`{invalid json`),
			wantErr: true,
			errMsg:  "",
		},
		{
			name:    "empty JSON object",
			data:    []byte(`{}`),
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "valid houses and storage",
			data:    []byte(`{"houses":{},"storage":{}}`),
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "houses with invalid structure for unmarshal",
			data:    []byte(`{"houses":{"test":"invalid"}}`),
			wantErr: true,
			errMsg:  "failed to unmarshal houses",
		},
		{
			name:    "storage with invalid structure for unmarshal",
			data:    []byte(`{"houses":{},"storage":{"test":"invalid"}}`),
			wantErr: true,
			errMsg:  "failed to unmarshal storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()
			err := manager.Load(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

// Benchmarks
func BenchmarkCreateGuildHouse(b *testing.B) {
	manager := NewManager()
	size := housing.SizeMedium
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateGuildHouse("guild-001", "player-001", size)
	}
}

func BenchmarkCheckPermission(b *testing.B) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckPermission(house.HouseID, guild.RankMember, PermissionUse)
	}
}

func BenchmarkDepositItem(b *testing.B) {
	manager := NewManager()
	storage := manager.CreateGuildStorage("guild-001", 10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.DepositItem(storage.StorageID, "player-001", "item-001", 1)
	}
}

func BenchmarkWithdrawItem(b *testing.B) {
	manager := NewManager()
	storage := manager.CreateGuildStorage("guild-001", 500)
	manager.DepositItem(storage.StorageID, "player-001", "item-001", 1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.WithdrawItem(storage.StorageID, "player-002", "item-001", 1)
	}
}

func BenchmarkGetUpgradeBonus(b *testing.B) {
	manager := NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)
	manager.UpgradeHouse(house.HouseID, 50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetUpgradeBonus(house.HouseID)
	}
}

func BenchmarkAddMemberToHall(b *testing.B) {
	manager := NewManager()
	hall := manager.CreateMeetingHall("guild-001", 100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.AddMemberToHall(hall, "player-001")
		hall.Members = hall.Members[:len(hall.Members)-1]
	}
}
