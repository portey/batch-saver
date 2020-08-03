package models

import "encoding/json"

type Event struct {
	ID      string
	GroupID string
	Data    []byte
}

func (e Event) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}
