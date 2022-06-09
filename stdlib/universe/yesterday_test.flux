package universe_test


import "array"
import "testing"

testcase yesterday_one_test {
    option now = () => 2022-06-01T12:20:11Z
    ret = yesterday()
    
    want = array.from(rows: [{_value_a: 2022-05-31T00:00:00Z, _value_b:2022-06-01T00:00:00Z }])
    got = array.from(rows: [{_value_a:  ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}

testcase yesterday_two_test {
    option now = () => 2018-10-12T14:20:11Z
    ret = yesterday()

    want = array.from(rows: [{_value_a: 2018-10-11T00:00:00Z, _value_b: 2018-10-12T00:00:00Z}])
    got = array.from(rows: [{_value_a:  ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}