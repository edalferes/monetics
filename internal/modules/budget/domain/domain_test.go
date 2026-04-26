package domain

import (
	"testing"
	"time"
)

// --- Budget ---

func TestBudget_RemainingAmount(t *testing.T) {
	b := &Budget{Amount: 1000, Spent: 250}
	if got := b.RemainingAmount(); got != 750 {
		t.Fatalf("RemainingAmount = %v, want 750", got)
	}
}

func TestBudget_PercentageUsed(t *testing.T) {
	tests := []struct {
		name string
		b    Budget
		want float64
	}{
		{"half spent", Budget{Amount: 200, Spent: 100}, 50},
		{"zero amount returns zero", Budget{Amount: 0, Spent: 100}, 0},
		{"none spent", Budget{Amount: 100, Spent: 0}, 0},
		{"over 100", Budget{Amount: 100, Spent: 150}, 150},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.PercentageUsed(); got != tt.want {
				t.Fatalf("PercentageUsed = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBudget_IsOverBudget(t *testing.T) {
	if (&Budget{Amount: 100, Spent: 99}).IsOverBudget() {
		t.Fatal("should not be over when spent < amount")
	}
	if (&Budget{Amount: 100, Spent: 100}).IsOverBudget() {
		t.Fatal("should not be over when spent == amount")
	}
	if !(&Budget{Amount: 100, Spent: 101}).IsOverBudget() {
		t.Fatal("should be over when spent > amount")
	}
}

func TestBudget_ShouldAlert(t *testing.T) {
	threshold := 80.0
	tests := []struct {
		name    string
		alertAt *float64
		spent   float64
		want    bool
	}{
		{"nil threshold never alerts", nil, 99, false},
		{"below threshold", &threshold, 70, false},
		{"at threshold", &threshold, 80, true},
		{"above threshold", &threshold, 90, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Budget{Amount: 100, Spent: tt.spent, AlertAt: tt.alertAt}
			if got := b.ShouldAlert(); got != tt.want {
				t.Fatalf("ShouldAlert = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBudget_Overlaps(t *testing.T) {
	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	apr := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)

	base := &Budget{UserID: 1, CategoryID: 1, StartDate: jan, EndDate: mar}

	tests := []struct {
		name  string
		other *Budget
		want  bool
	}{
		{"same window same user/category", &Budget{UserID: 1, CategoryID: 1, StartDate: jan, EndDate: mar}, true},
		{"partial overlap", &Budget{UserID: 1, CategoryID: 1, StartDate: feb, EndDate: apr}, true},
		{"adjacent (touching) does not overlap", &Budget{UserID: 1, CategoryID: 1, StartDate: mar, EndDate: apr}, false},
		{"different user", &Budget{UserID: 2, CategoryID: 1, StartDate: jan, EndDate: mar}, false},
		{"different category", &Budget{UserID: 1, CategoryID: 2, StartDate: jan, EndDate: mar}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := base.Overlaps(tt.other); got != tt.want {
				t.Fatalf("Overlaps = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Money / Period ---

func TestMoney_Construction(t *testing.T) {
	if _, err := NewMoney(100, ""); err != ErrInvalidCurrency {
		t.Fatalf("expected ErrInvalidCurrency, got %v", err)
	}
	m, err := NewMoneyFromFloat(12.345, "BRL")
	if err != nil {
		t.Fatal(err)
	}
	if m.AmountMinor() != 1235 {
		t.Fatalf("expected 1235 cents (rounded), got %d", m.AmountMinor())
	}
	if m.Currency() != "BRL" {
		t.Fatalf("expected BRL, got %s", m.Currency())
	}
}

func TestMoney_AddSub(t *testing.T) {
	a, _ := NewMoney(1000, "USD")
	b, _ := NewMoney(250, "USD")
	c, _ := NewMoney(100, "BRL")

	sum, err := a.Add(b)
	if err != nil || sum.AmountMinor() != 1250 {
		t.Fatalf("Add failed: %v %d", err, sum.AmountMinor())
	}
	if _, err := a.Add(c); err != ErrInvalidCurrency {
		t.Fatalf("expected ErrInvalidCurrency on cross-currency Add, got %v", err)
	}

	diff, err := a.Sub(b)
	if err != nil || diff.AmountMinor() != 750 {
		t.Fatalf("Sub failed: %v %d", err, diff.AmountMinor())
	}
	if _, err := a.Sub(c); err != ErrInvalidCurrency {
		t.Fatalf("expected ErrInvalidCurrency on cross-currency Sub, got %v", err)
	}
}

func TestPeriod_Validation(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := NewPeriod(now, now, BudgetPeriodMonthly); err != ErrInvalidPeriod {
		t.Fatalf("expected ErrInvalidPeriod for zero range, got %v", err)
	}
	if _, err := NewPeriod(now.Add(time.Hour), now, BudgetPeriodMonthly); err != ErrInvalidPeriod {
		t.Fatalf("expected ErrInvalidPeriod for inverted range, got %v", err)
	}
}

func TestPeriod_ContainsAndOverlaps(t *testing.T) {
	jan, _ := NewPeriod(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		BudgetPeriodMonthly,
	)
	feb, _ := NewPeriod(
		time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC),
		BudgetPeriodMonthly,
	)
	mid, _ := NewPeriod(
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC),
		BudgetPeriodMonthly,
	)

	if !jan.Contains(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("jan should contain 2025-01-15")
	}
	if jan.Overlaps(feb) {
		t.Fatal("jan and feb should not overlap")
	}
	if !jan.Overlaps(mid) {
		t.Fatal("jan and mid should overlap")
	}
}

// --- Transaction ---

func TestResolveStatus(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		date time.Time
		want TransactionStatus
	}{
		{"past date is completed", now.Add(-24 * time.Hour), TransactionStatusCompleted},
		{"same instant is completed", now, TransactionStatusCompleted},
		{"future date is pending", now.Add(24 * time.Hour), TransactionStatusPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveStatus(tc.date, now); got != tc.want {
				t.Fatalf("ResolveStatus(%v, %v) = %v, want %v", tc.date, now, got, tc.want)
			}
		})
	}
}
