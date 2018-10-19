package pci15

import (
	"encoding/json"
	"time"
)

type PCILogItem struct {
	Time    time.Duration `json:"time"`
	Name    string        `json:"name"`
	Content string        `json:"content"`
}

type PCILog struct {
	Log     []*PCILogItem `json:"log"`
	AbsTime time.Time     `json:"time"`
	Name    string        `json:"name"`
}

func NewPCILog(name string) *PCILog {
	return &PCILog{
		Log:     make([]*PCILogItem, 0),
		AbsTime: time.Now(),
		Name:    name,
	}
}

func (l *PCILog) Append(val string) {
	l.Log = append(l.Log, &PCILogItem{
		Time:    time.Now().Sub(l.AbsTime),
		Name:    l.Name,
		Content: val,
	})
}

func (l *PCILog) ToJSON() ([]byte, error) {
	return json.Marshal(l.Log)
}

func (l *PCILog) Merge(ll *PCILog) {
	timeSft := ll.AbsTime.Sub(l.AbsTime)
	for _, log := range ll.Log {
		l.Log = append(l.Log, &PCILogItem{
			Time:    log.Time + timeSft,
			Name:    ll.Name,
			Content: log.Content,
		})
	}
}
