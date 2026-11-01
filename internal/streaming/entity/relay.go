package entity

type RelayOffer struct {
	RoomID  string
	RelayID string
	SDP     string
}

type RelayAnswer struct {
	RoomID  string
	RelayID string
	SDP     string
}
