package universe_test


import "array"
import "testing"

testcase yesterday_test {
    option now = () => 2022-06-01T12:20:11Z
    ret = yesterday()
    
    want = array.from(rows: [{_value_a: 2022-05-31T00:00:00Z, _value_b:2022-06-01T00:00:00Z }])
    got = array.from(rows: [{_value_a:  ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}

testcase yesterday_test {
    option now = () => 2018-10-12T14:20:11Z
    ret = yesterday()

    want = array.from(rows: [{_value_a: 2018-10-11T00:00:00Z, _value_b: 2018-10-12T00:00:00Z}])
    got = array.from(rows: [{_value_a:  ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}


testcase week_start_default_sunday_test {
    option now = () => 2022-06-06T10:02:10Z
    ret = weekStart(start_sunday: true)

    want = array.from(rows: [{_value_a: 2022-06-05T00:00:00Z, _value_b: 2022-06-12T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}

testcase week_start_default_monday_test {
    option now = () => 2022-06-08T14:20:11Z
    ret = weekStart(start_sunday:false)

    want = array.from(rows: [{_value_a: 2022-06-06T00:00:00Z, _value_b: 2022-06-13T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}


testcase  month_start_one_test {
    option now = () => 2021-03-10T22:10:00Z
    ret = monthStart()

    want = array.from(rows: [{_value_a: 2021-03-01T00:00:00Z, _value_b: 2021-04-01T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}

testcase month_start_two_test {
    option now = () => 2020-12-10T22:10:00Z
    ret = monthStart()

    want = array.from(rows: [{_value_a: 2020-12-01T00:00:00Z, _value_b: 2021-01-01T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop }])
    
    testing.diff(want:want, got: got)
}

testcase  monday_test_timeable {
    option now = () => 2021-03-05T12:10:11Z
    ret = monday()

    want = array.from(rows: [{_value_a: 2021-03-01T00:00:00Z, _value_b: 2021-03-02T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop  }])

    testing.diff(want:want, got: got)
}

testcase monday_test_timeable {
    option now = () => 2021-12-30T00:40:44Z
    ret = monday()

    want = array.from(rows: [{_value_a: 2021-12-27T00:00:00Z, _value_b:2021-12-28T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}


testcase  tuesday_test_timeable {
    option now = () => 2021-03-05T12:10:11Z
    ret = tuesday()

    want = array.from(rows: [{_value_a: 2021-03-02T00:00:00Z, _value_b: 2021-03-03T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop  }])

    testing.diff(want:want, got: got)
}

testcase tuesday_test_timeable {
    option now = () => 2021-12-30T00:40:44Z
    ret = tuesday()

    want = array.from(rows: [{_value_a: 2021-12-28T00:00:00Z, _value_b:2021-12-29T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}


testcase wednesday_test_timeable {
    option now = () => 2022-01-01T00:40:44Z
    ret = wednesday()

    want = array.from(rows: [{_value_a: 2021-12-29T00:00:00Z, _value_b:2021-12-30T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}

testcase wednesday_test_timeable {
    option now = () => 2021-12-05T12:10:11Z
    ret = wednesday()

    want = array.from(rows: [{_value_a: 2021-12-01T00:00:00Z, _value_b:2021-12-02T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}





testcase thursday_test_timeable {
    option now = () => 2022-01-10T00:40:44Z
    ret = thursday()

    want = array.from(rows: [{_value_a: 2022-01-06T00:00:00Z, _value_b:2022-01-07T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}

testcase thursday_test_timeable {
    option now = () => 2022-01-21T12:10:11Z
    ret = thursday()

    want = array.from(rows: [{_value_a: 2022-01-20T00:00:00Z, _value_b:2022-01-21T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}

testcase friday_test_timeable {
    option now = () => 2022-01-23T12:10:11Z
    ret = friday()

    want = array.from(rows: [{_value_a: 2022-01-21T00:00:00Z, _value_b:2022-01-22T00:00:00Z }])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}

testcase friday_test_timeable {
    option now = () => 2022-01-03T12:10:11Z
    ret = friday()

    want = array.from(rows: [{_value_a: 2021-12-31T00:00:00Z, _value_b: 2022-01-01T00:00:00Z }])
    got = array.from( rows: [{_value_a: ret.start,  _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}

testcase friday_test_timeable { 
    option now = () => 2022-01-25T12:10:11Z
    ret = friday()

    want = array.from(rows: [{_value_a: 2022-01-21T00:00:00Z, _value_b: 2022-01-22T00:00:00Z }])
    got = array.from( rows: [{_value_a: ret.start,  _value_b: ret.stop}])

    testing.diff(want:want, got: got)
}


testcase saturday_test_timeable {
    option now = () => 2022-01-25T12:10:11Z
    ret = saturday()

    want = array.from(rows: [{_value_a: 2022-01-22T00:00:00Z, _value_b: 2022-01-23T00:00:00Z }])
    got = array.from( rows: [{_value_a: ret.start,  _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}

testcase saturday_test_timeable {
    option now = () => 2022-01-15T12:10:11Z
    ret = saturday()

    want = array.from(rows: [{_value_a: 2022-01-08T00:00:00Z, _value_b: 2022-01-09T00:00:00Z }])
    got = array.from( rows: [{_value_a: ret.start,  _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}

testcase sunday_test_timeable {
    option now = () => 2022-01-22T12:10:11Z
    ret = sunday()

    want = array.from(rows: [{_value_a: 2022-01-16T00:00:00Z, _value_b: 2022-01-17T00:00:00Z }])
    got = array.from( rows: [{_value_a: ret.start,  _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}

testcase sunday_test_timeable {
    option now = () => 2022-01-24T12:10:11Z
    ret = sunday()

    want = array.from(rows: [{_value_a: 2022-01-23T00:00:00Z, _value_b: 2022-01-24T00:00:00Z }])
    got = array.from( rows: [{_value_a: ret.start,  _value_b: ret.stop}])
    
    testing.diff(want:want, got: got)
}