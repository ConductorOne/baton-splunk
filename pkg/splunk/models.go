package splunk

type BaseResource struct {
	Id  string `json:"id"`
	ACL ACL    `json:"acl"`
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
	Name    string `json:"name"`
	Author  string `json:"author"`
	Content struct {
		Capabilities         []string `json:"capabilities"`
		ImportedCapabilities []string `json:"imported_capabilities"`
	} `json:"content"`
}

type Application struct {
	BaseResource
	Name    string `json:"name"`
	Author  string `json:"author"`
	Content struct {
		Description string `json:"description"`
	} `json:"content"`
}

type Capability struct {
	BaseResource
	Name    string `json:"name"`
	Content struct {
		Capabilities []string `json:"capabilities"`
	} `json:"content"`
}

type ACL struct {
	App   string `json:"app"`
	Perms struct {
		Read  []string `json:"read"`
		Write []string `json:"write"`
	} `json:"perms"`
}
