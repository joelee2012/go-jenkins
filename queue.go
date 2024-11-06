package jenkins

import (
	"fmt"
	"net/http"
)

type OneQueueItem struct {
	*Item
	ID    int
	build *BuildItem
}

func NewQueueItem(url string, client *Jenkins) *OneQueueItem {
	return &OneQueueItem{
		Item:  NewItem(url, "QueueItem", client),
		ID:    parseId(url),
		build: nil,
	}
}

func (q *OneQueueItem) GetJob() (*JobItem, error) {
	var queueJson QueueItem
	if err := q.ApiJson(&queueJson, nil); err != nil {
		return nil, err
	}
	if parseClass(queueJson.Class) == "BuildableItem" {
		return q.build.GetJob()
	}
	return NewJobItem(queueJson.Task.URL, queueJson.Task.Class, q.jenkins), nil
}

func (q *OneQueueItem) GetBuild() (*BuildItem, error) {
	if q.build != nil {
		return q.build, nil
	}
	var queueJson QueueItem
	if err := q.ApiJson(&queueJson, nil); err != nil {
		return nil, err
	}
	var err error
	switch parseClass(queueJson.Class) {
	case "LeftItem":
		q.build = NewBuildItem(queueJson.Executable.URL, queueJson.Executable.Class, q.jenkins)
	case "BuildableItem", "WaitingItem":
		q.build, err = q.getWaitingBuild()
	}
	return q.build, err
}

func (q *OneQueueItem) getWaitingBuild() (*BuildItem, error) {
	builds, err := q.jenkins.Nodes().GetBuilds()
	if err != nil {
		return nil, err
	}
	var buildJson struct {
		Class   string `json:"_class"`
		QueueId int    `json:"queueId"`
	}
	for _, build := range builds {
		if err := build.ApiJson(&buildJson, &ApiJsonOpts{Tree: "queueId"}); err != nil {
			return nil, err
		}
		if buildJson.QueueId == q.ID {
			return build, nil
		}
	}
	// build is not avaliable and no error
	return nil, nil
}

type QueueService struct {
	*Item
}

func (q *QueueService) List() ([]*OneQueueItem, error) {
	queue := &Queue{}
	if err := q.ApiJson(queue, nil); err != nil {
		return nil, err
	}
	var items []*OneQueueItem
	for _, item := range queue.Items {
		items = append(items, NewQueueItem(item.URL, q.jenkins))
	}
	return items, nil
}

func (q *QueueService) Get(id int) (*OneQueueItem, error) {
	var queue Queue
	if err := q.ApiJson(&queue, nil); err != nil {
		return nil, err
	}
	for _, item := range queue.Items {
		if item.ID == id {
			return NewQueueItem(item.URL, q.jenkins), nil
		}
	}
	return nil, nil
}

func (q *QueueService) Cancel(id int) (*http.Response, error) {
	return q.Request("POST", fmt.Sprintf("cancelItem?id=%d", id), nil)
}
