package resource

type AWXSource struct {
	Endpoint  string `json:"endpoint"`
	AuthToken string `json:"auth_token"`
}

type AWXVersion struct{}

type AWXMetadata struct{}

type AWXResource struct {
	source *AWXSource
	params *AWXParams
}

type AWXParams struct {
	Inventory  string `json:"inventory,omitempty"`
	TemplateId int    `json:"template_id"`
	Debug      bool   `json:"debug,omitempty"`
}

func NewAWXResource() *AWXResource {
	return &AWXResource{
		source: &AWXSource{},
		params: &AWXParams{},
	}
}

func (a *AWXResource) Source() (source interface{}) {
	return a.source
}

func (a *AWXResource) Params() (params interface{}) {
	return a.params
}
