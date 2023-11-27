package mr

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	// Your definitions here.

	files   []string
	nReduce int

	nMap             int
	mapfinished      int
	maptaskstatus    []int //long for maptasks , finished 2,waiting 1,not allocated 0
	reducefinished   int
	reducetaskstatus []int //long for reducetasks , finished 2,waiting 1,not allocated 0

}

var mutex sync.Mutex
var group2 sync.WaitGroup

var t = 10

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *StringReply) error {

	filename := "../main/pg-grimm.txt"

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()

	reply.S = string(content)
	return nil
}

func (c *Coordinator) FinishedMap(args *TaskArgs, reply *TaskReply) error {
	mutex.Lock()
	defer mutex.Unlock()

	c.mapfinished++
	c.maptaskstatus[args.Maptasknum] = 2 //finished
	return nil
}

func (c *Coordinator) FinishedReduce(args *TaskArgs, reply *TaskReply) error {
	mutex.Lock()
	defer mutex.Unlock()

	c.reducefinished++
	c.reducetaskstatus[args.Reducetasknum] = 2 //finished
	return nil
}

func (c *Coordinator) HandleWorker(args *TaskArgs, reply *TaskReply) error {
	mutex.Lock()
	//give maptask
	if c.mapfinished < c.nMap {
		id := -1
		for i := 0; i < c.nMap; i++ {
			if c.maptaskstatus[i] == 0 { //not allocated
				id = i
				break
			}
		}
		if id == -1 {
			reply.TaskType = "waiting"
			mutex.Unlock()
		} else {
			reply.TaskType = "map"
			reply.NReduce = c.nReduce
			reply.Maptasknum = id
			c.maptaskstatus[id] = 1 //waiting
			filename := c.files[id]
			reply.FileName = filename
			mutex.Unlock()

			go c.asyncCheck(t, id, "map")
			mutex.Unlock()
		}

		//map finished, give reducetask
	} else if c.mapfinished == c.nMap && c.reducefinished < c.nReduce {
		id := -1
		for i := 0; i < c.nReduce; i++ {
			if c.reducetaskstatus[i] == 0 { //not allocated
				id = i
				break
			}
		}
		if id == -1 {
			reply.TaskType = "waiting"
			mutex.Unlock()
		} else {
			reply.NMap = c.nMap
			reply.TaskType = "reduce"
			reply.Reducetasknum = id
			c.reducetaskstatus[id] = 1 //waiting
			mutex.Unlock()

			go c.asyncCheck(t, id, "reduce")
			mutex.Unlock()
		}
	} else {
		reply.TaskType = "done"
		mutex.Unlock()
	}
	return nil
}

func (c *Coordinator) asyncCheck(sleepSeconds int, id int, stage string) {
	time.Sleep(time.Duration(sleepSeconds) * time.Second)

	switch coordStage := stage; coordStage {

	case "map":
		mutex.Lock()
		if c.maptaskstatus[id] == 1 { //waiting
			c.maptaskstatus[id] = 0 //not allocated
			break
		}

	case "reduce":
		mutex.Lock()
		if c.reducetaskstatus[id] == 1 { //waiting
			c.reducetaskstatus[id] = 0 //not allocated
			break
		}
	}
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
	ret = c.reducefinished == c.nReduce

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.files = files
	c.nMap = len(files)
	c.maptaskstatus = make([]int, c.nMap)
	c.reducetaskstatus = make([]int, c.nReduce)
	c.nReduce = nReduce
	c.server()
	return &c
}

/*
	1. Listen for incomming work requests from workers

	2. Find some work and attach it to the reply struct in the RPC reply

	3. Give the worker some time to finish their work, if the work isn't done in a reasonable time then assign the work to someone else.

	4. Receive response from workers, combine results into a output file.




*/
