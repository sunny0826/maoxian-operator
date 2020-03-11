package maoxianbot

type botError struct {
	context string
}

//方法名字是Error()
func (e botError) Error() string {
	return e.context
}
