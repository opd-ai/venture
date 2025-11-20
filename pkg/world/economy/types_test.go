package economy

import (
	"testing"
	"time"
)

func TestSortCriteria_String(t *testing.T) {
	tests := []struct {
		name     string
		criteria SortCriteria
		want     string
	}{
		{"SortByPrice", SortByPrice, "Price (Ascending)"},
		{"SortByPriceDesc", SortByPriceDesc, "Price (Descending)"},
		{"SortByQuantity", SortByQuantity, "Quantity"},
		{"SortByDeliveryTime", SortByDeliveryTime, "Delivery Time"},
		{"SortByRelevance", SortByRelevance, "Relevance"},
		{"Unknown", SortCriteria(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.criteria.String(); got != tt.want {
				t.Errorf("SortCriteria.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeliveryMethod_String(t *testing.T) {
	tests := []struct {
		name   string
		method DeliveryMethod
		want   string
	}{
		{"DeliveryMail", DeliveryMail, "Mail"},
		{"DeliveryCourier", DeliveryCourier, "Courier"},
		{"Unknown", DeliveryMethod(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.String(); got != tt.want {
				t.Errorf("DeliveryMethod.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateTransactionFee(t *testing.T) {
	tests := []struct {
		name  string
		price int
		hops  int
		want  int
	}{
		{"Local transaction (0 hops)", 1000, 0, 50}, // 5%
		{"1 hop", 1000, 1, 70},                      // 7%
		{"2 hops", 1000, 2, 90},                     // 9%
		{"3 hops", 1000, 3, 110},                    // 11%
		{"4 hops", 1000, 4, 130},                    // 13%
		{"5 hops (capped at 15%)", 1000, 5, 150},    // 15% (capped)
		{"10 hops (capped at 15%)", 1000, 10, 150},  // 15% (capped)
		{"High price with hops", 10000, 2, 900},     // 9%
		{"Small price", 100, 1, 7},                  // 7%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTransactionFee(tt.price, tt.hops)
			if got != tt.want {
				t.Errorf("CalculateTransactionFee(%d, %d) = %d, want %d", tt.price, tt.hops, got, tt.want)
			}
		})
	}
}

func TestListing_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"Not expired", time.Now().Add(24 * time.Hour), false},
		{"Expired", time.Now().Add(-24 * time.Hour), true},
		{"Just expired", time.Now().Add(-1 * time.Second), true},
		{"Future expiration", time.Now().Add(7 * 24 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing := &Listing{
				ExpiresAt: tt.expiresAt,
			}
			if got := listing.IsExpired(); got != tt.want {
				t.Errorf("Listing.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListing_GetDeliveryTime(t *testing.T) {
	tests := []struct {
		name           string
		deliveryMethod DeliveryMethod
		hops           int
		want           int
	}{
		{"Mail delivery", DeliveryMail, 0, 0},
		{"Mail with hops (still instant)", DeliveryMail, 3, 0},
		{"Courier 0 hops", DeliveryCourier, 0, 600},             // 10 minutes
		{"Courier 1 hop", DeliveryCourier, 1, 900},              // 15 minutes
		{"Courier 2 hops", DeliveryCourier, 2, 1200},            // 20 minutes
		{"Courier 5 hops", DeliveryCourier, 5, 2100},            // 35 minutes
		{"Courier 10 hops (capped)", DeliveryCourier, 10, 3600}, // 60 minutes (capped)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing := &Listing{
				DeliveryMethod: tt.deliveryMethod,
				EstimatedHops:  tt.hops,
			}
			if got := listing.GetDeliveryTime(); got != tt.want {
				t.Errorf("Listing.GetDeliveryTime() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestListing_GetTotalCost(t *testing.T) {
	tests := []struct {
		name     string
		price    int
		hops     int
		quantity int
		want     int
	}{
		{"Single item, no hops", 1000, 0, 1, 1050},   // 1000 + 5%
		{"Single item, 1 hop", 1000, 1, 1, 1070},     // 1000 + 7%
		{"Multiple items, no hops", 500, 0, 3, 1575}, // 1500 + 5%
		{"Multiple items, 2 hops", 200, 2, 5, 1090},  // 1000 + 9%
		{"High value, 3 hops", 10000, 3, 1, 11100},   // 10000 + 11%
		{"Max fee cap", 1000, 10, 1, 1150},           // 1000 + 15% (capped)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing := &Listing{
				Price:         tt.price,
				EstimatedHops: tt.hops,
			}
			if got := listing.GetTotalCost(tt.quantity); got != tt.want {
				t.Errorf("Listing.GetTotalCost(%d) = %d, want %d", tt.quantity, got, tt.want)
			}
		})
	}
}

func BenchmarkCalculateTransactionFee(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateTransactionFee(1000, 2)
	}
}

func BenchmarkListing_GetDeliveryTime(b *testing.B) {
	listing := &Listing{
		DeliveryMethod: DeliveryCourier,
		EstimatedHops:  3,
	}
	for i := 0; i < b.N; i++ {
		listing.GetDeliveryTime()
	}
}

func BenchmarkListing_GetTotalCost(b *testing.B) {
	listing := &Listing{
		Price:         1000,
		EstimatedHops: 2,
	}
	for i := 0; i < b.N; i++ {
		listing.GetTotalCost(5)
	}
}
