package extensions

type BackendExtension struct {
	Name     string `json:"name" yaml:"name"`
	Method   string `json:"methodName" yaml:"method"`
	ActionID string `json:"actionId" yaml:"action"`
}

type ConfigExtension struct {
	Name        string `json:"name" yaml:"-"`
	Description string `json:"description" yaml:"description"`
	MethodName  string `json:"method" yaml:"method"`
	ActionName  string `json:"action" yaml:"action"`
}

type ConfigProjectExtensions struct {
	Specs map[string]*ConfigExtension `json:"specs" yaml:"specs"`
}
