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
)

// NewGame creates a new game with specified rule
func NewGame(rule GameRule) *Game {
	g := &Game{
		Board:       NewBoard(),
		Rule:        rule,
		PhaseConfig: DefaultPhaseConfig(),
		Status:      StatusWaiting,
		CurrentRow:  0,
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
// Bingo only records marking status, no turn concept
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
		if err := g.markPhase(cell, row, player); err != nil {
			return err
		}
	}

	// Check for winner (phase rule doesn't check here)
	if g.Rule != RulePhase {
		g.CheckWin()
	}

	return nil
}

// MarkCellForce marks a cell with force overwrite (for referee)
// Referee can operate after game ends, will check if state needs reset
func (g *Game) MarkCellForce(row, col int, player PlayerColor) error {
	if g.Status == StatusWaiting {
		return ErrGameNotStarted
	}

	if row < 0 || row > 4 || col < 0 || col > 4 {
		return errors.New("invalid cell position")
	}

	cell := &g.Board.Cells[row][col]

	// Referee force mark: set color directly, reset times
	cell.MarkedBy = player
	cell.Times = 0

	// Re-check winner status (phase rule doesn't check here)
	if g.Rule != RulePhase {
		g.CheckWin()
	}

	return nil
}

// markNormal handles marking for normal rule
// Normal rule: each cell can only be marked once, first come first served
func (g *Game) markNormal(cell *Cell, player PlayerColor) error {
	if cell.MarkedBy != ColorNone {
		return ErrCellAlreadyMarked
	}

	cell.MarkedBy = player
	return nil
}

// markBlackout handles marking for blackout rule
// Blackout rule: allows repeated marking, records mark count
func (g *Game) markBlackout(cell *Cell, player PlayerColor) error {
	cell.Times++
	cell.MarkedBy = player
	return nil
}

// markPhase handles marking for phase rule
// Phase rule: unlock by row, each row has mark limit
func (g *Game) markPhase(cell *Cell, row int, player PlayerColor) error {
	// Check if row is unlocked
	if row > g.CurrentRow {
		return ErrRowLocked
	}

	// Check row limits
	var rowMarks *int
	if player == ColorRed {
		rowMarks = &g.RedRowMarks[row]
	} else {
		rowMarks = &g.BlueRowMarks[row]
	}

	if *rowMarks >= g.PhaseConfig.CellsPerRow {
		return ErrRowLimitExceeded
	}

	// Mark the cell
	if cell.MarkedBy == ColorNone {
		cell.MarkedBy = player
	} else if cell.MarkedBy != player {
		// Cell already marked by opponent, record double mark
		cell.Times++
	}

	*rowMarks++

	// Check if should unlock next row
	if g.CurrentRow < 4 {
		totalMarks := g.RedRowMarks[g.CurrentRow] + g.BlueRowMarks[g.CurrentRow]
		if totalMarks >= g.PhaseConfig.UnlockThreshold {
			g.CurrentRow++
		}
	}

	return nil
}

// UnmarkCell removes a mark from a cell (for corrections)
// Referee can operate after game ends, will check if state needs reset
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
	}

	cell.MarkedBy = ColorNone
	cell.Times = 0

	// Re-check winner status (phase rule doesn't check here)
	if g.Rule != RulePhase {
		g.CheckWin()
	}

	return nil
}

// CalculatePhaseScore calculates scores for phase rule
func (g *Game) CalculatePhaseScore() (redScore, blueScore int) {
	for row := 0; row < 5; row++ {
		rowScore := g.PhaseConfig.RowScores[row]

		for col := 0; col < 5; col++ {
			cell := g.Board.Cells[row][col]

			if cell.MarkedBy == ColorRed {
				if cell.Times > 0 {
					// Red was first, Blue marked after
					redScore += rowScore
					blueScore += int(float64(rowScore) * g.PhaseConfig.SecondHalfRate)
				} else {
					redScore += rowScore
				}
			} else if cell.MarkedBy == ColorBlue {
				if cell.Times > 0 {
					// Blue was first, Red marked after
					blueScore += rowScore
					redScore += int(float64(rowScore) * g.PhaseConfig.SecondHalfRate)
				} else {
					blueScore += rowScore
				}
			}
		}
	}

	// Add final bonus
	if g.PhaseConfig.FinalBonus > 0 {
		// Final bonus goes to player with more marks
		redCount, blueCount := g.CountMarks()
		if redCount > blueCount {
			redScore += g.PhaseConfig.FinalBonus
		} else if blueCount > redCount {
			blueScore += g.PhaseConfig.FinalBonus
		}
	}

	return redScore, blueScore
}

// CountMarks counts total marks for each player
func (g *Game) CountMarks() (redCount, blueCount int) {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			cell := g.Board.Cells[i][j]
			if cell.MarkedBy == ColorRed {
				redCount++
			} else if cell.MarkedBy == ColorBlue {
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
	g.CurrentRow = 0
	g.RedRowMarks = [5]int{}
	g.BlueRowMarks = [5]int{}
}

// GetState returns the current game state
func (g *Game) GetState() *Game {
	return g
}

// CheckWin checks if there's a winner and updates game state
func (g *Game) CheckWin() *Winner {
	var winner *Winner

	switch g.Rule {
	case RuleNormal:
		winner = g.checkNormalWin()
	case RuleBlackout:
		winner = g.checkBlackoutWin()
	case RulePhase:
		// Not implemented yet
		return nil
	}

	if winner != nil {
		g.Status = StatusFinished
		g.Winner = winner
	} else {
		// If previously finished but now no winner, reset to playing
		if g.Status == StatusFinished {
			g.Status = StatusPlaying
		}
		g.Winner = nil
	}

	return winner
}

// checkNormalWin checks for winner in normal rule
func (g *Game) checkNormalWin() *Winner {
	// Check all lines (5 horizontal + 5 vertical + 2 diagonal = 12 lines)
	
	// Horizontal lines
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

	// Vertical lines
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

	// Diagonal lines
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

	// Check if board is full
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

	// If board is not full, return nil
	if total < 25 {
		return nil
	}

	// Board is full, player with more cells wins
	var winner PlayerColor
	if redCount > blueCount {
		winner = ColorRed
	} else if blueCount > redCount {
		winner = ColorBlue
	} else {
		winner = ColorNone // Draw
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

	// Player wins when completing all 25 cells
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
