package jenkins

type OneQueueItem struct {
	*Jenkins
	ID      int
	build   *BuildItem
	BaseURL string
}

func NewQueueItem(url string, client *Jenkins) *OneQueueItem {
	return &OneQueueItem{
		Jenkins: client,
		ID:      parseId(url),
		BaseURL: url,
		build:   nil,
	}
}

func (q *OneQueueItem) GetJob() (*JobItem, error) {
	var queueJson QueueItem
	if _, err := R().SetSuccessResult(&queueJson).Get("api/json"); err != nil {
		return nil, err
	}
	if parseClass(queueJson.Class) == "BuildableItem" {
		return q.build.GetJob()
	}
	return NewJobItem(queueJson.Task.URL, queueJson.Task.Class, q.Jenkins), nil
}

func (q *OneQueueItem) GetBuild() (*BuildItem, error) {
	if q.build != nil {
		return q.build, nil
	}
	var queueJson QueueItem
	if _, err := R().SetSuccessResult(&queueJson).Get("api/json"); err != nil {
		return nil, err
	}
	var err error
	switch parseClass(queueJson.Class) {
	case "LeftItem":
		q.build = NewBuildItem(queueJson.Executable.URL, queueJson.Executable.Class, q.Jenkins)
	case "BuildableItem", "WaitingItem":
		q.build, err = q.getWaitingBuild()
	}
	return q.build, err
}

func (q *OneQueueItem) getWaitingBuild() (*BuildItem, error) {
	builds, err := q.Nodes.GetBuilds()
	if err != nil {
		return nil, err
	}
	var buildJson struct {
		Class   string `json:"_class"`
		QueueId int    `json:"queueId"`
	}
	for _, build := range builds {
		if _, err := R().SetQueryParam("tree", "queueId").SetSuccessResult(&buildJson).Get("api/json"); err != nil {
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
	*Jenkins
	BaseURL string
}

func NewQueueService(c *Jenkins) *QueueService {
	return &QueueService{BaseURL: c.URL + "queue/", Jenkins: c}
}

func (q *QueueService) List() ([]*OneQueueItem, error) {
	queue := &Queue{}
	if _, err := R().SetSuccessResult(&queue).Get("api/json"); err != nil {
		return nil, err
	}
	var items []*OneQueueItem
	for _, item := range queue.Items {
		items = append(items, NewQueueItem(item.URL, q.Jenkins))
	}
	return items, nil
}

func (q *QueueService) Get(id int) (*OneQueueItem, error) {
	var queue Queue
	if _, err := R().SetSuccessResult(&queue).Get("api/json"); err != nil {
		return nil, err
	}
	for _, item := range queue.Items {
		if item.ID == id {
			return NewQueueItem(item.URL, q.Jenkins), nil
		}
	}
	return nil, nil
}

func (q *QueueService) Cancel(id string) error {
	_, err := R().SetQueryParam("id", id).Post("cancelItem")
	return err
}
