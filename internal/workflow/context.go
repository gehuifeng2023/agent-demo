package workflow

type Context struct {
	SessionID string
	Question  string
	Results   map[string]string
}

func NewContext(sessionID, question string) *Context {
	return &Context{
		SessionID: sessionID,
		Question:  question,
		Results:   make(map[string]string),
	}
}
