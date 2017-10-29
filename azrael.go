package main

import (
	"database/sql"
	"os/exec"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strconv"
	"fmt"
	"os/signal"
	"syscall"
	"time"
	"github.com/shirou/gopsutil/process"
	"strings"
	"log"
	"io"
)

func main() {

	sig := make(chan bool, 1)
	allProcs := make(map[string]*process.Process)
	go ProcessSignal(sig)
	db, err := sql.Open("sqlite3", "./config")
	checkErr(err)
	rows, err := db.Query("SELECT ROWID, command, args, freq, status, last_run, acceptable_runtime FROM jobs")
	defer rows.Close()
	var rowid string
	var command sql.NullString
	var args sql.NullString
	var freq sql.NullString
	var status sql.NullString
	var last_run sql.NullString
	var acceptable_runtime sql.NullString

	for rows.Next() {
		err = rows.Scan(&rowid, &command, &args, &freq, &status, &last_run, &acceptable_runtime)
		checkErr(err)
		go func(sig chan bool, rowid string, command string, args string, freq string, acceptable_runtime string) {
			logger("debug","Assigned process id #%s to command %s with args %s", rowid, command, args)
			myProcess := "--id=agares" + rowid
			for {
				//all processing done here
				proc := findProcess(myProcess)
				if (proc == nil) {
					waitTime, _ := strconv.Atoi(freq)
					logger("debug","Process #%s waiting %s seconds for its turn", rowid, freq)
					time.Sleep(time.Second * time.Duration(waitTime))
					logger("debug","Start spawning a new process for #%s", rowid)
					cmd := exec.Command(command, args, myProcess)
					cmdError, _ := cmd.StderrPipe()
					go ErrorPipeToPrint(rowid, cmdError)
					cmd.Start()

					db.Exec("UPDATE jobs set status = 'running', last_run = date('now') where ROWID = " + rowid)
					proc = findProcess(myProcess)
					if (proc != nil) {
						allProcs[rowid] = proc
					}
				} else {
					procCreateTime,_ := proc.CreateTime()
					acceptableRuntime,_ := strconv.Atoi(acceptable_runtime)
					logger("debug","Process #%s has been running for %d seconds", rowid, (time.Now().Unix()*1000 - procCreateTime)/1000)
					if (time.Now().Unix()*1000 - procCreateTime > int64(acceptableRuntime)*1000) {
						logger("debug","Process %s is probably hanged, will try to kill it gracefully ", rowid)
						KillGracefully(proc)
						proc = nil
					}
				}
				time.Sleep(time.Second * 1)
			}

		}(sig, rowid, command.String, args.String, freq.String, acceptable_runtime.String)

	}

	// main loop
	for {
		select {
		case <-sig:
			log.Println("Cleaning up...")
			for pid, proc := range allProcs {
				logger("debug","Killing Child #%s", pid)
				proc.Kill()
			}
			os.Exit(0)

		}
	}

}
func logger(level string, message string, v ...interface{}) {
	switch(level) {
	case "error":
		log.SetOutput(os.Stderr)
	case "debug":
		log.SetOutput(os.Stdout)
	}
	log.Printf(message, v...)
}
func KillGracefully(proc *process.Process) {
	proc.Terminate()
	time.Sleep(time.Second * 20)
	proc.Kill()
}
func findProcess(processName string) *process.Process {
	//var out, _ = exec.Command(fmt.Sprintf("ps aux | grep '%s' | grep -v grep | awk 'NR==1{print $2}'", processName)).Output()
	out, err := exec.Command("bash", "-c",
		fmt.Sprintf("ps aux | grep '%s' | grep -v grep | awk 'NR==1{print $2}'",
			strings.Replace(processName, "-", "\\-", -1))).Output()
	checkErr(err)
	if len(out) == 0 {
		return nil
	}
	var pidInt, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	var proc, _ = process.NewProcess(int32(pidInt))
	return proc
}
func ProcessSignal(sig chan bool) {
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	for {
		signalType := <-sigch
		fmt.Println("Received signal from channel : ", signalType)

		switch signalType {
		case syscall.SIGINT, syscall.SIGTERM:
			fmt.Println("got Termination signal")
			sig <- true
			break
		}
	}
}
func ErrorPipeToPrint(rowid string, r io.ReadCloser) {
	defer r.Close()

	for {
		b := make([]byte,100)
		_, err := r.Read(b)
		switch {
		case err == io.EOF:
			return
		case err != nil:
			fmt.Printf("READ_ERROR:%v\n", err)
			return
		}
		logger("error","Process #%s had errors: %s", rowid, b)

	}

}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
