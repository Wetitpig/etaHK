package GMB

import (
	"fmt"
)

type stopLoc struct {
	route_id, route_seq, stop_seq int
}

func formStopLoc(obj map[string]interface{}) stopLoc {
	return stopLoc{
		int(obj["route_id"].(float64)), int(obj["route_seq"].(float64)) - 1, int(obj["stop_seq"].(float64)),
	}
}

func (loc *stopLoc) marshal() string {
	return fmt.Sprintf("%d-%d-%d", loc.route_id, loc.route_seq, loc.stop_seq)
}

func (loc *stopLoc) unmarshal(input string) {
	fmt.Sscanf(input, "%d-%d-%d", &loc.route_id, &loc.route_seq, &loc.stop_seq)
}
