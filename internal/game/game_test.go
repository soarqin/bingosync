package game

import (
	"testing"
)

func TestPhaseRuleSecondSettlerNoCondition(t *testing.T) {
	// Create a new phase rule game
	g := NewGame(RulePhase)
	g.Start()

	// First, unlock all rows for red player
	// Row 0: mark 2 cells to unlock row 1
	g.MarkCell(0, 0, ColorRed)
	g.MarkCell(0, 1, ColorRed)
	// Row 1: mark 2 cells to unlock row 2
	g.MarkCell(1, 0, ColorRed)
	g.MarkCell(1, 1, ColorRed)
	// Row 2: mark 2 cells to unlock row 3
	g.MarkCell(2, 0, ColorRed)
	g.MarkCell(2, 1, ColorRed)
	// Row 3: mark 2 cells to unlock row 4
	g.MarkCell(3, 0, ColorRed)
	g.MarkCell(3, 1, ColorRed)

	// Now row 4 is unlocked, mark 2 cells in row 5 (index 4) for red player
	g.MarkCell(4, 0, ColorRed)
	g.MarkCell(4, 1, ColorRed)

	// Red player should be able to settle (meets conditions)
	err := g.Settle(ColorRed)
	if err != nil {
		t.Errorf("Red player should be able to settle, got error: %v", err)
	}

	if !g.RedSettled {
		t.Error("Red player should be settled")
	}

	if g.FirstSettler != ColorRed {
		t.Errorf("First settler should be red, got: %v", g.FirstSettler)
	}

	// Unlock all rows for blue player
	g.MarkCell(0, 2, ColorBlue)
	g.MarkCell(0, 3, ColorBlue)
	g.MarkCell(1, 2, ColorBlue)
	g.MarkCell(1, 3, ColorBlue)
	g.MarkCell(2, 2, ColorBlue)
	g.MarkCell(2, 3, ColorBlue)
	g.MarkCell(3, 2, ColorBlue)
	g.MarkCell(3, 3, ColorBlue)

	// Blue player has NOT marked any cells in row 5
	// But should still be able to settle because red already settled
	err = g.Settle(ColorBlue)
	if err != nil {
		t.Errorf("Blue player should be able to settle without conditions after red settled, got error: %v", err)
	}

	if !g.BlueSettled {
		t.Error("Blue player should be settled")
	}

	// Game should be finished now
	if g.Status != StatusFinished {
		t.Errorf("Game should be finished, got status: %v", g.Status)
	}
}

func TestPhaseRuleFirstSettlerNeedsCondition(t *testing.T) {
	// Create a new phase rule game
	g := NewGame(RulePhase)
	g.Start()

	// Unlock all rows for red player
	g.MarkCell(0, 0, ColorRed)
	g.MarkCell(0, 1, ColorRed)
	g.MarkCell(1, 0, ColorRed)
	g.MarkCell(1, 1, ColorRed)
	g.MarkCell(2, 0, ColorRed)
	g.MarkCell(2, 1, ColorRed)
	g.MarkCell(3, 0, ColorRed)
	g.MarkCell(3, 1, ColorRed)

	// Red player has NOT marked enough cells in row 5 (only 1 cell)
	g.MarkCell(4, 0, ColorRed)

	// Should NOT be able to settle
	err := g.Settle(ColorRed)
	if err == nil {
		t.Error("Red player should NOT be able to settle without meeting conditions")
	}

	if g.RedSettled {
		t.Error("Red player should NOT be settled")
	}

	if g.FirstSettler != ColorNone {
		t.Errorf("First settler should be none, got: %v", g.FirstSettler)
	}
}
