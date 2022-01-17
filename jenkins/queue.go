package jenkins

type QueueItem struct {
	*Item
	ID    int
	build *Build
}

func NewQueueItem(url string, client *Client) *QueueItem {
	return &QueueItem{
		Item:  NewItem(url, "QueueItem", client),
		ID:    parseId(url),
		build: nil,
	}
}

func (q *QueueItem) GetJob() (*Job, error) {
	var queueJson QueueItemJson
	if err := q.BindAPIJson(ReqParams{}, &queueJson); err != nil {
		return nil, err
	}
	if parseClass(queueJson.Class) == "BuildableItem" {
		return q.build.GetJob()
	}
	return NewJob(queueJson.Task.URL, queueJson.Task.Class, q.client), nil
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
		q.build = NewBuild(queueJson.Executable.URL, queueJson.Executable.Class, q.client)
	case "BuildableItem", "WaitingItem":
		q.build, err = q.getWaitingBuild()
	}
	return q.build, err
}

func (q *QueueItem) getWaitingBuild() (*Build, error) {
	cs := q.client.ComputerSet()
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

type Queue struct {
	Item
}

func (q *Queue) List() ([]*QueueItem, error) {
	var qJson QueueJson
	if err := q.BindAPIJson(ReqParams{}, &qJson); err != nil {
		return nil, err
	}
	var items []*QueueItem
	for _, item := range qJson.Items {
		items = append(items, NewQueueItem(item.URL, q.client))
	}
	return items, nil
}

func (q *Queue) Get(id int) (*QueueItem, error) {
	var qJson QueueJson
	if err := q.BindAPIJson(ReqParams{}, &qJson); err != nil {
		return nil, err
	}
	for _, item := range qJson.Items {
		if item.ID == id {
			return NewQueueItem(item.URL, q.client), nil
		}
	}
	return nil, nil
}

func (q *Queue) Cancel(id int) error {
	return doRequestAndDropResp(q, "POST", "cancelItem", ReqParams{"id": id})
}
