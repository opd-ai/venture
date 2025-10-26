package network

import "testing"

// TestComponentData_Structure verifies ComponentData struct initialization and fields.
func TestComponentData_Structure(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		data          []byte
	}{
		{"position_component", "position", []byte{1, 2, 3, 4}},
		{"velocity_component", "velocity", []byte{5, 6, 7, 8}},
		{"empty_data", "empty", []byte{}},
		{"nil_data", "nil", nil},
		{"large_data", "large", make([]byte, 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := ComponentData{
				Type: tt.componentType,
				Data: tt.data,
			}

			if cd.Type != tt.componentType {
				t.Errorf("Expected type %s, got %s", tt.componentType, cd.Type)
			}

			if tt.data == nil {
				if cd.Data != nil {
					t.Error("Expected nil data")
				}
			} else {
				if len(cd.Data) != len(tt.data) {
					t.Errorf("Expected data length %d, got %d", len(tt.data), len(cd.Data))
				}
			}
		})
	}
}

// TestStateUpdate_Structure verifies StateUpdate struct initialization.
func TestStateUpdate_Structure(t *testing.T) {
	components := []ComponentData{
		{Type: "position", Data: []byte{1, 2, 3}},
		{Type: "velocity", Data: []byte{4, 5, 6}},
	}

	update := StateUpdate{
		Timestamp:      1234567890,
		EntityID:       42,
		Components:     components,
		Priority:       128,
		SequenceNumber: 100,
	}

	if update.Timestamp != 1234567890 {
		t.Errorf("Expected timestamp 1234567890, got %d", update.Timestamp)
	}

	if update.EntityID != 42 {
		t.Errorf("Expected entity ID 42, got %d", update.EntityID)
	}

	if len(update.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(update.Components))
	}

	if update.Priority != 128 {
		t.Errorf("Expected priority 128, got %d", update.Priority)
	}

	if update.SequenceNumber != 100 {
		t.Errorf("Expected sequence number 100, got %d", update.SequenceNumber)
	}
}

// TestStateUpdate_PriorityLevels verifies priority value ranges.
func TestStateUpdate_PriorityLevels(t *testing.T) {
	tests := []struct {
		name     string
		priority uint8
	}{
		{"low_priority", 0},
		{"medium_priority", 128},
		{"high_priority", 200},
		{"critical_priority", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := StateUpdate{
				Priority: tt.priority,
			}

			if update.Priority != tt.priority {
				t.Errorf("Expected priority %d, got %d", tt.priority, update.Priority)
			}
		})
	}
}

// TestStateUpdate_EmptyComponents verifies behavior with no components.
func TestStateUpdate_EmptyComponents(t *testing.T) {
	update := StateUpdate{
		Timestamp:      100,
		EntityID:       1,
		Components:     []ComponentData{},
		Priority:       0,
		SequenceNumber: 0,
	}

	if len(update.Components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(update.Components))
	}
}

// TestStateUpdate_MultipleComponents verifies handling multiple components.
func TestStateUpdate_MultipleComponents(t *testing.T) {
	components := []ComponentData{
		{Type: "position", Data: []byte{1}},
		{Type: "velocity", Data: []byte{2}},
		{Type: "health", Data: []byte{3}},
		{Type: "sprite", Data: []byte{4}},
		{Type: "collider", Data: []byte{5}},
	}

	update := StateUpdate{
		EntityID:   1,
		Components: components,
	}

	if len(update.Components) != 5 {
		t.Errorf("Expected 5 components, got %d", len(update.Components))
	}

	// Verify component types
	expectedTypes := []string{"position", "velocity", "health", "sprite", "collider"}
	for i, comp := range update.Components {
		if comp.Type != expectedTypes[i] {
			t.Errorf("Component %d: expected type %s, got %s", i, expectedTypes[i], comp.Type)
		}
	}
}

