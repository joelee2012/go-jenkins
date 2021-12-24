package jenkins

type ComputerSet struct {
	Item
}

func NewComputerSet(url string, jenkins *Jenkins) *ComputerSet {
	return &ComputerSet{Item: *NewItem(url, "ComputerSet", jenkins)}
}

func (cs *ComputerSet) GetBuilds() ([]*Build, error) {
	var csJson ComputerSetJson
	var builds []*Build
	if err := cs.BindAPIJson(ReqParams{"tree": "computer[oneOffExecutors[currentExecutable[url]]]"}, &csJson); err != nil {
		return builds, err
	}
	for _, c := range csJson.Computers {
		for _, e := range c.OneOffExecutors {
			builds = append(builds, NewBuild(e.CurrentExecutable.URL, e.CurrentExecutable.Class, cs.jenkins))
		}

	}
	return builds, nil
}

type Computer struct {
	Item
}
