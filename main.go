package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbMaster = flag.String("master", "root:@tcp(localhost:3306)/?charset=utf8mb4", "connection string for master")
	dbSlave  = flag.String("slave", "", "connection string for slave")
	dbMaxLag = flag.Int("maxlag", 10, "maximum allowed slave lag")
	dbPause  = flag.Duration("pause", 10*time.Millisecond, "time to sleep between queries")
)

// return current master "replication position"
func getMasterPost(master *sql.DB) (string, uint64, error) {
	var (
		file     string
		pos      uint64
		doDb     string
		ignoreDb string
		gtid     string
	)
	err := master.QueryRow("SHOW MASTER STATUS").Scan(&file, &pos, &doDb, &ignoreDb, &gtid)
	if nil != err {
		return "", 0, err
	}
	return file, pos, nil
}

// wait until slave processes all statements up to specified position
func waitForSlave(slave *sql.DB, file string, post uint64) error {
	// wait up to 1 hour in loop

	for {
		var result sql.NullInt64
		err := slave.QueryRow("SELECT MASTER_POS_WAIT(?, ?, ?)", file, post, 3600).Scan(&result)
		//		err := slave.QueryRow(fmt.Sprintf("MASTER_POS_WAIT('%s', %d, 3600)", file, post)).Scan(&result)
		if nil != err {
			return err
		}
		if !result.Valid {
			return errors.New("Replication error")
		}
		if result.Int64 != -1 {
			return nil
		}
	}
}

func getSlaveLag(slave *sql.DB) (int, error) {
	var slaveStatus struct {
		SlaveIOState              *sql.RawBytes
		MasterHost                *sql.RawBytes
		MasterUser                *sql.RawBytes
		MasterPort                *sql.RawBytes
		ConnectRetry              *sql.RawBytes
		MasterLogFile             *sql.RawBytes
		ReadMasterLogPos          *sql.RawBytes
		RelayLogFile              *sql.RawBytes
		RelayLogPos               *sql.RawBytes
		RelayMasterLogFile        *sql.RawBytes
		SlaveIORunning            string
		SlaveSQLRunning           string
		ReplicateDoDB             *sql.RawBytes
		ReplicateIgnoreDB         *sql.RawBytes
		ReplicateDoTable          *sql.RawBytes
		ReplicateIgnoreTable      *sql.RawBytes
		ReplicateWildDoTable      *sql.RawBytes
		ReplicateWildIgnoreTable  *sql.RawBytes
		LastErrno                 *sql.RawBytes
		LastError                 *sql.RawBytes
		SkipCounter               *sql.RawBytes
		ExecMasterLogPos          *sql.RawBytes
		RelayLogSpace             *sql.RawBytes
		UntilCondition            *sql.RawBytes
		UntilLogFile              *sql.RawBytes
		UntilLogPos               *sql.RawBytes
		MasterSSLAllowed          *sql.RawBytes
		MasterSSLCAFile           *sql.RawBytes
		MasterSSLCAPath           *sql.RawBytes
		MasterSSLCert             *sql.RawBytes
		MasterSSLCipher           *sql.RawBytes
		MasterSSLKey              *sql.RawBytes
		SecondsBehindMaster       int
		MasterSSLVerifyServerCert *sql.RawBytes
		LastIOErrno               *sql.RawBytes
		LastIOError               *sql.RawBytes
		LastSQLErrno              *sql.RawBytes
		LastSQLError              *sql.RawBytes
		ReplicateIgnoreServerIds  *sql.RawBytes
		MasterServerID            *sql.RawBytes
		MasterUUID                *sql.RawBytes
		MasterInfoFile            *sql.RawBytes
		SQLDelay                  *sql.RawBytes
		SQLRemainingDelay         *sql.RawBytes
		SlaveSQLRunningState      *sql.RawBytes
		MasterRetryCount          *sql.RawBytes
		MasterBind                *sql.RawBytes
		LastIOErrorTimestamp      *sql.RawBytes
		LastSQLErrorTimestamp     *sql.RawBytes
		MasterSSLCrl              *sql.RawBytes
		MasterSSLCrlpath          *sql.RawBytes
		RetrievedGtidSet          *sql.RawBytes
		ExecutedGtidSet           *sql.RawBytes
		AutoPosition              *sql.RawBytes
	}

	err := slave.QueryRow("SHOW SLAVE STATUS").Scan(&slaveStatus.SlaveIOState, &slaveStatus.MasterHost, &slaveStatus.MasterUser, &slaveStatus.MasterPort, &slaveStatus.ConnectRetry, &slaveStatus.MasterLogFile, &slaveStatus.ReadMasterLogPos, &slaveStatus.RelayLogFile, &slaveStatus.RelayLogPos, &slaveStatus.RelayMasterLogFile, &slaveStatus.SlaveIORunning, &slaveStatus.SlaveSQLRunning, &slaveStatus.ReplicateDoDB, &slaveStatus.ReplicateIgnoreDB, &slaveStatus.ReplicateDoTable, &slaveStatus.ReplicateIgnoreTable, &slaveStatus.ReplicateWildDoTable, &slaveStatus.ReplicateWildIgnoreTable, &slaveStatus.LastErrno, &slaveStatus.LastError, &slaveStatus.SkipCounter, &slaveStatus.ExecMasterLogPos, &slaveStatus.RelayLogSpace, &slaveStatus.UntilCondition, &slaveStatus.UntilLogFile, &slaveStatus.UntilLogPos, &slaveStatus.MasterSSLAllowed, &slaveStatus.MasterSSLCAFile, &slaveStatus.MasterSSLCAPath, &slaveStatus.MasterSSLCert, &slaveStatus.MasterSSLCipher, &slaveStatus.MasterSSLKey, &slaveStatus.SecondsBehindMaster, &slaveStatus.MasterSSLVerifyServerCert, &slaveStatus.LastIOErrno, &slaveStatus.LastIOError, &slaveStatus.LastSQLErrno, &slaveStatus.LastSQLError, &slaveStatus.ReplicateIgnoreServerIds, &slaveStatus.MasterServerID, &slaveStatus.MasterUUID, &slaveStatus.MasterInfoFile, &slaveStatus.SQLDelay, &slaveStatus.SQLRemainingDelay, &slaveStatus.SlaveSQLRunningState, &slaveStatus.MasterRetryCount, &slaveStatus.MasterBind, &slaveStatus.LastIOErrorTimestamp, &slaveStatus.LastSQLErrorTimestamp, &slaveStatus.MasterSSLCrl, &slaveStatus.MasterSSLCrlpath, &slaveStatus.RetrievedGtidSet, &slaveStatus.ExecutedGtidSet, &slaveStatus.AutoPosition)

	if sql.ErrNoRows == err {
		return 0, errors.New("No slave running")
	}

	if nil != err {
		return 0, err
	}

	if slaveStatus.SlaveIORunning != "Yes" {
		return 0, errors.New("Slave IO not running")
	}

	if slaveStatus.SlaveSQLRunning != "Yes" {
		return 0, errors.New("Slave SQL not running")
	}

	return slaveStatus.SecondsBehindMaster, nil
}

