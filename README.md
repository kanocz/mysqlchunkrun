# mysqlchunkrun
Small tool to run DELETE or UPDATE queries in small portions (with LIMIT)

# usage
something like this:
```
mysqlchunkrun \
    -host "localhost" \
    -user "user" \
    -password "mysqlpassword" \
    -database "stats" \
    "DELETE FROM stats5 WHERE ts < '2016-01-01 00:00:00' LIMIT 100000"
```

this will run `DELETE FROM stats5 WHERE ts < '2016-01-01 00:00:00' LIMIT 100000` until no more rows affected
