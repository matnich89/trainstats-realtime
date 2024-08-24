package model

type NationalData struct {
	OnTime              string `json:"on_time"`
	CancelledOrVeryLate string `json:"cancelled_or_very_late"`
	Late                string `json:"late"`
	Total               string `json:"total"`
}
