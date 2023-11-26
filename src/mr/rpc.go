package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type StringReply struct {
	S string
}

type TaskArgs struct {
	Maptasknum    int // map task number only
	Reducetasknum int // reduce task number only
}

type TaskReply struct {
	TaskType string // map or reduce or waiting
	FileName string
	NReduce  int // number of reduce tasks
	NMap     int // number of map tasks

	Maptasknum    int // map task number only
	Reducetasknum int // reduce task number only

}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
