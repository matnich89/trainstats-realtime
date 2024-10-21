package model

type NationalData struct {
	OnTime                        int     `json:"on_time"`
	CancelledOrVeryLate           int     `json:"cancelled_or_very_late"`
	Late                          int     `json:"late"`
	Total                         int     `json:"total"`
	OnTimePercentage              float64 `json:"on_time_percentage"`
	CancelledOrVeryLatePercentage float64 `json:"cancelled_or_very_late_percentage"`
	LatePercentage                float64 `json:"late_percentage"`
}
