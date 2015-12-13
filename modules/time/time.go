package time

import (
	"encoding/json"
	"github.com/andir/go3status/modules"
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("go3status.time")

type TimeInstance struct {
	name   string
	config map[string]interface{}
	format string
}

func (t TimeInstance) RefreshInterval() int {
	return 1
}

func (t TimeInstance) Name() (n string) {
	n = t.name
	return
}

func (t TimeInstance) Config() (m map[string]interface{}) {
	m = t.config
	return
}

func (t TimeInstance) String() (s string) {
	s = t.Name()
	return
}

func (t TimeInstance) Render() (item modules.Item) {
	item = RenderInstance(t)
	return
}

type TimeItem struct {
	Name string `json:"name"`
	Text string `json:"full_text"`
}

func (e TimeItem) Marshal() (bytes []byte) {
	var err error
	if bytes, err = json.Marshal(e); err != nil {
		log.Error(err.Error())
	}
	return
}

func RenderInstance(i modules.ModuleInstance) (t modules.Item) {

	instance := i.(TimeInstance)

	now := time.Now()
	formatted := now.Format(instance.format)
	t = modules.Item(TimeItem{Name: instance.name, Text: formatted})

	return
}

func CreateInstance(name string, config map[string]interface{}) (m modules.ModuleInstance) {
	f := TimeInstance{
		name:   name,
		config: config,
	}

	if i, ok := config["format"]; ok {
		f.format = i.(string)
	} else {
		f.format = "Mon, 02.01.2006 15:04:05 MST"
	}

	m = modules.ModuleInstance(f)

	return
}

var Module = modules.Module{
	Name:           "time",
	CreateInstance: CreateInstance,
}
