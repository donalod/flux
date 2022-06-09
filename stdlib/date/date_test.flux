package date_test


import "date"
import "testing"
import "array"

testcase date_timeable {
    option now = () => 2021-03-01T00:00:00Z

    want = array.from(rows: [{_value: 2}])
    got = array.from(rows: [{_value: date.month(t: -1mo)}])

    testing.diff(want: want, got: got)
}

testcase week_start_default_sunday_test {
    ret = date.weekStart(d:2022-06-08T12:20:11Z, start_sunday: true)

    want = array.from(rows: [{_value_a: 2022-06-05T00:00:00Z, _value_b: 2022-06-12T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}

testcase week_start_default_monday_test {
    ret = date.weekStart(d:2022-06-08T12:20:11Z, start_sunday:false)

    want = array.from(rows: [{_value_a: 2022-06-06T00:00:00Z, _value_b: 2022-06-13T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}


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

testcase  monday_test_one_timeable {
    ret = date.monday(d:2021-03-06T00:20:11Z)

    want = array.from(rows: [{_value_a: 2021-03-01T00:00:00Z, _value_b: 2021-03-02T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop  }])

    testing.diff(want:want, got: got)
}

testcase monday_test_two_timeable {
    ret = date.monday(d: 2022-01-01T00:20:11Z)

    want = array.from(rows: [{_value_a: 2021-12-27T00:00:00Z, _value_b:2021-12-28T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}
