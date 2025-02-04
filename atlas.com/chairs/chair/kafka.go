package chair

const (
	EnvEventTopicStatus      = "EVENT_TOPIC_CHAIR_STATUS"
	EventStatusTypeUsed      = "USED"
	EventStatusTypeCancelled = "CANCELLED"

	ChairTypeFixed    = "FIXED"
	ChairTypePortable = "PORTABLE"
)

type statusEvent[E any] struct {
	WorldId   byte   `json:"worldId"`
	ChannelId byte   `json:"channelId"`
	MapId     uint32 `json:"mapId"`
	ChairType string `json:"chairType"`
	ChairId   uint32 `json:"chairId"`
	Type      string `json:"type"`
	Body      E      `json:"body"`
}

type statusEventUsedBody struct {
	CharacterId uint32 `json:"characterId"`
}

type statusEventCancelledBody struct {
	CharacterId uint32 `json:"characterId"`
}
