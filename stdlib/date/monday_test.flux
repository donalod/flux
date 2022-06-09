package date_test

import "testing"
import "date"
import "array"

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