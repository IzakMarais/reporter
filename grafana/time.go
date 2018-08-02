/*
   Copyright 2016 Vastech SA (PTY) LTD

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package grafana

import (
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"time"
)

type TimeRange struct {
	From string
	To   string
}

// Used to parse grafana time specifications. These can take various forms:
//	 * relative: "now", "now-1h", "now-2d", "now-3w", "now-5M", "now-1y"
//   * human friendly boundary:
// 			From:"now/d" -> start of today
//			To:  "now/d" -> end of today
//			To:  "now/w" -> end of the week
//			To:  "now-1d/d" -> end of yesterday
//			When used as boundary, the same string will evaluate to a different time if used in 'From' or 'To'
//	 * absolute unix time: "142321234"
//
// The required behaviour is clearly documented in the unit tests, time_test.go.
type now time.Time

type boundary int

const (
	From boundary = iota
	To
)

const (
	relTimeRegExp      = "^now([+-][0-9]+)([mhdwMy])$"
	boundaryTimeRegExp = "^(.*?)/([dwMy])$"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func NewTimeRange(from, to string) TimeRange {
	if from == "" {
		from = "now-1h"
	}
	if to == "" {
		to = "now"
	}
	return TimeRange{from, to}
}

// Formats Grafana 'From' time spec into absolute printable time
func (tr TimeRange) FromFormatted() string {
	n := newNow()
	return n.parseFrom(tr.From).Format(time.UnixDate)
}

// Formats Grafana 'To' time spec into absolute printable time
func (tr TimeRange) ToFormatted() string {
	n := newNow()
	return n.parseTo(tr.To).Format(time.UnixDate)
}

func newNow() now {
	return now(time.Now())
}

func (n now) asTime() time.Time {
	return time.Time(n)
}

func (n now) parseFrom(s string) time.Time {
	return n.parseHumanFriendlyBoundary(s, From)
}

func (n now) parseTo(s string) time.Time {
	return n.parseHumanFriendlyBoundary(s, To)
}

func (n now) parseHumanFriendlyBoundary(s string, b boundary) time.Time {
	if !isHumanFriendlyBoundray(s) {
		return n.parseMoment(s)
	} else {
		moment, boundaryUnit := n.parseMomentAndBoundaryUnit(s)
		return roundMomentToBoundary(moment, b, boundaryUnit)
	}
}

func (n now) parseMomentAndBoundaryUnit(s string) (time.Time, string) {
	re := regexp.MustCompile(boundaryTimeRegExp)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 3 {
		panic(unrecognized(s))
	}
	moment := n.parseMoment(matches[1])
	boundaryUnit := matches[2]
	return moment, boundaryUnit
}

func roundMomentToBoundary(moment time.Time, b boundary, boundaryUnit string) time.Time {
	y := moment.Year()
	M := moment.Month()
	d := moment.Day()

	switch boundaryUnit {
	case "d":
		d += add(b)
	case "w":
		d += daysToWeekBoundary(moment.Weekday(), b)
	case "M":
		d = 1
		M = time.Month(int(M) + add(b))
	case "y":
		d = 1
		M = time.January
		y += add(b)
	}

	return time.Date(y, M, d, 0, 0, 0, 0, moment.Location())
}

func add(b boundary) int {
	if b == To {
		return 1
	}
	// b == From
	return 0
}

func daysToWeekBoundary(wd time.Weekday, b boundary) int {
	if b == To {
		return 1 + int(time.Saturday) - int(wd)
	} else {
		//b == From
		return -int(wd)
	}
}

func (n now) parseMoment(s string) time.Time {
	if s == "now" {
		return n.asTime()
	} else if isRelativeTime(s) {
		return n.parseRelativeTime(s)
	} else {
		return parseAbsTime(s)
	}
}

func (n now) parseRelativeTime(s string) time.Time {
	re := regexp.MustCompile(relTimeRegExp)

	matches := re.FindStringSubmatch(s)
	if len(matches) != 3 {
		panic(unrecognized(s))
	}
	unit := matches[2]
	number := matches[1]

	i, err := strconv.Atoi(number)
	stopIf(err)

	switch unit {
	case "m", "h":
		d, err := time.ParseDuration(number + unit)
		stopIf(err)
		return n.asTime().Add(d)
	case "d":
		return n.asTime().AddDate(0, 0, i)
	case "w":
		return n.asTime().AddDate(0, 0, i*7)
	case "M":
		return n.asTime().AddDate(0, i, 0)
	case "y":
		return n.asTime().AddDate(i, 0, 0)
	}

	return n.asTime()
}

func parseAbsTime(s string) time.Time {
	if timeInMs, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(int64(timeInMs)/1000, 0)
	}

	panic(unrecognized(s))
}

func unrecognized(s string) string {
	return s + " is not a recognised time format"
}

func stopIf(err error) {
	if err != nil {
		panic(err)
	}
}

func isRelativeTime(s string) bool {
	matched, _ := regexp.MatchString(relTimeRegExp, s)
	return matched
}

func isHumanFriendlyBoundray(s string) bool {
	matched, _ := regexp.MatchString(boundaryTimeRegExp, s)
	return matched
}
