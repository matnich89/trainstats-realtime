package model

type NationalData struct {
	OnTime              int `json:"on_time"`
	CancelledOrVeryLate int `json:"cancelled_or_very_late"`
	Late                int `json:"late"`
	Total               int `json:"total"`
}
