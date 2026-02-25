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
	RowScores       [5]int   `json:"row_scores"`        // Score per row, default: [2, 2, 4, 4, 6]
	SecondHalfRate  float64  `json:"second_half_rate"`  // Rate for second player marking same cell, default: 0.5
	FinalBonus      int      `json:"final_bonus"`       // Bonus for triggering final calculation, default: 0
	FinalBonusType  string   `json:"final_bonus_type"`  // Type of final bonus calculation
	CellsPerRow     int      `json:"cells_per_row"`     // Max cells each player can mark per row, default: 3
	UnlockThreshold int      `json:"unlock_threshold"`  // Cells needed to unlock next row, default: 2
}

// DefaultPhaseConfig returns the default phase configuration
func DefaultPhaseConfig() PhaseConfig {
	return PhaseConfig{
		RowScores:       [5]int{2, 2, 4, 4, 6},
		SecondHalfRate:  0.5,
		FinalBonus:      0,
		FinalBonusType:  "fixed",
		CellsPerRow:     3,
		UnlockThreshold: 2,
	}
}

// Cell represents a single cell on the board
type Cell struct {
	MarkedBy PlayerColor `json:"marked_by"` // Which player marked this cell
	Times    int         `json:"times"`     // How many times marked (for blackout)
	Text     string      `json:"text"`      // Text displayed in the cell
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

	// For phase rule tracking
	RedRowMarks  [5]int `json:"red_row_marks"`  // Marks per row for red
	BlueRowMarks [5]int `json:"blue_row_marks"` // Marks per row for blue
	CurrentRow   int    `json:"current_row"`    // Current unlocked row (phase rule)
}
