package game

import (
	"errors"
)

var (
	ErrGameNotStarted    = errors.New("game has not started")
	ErrGameFinished      = errors.New("game already finished")
	ErrCellAlreadyMarked = errors.New("cell already marked")
	ErrRowLocked         = errors.New("row is locked")
	ErrRowLimitExceeded  = errors.New("row mark limit exceeded")
	ErrAlreadySettled    = errors.New("player already settled")
	ErrCannotSettleYet   = errors.New("need at least 2 cells in row 5 to settle")
)

// NewGame creates a new game with specified rule
func NewGame(rule GameRule) *Game {
	g := &Game{
		Board:           NewBoard(),
		Rule:            rule,
		PhaseConfig:     DefaultPhaseConfig(),
		Status:          StatusWaiting,
		BingoLine:       -1,
		RedUnlockedRow:  0,
		BlueUnlockedRow: 0,
	}
	return g
}

// Start begins the game
func (g *Game) Start() error {
	if g.Status == StatusPlaying {
		return errors.New("game already in progress")
	}
	g.Status = StatusPlaying
	return nil
}

// MarkCell marks a cell for a player
func (g *Game) MarkCell(row, col int, player PlayerColor) error {
	if g.Status == StatusWaiting {
		return ErrGameNotStarted
	}
	if g.Status == StatusFinished {
		return ErrGameFinished
	}

	if row < 0 || row > 4 || col < 0 || col > 4 {
		return errors.New("invalid cell position")
	}

	cell := &g.Board.Cells[row][col]

	switch g.Rule {
	case RuleNormal:
		if err := g.markNormal(cell, player); err != nil {
			return err
		}
	case RuleBlackout:
		if err := g.markBlackout(cell, player); err != nil {
			return err
		}
	case RulePhase:
		if err := g.markPhase(row, col, player); err != nil {
			return err
		}
	}

	// Check for winner (phase rule checks after mark)
	if g.Rule != RulePhase {
		g.CheckWin()
	}

	return nil
}

// MarkCellForce marks a cell with force overwrite (for referee)
func (g *Game) MarkCellForce(row, col int, player PlayerColor) error {
	if g.Status == StatusWaiting {
		return ErrGameNotStarted
	}

	if row < 0 || row > 4 || col < 0 || col > 4 {
		return errors.New("invalid cell position")
	}

	cell := &g.Board.Cells[row][col]

	cell.MarkedBy = player
	cell.SecondMark = ColorNone
	cell.Times = 0

	if g.Rule != RulePhase {
		g.CheckWin()
	}

	return nil
}

// markNormal handles marking for normal rule
func (g *Game) markNormal(cell *Cell, player PlayerColor) error {
	if cell.MarkedBy != ColorNone {
		return ErrCellAlreadyMarked
	}

	cell.MarkedBy = player
	return nil
}

// markBlackout handles marking for blackout rule
// Both players can mark the same cell, first marker in MarkedBy, second in SecondMark
func (g *Game) markBlackout(cell *Cell, player PlayerColor) error {
	// Check if player already marked this cell
	if cell.MarkedBy == player {
		return errors.New("player already marked this cell")
	}
	if cell.SecondMark == player {
		return errors.New("player already marked this cell")
	}

	// First marker
	if cell.MarkedBy == ColorNone {
		cell.MarkedBy = player
		cell.Times = 1
		return nil
	}

	// Second marker (different color from first)
	if cell.SecondMark == ColorNone {
		cell.SecondMark = player
		cell.Times = 2
		return nil
	}

	// Cell already has both colors marked
	return ErrCellAlreadyMarked
}

