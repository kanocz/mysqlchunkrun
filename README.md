# mysqlchunkrun
Small tool to run DELETE or UPDATE queries in small portions (with LIMIT) and controls replication lag
OR run list of queries with control of replication lag after each of it

# usage
something like this:
```
mysqlchunkrun \
    -master 'root:password@tcp(masterhost:3306)/stats?charset=utf8mb4' \
    -slave 'root:password@tcp(slavehost:3306)/?charset=utf8mb4' \
    -maxlag 5 \
    -pause 0s \
    "DELETE FROM stats5 WHERE ts < '2016-01-01 00:00:00' LIMIT 100000"
```

this will run `DELETE FROM stats5 WHERE ts < '2016-01-01 00:00:00' LIMIT 100000` until no more rows affected, checking that slave is synced with master after each chunk (master_pos_wait + control seconds_behind_master value)

or

```
mysqlchunkrun \
    -master 'root:password@tcp(masterhost:3306)/stats?charset=utf8mb4' \
    -slave 'root:password@tcp(slavehost:3306)/?charset=utf8mb4' \
    -maxlag 5 \
    -pause 1s \
    -bash list.sql
```

this will run queries from `list.sql` one by one, checking that slave is synced with master after each query (master_pos_wait + control seconds_behind_master value) with pause 1s after each query
