package Bus

import (
	"github.com/Wetitpig/etaHK/ui"
	"golang.org/x/exp/slices"
)

func (f *busFile) buildStopIndex() {
	s := f.Stops[len(f.Stops)-1]
	if _, ok := f.StopIndex[s.SId]; !ok {
		f.StopIndex[s.SId] = stop{[]ui.Lang{s.Name}, []*stopEntry{&s}}
	} else {
		si := f.StopIndex[s.SId]
		if slices.Index(si.Name, s.Name) == -1 {
			si.Name = append(si.Name, s.Name)
		}
		si.StopEntries = append(si.StopEntries, &s)
		f.StopIndex[s.SId] = si
	}
}

func (f *busFile) buildRouteIndex(op operator, feature map[string]interface{}) {
	s := f.Stops[len(f.Stops)-1]
	if _, ok := f.RouteIndex[s.RId]; !ok {
		f.RouteIndex[s.RId] = subroute{op, feature["routeNameC"].(string), formLang(feature, "locStartName"), formLang(feature, "locEndName"), make([][]*stopEntry, 2)}
	}
	ri := f.RouteIndex[s.RId]
	if len(ri.Dir[s.RSeq]) == 0 {
		ri.Dir[s.RSeq] = []*stopEntry{&s}
	} else {
		ri.Dir[s.RSeq] = append(ri.Dir[s.RSeq], &s)
	}
	f.RouteIndex[s.RId] = ri
}