// markPhase handles marking for phase rule
func (g *Game) markPhase(row, col int, player PlayerColor) error {
	cell := &g.Board.Cells[row][col]

	var unlockedRow *int
	var rowMarks *int
	if player == ColorRed {
		unlockedRow = &g.RedUnlockedRow
		rowMarks = &g.RedRowMarks[row]
	} else {
		unlockedRow = &g.BlueUnlockedRow
		rowMarks = &g.BlueRowMarks[row]
	}

	// Check if row is locked
	if row > *unlockedRow {
		return ErrRowLocked
	}

	// Check per-row limit
	if *rowMarks >= g.PhaseConfig.CellsPerRow {
		return ErrRowLimitExceeded
	}

	// Check if player already marked this cell
	if cell.MarkedBy == player || cell.SecondMark == player {
		return errors.New("player already marked this cell")
	}

	// Mark the cell
	if cell.MarkedBy == ColorNone {
		cell.MarkedBy = player
	} else if cell.SecondMark == ColorNone {
		cell.SecondMark = player
		cell.Times = 1
	}

	// Update row marks count
	*rowMarks++

	// Check for row unlock: only when marking the current highest unlocked row
	// and reaching the threshold, unlock the next row
	if row == *unlockedRow && *unlockedRow < 4 {
		if *rowMarks >= g.PhaseConfig.UnlockThreshold {
			*unlockedRow++
		}
	}

	// Check for Bingo
	if g.BingoAchiever == ColorNone {
		g.checkPhaseBingo()
	}

	return nil
}

// checkPhaseBingo checks for vertical and diagonal Bingo
func (g *Game) checkPhaseBingo() bool {
	for col := 0; col < 5; col++ {
		if g.checkPhaseLineBingo(0, col, 1, 0, col) {
			return true
		}
	}

	if g.checkPhaseLineBingo(0, 0, 1, 1, 5) {
		return true
	}
	if g.checkPhaseLineBingo(0, 4, 1, -1, 6) {
		return true
	}

	return false
}

// checkPhaseLineBingo checks if a line has Bingo
func (g *Game) checkPhaseLineBingo(startRow, startCol, dRow, dCol, lineIndex int) bool {
	redCount := 0
	blueCount := 0

	for i := 0; i < 5; i++ {
		cell := g.Board.Cells[startRow+i*dRow][startCol+i*dCol]
		if cell.MarkedBy == ColorRed || cell.SecondMark == ColorRed {
			redCount++
		}
		if cell.MarkedBy == ColorBlue || cell.SecondMark == ColorBlue {
			blueCount++
		}
	}

	if redCount == 5 && g.BingoAchiever == ColorNone {
		g.BingoAchiever = ColorRed
		g.BingoLine = lineIndex
		return true
	}
	if blueCount == 5 && g.BingoAchiever == ColorNone {
		g.BingoAchiever = ColorBlue
		g.BingoLine = lineIndex
		return true
	}

	return false
}

// CanSettle checks if a player can trigger settlement
func (g *Game) CanSettle(player PlayerColor) bool {
	var rowMarks int
	if player == ColorRed {
		rowMarks = g.RedRowMarks[4]
	} else {
		rowMarks = g.BlueRowMarks[4]
	}
	return rowMarks >= 2
}

// Settle triggers settlement for a player
func (g *Game) Settle(player PlayerColor) error {
	if g.Status != StatusPlaying {
		return ErrGameNotStarted
	}

	if player == ColorRed && g.RedSettled {
		return ErrAlreadySettled
	}
	if player == ColorBlue && g.BlueSettled {
		return ErrAlreadySettled
	}

	// First settler must meet conditions, second settler can settle without conditions
	if g.FirstSettler == ColorNone {
		// This is the first settler - must meet conditions
		if !g.CanSettle(player) {
			return ErrCannotSettleYet
		}
		g.FirstSettler = player
	}
	// Second settler doesn't need to meet any conditions

	if player == ColorRed {
		g.RedSettled = true
	} else {
		g.BlueSettled = true
	}

	if g.RedSettled && g.BlueSettled {
		g.checkPhaseWin()
	}

	return nil
}

