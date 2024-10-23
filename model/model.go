package model

type Carries string

const (
	Freight   Carries = "freight"
	Passenger Carries = "passenger"
)

type PerformanceData struct {
	OnTime                        int                  `json:"on_time"`
	CancelledOrVeryLate           int                  `json:"cancelled_or_very_late"`
	Late                          int                  `json:"late"`
	Total                         int                  `json:"total"`
	OnTimePercentage              float64              `json:"on_time_percentage"`
	CancelledOrVeryLatePercentage float64              `json:"cancelled_or_very_late_percentage"`
	LatePercentage                float64              `json:"late_percentage"`
	WorstOperator                 *OperatorPerformance `json:"worst_operator"`
	BestOperator                  *OperatorPerformance `json:"best_operator"`
}

type OperatorPerformance struct {
	Code                          string  `json:"code"`
	Name                          string  `json:"name"`
	OnTime                        int     `json:"on_time"`
	Late                          int     `json:"late"`
	CancelledOrVeryLate           int     `json:"cancelled_or_very_late"`
	Total                         int     `json:"total"`
	OnTimePercentage              float64 `json:"on_time_percentage"`
	LatePercentage                float64 `json:"late_percentage"`
	CancelledOrVeryLatePercentage float64 `json:"cancelled_or_very_late_percentage"`
	PerformanceScore              float64 `json:"performance_score"`
}

type TrainOperator struct {
	Code    string
	Name    string
	Carries Carries
}

var TrainOperators = map[string]TrainOperator{
	"22":  {Code: "22", Name: "Grand Central", Carries: Passenger},
	"55":  {Code: "55", Name: "Hull Trains", Carries: Passenger},
	"45":  {Code: "45", Name: "Lumo", Carries: Passenger},
	"28":  {Code: "28", Name: "East Midlands Railway", Carries: Passenger},
	"61":  {Code: "61", Name: "London North Eastern Railway", Carries: Passenger},
	"23":  {Code: "23", Name: "Northern Trains", Carries: Passenger},
	"29":  {Code: "29", Name: "West Midlands Trains", Carries: Passenger},
	"71":  {Code: "71", Name: "Transport for Wales", Carries: Passenger},
	"65":  {Code: "65", Name: "Avanti West Coast", Carries: Passenger},
	"80":  {Code: "80", Name: "Southeastern", Carries: Passenger},
	"27":  {Code: "27", Name: "CrossCountry", Carries: Passenger},
	"97":  {Code: "97", Name: "Direct Rail Services", Carries: Freight},
	"20":  {Code: "20", Name: "TransPennine Trains", Carries: Passenger},
	"42":  {Code: "42", Name: "Colas Freight", Carries: Freight},
	"64":  {Code: "64", Name: "Merseyrail", Carries: Passenger},
	"54":  {Code: "54", Name: "GB Railfreight", Carries: Freight},
	"25":  {Code: "25", Name: "Great Western Railway", Carries: Passenger},
	"60":  {Code: "60", Name: "ScotRail", Carries: Passenger},
	"NRM": {Code: "NRM", Name: "Network Rail - Materials", Carries: Freight},
	"05":  {Code: "05", Name: "DB Cargo", Carries: Freight},
	"88":  {Code: "88", Name: "Govia Thameslink Railway", Carries: Passenger},
	"84":  {Code: "84", Name: "South Western Railway", Carries: Passenger},
	"FHH": {Code: "FHH", Name: "Freightliner Heavy Haul", Carries: Freight},
	"79":  {Code: "79", Name: "c2c", Carries: Passenger},
	"FLI": {Code: "FLI", Name: "Freightliner Intermodal", Carries: Freight},
	"74":  {Code: "74", Name: "Chiltern", Carries: Passenger},
	"21":  {Code: "21", Name: "Greater Anglia", Carries: Passenger},
	"33":  {Code: "33", Name: "Elizabeth line", Carries: Passenger},
	"30":  {Code: "30", Name: "London Overground", Carries: Passenger},
	"35":  {Code: "35", Name: "Caledonian Sleeper Limited", Carries: Passenger},
	"86":  {Code: "86", Name: "Heathrow Express", Carries: Passenger},
}
