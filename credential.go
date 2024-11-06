package jenkins

import (
	"io"
	"net/http"
)

type CredentialService struct {
	*Item
}

func (cs *CredentialService) Get(name string) (*Credential, error) {
	var credsJson Credentials
	if err := cs.ApiJson(&credsJson, &ApiJsonOpts{Depth: 1}); err != nil {
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

func (cs *CredentialService) Create(xml io.Reader) (*http.Response, error) {
	return cs.Request("POST", "createCredentials", xml)
}

func (cs *CredentialService) Delete(name string) (*http.Response, error) {
	return cs.Request("POST", "credential/"+name+"/doDelete", nil)
}

func (cs *CredentialService) GetConfigure(name string) (string, error) {
	return readResponseToString(cs, "GET", "credential/"+name+"/config.xml", nil)
}

func (cs *CredentialService) SetConfigure(name string, xml io.Reader) (*http.Response, error) {
	return cs.Request("POST", "credential/"+name+"/config.xml", xml)
}

func (cs *CredentialService) List() ([]*Credential, error) {
	var credsJson Credentials
	if err := cs.ApiJson(&credsJson, &ApiJsonOpts{Depth: 1}); err != nil {
		return nil, err
	}
	return credsJson.Credentials, nil
}