// CalculatePhaseScore calculates scores for phase rule
func (g *Game) CalculatePhaseScore() (redScore, blueScore int) {
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			cell := g.Board.Cells[row][col]

			if cell.MarkedBy == ColorRed {
				redScore += g.PhaseConfig.RowScores[row]
			} else if cell.MarkedBy == ColorBlue {
				blueScore += g.PhaseConfig.RowScores[row]
			}

			if cell.SecondMark == ColorRed {
				redScore += g.PhaseConfig.SecondHalfScores[row]
			} else if cell.SecondMark == ColorBlue {
				blueScore += g.PhaseConfig.SecondHalfScores[row]
			}
		}
	}

	if g.BingoAchiever == ColorRed {
		redScore += g.PhaseConfig.BingoBonus
	} else if g.BingoAchiever == ColorBlue {
		blueScore += g.PhaseConfig.BingoBonus
	}

	if g.FirstSettler == ColorRed {
		redScore += g.PhaseConfig.FinalBonus
	} else if g.FirstSettler == ColorBlue {
		blueScore += g.PhaseConfig.FinalBonus
	}

	return redScore, blueScore
}

// CountMarks counts total marks for each player
func (g *Game) CountMarks() (redCount, blueCount int) {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			cell := g.Board.Cells[i][j]
			if cell.MarkedBy == ColorRed || cell.SecondMark == ColorRed {
				redCount++
			}
			if cell.MarkedBy == ColorBlue || cell.SecondMark == ColorBlue {
				blueCount++
			}
		}
	}
	return
}

// Reset resets the game board
func (g *Game) Reset() {
	g.Board = NewBoard()
	g.Status = StatusWaiting
	g.Winner = nil
	g.RedRowMarks = [5]int{}
	g.BlueRowMarks = [5]int{}
	g.RedUnlockedRow = 0
	g.BlueUnlockedRow = 0
	g.BingoAchiever = ColorNone
	g.BingoLine = -1
	g.RedSettled = false
	g.BlueSettled = false
	g.FirstSettler = ColorNone
}

// GetState returns the current game state
func (g *Game) GetState() *Game {
	return g
}

// CheckWin checks if there is a winner and updates game state
func (g *Game) CheckWin() *Winner {
	var winner *Winner

	switch g.Rule {
	case RuleNormal:
		winner = g.checkNormalWin()
	case RuleBlackout:
		winner = g.checkBlackoutWin()
	case RulePhase:
		return nil
	}

	if winner != nil {
		g.Status = StatusFinished
		g.Winner = winner
	} else {
		if g.Status == StatusFinished {
			g.Status = StatusPlaying
		}
		g.Winner = nil
	}

	return winner
}

// checkPhaseWin checks and sets winner for phase rule after both settled
func (g *Game) checkPhaseWin() *Winner {
	redScore, blueScore := g.CalculatePhaseScore()

	var winner PlayerColor
	if redScore > blueScore {
		winner = ColorRed
	} else if blueScore > redScore {
		winner = ColorBlue
	} else {
		winner = g.FirstSettler
	}

	g.Winner = &Winner{
		Winner:    winner,
		Reason:    WinReasonPhase,
		RedScore:  redScore,
		BlueScore: blueScore,
	}
	g.Status = StatusFinished

	return g.Winner
}