func runOneQuery(master *sql.DB, slave *sql.DB, query string) {
	stmt, err := master.Prepare(query)
	if nil != err {
		log.Fatalln("Error preparing query: ", err)
	}

	for {
		res, err := stmt.Exec()
		if nil != err {
			log.Fatalln("Error executing query: ", err)
		}

		if rows, err := res.RowsAffected(); 0 == rows {
			if nil != err {
				log.Fatalln("Error on RowsAffected(): ", err)
			}
			fmt.Println("0 affected rows, probably done")
			break
		}

		logFile, logPos, err := getMasterPost(master)
		if nil != err {
			log.Fatalln("Error reading master position: ", err)
		}

		err = waitForSlave(slave, logFile, logPos)
		if nil != err {
			log.Fatalln("Error waiting for slave: ", err)
		}

		for {
			lag, err := getSlaveLag(slave)
			if nil != err {
				log.Fatalln("get lag error: ", err)
			}
			if lag <= *dbMaxLag {
				break
			}
			time.Sleep(time.Second)
		}

		os.Stdout.Write([]byte("."))
		time.Sleep(*dbPause)
	}
}

func main() {

	flag.Parse()

	master, err := sql.Open("mysql", *dbMaster)
	if nil != err {
		log.Fatalln("Error connecting to database (master): ", err)
	}
	defer master.Close()

	slave, err := sql.Open("mysql", *dbSlave)
	if nil != err {
		log.Fatalln("Error connecting to database (slave): ", err)
	}
	defer slave.Close()

	query := flag.Arg(0)
	if "" == query {
		fmt.Println("Please specify query as last argument")
		return
	}

	runOneQuery(master, slave, query)

}
