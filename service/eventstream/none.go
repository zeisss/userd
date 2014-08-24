package eventstream

func NewNoneEventLog() *noneEventLog {
	return &noneEventLog{}
}

type noneEventLog struct {
}

func (log *noneEventLog) Publish(tag string, data []byte) {

}
