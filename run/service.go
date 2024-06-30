package run

import (
	"fmt"
)

type RunService interface {
	AddCodeInstanceToQueue(inst CodeInstance) (int, error)
	GetQueueSize() int
}

type runServiceInternal struct {
	codeChan      chan CodeInstance
	queueSize     int
	queueSizeChan chan QueueSizeRequest
}

type QueueSizeRequest struct {
	Modification int
	Response     chan int
}

type CodeInstance struct {
	Files     []DafnyFile
	Requester string
	Result    chan RunResult
}

type RunResult struct {
	Status  int
	Content string
}

type DafnyFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (c *runServiceInternal) AddCodeInstanceToQueue(inst CodeInstance) (int, error) {
	go func() {
		c.codeChan <- inst
	}()
	return c.increaseQueueSize(), nil
}

func (c *runServiceInternal) GetQueueSize() int {
	return c.queueSize
}

func (c *runServiceInternal) increaseQueueSize() int {
	response := make(chan int)
	c.queueSizeChan <- QueueSizeRequest{
		Modification: 1,
		Response:     response,
	}
	return <-response
}

func (c runServiceInternal) decreaseQueueSize() int {
	response := make(chan int)
	c.queueSizeChan <- QueueSizeRequest{
		Modification: -1,
		Response:     response,
	}
	return <-response
}

func (c *runServiceInternal) StartQueueIncrease() {
	for request := range c.queueSizeChan {
		c.queueSize += request.Modification
		request.Response <- c.queueSize
	}
}

func (c *runServiceInternal) StartRunQueue() {
	for inst := range c.codeChan {
		c.decreaseQueueSize()
		err := prepareRunEnvironment(inst)
		if err != nil {
			fmt.Printf("Error preparing running environment: %s\n", err.Error())
		}
		result, err, status := runAtTmp()
		inst.Result <- RunResult{
			Status:  status,
			Content: result,
		}
		fmt.Printf("Ran request by %s, queue now %d\n", inst.Requester, c.GetQueueSize())
	}
}

func StartRunService() (RunService, error) {
	c := runServiceInternal{
		codeChan:      make(chan CodeInstance),
		queueSize:     0,
		queueSizeChan: make(chan QueueSizeRequest),
	}

	// syncronized queue size
	go c.StartQueueIncrease()
	go c.StartRunQueue()
	return &c, nil
}
