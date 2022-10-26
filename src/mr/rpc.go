package mr

//
// RPC definitions.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type RequestTaskArgs struct {
	X int
}

type RequestTaskReply struct {
	Filename       string
	ReduceFileList []string
	TaskType       string
	Pattern        string
	TaskNumber     int
	NReduce        int
}

type FinishedTaskArgs struct {
	FilesArray   []string
	TaskNumber   int
	TaskType     string
}

/*
type SendReduceFileReply {

}
*/

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
