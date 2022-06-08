package universe_test


import "array"
import "testing"

testcase yesterday_one_test {
    option now = () => 2022-06-01T12:20:11Z

    want = array.from(rows: [{_value: 2022-05-31T00:00:00Z}])
    got = array.from(rows: [{_value:  yesterday()}])

    testing.diff(want:want, got: got)
}

testcase yesterday_two_test {
    option now = () => 2018-10-12T14:20:11Z

    want = array.from(rows: [{_value: 2018-10-11T00:00:00Z}])
    got = array.from(rows: [{_value:  yesterday()}])

    testing.diff(want:want, got: got)
}