// TestInputCommand_Structure verifies InputCommand struct initialization.
func TestInputCommand_Structure(t *testing.T) {
	cmd := InputCommand{
		PlayerID:       999,
		Timestamp:      1111111111,
		SequenceNumber: 50,
		InputType:      "move",
		Data:           []byte{10, 20, 30},
	}

	if cmd.PlayerID != 999 {
		t.Errorf("Expected player ID 999, got %d", cmd.PlayerID)
	}

	if cmd.Timestamp != 1111111111 {
		t.Errorf("Expected timestamp 1111111111, got %d", cmd.Timestamp)
	}

	if cmd.SequenceNumber != 50 {
		t.Errorf("Expected sequence number 50, got %d", cmd.SequenceNumber)
	}

	if cmd.InputType != "move" {
		t.Errorf("Expected input type 'move', got %s", cmd.InputType)
	}

	if len(cmd.Data) != 3 {
		t.Errorf("Expected data length 3, got %d", len(cmd.Data))
	}
}

// TestInputCommand_InputTypes verifies different input types.
func TestInputCommand_InputTypes(t *testing.T) {
	inputTypes := []string{
		"move",
		"attack",
		"use_item",
		"interact",
		"jump",
		"crouch",
		"inventory",
	}

	for _, inputType := range inputTypes {
		t.Run(inputType, func(t *testing.T) {
			cmd := InputCommand{
				PlayerID:  1,
				InputType: inputType,
			}

			if cmd.InputType != inputType {
				t.Errorf("Expected input type %s, got %s", inputType, cmd.InputType)
			}
		})
	}
}

// TestInputCommand_SequenceOrdering verifies sequence number ordering.
func TestInputCommand_SequenceOrdering(t *testing.T) {
	commands := []InputCommand{
		{SequenceNumber: 1},
		{SequenceNumber: 2},
		{SequenceNumber: 3},
		{SequenceNumber: 4},
		{SequenceNumber: 5},
	}

	for i, cmd := range commands {
		expectedSeq := uint32(i + 1)
		if cmd.SequenceNumber != expectedSeq {
			t.Errorf("Command %d: expected sequence %d, got %d", i, expectedSeq, cmd.SequenceNumber)
		}
	}
}

// TestConnectionInfo_Structure verifies ConnectionInfo struct initialization.
func TestConnectionInfo_Structure(t *testing.T) {
	conn := ConnectionInfo{
		PlayerID:  12345,
		Address:   "192.168.1.100:8080",
		Latency:   45.5,
		Connected: true,
	}

	if conn.PlayerID != 12345 {
		t.Errorf("Expected player ID 12345, got %d", conn.PlayerID)
	}

	if conn.Address != "192.168.1.100:8080" {
		t.Errorf("Expected address '192.168.1.100:8080', got %s", conn.Address)
	}

	if conn.Latency != 45.5 {
		t.Errorf("Expected latency 45.5, got %f", conn.Latency)
	}

	if !conn.Connected {
		t.Error("Expected connected to be true")
	}
}

// TestConnectionInfo_DisconnectedState verifies disconnected state.
func TestConnectionInfo_DisconnectedState(t *testing.T) {
	conn := ConnectionInfo{
		PlayerID:  1,
		Address:   "0.0.0.0:0",
		Latency:   0,
		Connected: false,
	}

	if conn.Connected {
		t.Error("Expected connected to be false")
	}

	if conn.Latency != 0 {
		t.Errorf("Expected latency 0 for disconnected, got %f", conn.Latency)
	}
}

// TestConnectionInfo_LatencyValues verifies various latency values.
func TestConnectionInfo_LatencyValues(t *testing.T) {
	tests := []struct {
		name    string
		latency float64
	}{
		{"excellent", 10.0},
		{"good", 50.0},
		{"moderate", 100.0},
		{"poor", 200.0},
		{"very_poor", 500.0},
		{"zero", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := ConnectionInfo{
				Latency: tt.latency,
			}

			if conn.Latency != tt.latency {
				t.Errorf("Expected latency %f, got %f", tt.latency, conn.Latency)
			}
		})
	}
}

// TestConnectionInfo_AddressFormats verifies different address formats.
func TestConnectionInfo_AddressFormats(t *testing.T) {
	addresses := []string{
		"127.0.0.1:8080",
		"192.168.1.1:9000",
		"10.0.0.1:3000",
		"example.com:8080",
		"[::1]:8080",
		"localhost:3000",
	}

	for _, addr := range addresses {
		t.Run(addr, func(t *testing.T) {
			conn := ConnectionInfo{
				Address: addr,
			}

			if conn.Address != addr {
				t.Errorf("Expected address %s, got %s", addr, conn.Address)
			}
		})
	}
}

