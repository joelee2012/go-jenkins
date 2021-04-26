package jenkins

import (
	"github.com/imroc/req"
)

type QueueItem struct {
	Item
	id    int
	build Build
}

func NewQueueItem(url string, jenkins *Jenkins) *QueueItem {
	return &QueueItem{
		Item:  Item{Url: url, Class: "Queue", jenkins: jenkins},
		id:    getId(url),
		build: nil,
	}
}

func (q *QueueItem) GetBuild() (Build, error) {
	if q.build != nil {
		return q.build, nil
	}
	result, err := q.APIJson(req.Param{"tree": "_class"})
	if err != nil {
		return nil, err
	}
	class := GetClassName(result.Get("_class").String())
	var method func() (Build, error)
	switch class {
	case "LeftItem":
		method = q.getLeftBuild
	case "BuildableItem", "WaitingItem":
		method = q.getWaitingBuild
	}
	build, err := method()
	if err != nil {
		return nil, err
	}
	q.build = build
	return q.build, nil
}

func (q *QueueItem) getLeftBuild() (Build, error) {
	executable, err := q.APIJson(req.Param{"tree": "executable[url]"})
	if err != nil {
		return nil, err
	}
	class := GetClassName(executable.Get("executable._class").String())
	url := executable.Get("executable.url").String()
	return NewBuild(url, class, q.jenkins), nil
}

func (q *QueueItem) getWaitingBuild() (Build, error) {
	cs := q.jenkins.GetComputerSet()
	builds, err := cs.GetBuilds()
	if err != nil {
		return nil, err
	}
	for _, build := range builds {
		buildQId, err := build.APIJson(req.Param{"tree": "queueId"})
		if err != nil {
			return nil, err
		}
		buildQIdInt := buildQId.Get("queueId").Int()
		if int(buildQIdInt)+1 == q.id {
			return build, nil
		}
	}
	return nil, nil
}

func (q *QueueItem) GetId() int {
	return q.id
}
