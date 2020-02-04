# cloud_watcher
check cloudwatch log group retention dates, and collect in a google sheet
super P.O.C

`dep ensure`
`go run go.go`

you will need to authorise your google sheets API stuff and get a spreadsheetID from a sheet
then the `go` `aws` sdk will (if you have your life sorted out credentials-wise) will get all your cloudwatch loggroups
and add the groups that have `Never Expire` 'set' on to the google sheet. it'll even get tags if you added them, it even
might show you what is terraformed. but who knows. 


TO:DO 
docker/lambdaize
make cool
eat rice
