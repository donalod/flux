package date_test

import "testing"
import "date"
import "array"

testcase  month_start_one_test {
    ret = date.monthStart(d:2021-03-20T20:20:11Z)

    want = array.from(rows: [{_value_a: 2021-03-01T00:00:00Z, _value_b: 2021-04-01T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}

testcase month_start_two_test {
    ret = date.monthStart(d:2021-01-01T01:01:44Z)

    want = array.from(rows: [{_value_a: 2021-01-01T00:00:00Z, _value_b: 2021-02-01T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop }])
    
    testing.diff(want:want, got: got)
}