package service

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

func NewEventCollector(maxItems int) *EventCollector {
	return &EventCollector{
		MaxItems: maxItems,
		Items:    make([]Item, 0, maxItems),
		Lock:     &sync.Mutex{},
	}
}

type Item struct {
	Tag       string
	Json      []byte
	Timestamp time.Time

	marshalCache []byte
}

func (i *Item) MarshalJSON() ([]byte, error) {
	if i.marshalCache != nil {
		return i.marshalCache, nil
	}

	msg := json.RawMessage(i.Json)
	item := map[string]interface{}{
		"tag":       i.Tag,
		"timestamp": i.Timestamp,
		"message":   &msg,
	}
	bytes, err := json.Marshal(item)
	if err == nil {
		i.marshalCache = bytes
	}
	return bytes, err
}

type EventCollector struct {
	MaxItems int
	Items    []Item
	Lock     *sync.Mutex
}

func (esc *EventCollector) publish(tag string, json []byte) {
	esc.Lock.Lock()
	defer esc.Lock.Unlock()

	item := Item{Tag: tag, Json: json, Timestamp: time.Now()}

	esc.Items = append(esc.Items, item)

	if len(esc.Items) > esc.MaxItems {
		esc.Items = esc.Items[1:]
	}
}

func (esc *EventCollector) Get() []Item {
	esc.Lock.Lock()
	defer esc.Lock.Unlock()
	return esc.Items
}

// WriteJSONOnce writes the current collected items into the writer.
// Each item is written on its own line.
func (esc *EventCollector) WriteJSONOnce(w io.Writer) error {
	items := esc.Get() // get a copy of the item array, should be conflict free for parallel access
	encoder := json.NewEncoder(w)
	w.Write([]byte("[\n"))
	for index, item := range items {
		if index > 0 {
			w.Write([]byte(",\n"))
		}
		if err := encoder.Encode(&item); err != nil {
			return err
		}

	}
	w.Write([]byte("]\n"))
	return nil
}
