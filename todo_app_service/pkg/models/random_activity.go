package models

type RandomActivity struct {
	Activity     string `json:"activity"`
	Type         string `json:"type"`
	Participants int    `json:"participants"`
	KidFriendly  bool   `json:"kidFriendly"`
}