// TestStateUpdate_SequenceNumberOverflow verifies large sequence numbers.
func TestStateUpdate_SequenceNumberOverflow(t *testing.T) {
	tests := []struct {
		name     string
		sequence uint32
	}{
		{"zero", 0},
		{"small", 100},
		{"medium", 10000},
		{"large", 1000000},
		{"max", 4294967295}, // uint32 max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := StateUpdate{
				SequenceNumber: tt.sequence,
			}

			if update.SequenceNumber != tt.sequence {
				t.Errorf("Expected sequence %d, got %d", tt.sequence, update.SequenceNumber)
			}
		})
	}
}

// TestInputCommand_EmptyData verifies behavior with empty data.
func TestInputCommand_EmptyData(t *testing.T) {
	cmd := InputCommand{
		PlayerID:  1,
		InputType: "ping",
		Data:      []byte{},
	}

	if len(cmd.Data) != 0 {
		t.Errorf("Expected empty data, got length %d", len(cmd.Data))
	}
}

// TestInputCommand_NilData verifies behavior with nil data.
func TestInputCommand_NilData(t *testing.T) {
	cmd := InputCommand{
		PlayerID:  1,
		InputType: "disconnect",
		Data:      nil,
	}

	if cmd.Data != nil {
		t.Error("Expected nil data")
	}
}

// TestStateUpdate_TimestampProgression verifies timestamp ordering.
func TestStateUpdate_TimestampProgression(t *testing.T) {
	updates := []StateUpdate{
		{Timestamp: 1000},
		{Timestamp: 2000},
		{Timestamp: 3000},
		{Timestamp: 4000},
		{Timestamp: 5000},
	}

	for i := 1; i < len(updates); i++ {
		if updates[i].Timestamp <= updates[i-1].Timestamp {
			t.Errorf("Update %d timestamp not increasing: %d -> %d",
				i, updates[i-1].Timestamp, updates[i].Timestamp)
		}
	}
}

// TestConnectionInfo_MultipleConnections verifies handling multiple connections.
func TestConnectionInfo_MultipleConnections(t *testing.T) {
	connections := []ConnectionInfo{
		{PlayerID: 1, Address: "192.168.1.1:8080", Connected: true},
		{PlayerID: 2, Address: "192.168.1.2:8080", Connected: true},
		{PlayerID: 3, Address: "192.168.1.3:8080", Connected: false},
		{PlayerID: 4, Address: "192.168.1.4:8080", Connected: true},
	}

	if len(connections) != 4 {
		t.Errorf("Expected 4 connections, got %d", len(connections))
	}

	connectedCount := 0
	for _, conn := range connections {
		if conn.Connected {
			connectedCount++
		}
	}

	if connectedCount != 3 {
		t.Errorf("Expected 3 connected, got %d", connectedCount)
	}
}

// TestComponentData_ZeroValue verifies zero-value initialization.
func TestComponentData_ZeroValue(t *testing.T) {
	var cd ComponentData

	if cd.Type != "" {
		t.Errorf("Expected empty type, got %s", cd.Type)
	}

	if cd.Data != nil {
		t.Error("Expected nil data for zero value")
	}
}

// TestStateUpdate_ZeroValue verifies zero-value initialization.
func TestStateUpdate_ZeroValue(t *testing.T) {
	var update StateUpdate

	if update.Timestamp != 0 {
		t.Errorf("Expected timestamp 0, got %d", update.Timestamp)
	}

	if update.EntityID != 0 {
		t.Errorf("Expected entity ID 0, got %d", update.EntityID)
	}

	if update.Components != nil {
		t.Error("Expected nil components for zero value")
	}

	if update.Priority != 0 {
		t.Errorf("Expected priority 0, got %d", update.Priority)
	}

	if update.SequenceNumber != 0 {
		t.Errorf("Expected sequence number 0, got %d", update.SequenceNumber)
	}
}

