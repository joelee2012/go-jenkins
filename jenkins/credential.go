package jenkins

import (
	"github.com/imroc/req"
)

type CredentialService struct {
	*Item
}

func NewCredentialService(v interface{}) *CredentialService {
	if c, ok := v.(*Client); ok {
		return &CredentialService{Item: NewItem(c.URL+"credentials/store/system/domain/_/", "Credentials", c)}
	}

	if c, ok := v.(*JobItem); ok {
		return &CredentialService{Item: NewItem(c.URL+"credentials/store/folder/domain/_/", "Credentials", c.client)}
	}
	return nil
}

func (cs *CredentialService) Get(name string) (*Credential, error) {
	var credsJson Credentials
	if err := cs.BindAPIJson(ReqParams{"depth": "1"}, &credsJson); err != nil {
		return nil, err
	}
	if credsJson.Credentials != nil {
		for _, cred := range credsJson.Credentials {
			if cred.ID == name {
				return cred, nil
			}
		}
	}
	return nil, nil
}

func (cs *CredentialService) Create(xml string) error {
	_, err := cs.Request("POST", "createCredentials", req.BodyXML(xml))
	return err
}

func (cs *CredentialService) Delete(name string) error {
	_, err := cs.Request("POST", "credential/"+name+"/doDelete")
	return err
}

func (cs *CredentialService) GetConfigure(name string) (string, error) {
	resp, err := cs.Request("GET", name+"/config.xml")
	return resp.String(), err
}

func (cs *CredentialService) SetConfigure(name, xml string) error {
	_, err := cs.Request("POST", name+"/config.xml", req.BodyXML(xml))
	return err
}

func (cs *CredentialService) List() ([]*Credential, error) {
	var credsJson Credentials
	if err := cs.BindAPIJson(ReqParams{"depth": "1"}, &credsJson); err != nil {
		return nil, err
	}
	return credsJson.Credentials, nil
}
