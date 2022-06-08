package date_test



import "testing"
import "date"
import "array"

testcase week_start_default_sunday_test {

    want = array.from(rows: [{_value: 2022-06-05T00:00:00Z}])
    got = array.from(rows: [{_value:  date.weekStart(d:2022-06-08T12:20:11Z, start_sunday: true)}])

    testing.diff(want:want, got: got)
}

testcase week_start_default_monday_test {

    want = array.from(rows: [{_value: 2022-06-06T00:00:00Z}])
    got = array.from(rows: [{_value:  date.weekStart(d:2022-06-08T12:20:11Z, start_sunday:false)}])

    testing.diff(want:want, got: got)
}



// testcase week_start_two_test {
//     option now = () => 2018-10-12T14:20:11Z

//     want = array.from(rows: [{_value: 2018-10-11T00:00:00Z}])
//     got = array.from(rows: [{_value:  yesterday()}])

//     testing.diff(want:want, got: got)
// }