// TestInputCommand_ZeroValue verifies zero-value initialization.
func TestInputCommand_ZeroValue(t *testing.T) {
	var cmd InputCommand

	if cmd.PlayerID != 0 {
		t.Errorf("Expected player ID 0, got %d", cmd.PlayerID)
	}

	if cmd.Timestamp != 0 {
		t.Errorf("Expected timestamp 0, got %d", cmd.Timestamp)
	}

	if cmd.SequenceNumber != 0 {
		t.Errorf("Expected sequence number 0, got %d", cmd.SequenceNumber)
	}

	if cmd.InputType != "" {
		t.Errorf("Expected empty input type, got %s", cmd.InputType)
	}

	if cmd.Data != nil {
		t.Error("Expected nil data for zero value")
	}
}

// TestConnectionInfo_ZeroValue verifies zero-value initialization.
func TestConnectionInfo_ZeroValue(t *testing.T) {
	var conn ConnectionInfo

	if conn.PlayerID != 0 {
		t.Errorf("Expected player ID 0, got %d", conn.PlayerID)
	}

	if conn.Address != "" {
		t.Errorf("Expected empty address, got %s", conn.Address)
	}

	if conn.Latency != 0 {
		t.Errorf("Expected latency 0, got %f", conn.Latency)
	}

	if conn.Connected {
		t.Error("Expected connected to be false for zero value")
	}
}

// TestOpenShopMessage_Structure verifies OpenShopMessage struct initialization.
func TestOpenShopMessage_Structure(t *testing.T) {
	tests := []struct {
		name       string
		playerID   uint64
		merchantID uint64
		sequence   uint32
	}{
		{"valid_shop_open", 123, 456, 789},
		{"zero_values", 0, 0, 0},
		{"max_values", ^uint64(0), ^uint64(0), ^uint32(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := OpenShopMessage{
				PlayerID:       tt.playerID,
				MerchantID:     tt.merchantID,
				SequenceNumber: tt.sequence,
			}

			if msg.PlayerID != tt.playerID {
				t.Errorf("Expected PlayerID %d, got %d", tt.playerID, msg.PlayerID)
			}
			if msg.MerchantID != tt.merchantID {
				t.Errorf("Expected MerchantID %d, got %d", tt.merchantID, msg.MerchantID)
			}
			if msg.SequenceNumber != tt.sequence {
				t.Errorf("Expected SequenceNumber %d, got %d", tt.sequence, msg.SequenceNumber)
			}
		})
	}
}

// TestShopInventoryMessage_Structure verifies ShopInventoryMessage struct initialization.
func TestShopInventoryMessage_Structure(t *testing.T) {
	tests := []struct {
		name          string
		merchantID    uint64
		merchantName  string
		priceMulti    float64
		buyBackPct    float64
		itemIDs       []uint64
		itemPrices    []int
		sequence      uint32
		expectValid   bool
		checkParallel bool
	}{
		{
			name:          "valid_inventory",
			merchantID:    100,
			merchantName:  "Test Merchant",
			priceMulti:    1.5,
			buyBackPct:    0.5,
			itemIDs:       []uint64{1, 2, 3},
			itemPrices:    []int{100, 200, 300},
			sequence:      1,
			expectValid:   true,
			checkParallel: true,
		},
		{
			name:          "empty_inventory",
			merchantID:    200,
			merchantName:  "Empty Shop",
			priceMulti:    1.8,
			buyBackPct:    0.4,
			itemIDs:       []uint64{},
			itemPrices:    []int{},
			sequence:      2,
			expectValid:   true,
			checkParallel: true,
		},
		{
			name:          "nil_slices",
			merchantID:    300,
			merchantName:  "Nil Shop",
			priceMulti:    2.0,
			buyBackPct:    0.3,
			itemIDs:       nil,
			itemPrices:    nil,
			sequence:      3,
			expectValid:   true,
			checkParallel: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ShopInventoryMessage{
				MerchantID:        tt.merchantID,
				MerchantName:      tt.merchantName,
				PriceMultiplier:   tt.priceMulti,
				BuyBackPercentage: tt.buyBackPct,
				ItemIDs:           tt.itemIDs,
				ItemPrices:        tt.itemPrices,
				SequenceNumber:    tt.sequence,
			}

			if msg.MerchantID != tt.merchantID {
				t.Errorf("Expected MerchantID %d, got %d", tt.merchantID, msg.MerchantID)
			}
			if msg.MerchantName != tt.merchantName {
				t.Errorf("Expected MerchantName %s, got %s", tt.merchantName, msg.MerchantName)
			}
			if msg.PriceMultiplier != tt.priceMulti {
				t.Errorf("Expected PriceMultiplier %.2f, got %.2f", tt.priceMulti, msg.PriceMultiplier)
			}
			if msg.BuyBackPercentage != tt.buyBackPct {
				t.Errorf("Expected BuyBackPercentage %.2f, got %.2f", tt.buyBackPct, msg.BuyBackPercentage)
			}

			if tt.checkParallel && len(msg.ItemIDs) != len(msg.ItemPrices) {
				t.Errorf("ItemIDs and ItemPrices length mismatch: %d vs %d",
					len(msg.ItemIDs), len(msg.ItemPrices))
			}
		})
	}
}

