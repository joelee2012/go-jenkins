package jenkins

import (
	"fmt"

	"github.com/imroc/req"
)

func doRequest(client *Client, method, url string, v ...interface{}) (*req.Resp, error) {
	if _, err := client.GetCrumb(); err != nil {
		return nil, err
	}
	v = append(v, client.Header)
	resp, err := client.Req.Do(method, url, v...)
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode >= 400 {
		return nil, fmt.Errorf(resp.Dump())
	}
	return resp, nil
}

func doRequestAndDropResp(r Requester, method, entry string, v ...interface{}) error {
	_, err := r.Request(method, entry, v...)
	return err
}

func doDelete(r Requester) error {
	return doRequestAndDropResp(r, "POST", "doDelete")
}

func doGetConfigure(r Requester) (string, error) {
	resp, err := r.Request("GET", "config.xml")
	return resp.String(), err
}

func doSetConfigure(r Requester, xml string) error {
	return doRequestAndDropResp(r, "POST", "config.xml", req.BodyXML(xml))
}

func doSetDescription(r Requester, description string) error {
	return doRequestAndDropResp(r, "POST", "submitDescription", ReqParams{"description": description})
}

func doGetDescription(r Requester) (string, error) {
	data := make(map[string]string)
	if err := doBindAPIJson(r, ReqParams{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func doBindAPIJson(r Requester, params ReqParams, v interface{}) error {
	resp, err := r.Request("GET", "api/json", params)
	if err != nil {
		return err
	}
	return resp.ToJSON(v)
}

func doDisable(r Requester) error {
	return doRequestAndDropResp(r, "POST", "disable")
}

func doEnable(r Requester) error {
	return doRequestAndDropResp(r, "POST", "enable")
}

func doRunScript(r Requester, script string) (string, error) {
	resp, err := r.Request("POST", "scriptText", ReqParams{"script": script})
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}
