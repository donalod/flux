package testdata_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102,load1,system,host.local
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,long
#group,false,false,false,false,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,_newValue
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local,200
,,0,2018-05-22T19:53:36Z,101,load1,system,host.local,202
,,0,2018-05-22T19:53:46Z,102,load1,system,host.local,204
"

testcase map {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-15T00:00:00Z)
            |> drop(columns: ["_start", "_stop"])
            |> map(fn: (r) => ({r with _newValue: 2 * r._value}))
    want = csv.from(csv: outData)

    testing.diff(got, want)
}// Equivalent TICKscript query:
// stream
//  |eval(lambda: 2 * _value)
//.as('_newValue')
