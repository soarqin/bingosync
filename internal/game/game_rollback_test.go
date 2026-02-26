package game

import (
	"testing"
)

func TestPhaseRuleRowUnlockRollback(t *testing.T) {
	// Create a new phase rule game
	g := NewGame(RulePhase)
	g.Start()

	// Row 0: mark 2 cells to unlock row 1
	g.MarkCell(0, 0, ColorRed)
	g.MarkCell(0, 1, ColorRed)

	if g.RedUnlockedRow != 1 {
		t.Errorf("Red unlocked row should be 1, got: %d", g.RedUnlockedRow)
	}

	// Row 1: mark 2 cells to unlock row 2
	g.MarkCell(1, 0, ColorRed)
	g.MarkCell(1, 1, ColorRed)

	if g.RedUnlockedRow != 2 {
		t.Errorf("Red unlocked row should be 2, got: %d", g.RedUnlockedRow)
	}

	// Clear one mark from row 1 (now only 1 mark, below threshold)
	g.ClearCellMark(1, 1, ColorRed)

	// Row should be rolled back to 1
	if g.RedUnlockedRow != 1 {
		t.Errorf("Red unlocked row should be rolled back to 1, got: %d", g.RedUnlockedRow)
	}

	// Clear one mark from row 0 (now only 1 mark, below threshold)
	g.ClearCellMark(0, 1, ColorRed)

	// Row should be rolled back to 0
	if g.RedUnlockedRow != 0 {
		t.Errorf("Red unlocked row should be rolled back to 0, got: %d", g.RedUnlockedRow)
	}

	// Try to mark row 1 again, should fail because it's locked
	err := g.MarkCell(1, 2, ColorRed)
	if err != ErrRowLocked {
		t.Errorf("Expected ErrRowLocked, got: %v", err)
	}
}

func TestPhaseRuleUnmarkCellRollback(t *testing.T) {
	// Create a new phase rule game
	g := NewGame(RulePhase)
	g.Start()

	// Red player marks row 0
	g.MarkCell(0, 0, ColorRed)
	g.MarkCell(0, 1, ColorRed)

	// Blue player marks row 0
	g.MarkCell(0, 2, ColorBlue)
	g.MarkCell(0, 3, ColorBlue)

	if g.RedUnlockedRow != 1 {
		t.Errorf("Red unlocked row should be 1, got: %d", g.RedUnlockedRow)
	}
	if g.BlueUnlockedRow != 1 {
		t.Errorf("Blue unlocked row should be 1, got: %d", g.BlueUnlockedRow)
	}

	// Use UnmarkCell (referee action) to clear a cell with red mark
	g.UnmarkCell(0, 1) // This was a red mark

	// Red should be rolled back, blue should still be unlocked
	if g.RedUnlockedRow != 0 {
		t.Errorf("Red unlocked row should be rolled back to 0, got: %d", g.RedUnlockedRow)
	}
	if g.BlueUnlockedRow != 1 {
		t.Errorf("Blue unlocked row should still be 1, got: %d", g.BlueUnlockedRow)
	}
}
