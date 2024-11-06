package jenkins

import (
	"io"
	"net/http"
	"net/url"
)

type ViewService struct {
	*Item
}

func (v *ViewService) Get(name string) (*View, error) {
	jobJson := &Job{}
	if err := v.BindAPIJson(jobJson, &ApiJsonOpts{Tree: "views[name,url,description]"}); err != nil {
		return nil, err
	}
	for _, view := range jobJson.Views {
		if view.Name == name {
			return view, nil
		}
	}
	return nil, nil
}

func (v *ViewService) Create(name string, xml io.Reader) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", name)
	return v.Request("POST", "createView?"+p.Encode(), xml)
}
func (v *ViewService) Delete(name string) (*http.Response, error) {
	return v.Request("POST", "view/"+name+"/doDelete", nil)
}

func (v *ViewService) AddJobToView(name, jobName string) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", jobName)
	return v.Request("POST", "view/"+name+"/addJobToView?"+p.Encode(), nil)
}

func (v *ViewService) RemoveJobFromView(name, jobName string) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", jobName)
	return v.Request("POST", "view/"+name+"/removeJobFromView?"+p.Encode(), nil)
}

func (v *ViewService) GetConfigure(name string) (string, error) {
	return readResponseToString(v, "GET", "view/"+name+"/config.xml", nil)
}

func (v *ViewService) SetConfigure(name string, xml io.Reader) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", name)
	return v.Request("POST", "view/"+name+"/config.xml?"+p.Encode(), xml)
}

func (v *ViewService) SetDescription(name, description string) (*http.Response, error) {
	p := url.Values{}
	p.Add("description", description)
	return v.Request("POST", "view/"+name+"/submitDescription?"+p.Encode(), nil)
}

func (v *ViewService) List() ([]*View, error) {
	jobJson := &Job{}
	if err := v.BindAPIJson(jobJson, &ApiJsonOpts{Tree: "views[name,url,description]"}); err != nil {
		return nil, err
	}
	return jobJson.Views, nil
}

// func (v *ViewService) bindViewAPIJson(name string, view interface{}) error {
// 	resp, err := v.Request("GET", "view/"+name+"/api/json")
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	return unmarshalResponse(resp.Body, view)
// }

// func (v *ViewService) GetJobFromView(name, jobName string) (*JobItem, error) {
// 	view := &View{}
// 	if err := v.bindViewAPIJson(name, view); err != nil {
// 		return nil, err
// 	}
// 	for _, job := range view.Jobs {
// 		if job.Name == jobName {
// 			return NewJobItem(job.URL, job.Class, v.jenkins), nil
// 		}
// 	}
// 	return nil, nil
// }

// func (v *ViewService) ListJobInView(name string) ([]*JobItem, error) {
// 	view := &View{}
// 	if err := v.bindViewAPIJson(name, view); err != nil {
// 		return nil, err
// 	}
// 	var jobs []*JobItem
// 	for _, job := range view.Jobs {
// 		jobs = append(jobs, NewJobItem(job.URL, job.Class, v.jenkins))
// 	}
// 	return jobs, nil
// }
