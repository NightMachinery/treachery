package app

import "time"

type ForensicCardResource struct {
	CauseCards    []ForensicCard `json:"causeCards"`
	LocationCards []ForensicCard `json:"locationCards"`
	OtherCards    []ForensicCard `json:"otherCards"`
}

type Card struct {
	ID        string   `json:"id,omitempty"`
	ImgURL    string   `json:"imgUrl,omitempty"`
	AltImgURL string   `json:"altImgUrl,omitempty"`
	Name      string   `json:"name"`
	HasImage  bool     `json:"hasImage,omitempty"`
	GuessedBy []string `json:"guessedBy,omitempty"`
}

type ForensicCard struct {
	CardID           string   `json:"cardId,omitempty"`
	CardName         string   `json:"cardName"`
	Choices          []string `json:"choices"`
	ChoiceIDs        []string `json:"choiceIds,omitempty"`
	SelectedChoiceID string   `json:"selectedChoiceId,omitempty"`
	SelectedChoice   string   `json:"selectedChoice,omitempty"`
	Replaced         bool     `json:"replaced"`
}

type ParticipantRole string

const (
	ParticipantRolePlayer   ParticipantRole = "player"
	ParticipantRoleObserver ParticipantRole = "observer"
)

type SecretRole string

const (
	SecretRoleInvestigator SecretRole = "investigator"
	SecretRoleMurderer     SecretRole = "murderer"
	SecretRoleAccomplice   SecretRole = "accomplice"
	SecretRoleWitness      SecretRole = "witness"
)

type Winner string

const (
	WinnerNone             Winner = "none"
	WinnerInvestigatorTeam Winner = "investigatorTeam"
	WinnerMurdererTeam     Winner = "murdererTeam"
)

type Participant struct {
	Name              string          `json:"name"`
	UID               string          `json:"uid"`
	Role              ParticipantRole `json:"role"`
	IsCreator         bool            `json:"isCreator,omitempty"`
	IsScientist       bool            `json:"isScientist,omitempty"`
	IsMarkedScientist bool            `json:"isMarkedScientist,omitempty"`
}

type Viewer struct {
	UID           string          `json:"uid"`
	Name          string          `json:"name,omitempty"`
	Role          ParticipantRole `json:"role,omitempty"`
	IsParticipant bool            `json:"isParticipant"`
	IsCreator     bool            `json:"isCreator"`
	IsScientist   bool            `json:"isScientist"`
}

type Player struct {
	Name       string `json:"name"`
	UID        string `json:"uid"`
	ClueCards  []Card `json:"clueCards"`
	MeansCards []Card `json:"meansCards"`
	Guessed    bool   `json:"guessed,omitempty"`
}

type Guess struct {
	GuessedByUID  string `json:"guessedByUid"`
	MurdererUID   string `json:"murdererUid"`
	MeansCardID   string `json:"meansCardId,omitempty"`
	MeansCardName string `json:"meansCardName"`
	ClueCardID    string `json:"clueCardId,omitempty"`
	ClueCardName  string `json:"clueCardName"`
	Correct       bool   `json:"correct"`
	CreatedAt     string `json:"createdTimestamp,omitempty"`
}

type MessageType string

const (
	MessageTypeChat     MessageType = "chat"
	MessageTypeGuess    MessageType = "guess"
	MessageTypeForensic MessageType = "forensic"
)

type Message struct {
	PlayerUID string      `json:"playerUid,omitempty"`
	Message   string      `json:"message"`
	Timestamp string      `json:"timestamp"`
	Type      MessageType `json:"type"`
}

type RoomTimer struct {
	DurationSeconds        int    `json:"durationSeconds"`
	ExpiresAt              string `json:"expiresAt,omitempty"`
	PausedRemainingSeconds int    `json:"pausedRemainingSeconds,omitempty"`
	RunID                  int    `json:"runId"`
}

