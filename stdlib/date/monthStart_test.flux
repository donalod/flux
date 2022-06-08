package date_test

import "testing"
import "date"
import "array"

testcase  month_start_one_test {

    want = array.from(rows: [{_value: 2021-03-01T00:00:00Z}])
    got = array.from(rows: [{_value:  date.monthStart(d:2021-03-20T20:20:11Z)}])

    testing.diff(want:want, got: got)
}

testcase month_start_two_test {

    want = array.from(rows: [{_value: 2021-01-01T00:00:00Z }])
    got = array.from(rows: [{_value: date.monthStart(d:2022-01-01T00:00:00Z)}])
    
    testing.diff(want:want, got: got)
}