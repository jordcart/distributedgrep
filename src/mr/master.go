package mr

import "log"
import "net"
import "os"
import "fmt"
import "sync"
import "errors"
import "net/rpc"
import "net/http"


type MapTask struct {
	filename   string
	// state types; too lazy to make enum rn
	// "state", "in-progress", "finished"
	state      string 
	taskNumber int
}

type ReduceTask struct {
	files       []string
	state       string
	taskNumber  int
}

type Master struct {
	mapTasks        []MapTask
	reduceTasks     []ReduceTask
	nReduce         int
	MapTaskCount    int
	MapTasksDone    int
	ReduceTaskCount int
	ReduceTasksDone int
	mu              sync.Mutex
}

func (m *Master) GetTask(args *ExampleArgs, reply *RequestTaskReply) error {
	reply.NReduce = m.nReduce

	// only send reduce tasks once all map tasks done
	if m.MapTasksDone != len(m.mapTasks) {
		for i := 0; i < len(m.mapTasks); i++ {
			currentMapTask := m.mapTasks[i]
			if currentMapTask.state == "idle" {
				reply.Filename = currentMapTask.filename
				reply.TaskType = "map"
				reply.TaskNumber = i

				// change state of task
				m.mu.Lock()
				m.mapTasks[i].state = "in-progress"
				m.MapTaskCount++
				m.mu.Unlock()
				reply.CurMapIndex = m.MapTaskCount
				break
			}
		}
	} else {
		for i := 0; i < len(m.reduceTasks); i++ {
			currentReduceTask := m.reduceTasks[i]
			if currentReduceTask.state == "idle" {
				reply.ReduceFileList = currentReduceTask.files
				reply.TaskType = "reduce"
				reply.TaskNumber = i

				// change state of task
				m.mu.Lock()
				m.reduceTasks[i].state = "in-progress"
				m.ReduceTaskCount++
				m.mu.Unlock()
				reply.CurReduceIndex = m.ReduceTaskCount
				break
			}
		}
	}

	if reply.Filename == "" && reply.TaskType == "" {
		return errors.New("No tasks") 
	}

	return nil
}

func (m *Master) ReportFinishedTask(args *FinishedTaskArgs, reply *ExampleReply) error {

	if args.TaskType == "map" {
		m.mu.Lock()
		for i, file := range args.FilesArray {
			fmt.Println(i, file)
			m.reduceTasks[i].files = append(m.reduceTasks[i].files, file) 
		}

		m.MapTasksDone++
		m.mapTasks[args.TaskNumber].state = "finished"
		m.mu.Unlock()
	} else if args.TaskType == "reduce" {
		m.mu.Lock()
		m.ReduceTasksDone++
		m.reduceTasks[args.TaskNumber].state = "finished"
		m.mu.Unlock()
	}

	return nil
}

func (m *Master) taskTimer() {
	// todo
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	if m.ReduceTasksDone == len(m.reduceTasks) && m.MapTasksDone == len(m.mapTasks) {
		ret = true
	}

	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}
	m.mapTasks = make([]MapTask, 0, len(files))
	m.reduceTasks = make([]ReduceTask, nReduce)
	m.nReduce = nReduce

	for i := 0; i < len(files); i++ {
		m.mapTasks = append(m.mapTasks, MapTask{files[i], "idle", i})
	}

	for i := 0; i < nReduce; i++ {
		// populate Reduce Tasks
		m.reduceTasks[i] = ReduceTask{make([]string, 0, len(files)), "idle", i}
	}

	m.server()
	return &m
}
