package model

type NationalData struct {
	OnTime              int `json:"on_time"`
	CancelledOrVeryLate int `json:"cancelled_or_very_late"`
	Late                int `json:"late"`
	Total               int `json:"total"`
}

type TrainOperator struct {
	Name                string `json:"name"`
	Total               int    `json:"total"`
	OnTime              int    `json:"on_time"`
	Late                int    `json:"late"`
	CancelledOrVeryLate int    `json:"cancelled_or_very_late"`
	Percentage          int    `json:"percentage"`
	Position            int    `json:"position"`
}

type TrainOperatorTableEntry struct {
	Position int    `json:"position"`
	Name     string `json:"name"`
}

type TrainOperatorTable struct {
	TrainOperatorTableEntry int `json:"train_operator_table_entry"`
}