// checkNormalWin checks for winner in normal rule
func (g *Game) checkNormalWin() *Winner {
	for row := 0; row < 5; row++ {
		if winner := g.checkLineWin(row, 0, 0, 1); winner != ColorNone {
			redCount, blueCount := g.CountMarks()
			return &Winner{
				Winner:    winner,
				Reason:    WinReasonBingo,
				RedScore:  redCount,
				BlueScore: blueCount,
			}
		}
	}

	for col := 0; col < 5; col++ {
		if winner := g.checkLineWin(0, col, 1, 0); winner != ColorNone {
			redCount, blueCount := g.CountMarks()
			return &Winner{
				Winner:    winner,
				Reason:    WinReasonBingo,
				RedScore:  redCount,
				BlueScore: blueCount,
			}
		}
	}

	if winner := g.checkLineWin(0, 0, 1, 1); winner != ColorNone {
		redCount, blueCount := g.CountMarks()
		return &Winner{
			Winner:    winner,
			Reason:    WinReasonBingo,
			RedScore:  redCount,
			BlueScore: blueCount,
		}
	}
	if winner := g.checkLineWin(0, 4, 1, -1); winner != ColorNone {
		redCount, blueCount := g.CountMarks()
		return &Winner{
			Winner:    winner,
			Reason:    WinReasonBingo,
			RedScore:  redCount,
			BlueScore: blueCount,
		}
	}

	return g.checkFullBoard()
}

// checkLineWin checks if a line is completely marked by one player
func (g *Game) checkLineWin(startRow, startCol, dRow, dCol int) PlayerColor {
	firstCell := g.Board.Cells[startRow][startCol]
	if firstCell.MarkedBy == ColorNone {
		return ColorNone
	}

	for i := 1; i < 5; i++ {
		cell := g.Board.Cells[startRow+i*dRow][startCol+i*dCol]
		if cell.MarkedBy != firstCell.MarkedBy {
			return ColorNone
		}
	}

	return firstCell.MarkedBy
}

// checkFullBoard checks if the board is full and determines winner
func (g *Game) checkFullBoard() *Winner {
	redCount, blueCount := g.CountMarks()
	total := redCount + blueCount

	if total < 25 {
		return nil
	}

	var winner PlayerColor
	if redCount > blueCount {
		winner = ColorRed
	} else if blueCount > redCount {
		winner = ColorBlue
	} else {
		winner = ColorNone
	}

	return &Winner{
		Winner:    winner,
		Reason:    WinReasonFullBoard,
		RedScore:  redCount,
		BlueScore: blueCount,
	}
}

// checkBlackoutWin checks for winner in blackout rule
func (g *Game) checkBlackoutWin() *Winner {
	redCount, blueCount := g.CountMarks()

	if redCount == 25 {
		return &Winner{
			Winner:    ColorRed,
			Reason:    WinReasonBlackout,
			RedScore:  redCount,
			BlueScore: blueCount,
		}
	}
	if blueCount == 25 {
		return &Winner{
			Winner:    ColorBlue,
			Reason:    WinReasonBlackout,
			RedScore:  redCount,
			BlueScore: blueCount,
		}
	}

	return nil
}

// SetCellText sets the text of a cell
func (g *Game) SetCellText(row, col int, text string) error {
	if row < 0 || row > 4 || col < 0 || col > 4 {
		return errors.New("invalid cell position")
	}

	g.Board.Cells[row][col].Text = text
	return nil
}

// SetAllCellTexts sets all cell texts at once
func (g *Game) SetAllCellTexts(texts []string) error {
	if len(texts) != 25 {
		return errors.New("must provide exactly 25 texts")
	}

	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			g.Board.Cells[row][col].Text = texts[row*5+col]
		}
	}
	return nil
}

// UnmarkCell removes all marks from a cell (for referee)
// For clearing a specific color, use ClearCellMark
func (g *Game) UnmarkCell(row, col int) error {
	if g.Status == StatusWaiting {
		return ErrGameNotStarted
	}

	if row < 0 || row > 4 || col < 0 || col > 4 {
		return errors.New("invalid cell position")
	}

	cell := &g.Board.Cells[row][col]

	if g.Rule == RulePhase {
		// Update row marks count
		if cell.MarkedBy == ColorRed && g.RedRowMarks[row] > 0 {
			g.RedRowMarks[row]--
		} else if cell.MarkedBy == ColorBlue && g.BlueRowMarks[row] > 0 {
			g.BlueRowMarks[row]--
		}
		if cell.SecondMark == ColorRed && g.RedRowMarks[row] > 0 {
			g.RedRowMarks[row]--
		} else if cell.SecondMark == ColorBlue && g.BlueRowMarks[row] > 0 {
			g.BlueRowMarks[row]--
		}
	}

	cell.MarkedBy = ColorNone
	cell.SecondMark = ColorNone
	cell.Times = 0

	// Re-check winner status (phase rule doesn't check here)
	if g.Rule != RulePhase {
		g.CheckWin()
	}

	return nil
}