type Game struct {
	CreatorUID              string         `json:"creatorUid"`
	ScientistUID            string         `json:"scientistUid,omitempty"`
	MarkedScientistUID      string         `json:"markedScientistUid,omitempty"`
	CauseCard               *ForensicCard  `json:"causeCard,omitempty"`
	LocationCard            *ForensicCard  `json:"locationCard,omitempty"`
	OtherCards              []ForensicCard `json:"otherCards"`
	GameID                  string         `json:"gameId"`
	CreatedTimestamp        string         `json:"createdTimestamp"`
	StartedTimestamp        string         `json:"startedTimestamp,omitempty"`
	MurdererSelected        bool           `json:"murdererSelected,omitempty"`
	MurdererCardsSelected   bool           `json:"murdererCardsSelected"`
	MurdererClueCardID      string         `json:"murdererClueCardId,omitempty"`
	MurdererClueCardName    string         `json:"murdererClueCardName,omitempty"`
	MurdererMeansCardID     string         `json:"murdererMeansCardId,omitempty"`
	MurdererMeansCardName   string         `json:"murdererMeansCardName,omitempty"`
	MurdererUID             string         `json:"murdererUid,omitempty"`
	StartedOn               string         `json:"startedOn,omitempty"`
	Finished                bool           `json:"finished"`
	MeansCardsPerPlayer     int            `json:"meansCardsPerPlayer"`
	ClueCardsPerPlayer      int            `json:"clueCardsPerPlayer"`
	LinkClueCountToMeans    bool           `json:"linkClueCountToMeans"`
	MeansCluesTextOnly      bool           `json:"meansCluesTextOnly"`
	CrimePackID             string         `json:"crimePackId"`
	CrimePackLanguage       string         `json:"crimePackLanguage"`
	CrimePackAssetSetID     string         `json:"crimePackAssetSetId,omitempty"`
	HintPackID              string         `json:"hintPackId"`
	HintPackLanguage        string         `json:"hintPackLanguage"`
	AccompliceCount         int            `json:"accompliceCount"`
	WitnessCount            int            `json:"witnessCount"`
	WitnessesToFind         int            `json:"witnessesToFind"`
	PendingWitnessSelection bool           `json:"pendingWitnessSelection"`
	Winner                  Winner         `json:"winner"`
	FinishedReason          string         `json:"finishedReason,omitempty"`
	ResultMessage           string         `json:"resultMessage,omitempty"`
	RoomTimer               *RoomTimer     `json:"roomTimer,omitempty"`
	RoomTimerRunID          int            `json:"-"`
}

type KnownRolePlayer struct {
	UID  string     `json:"uid"`
	Name string     `json:"name"`
	Role SecretRole `json:"role"`
}

type WitnessSelectionPromptState struct {
	RequiredSelections int  `json:"requiredSelections"`
	Dismissible        bool `json:"dismissible"`
	CreatorInitiated   bool `json:"creatorInitiated"`
	Active             bool `json:"active"`
}

type PlayerPrivateData struct {
	IsMurderer                   bool                         `json:"isMurderer"`
	ClueCardID                   string                       `json:"clueCardId,omitempty"`
	ClueCardName                 string                       `json:"clueCardName,omitempty"`
	MeansCardID                  string                       `json:"meansCardId,omitempty"`
	MeansCardName                string                       `json:"meansCardName,omitempty"`
	Role                         SecretRole                   `json:"role,omitempty"`
	KnownMurdererTeam            []KnownRolePlayer            `json:"knownMurdererTeam,omitempty"`
	KnownMurdererClueCardID      string                       `json:"knownMurdererClueCardId,omitempty"`
	KnownMurdererClueCardName    string                       `json:"knownMurdererClueCardName,omitempty"`
	KnownMurdererMeansCardID     string                       `json:"knownMurdererMeansCardId,omitempty"`
	KnownMurdererMeansCardName   string                       `json:"knownMurdererMeansCardName,omitempty"`
	ActiveWitnessSelectionPrompt *WitnessSelectionPromptState `json:"activeWitnessSelectionPrompt,omitempty"`
}

type ForensicPrivateData struct {
	Murderer              *Player `json:"murderer,omitempty"`
	MurdererClueCardID    string  `json:"murdererClueCardId,omitempty"`
	MurdererClueCardName  string  `json:"murdererClueCardName,omitempty"`
	MurdererMeansCardID   string  `json:"murdererMeansCardId,omitempty"`
	MurdererMeansCardName string  `json:"murdererMeansCardName,omitempty"`
}

type WitnessPromptCandidate struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type ModeratorPrivateData struct {
	WitnessPromptCandidates []WitnessPromptCandidate `json:"witnessPromptCandidates"`
}

type RoleRevealEntry struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type GameSnapshot struct {
	Game                 *Game                 `json:"game"`
	Participants         []Participant         `json:"participants"`
	Players              []Player              `json:"players"`
	Guesses              []Guess               `json:"guesses"`
	Messages             []Message             `json:"messages"`
	Viewer               Viewer                `json:"viewer"`
	PlayerPrivateData    PlayerPrivateData     `json:"playerPrivateData"`
	ForensicPrivateData  *ForensicPrivateData  `json:"forensicPrivateData,omitempty"`
	ModeratorPrivateData *ModeratorPrivateData `json:"moderatorPrivateData,omitempty"`
	RoleReveal           []RoleRevealEntry     `json:"roleReveal,omitempty"`
	ServerTimestamp      string                `json:"serverTimestamp"`
}

type Session struct {
	Token            string `json:"token,omitempty"`
	UID              string `json:"uid"`
	DisplayName      string `json:"displayName,omitempty"`
	CreatedTimestamp string `json:"createdTimestamp,omitempty"`
}

type RoomAuthToken struct {
	Token            string `json:"token"`
	GameID           string `json:"gameId"`
	UID              string `json:"uid,omitempty"`
	CreatedTimestamp string `json:"createdTimestamp,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func nowTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
