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
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func sameTimeAs(actual interface{}, expected ...interface{}) string {
	if actual == expected[0] {
		return ""
	} else {
		a := actual.(time.Time)
		e := expected[0].(time.Time)
		return fmt.Sprintf("Times differ:\n\t Actual: %q\n\tExpected: %q\n", a, e)
	}
}

func TestTimeParsing(tst *testing.T) {
	testNow, _ := time.Parse(time.RFC1123, "Wed, 06 Jan 2016 16:34:32 UTC")
	t := now(testNow)

	Convey("When parsing relative time", tst, func() {
		Convey("'now' should return the time it was initialised with", func() {
			So(t.parseTo("now"), sameTimeAs, testNow)
		})

		Convey("Minutes are supported", func() {
			d, _ := time.ParseDuration("-1m")
			So(t.parseTo("now-1m"), sameTimeAs, testNow.Add(d))

			d, _ = time.ParseDuration("-58m")
			So(t.parseTo("now-58m"), sameTimeAs, testNow.Add(d))
		})

		Convey("Positive relative time is supported", func() {
			d, _ := time.ParseDuration("+1m")
			So(t.parseTo("now+1m"), sameTimeAs, testNow.Add(d))

			d, _ = time.ParseDuration("+58m")
			So(t.parseTo("now+58m"), sameTimeAs, testNow.Add(d))
		})

		Convey("Hours are supported", func() {
			d, _ := time.ParseDuration("-3h")
			So(t.parseTo("now-3h"), sameTimeAs, testNow.Add(d))

			d, _ = time.ParseDuration("-82h")
			So(t.parseTo("now-82h"), sameTimeAs, testNow.Add(d))
		})

		Convey("Days are supported", func() {
			So(t.parseTo("now-1d"), sameTimeAs, testNow.AddDate(0, 0, -1))
			So(t.parseTo("now-105d"), sameTimeAs, testNow.AddDate(0, 0, -105))
		})

		Convey("Weeks are supported", func() {
			So(t.parseTo("now-1w"), sameTimeAs, testNow.AddDate(0, 0, -1*7))
			So(t.parseTo("now-33w"), sameTimeAs, testNow.AddDate(0, 0, -33*7))
		})

		Convey("Months are supported", func() {
			So(t.parseTo("now-1M"), sameTimeAs, testNow.AddDate(0, -1, 0))
			So(t.parseTo("now-33M"), sameTimeAs, testNow.AddDate(0, -33, 0))
		})

		Convey("Years are supported", func() {
			So(t.parseTo("now-1y"), sameTimeAs, testNow.AddDate(-1, 0, 0))
			So(t.parseTo("now-33y"), sameTimeAs, testNow.AddDate(-33, 0, 0))
		})

	})

	//?from=1463464226537&to=1463472462258
	Convey("Should be able to parse absolute time ", tst, func() {
		So(t.parseTo("1463464226537"), sameTimeAs, time.Unix(1463464226537/1000, 0))
	})

	Convey("Should panic on accept unrecognised formats", tst, func() {
		So(func() { t.parseTo("not-a-time") }, ShouldPanic)
		So(func() { t.parseTo("now-43k") }, ShouldPanic)
		So(func() { t.parseTo("1235032k") }, ShouldPanic)
	})

	Convey("When parsing human frienly start time boundaries, parseFrom()", tst, func() {
		Convey("Should return the same time as parseTo() if boundary specifier ('/') is missing", func() {
			So(t.parseFrom("now"), sameTimeAs, t.parseTo("now"))
			So(t.parseFrom("now-3M"), sameTimeAs, t.parseTo("now-3M"))
			So(t.parseFrom("14123456789"), sameTimeAs, t.parseTo("14123456789"))
		})

		//now = Wed, 06 Jan 2016 16:34:32 UTC
		Convey("Should support days", func() {
			startOfTheDay, _ := time.Parse(time.RFC1123, "Wed, 06 Jan 2016 00:00:00 UTC")
			So(t.parseFrom("now/d"), sameTimeAs, startOfTheDay)
			So(t.parseFrom("now-1m/d"), sameTimeAs, startOfTheDay)
			So(t.parseFrom("now-72m/d"), sameTimeAs, startOfTheDay)

			startOfYesterday, _ := time.Parse(time.RFC1123, "Tue, 05 Jan 2016 00:00:00 UTC")
			So(t.parseFrom("now-1d/d"), sameTimeAs, startOfYesterday)
			So(t.parseFrom("now-24h/d"), sameTimeAs, startOfYesterday)
		})

		Convey("Should support weeks", func() {
			startOfTheWeek, _ := time.Parse(time.RFC1123, "Sun, 03 Jan 2016 00:00:00 UTC")
			So(t.parseFrom("now/w"), sameTimeAs, startOfTheWeek)
			So(t.parseFrom("now-82m/w"), sameTimeAs, startOfTheWeek)
			So(t.parseFrom("now-33h/w"), sameTimeAs, startOfTheWeek)
			So(t.parseFrom("now-2d/w"), sameTimeAs, startOfTheWeek)

			startOfLastWeek, _ := time.Parse(time.RFC1123, "Sun, 27 Dec 2015 00:00:00 UTC")
			So(t.parseFrom("now-1w/w"), sameTimeAs, startOfLastWeek)
		})

		Convey("Should support months", func() {
			startOfTheMonth, _ := time.Parse(time.RFC1123, "Fri, 01 Jan 2016 00:00:00 UTC")
			So(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC), sameTimeAs, startOfTheMonth)

			So(t.parseFrom("now/M"), sameTimeAs, startOfTheMonth)
			So(t.parseFrom("now-82m/M"), sameTimeAs, startOfTheMonth)
			So(t.parseFrom("now-33h/M"), sameTimeAs, startOfTheMonth)
			So(t.parseFrom("now-2d/M"), sameTimeAs, startOfTheMonth)

			startOfLastMonth, _ := time.Parse(time.RFC1123, "Tue, 01 Dec 2015 00:00:00 UTC")
			So(t.parseFrom("now-1M/M"), sameTimeAs, startOfLastMonth)
		})

		Convey("Should support years", func() {
			startOfTheYear, _ := time.Parse(time.RFC1123, "Fri, 01 Jan 2016 00:00:00 UTC")
			So(t.parseFrom("now/y"), sameTimeAs, startOfTheYear)
			So(t.parseFrom("now-82m/y"), sameTimeAs, startOfTheYear)
			So(t.parseFrom("now-33h/y"), sameTimeAs, startOfTheYear)
			So(t.parseFrom("now-2d/y"), sameTimeAs, startOfTheYear)

			startOfLastYear, _ := time.Parse(time.RFC1123, "Thu, 01 Jan 2015 00:00:00 UTC")
			So(t.parseFrom("now-1y/y"), sameTimeAs, startOfLastYear)
		})

	})

	Convey("When parsing human frienly end time boundaries, parseTo()", tst, func() {
		//now = Wed, 06 Jan 2016 16:34:32 UTC
		Convey("Should support days", func() {
			endOfToday, _ := time.Parse(time.RFC1123, "Thu, 07 Jan 2016 00:00:00 UTC")
			So(t.parseTo("now/d"), sameTimeAs, endOfToday)
			So(t.parseTo("now-1m/d"), sameTimeAs, endOfToday)
			So(t.parseTo("now-72m/d"), sameTimeAs, endOfToday)

			endOfYesterday, _ := time.Parse(time.RFC1123, "Wed, 06 Jan 2016 00:00:00 UTC")
			So(t.parseTo("now-1d/d"), sameTimeAs, endOfYesterday)
		})

		Convey("Should support weeks", func() {
			endOfTheWeek, _ := time.Parse(time.RFC1123, "Sun, 10 Jan 2016 00:00:00 UTC")
			So(t.parseTo("now/w"), sameTimeAs, endOfTheWeek)
			So(t.parseTo("now-82m/w"), sameTimeAs, endOfTheWeek)
			So(t.parseTo("now-33h/w"), sameTimeAs, endOfTheWeek)
			So(t.parseTo("now-2d/w"), sameTimeAs, endOfTheWeek)

			endOfLastWeek, _ := time.Parse(time.RFC1123, "Sun, 03 Jan 2016 00:00:00 UTC")
			So(t.parseTo("now-1w/w"), sameTimeAs, endOfLastWeek)
		})

		Convey("Should support months", func() {
			endOfTheMonth, _ := time.Parse(time.RFC1123, "Mon, 01 Feb 2016 00:00:00 UTC")
			So(t.parseTo("now/M"), sameTimeAs, endOfTheMonth)
			So(t.parseTo("now-82m/M"), sameTimeAs, endOfTheMonth)
			So(t.parseTo("now-33h/M"), sameTimeAs, endOfTheMonth)
			So(t.parseTo("now-2d/M"), sameTimeAs, endOfTheMonth)

			endOfLastMonth, _ := time.Parse(time.RFC1123, "Fri, 01 Jan 2016 00:00:00 UTC")
			So(t.parseTo("now-1M/M"), sameTimeAs, endOfLastMonth)
		})

		Convey("Should support years", func() {
			endOfTheYear, _ := time.Parse(time.RFC1123, "Sun, 01 Jan 2017 00:00:00 UTC")
			So(t.parseTo("now/y"), sameTimeAs, endOfTheYear)
			So(t.parseTo("now-82m/y"), sameTimeAs, endOfTheYear)
			So(t.parseTo("now-33h/y"), sameTimeAs, endOfTheYear)
			So(t.parseTo("now-2d/y"), sameTimeAs, endOfTheYear)

			endOfLastYear, _ := time.Parse(time.RFC1123, "Fri, 01 Jan 2016 00:00:00 UTC")
			So(t.parseTo("now-1y/y"), sameTimeAs, endOfLastYear)
		})

	})
}
