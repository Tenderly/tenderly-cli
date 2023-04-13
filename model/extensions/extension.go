package extensions

type Extension struct {
	Description string `json:"description" yaml:"description"`
	Method      string `json:"method" yaml:"method"`
	Action      string `json:"action" yaml:"action"`
}

type ProjectExtensions struct {
	Specs map[string]*Extension `json:"specs" yaml:"specs"`
}
