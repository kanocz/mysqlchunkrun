# mysqlchunkrun
Small tool to run DELETE or UPDATE queries in small portions (with LIMIT) and controls replication lag

# usage
something like this:
```
mysqlchunkrun \
    -master 'root:password@tcp(masterhost:3306)/stats?charset=utf8mb4' \
    -slave 'root:password@tcp(slavehost:3306)/?charset=utf8mb4' \
    -maxlag 5 \
    "DELETE FROM stats5 WHERE ts < '2016-01-01 00:00:00' LIMIT 100000"
```

this will run `DELETE FROM stats5 WHERE ts < '2016-01-01 00:00:00' LIMIT 100000` until no more rows affected, checking that slave is synced with master after each chunk (master_pos_wait + control seconds_behind_master value)