// ClearCellMark clears a specific color mark from a cell
// Used for blackout and phase rules where both colors can be on the same cell
func (g *Game) ClearCellMark(row, col int, player PlayerColor) error {
	if g.Status == StatusWaiting {
		return ErrGameNotStarted
	}

	if row < 0 || row > 4 || col < 0 || col > 4 {
		return errors.New("invalid cell position")
	}

	cell := &g.Board.Cells[row][col]

	// Handle based on which mark to clear
	if cell.MarkedBy == player {
		// First mark is the one to clear
		// Promote second mark to first if exists
		cell.MarkedBy = cell.SecondMark
		cell.SecondMark = ColorNone
		if cell.Times > 0 {
			cell.Times--
		}

		// Update phase rule tracking
		if g.Rule == RulePhase {
			if player == ColorRed && g.RedRowMarks[row] > 0 {
				g.RedRowMarks[row]--
			} else if player == ColorBlue && g.BlueRowMarks[row] > 0 {
				g.BlueRowMarks[row]--
			}
		}
	} else if cell.SecondMark == player {
		// Second mark is the one to clear
		cell.SecondMark = ColorNone
		if cell.Times > 0 {
			cell.Times--
		}

		// Update phase rule tracking
		if g.Rule == RulePhase {
			if player == ColorRed && g.RedRowMarks[row] > 0 {
				g.RedRowMarks[row]--
			} else if player == ColorBlue && g.BlueRowMarks[row] > 0 {
				g.BlueRowMarks[row]--
			}
		}
	}
	// If player has no mark on this cell, do nothing

	// For phase rule, recheck Bingo status after clearing
	if g.Rule == RulePhase {
		g.recheckPhaseBingo()
	}

	return nil
}

// recheckPhaseBingo rechecks Bingo status after a mark is cleared
// If the current Bingo line is broken, clear it and try to find a new one
func (g *Game) recheckPhaseBingo() {
	if g.BingoAchiever == ColorNone {
		return
	}

	// Check if current Bingo line is still valid
	if g.BingoLine >= 0 && g.BingoLine < 5 {
		// Vertical line
		col := g.BingoLine
		valid := true
		for row := 0; row < 5; row++ {
			cell := g.Board.Cells[row][col]
			if cell.MarkedBy != g.BingoAchiever && cell.SecondMark != g.BingoAchiever {
				valid = false
				break
			}
		}
		if valid {
			return // Bingo still valid
		}
	} else if g.BingoLine == 5 {
		// Diagonal top-left to bottom-right
		valid := true
		for i := 0; i < 5; i++ {
			cell := g.Board.Cells[i][i]
			if cell.MarkedBy != g.BingoAchiever && cell.SecondMark != g.BingoAchiever {
				valid = false
				break
			}
		}
		if valid {
			return // Bingo still valid
		}
	} else if g.BingoLine == 6 {
		// Diagonal top-right to bottom-left
		valid := true
		for i := 0; i < 5; i++ {
			cell := g.Board.Cells[i][4-i]
			if cell.MarkedBy != g.BingoAchiever && cell.SecondMark != g.BingoAchiever {
				valid = false
				break
			}
		}
		if valid {
			return // Bingo still valid
		}
	}

	// Current Bingo is broken, clear it
	g.BingoAchiever = ColorNone
	g.BingoLine = -1

	// Try to find a new Bingo (first one found wins)
	g.checkPhaseBingo()
}
