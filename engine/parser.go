package engine

// Parse get target urls from respones
func (e *Engine) Parse(content []byte) ([]string, error) {
	return e.parser.Parse(content)
}

func (e *Engine) Save() error {
	return e.parser.Save()
}
