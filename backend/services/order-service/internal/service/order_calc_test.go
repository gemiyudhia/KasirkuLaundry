package service

import "testing"

func TestCalculateItems(t *testing.T) {
	tests := []struct {
		name          string
		in            []CreateOrderItemInput
		wantTotal     int64
		wantQuantity  []float64
		wantSubtotals []int64
	}{
		{
			name:          "single item with decimal quantity",
			in:            []CreateOrderItemInput{{ServiceName: "Cuci Kiloan", UnitPrice: 7000, Quantity: 2.5}},
			wantTotal:     17500,
			wantQuantity:  []float64{2.5},
			wantSubtotals: []int64{17500},
		},
		{
			name:          "quantity rounded to 2 decimals",
			in:            []CreateOrderItemInput{{ServiceName: "Cuci Kiloan", UnitPrice: 7000, Quantity: 2.556}},
			wantTotal:     17920,
			wantQuantity:  []float64{2.56},
			wantSubtotals: []int64{17920},
		},
		{
			name:          "subtotal rounded to nearest rupiah",
			in:            []CreateOrderItemInput{{ServiceName: "Cuci Satuan", UnitPrice: 3333, Quantity: 1.5}},
			wantTotal:     5000, // 4999.5 dibulatkan ke atas
			wantQuantity:  []float64{1.5},
			wantSubtotals: []int64{5000},
		},
		{
			name: "multiple items are summed",
			in: []CreateOrderItemInput{
				{ServiceName: "Cuci Kiloan", UnitPrice: 7000, Quantity: 2.5},
				{ServiceName: "Setrika", UnitPrice: 4000, Quantity: 2},
			},
			wantTotal:     25500,
			wantQuantity:  []float64{2.5, 2},
			wantSubtotals: []int64{17500, 8000},
		},
		{
			name:          "empty input",
			in:            nil,
			wantTotal:     0,
			wantQuantity:  []float64{},
			wantSubtotals: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total := calculateItems(tt.in)

			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if len(items) != len(tt.wantSubtotals) {
				t.Fatalf("len(items) = %d, want %d", len(items), len(tt.wantSubtotals))
			}
			for i, it := range items {
				if it.Subtotal != tt.wantSubtotals[i] {
					t.Errorf("items[%d].Subtotal = %d, want %d", i, it.Subtotal, tt.wantSubtotals[i])
				}
				if it.Quantity != tt.wantQuantity[i] {
					t.Errorf("items[%d].Quantity = %v, want %v", i, it.Quantity, tt.wantQuantity[i])
				}
			}
		})
	}
}