// TestBuyItemMessage_Structure verifies BuyItemMessage struct initialization.
func TestBuyItemMessage_Structure(t *testing.T) {
	tests := []struct {
		name          string
		playerID      uint64
		merchantID    uint64
		itemIndex     int
		expectedPrice int
		sequence      uint32
	}{
		{"valid_purchase", 123, 456, 0, 100, 1},
		{"large_index", 999, 888, 999, 5000, 2},
		{"zero_price", 111, 222, 5, 0, 3},
		{"negative_index", 333, 444, -1, 200, 4}, // Invalid but test structure
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := BuyItemMessage{
				PlayerID:       tt.playerID,
				MerchantID:     tt.merchantID,
				ItemIndex:      tt.itemIndex,
				ExpectedPrice:  tt.expectedPrice,
				SequenceNumber: tt.sequence,
			}

			if msg.PlayerID != tt.playerID {
				t.Errorf("Expected PlayerID %d, got %d", tt.playerID, msg.PlayerID)
			}
			if msg.MerchantID != tt.merchantID {
				t.Errorf("Expected MerchantID %d, got %d", tt.merchantID, msg.MerchantID)
			}
			if msg.ItemIndex != tt.itemIndex {
				t.Errorf("Expected ItemIndex %d, got %d", tt.itemIndex, msg.ItemIndex)
			}
			if msg.ExpectedPrice != tt.expectedPrice {
				t.Errorf("Expected ExpectedPrice %d, got %d", tt.expectedPrice, msg.ExpectedPrice)
			}
		})
	}
}

// TestSellItemMessage_Structure verifies SellItemMessage struct initialization.
func TestSellItemMessage_Structure(t *testing.T) {
	tests := []struct {
		name          string
		playerID      uint64
		merchantID    uint64
		itemIndex     int
		expectedPrice int
		sequence      uint32
	}{
		{"valid_sale", 123, 456, 0, 50, 1},
		{"large_index", 999, 888, 100, 2500, 2},
		{"zero_price", 111, 222, 5, 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := SellItemMessage{
				PlayerID:       tt.playerID,
				MerchantID:     tt.merchantID,
				ItemIndex:      tt.itemIndex,
				ExpectedPrice:  tt.expectedPrice,
				SequenceNumber: tt.sequence,
			}

			if msg.PlayerID != tt.playerID {
				t.Errorf("Expected PlayerID %d, got %d", tt.playerID, msg.PlayerID)
			}
			if msg.MerchantID != tt.merchantID {
				t.Errorf("Expected MerchantID %d, got %d", tt.merchantID, msg.MerchantID)
			}
			if msg.ItemIndex != tt.itemIndex {
				t.Errorf("Expected ItemIndex %d, got %d", tt.itemIndex, msg.ItemIndex)
			}
			if msg.ExpectedPrice != tt.expectedPrice {
				t.Errorf("Expected ExpectedPrice %d, got %d", tt.expectedPrice, msg.ExpectedPrice)
			}
		})
	}
}

