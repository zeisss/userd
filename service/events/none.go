package events

func NewNoneEventLog() *noneEventLog {
	return &noneEventLog{}
}

type noneEventLog struct {
}

func (log *noneEventLog) Log(tag string, data []byte) {

}
