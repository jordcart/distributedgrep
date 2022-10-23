package mr

import "fmt"
import "log"
import "os"
import "time"
import "strconv"
import "sort"
import "path/filepath"
import "io/ioutil"
import "encoding/json"
import "net/rpc"
import "hash/fnv"


// functions for sorting
type SortedKeyValue []KeyValue
func (s SortedKeyValue) Len() int {
	return len(s)
}
func (s SortedKeyValue) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortedKeyValue) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}


type KeyValue struct {
	Key   string
	Value string
}

//
// to choose task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func workerMap(reply RequestTaskReply, mapf func(string, string) []KeyValue) {
	file, err := os.Open(reply.Filename)
	if err != nil {
		log.Fatalf("cannot open %v", reply.Filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", reply.Filename)
	}
	file.Close()

	keyVal := mapf(reply.Filename, string(content))

	nBuckets := divideIntoBuckets(keyVal, reply.NReduce)

	files := make([]string, len(nBuckets))
	// write keyval to file
	for reduceIndex, arr := range nBuckets {
		filename := "temp/mr-" + strconv.Itoa(reply.CurMapIndex) + "-" + strconv.Itoa(reduceIndex)
		WriteJSONToFile(filename, arr)
		files[reduceIndex] = filename
	}

	args := FinishedTaskArgs{files, reply.TaskNumber, "map"}
	// we need to report back to the master program
	sendFinishedToMaster(args)
}

func workerReduce(reply RequestTaskReply, reducef func(string, []KeyValue) string) {
	interKeyVal	:= []KeyValue{}
	for _, filename := range reply.ReduceFileList {
		temp := ReadJSONFromFile(filename)
		interKeyVal = append(interKeyVal, temp...)
	}

	sort.Sort(SortedKeyValue(interKeyVal))

	var prev string
	start := 0

	// tomororw add the bullshit here
	outputFile, _ := os.Create("mr-out-" + strconv.Itoa(reply.CurReduceIndex))

	for i, cur := range interKeyVal {
		if cur.Key != prev && prev != "" {
			num := reducef(prev, interKeyVal[start:i])
			fmt.Fprintf(outputFile, "%v %v\n", prev, num)
			start = i
		}
		prev = cur.Key
	}

	num := reducef(prev, interKeyVal[start:])
	fmt.Fprintf(outputFile, "%v %v\n", prev, num)

	args := FinishedTaskArgs{nil, reply.TaskNumber, "reduce"}
	// we need to report back to the master program
	sendFinishedToMaster(args)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []KeyValue) string) {

	// Your worker implementation here.

	for {
		reply := CallGetTask()
		// sleep and continue loop from top if there was no tasks available

		if reply.TaskType == "map" {
			fmt.Println("Received map task")
			workerMap(reply, mapf)
		} else if reply.TaskType == "reduce" {
			fmt.Println("Received reduce task")
			workerReduce(reply, reducef)
		} else if reply.TaskType == "sleep" {
			time.Sleep(5 * time.Second)
			continue
		} else if reply.TaskType == "exit" {
			break
		}
	}
}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallGetTask() RequestTaskReply {

	// declare an argument and reply structure.
	args := ExampleArgs{}
	reply := RequestTaskReply{}

	// send the RPC request, wait for the reply.
	ret := call("Master.GetTask", &args, &reply)

	if !ret {
		return RequestTaskReply{"", nil, "sleep", 0, 0, 0, 0}
	}

	return reply
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

func sendFinishedToMaster(args FinishedTaskArgs) {
	// declare an argument and reply structure.
	reply := ExampleReply{}

	call("Master.ReportFinishedTask", args, reply)
}

func WriteJSONToFile(filename string, keyValArray []KeyValue) {
	tempPath := filepath.Join(".", "temp")
	os.MkdirAll(tempPath, os.ModePerm)
	file, _ := json.MarshalIndent(keyValArray, "", "")
	ioutil.WriteFile(filename, file, 0644)
}

func ReadJSONFromFile(filename string) []KeyValue {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var keyVal []KeyValue
	err = json.Unmarshal(content, &keyVal)

	if err != nil {
		log.Fatal("Failed to unmarshal data ", err)
	}

	return keyVal 
}

func divideIntoBuckets(keyValArray []KeyValue, nReduce int) [][]KeyValue {
	nBuckets := make([][]KeyValue, nReduce)

	for _, arr := range keyValArray {
		index := ihash(arr.Key) % nReduce
		nBuckets[index] = append(nBuckets[index], arr)
	}

	return nBuckets
}