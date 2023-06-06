package splunk

type BaseResource struct {
	Id string `json:"id"`
}

type Content struct {
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	Capabilities []string `json:"capabilities"`
}

type User struct {
	BaseResource
	Name    string `json:"name"`
	Content struct {
		Email        string   `json:"email"`
		Roles        []string `json:"roles"`
		Capabilities []string `json:"capabilities"`
	} `json:"content"`
}

type Role struct {
	BaseResource
	Name   string `json:"name"`
	Author string `json:"author"`
}

type Capability struct {
	Name string `json:"name"`
}

type Vhost interface {
	GetName() string
}
