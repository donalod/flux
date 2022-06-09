package date_test



import "testing"
import "date"
import "array"

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