// TestTransactionResultMessage_Structure verifies TransactionResultMessage struct initialization.
func TestTransactionResultMessage_Structure(t *testing.T) {
	tests := []struct {
		name             string
		playerID         uint64
		merchantID       uint64
		success          bool
		errorMsg         string
		transType        string
		itemID           uint64
		goldAmount       int
		updatedGold      int
		updatedInventory bool
		sequence         uint32
	}{
		{
			name:             "successful_buy",
			playerID:         123,
			merchantID:       456,
			success:          true,
			errorMsg:         "",
			transType:        "buy",
			itemID:           789,
			goldAmount:       -100,
			updatedGold:      900,
			updatedInventory: true,
			sequence:         1,
		},
		{
			name:             "successful_sell",
			playerID:         123,
			merchantID:       456,
			success:          true,
			errorMsg:         "",
			transType:        "sell",
			itemID:           999,
			goldAmount:       50,
			updatedGold:      1050,
			updatedInventory: true,
			sequence:         2,
		},
		{
			name:             "failed_insufficient_gold",
			playerID:         123,
			merchantID:       456,
			success:          false,
			errorMsg:         "insufficient gold",
			transType:        "buy",
			itemID:           0,
			goldAmount:       0,
			updatedGold:      100,
			updatedInventory: false,
			sequence:         3,
		},
		{
			name:             "failed_invalid_item",
			playerID:         123,
			merchantID:       456,
			success:          false,
			errorMsg:         "item not found",
			transType:        "sell",
			itemID:           0,
			goldAmount:       0,
			updatedGold:      1000,
			updatedInventory: false,
			sequence:         4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := TransactionResultMessage{
				PlayerID:          tt.playerID,
				MerchantID:        tt.merchantID,
				Success:           tt.success,
				ErrorMessage:      tt.errorMsg,
				TransactionType:   tt.transType,
				ItemID:            tt.itemID,
				GoldAmount:        tt.goldAmount,
				UpdatedPlayerGold: tt.updatedGold,
				UpdatedInventory:  tt.updatedInventory,
				SequenceNumber:    tt.sequence,
			}

			if msg.PlayerID != tt.playerID {
				t.Errorf("Expected PlayerID %d, got %d", tt.playerID, msg.PlayerID)
			}
			if msg.MerchantID != tt.merchantID {
				t.Errorf("Expected MerchantID %d, got %d", tt.merchantID, msg.MerchantID)
			}
			if msg.Success != tt.success {
				t.Errorf("Expected Success %v, got %v", tt.success, msg.Success)
			}
			if msg.ErrorMessage != tt.errorMsg {
				t.Errorf("Expected ErrorMessage %s, got %s", tt.errorMsg, msg.ErrorMessage)
			}
			if msg.TransactionType != tt.transType {
				t.Errorf("Expected TransactionType %s, got %s", tt.transType, msg.TransactionType)
			}
			if msg.GoldAmount != tt.goldAmount {
				t.Errorf("Expected GoldAmount %d, got %d", tt.goldAmount, msg.GoldAmount)
			}
			if msg.UpdatedInventory != tt.updatedInventory {
				t.Errorf("Expected UpdatedInventory %v, got %v", tt.updatedInventory, msg.UpdatedInventory)
			}

			// Verify success/error message correlation
			if tt.success && msg.ErrorMessage != "" {
				t.Error("Successful transaction should have empty error message")
			}
			if !tt.success && msg.ErrorMessage == "" {
				t.Error("Failed transaction should have error message")
			}
		})
	}
}

// TestCloseShopMessage_Structure verifies CloseShopMessage struct initialization.
func TestCloseShopMessage_Structure(t *testing.T) {
	tests := []struct {
		name       string
		playerID   uint64
		merchantID uint64
		sequence   uint32
	}{
		{"valid_close", 123, 456, 789},
		{"zero_values", 0, 0, 0},
		{"max_values", ^uint64(0), ^uint64(0), ^uint32(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := CloseShopMessage{
				PlayerID:       tt.playerID,
				MerchantID:     tt.merchantID,
				SequenceNumber: tt.sequence,
			}

			if msg.PlayerID != tt.playerID {
				t.Errorf("Expected PlayerID %d, got %d", tt.playerID, msg.PlayerID)
			}
			if msg.MerchantID != tt.merchantID {
				t.Errorf("Expected MerchantID %d, got %d", tt.merchantID, msg.MerchantID)
			}
			if msg.SequenceNumber != tt.sequence {
				t.Errorf("Expected SequenceNumber %d, got %d", tt.sequence, msg.SequenceNumber)
			}
		})
	}
}

