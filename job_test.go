package mediarails

import "testing"

func TestJobStatus_IsTerminal(t *testing.T) {
	if !JobCompleted.IsTerminal() {
		t.Error("completed should be terminal")
	}
	if !JobFailed.IsTerminal() {
		t.Error("failed should be terminal")
	}
	if JobProcessing.IsTerminal() {
		t.Error("processing should not be terminal")
	}
	if JobQueued.IsTerminal() {
		t.Error("queued should not be terminal")
	}
}

func TestUsage_Cost(t *testing.T) {
	u := &Usage{Unit: "characters", Quantity: 100}
	cost := u.Cost(0.01)
	if cost != 1.0 {
		t.Errorf("expected 1.0, got %f", cost)
	}
}

func TestUsage_Cost_Nil(t *testing.T) {
	var u *Usage
	if u.Cost(0.01) != 0 {
		t.Error("nil usage should return 0 cost")
	}
}
