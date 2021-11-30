package jenkins

type QueueItem struct {
	Item
	ID    int
	build *Build
}

func NewQueueItem(url string, jenkins *Jenkins) *QueueItem {
	return &QueueItem{
		Item:  *NewItem(url, "QueueItem", jenkins),
		ID:    parseId(url),
		build: nil,
	}
}

func (q *QueueItem) GetBuild() (*Build, error) {
	if q.build != nil {
		return q.build, nil
	}
	var queueJson QueueItemJson
	if err := q.BindAPIJson(ReqParams{}, &queueJson); err != nil {
		return nil, err
	}
	var err error
	switch parseClass(queueJson.Class) {
	case "LeftItem":
		q.build = NewBuild(queueJson.Executable.URL, queueJson.Executable.Class, q.jenkins)
	case "BuildableItem", "WaitingItem":
		q.build, err = q.getWaitingBuild()
	}
	return q.build, err
}

func (q *QueueItem) getWaitingBuild() (*Build, error) {
	cs := q.jenkins.GetComputerSet()
	builds, err := cs.GetBuilds()
	if err != nil {
		return nil, err
	}
	var buildJson struct {
		Class   string `json:"_class"`
		QueueId int    `json:"queueId"`
	}
	for _, build := range builds {
		if err := build.BindAPIJson(ReqParams{"tree": "queueId"}, &buildJson); err != nil {
			return nil, err
		}
		if buildJson.QueueId == q.ID {
			return build, nil
		}
	}
	return nil, nil
}
