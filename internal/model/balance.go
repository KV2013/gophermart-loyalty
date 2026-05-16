package model

type Balance struct {
	Current   float64 `db:"current"   json:"current"`
	Withdrawn float64 `db:"withdrawn" json:"withdrawn"`
}
