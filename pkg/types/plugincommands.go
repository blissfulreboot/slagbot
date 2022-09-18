package types

type ParameterType string

const (
	Before ParameterType = "before"
	After  ParameterType = "after"
	Flag   ParameterType = "flag"
)

type Parameter struct {
	Keyword     string
	Description string
	Type        ParameterType
}

type Arguments map[string]interface{}

type Command struct {
	Keyword     string
	Description string
	Params      []Parameter
}

type ParsedCommand struct {
	Channel   string
	Command   string
	Arguments Arguments
}
