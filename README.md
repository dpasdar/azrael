# Azrael Process Manager
As :angel: of Death(pun intended), the following small golang program watches over a list of processes, create and kill them gracefully when their time is up! The config is listed in a small sqlite database. 

The sqllite config database has only one table (jobs) and the following simple structure:
 
 
 jobs>
 - command : the command to be executed without the arguments
 - args : all the arguments of the command (e.g. `--id=1 --time=2`)
 - freq : how many seconds to wait before restarting the process
 - acceptable_runtime : how many seconds is acceptable for the process to be working, after this number of seconds azrael will kill the child process and restart it
 - status : either running or stopped
 - last_run: date and time of the last successful run
 
 
 To compile simply install go and run the following command inside the directory
 `go install`
 
 The binary will be located in default `$GOPATH/bin`
 
 The program will read the jobs table from the sqlite config, and for every row creates a subprocess which will the handle the creation and termination of the commands.
 
 To configure and access the config database run the following:
 
 ```
 sqlite3
 .open config
 select * from jobs;
 .quit
 ```
 
 As an example of long running process to play around with, use the following rand program :
 [Rand](https://github.com/dpasdar/rand)