package model

type Balance struct {
	Current   float32 `db:"current"   json:"current"`
	Withdrawn float32 `db:"withdrawn" json:"withdrawn"`
}
