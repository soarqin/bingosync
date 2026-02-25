package game

// PlayerColor represents the color of a player
type PlayerColor int

const (
	ColorNone PlayerColor = iota
	ColorRed
	ColorBlue
)

func (c PlayerColor) String() string {
	switch c {
	case ColorRed:
		return "red"
	case ColorBlue:
		return "blue"
	default:
		return "none"
	}
}

func PlayerColorFromString(s string) PlayerColor {
	switch s {
	case "red":
		return ColorRed
	case "blue":
		return ColorBlue
	default:
		return ColorNone
	}
}

// GameRule represents the type of game rule
type GameRule int

const (
	RuleNormal GameRule = iota // Normal rule: each cell can only be marked once
	RuleBlackout               // Blackout: allow duplicate marks, record times
	RulePhase                  // Phase rule: row-by-row with limits and scoring
)

func (r GameRule) String() string {
	switch r {
	case RuleNormal:
		return "normal"
	case RuleBlackout:
		return "blackout"
	case RulePhase:
		return "phase"
	default:
		return "unknown"
	}
}

func GameRuleFromString(s string) GameRule {
	switch s {
	case "normal":
		return RuleNormal
	case "blackout":
		return RuleBlackout
	case "phase":
		return RulePhase
	default:
		return RuleNormal
	}
}

// PhaseConfig holds configuration for phase rule
type PhaseConfig struct {
	RowScores        [5]int `json:"row_scores"`         // A[n]: Score per row, default: [2, 2, 4, 4, 6]
	SecondHalfScores [5]int `json:"second_half_scores"` // B[n]: Score for second player, default: [1, 1, 2, 2, 3]
	CellsPerRow      int    `json:"cells_per_row"`      // C: Max cells each player can mark per row, default: 3
	UnlockThreshold  int    `json:"unlock_threshold"`   // D: Cells needed to unlock next row, default: 2
	BingoBonus       int    `json:"bingo_bonus"`        // E: Bonus for first Bingo, default: 3
	FinalBonus       int    `json:"final_bonus"`        // F: Bonus for first settlement, default: 3
}

// DefaultPhaseConfig returns the default phase configuration
func DefaultPhaseConfig() PhaseConfig {
	return PhaseConfig{
		RowScores:        [5]int{2, 2, 4, 4, 6},
		SecondHalfScores: [5]int{1, 1, 2, 2, 3},
		CellsPerRow:      3,
		UnlockThreshold:  2,
		BingoBonus:       3,
		FinalBonus:       3,
	}
}

// Cell represents a single cell on the board
type Cell struct {
	MarkedBy   PlayerColor `json:"marked_by"`   // Which player marked this cell first
	SecondMark PlayerColor `json:"second_mark"` // Which player marked this cell second (for phase rule)
	Times      int         `json:"times"`       // How many times marked (for blackout/phase)
	Text       string      `json:"text"`        // Text displayed in the cell
}

// Board represents the 5x5 bingo board
type Board struct {
	Cells [5][5]Cell `json:"cells"`
}

// NewBoard creates a new empty board
func NewBoard() *Board {
	return &Board{}
}

// GameStatus represents the current status of the game
type GameStatus int

const (
	StatusWaiting GameStatus = iota
	StatusPlaying
	StatusFinished
)

func (s GameStatus) String() string {
	switch s {
	case StatusWaiting:
		return "waiting"
	case StatusPlaying:
		return "playing"
	case StatusFinished:
		return "finished"
	default:
		return "unknown"
	}
}

// WinReason represents the reason for winning
type WinReason string

const (
	WinReasonBingo     WinReason = "bingo"
	WinReasonFullBoard WinReason = "full_board"
	WinReasonBlackout  WinReason = "blackout"
	WinReasonPhase     WinReason = "phase" // Phase rule: settlement complete
)

// Winner represents the game result
type Winner struct {
	Winner    PlayerColor `json:"winner"`
	Reason    WinReason   `json:"reason"`
	RedScore  int         `json:"red_score"`
	BlueScore int         `json:"blue_score"`
}

// Game represents a complete game state
type Game struct {
	Board       *Board      `json:"board"`
	Rule        GameRule    `json:"rule"`
	PhaseConfig PhaseConfig `json:"phase_config"`
	Status      GameStatus  `json:"status"`
	Winner      *Winner     `json:"winner,omitempty"`

	// For phase rule tracking - per-row marks
	RedRowMarks  [5]int `json:"red_row_marks"`  // Marks per row for red
	BlueRowMarks [5]int `json:"blue_row_marks"` // Marks per row for blue

	// Per-player row unlock tracking
	RedUnlockedRow  int `json:"red_unlocked_row"`  // Highest row unlocked by red
	BlueUnlockedRow int `json:"blue_unlocked_row"` // Highest row unlocked by blue

	// Bingo tracking (phase rule)
	BingoAchiever PlayerColor `json:"bingo_achiever"` // Who achieved Bingo first
	BingoLine     int         `json:"bingo_line"`     // Which line: 0-4 vertical, 5=diag\, 6=diag/

	// Settlement tracking (phase rule)
	RedSettled   bool        `json:"red_settled"`   // Whether red has settled
	BlueSettled  bool        `json:"blue_settled"`  // Whether blue has settled
	FirstSettler PlayerColor `json:"first_settler"` // Who settled first (for tie-breaking)
}