// TestCommerceProtocolWorkflow tests a complete commerce transaction flow.
func TestCommerceProtocolWorkflow(t *testing.T) {
	// Simulate client-server commerce interaction

	// Step 1: Client opens shop
	openMsg := OpenShopMessage{
		PlayerID:       100,
		MerchantID:     200,
		SequenceNumber: 1,
	}

	if openMsg.PlayerID != 100 || openMsg.MerchantID != 200 {
		t.Fatal("Failed to create OpenShopMessage")
	}

	// Step 2: Server responds with inventory
	inventoryMsg := ShopInventoryMessage{
		MerchantID:        200,
		MerchantName:      "Test Merchant",
		PriceMultiplier:   1.5,
		BuyBackPercentage: 0.5,
		ItemIDs:           []uint64{1001, 1002, 1003},
		ItemPrices:        []int{100, 200, 300},
		SequenceNumber:    2,
	}

	if len(inventoryMsg.ItemIDs) != len(inventoryMsg.ItemPrices) {
		t.Fatal("Inventory item count mismatch")
	}

	// Step 3: Client buys an item
	buyMsg := BuyItemMessage{
		PlayerID:       100,
		MerchantID:     200,
		ItemIndex:      0,
		ExpectedPrice:  100,
		SequenceNumber: 3,
	}

	if buyMsg.ItemIndex >= len(inventoryMsg.ItemIDs) {
		t.Error("Item index out of bounds")
	}

	// Step 4: Server responds with transaction result
	resultMsg := TransactionResultMessage{
		PlayerID:          100,
		MerchantID:        200,
		Success:           true,
		ErrorMessage:      "",
		TransactionType:   "buy",
		ItemID:            1001,
		GoldAmount:        -100,
		UpdatedPlayerGold: 900,
		UpdatedInventory:  true,
		SequenceNumber:    4,
	}

	if !resultMsg.Success {
		t.Error("Expected successful transaction")
	}
	if resultMsg.GoldAmount >= 0 {
		t.Error("Buy transaction should have negative gold amount")
	}

	// Step 5: Client closes shop
	closeMsg := CloseShopMessage{
		PlayerID:       100,
		MerchantID:     200,
		SequenceNumber: 5,
	}

	if closeMsg.PlayerID != openMsg.PlayerID {
		t.Error("Player ID mismatch in workflow")
	}
}

// TestCommerceProtocolFailureScenarios tests error handling in commerce protocol.
func TestCommerceProtocolFailureScenarios(t *testing.T) {
	tests := []struct {
		name         string
		resultMsg    TransactionResultMessage
		expectError  bool
		errorPattern string
	}{
		{
			name: "insufficient_gold",
			resultMsg: TransactionResultMessage{
				Success:      false,
				ErrorMessage: "insufficient gold",
			},
			expectError:  true,
			errorPattern: "gold",
		},
		{
			name: "invalid_item_index",
			resultMsg: TransactionResultMessage{
				Success:      false,
				ErrorMessage: "item index out of range",
			},
			expectError:  true,
			errorPattern: "index",
		},
		{
			name: "inventory_full",
			resultMsg: TransactionResultMessage{
				Success:      false,
				ErrorMessage: "inventory full",
			},
			expectError:  true,
			errorPattern: "inventory",
		},
		{
			name: "price_mismatch",
			resultMsg: TransactionResultMessage{
				Success:      false,
				ErrorMessage: "price changed",
			},
			expectError:  true,
			errorPattern: "price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resultMsg.Success {
				t.Error("Expected failed transaction")
			}
			if tt.resultMsg.ErrorMessage == "" {
				t.Error("Expected error message for failed transaction")
			}
		})
	}
}
