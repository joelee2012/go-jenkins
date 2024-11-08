package jenkins

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Views struct {
	*Item
}

func (v *Views) Get(name string) (*ViewJson, error) {
	jobJson := &JobJson{}
	if err := v.ApiJson(jobJson, &ApiJsonOpts{Tree: "views[name,url,description]"}); err != nil {
		return nil, err
	}
	for _, view := range jobJson.Views {
		if view.Name == name {
			return view, nil
		}
	}
	return nil, fmt.Errorf("%s has no view [%s]", v, name)
}

func (v *Views) Create(name string, xml io.Reader) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", name)
	return v.Request("POST", "createView?"+p.Encode(), xml)
}
func (v *Views) Delete(name string) (*http.Response, error) {
	return v.Request("POST", "view/"+name+"/doDelete", nil)
}

func (v *Views) AddJobToView(name, jobName string) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", jobName)
	return v.Request("POST", "view/"+name+"/addJobToView?"+p.Encode(), nil)
}

func (v *Views) RemoveJobFromView(name, jobName string) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", jobName)
	return v.Request("POST", "view/"+name+"/removeJobFromView?"+p.Encode(), nil)
}

func (v *Views) GetConfigure(name string) (string, error) {
	return readResponseToString(v, "GET", "view/"+name+"/config.xml", nil)
}

func (v *Views) SetConfigure(name string, xml io.Reader) (*http.Response, error) {
	p := url.Values{}
	p.Add("name", name)
	return v.Request("POST", "view/"+name+"/config.xml?"+p.Encode(), xml)
}

func (v *Views) SetDescription(name, description string) (*http.Response, error) {
	p := url.Values{}
	p.Add("description", description)
	return v.Request("POST", "view/"+name+"/submitDescription?"+p.Encode(), nil)
}

func (v *Views) List() ([]*ViewJson, error) {
	jobJson := &JobJson{}
	if err := v.ApiJson(jobJson, &ApiJsonOpts{Tree: "views[name,url,description]"}); err != nil {
		return nil, err
	}
	return jobJson.Views, nil
}

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
// 	return nil, fmt.Errorf("%s has no job %s", v.URL, name)
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
