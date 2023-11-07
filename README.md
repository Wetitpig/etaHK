# ETA@HK
Command line utility to display "Next Bus" times for public transport, written in Golang.

General key commands
* `t` for switching to traditional Chinese
* `s` for switching to simplified Chinese
* `e` for switching to English
* `Tab` for switching between input fields.
* `h` for returning to homepage.
* `b` for returning to the last page.
* `q` for quitting the utility.

## Build
```
go install github.com/Wetitpig/etaHK
```
No external dependencies are required except the `go` compiler.

## Green Minibus (GMB)
The homepage is a list of minibus routes, filtered by selected region and sorted by route code.
Origins and destinations are shown with special remarks on different route variations.

After selecting the route a list of stops with the estimated time of arrival (ETA) would appear on the screen. The origin and destination would be shown at the top.

Commands available
* `f` for switching between stop list and fare table
* `r` for reversing the route direction (not available in circular / unidirectional routes)
* `Enter` for showing the list and ETA of minibuses of the routes of a particular stop.
  * Another route can be selected by pressing `Enter` on the selected route.
* `↑` / `↓` for changing the selected route / stop.

### Source
[DATA.gov.hk](https://data.gov.hk/en-data/dataset/hk-td-sm_7-real-time-arrival-data-of-gmb)

## TO-DO
Bus arrival times and fares
* KMB / LWB
* CTB
* NLB
