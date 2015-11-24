package mpd

import (
	"bytes"
	"encoding/json"
	"github.com/andir/go3status/modules"
	go_mpd "github.com/fhs/gompd/mpd"
	"github.com/op/go-logging"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

var log = logging.MustGetLogger("go3status.mpd")

type MPDItem struct {
	Name   string `json:"name"`
	Text   string `json:"full_text"`
	Markup string `json:"markup"`
}

func (e MPDItem) Marshal() (bytes []byte) {
	var err error
	if bytes, err = json.Marshal(e); err != nil {
		log.Error(err.Error())
	}
	return
}

type MPDInstance struct {
	name      string
	host_name string
	port      int
	template  *template.Template
}

func (m MPDInstance) Name() string {
	return m.name
}

func (m MPDInstance) String() (s string) {
	s = m.name
	s += " - "
	s += m.host_name
	s += ":"
	s += strconv.Itoa(m.port)
	s += " - "
	s += string(m.template.Name())
	return
}

type MPDFormatData struct {
	Artist   string
	Song     string
	Album    string
	Date     string
	Time     string
	Id       string
	File     string
	Title    string
	Composer string
	Disc     string
	Pos      string
	State    string
}

func (m MPDInstance) Render() (item modules.Item) {
	mpdItem := MPDItem{Name: m.name, Markup: "pango"}
	mpdFormatData := MPDFormatData{}
	var client *go_mpd.Client

	if c, err := go_mpd.Dial("tcp", m.host_name+":"+strconv.Itoa(m.port)); err != nil {
		log.Error(err.Error())
		return nil
	} else {
		client = c
	}

	defer func(client *go_mpd.Client) {
		if err := client.Close(); err != nil {
			log.Error("Failed to disconnect: " + err.Error())
		}
	}(client)

	if attrs, err := client.Status(); err == nil {
		if state, ok := attrs["state"]; ok {
			mpdFormatData.State = state
		} else {
			log.Error("Failed to read state.")
			return nil
		}
	} else {
		log.Error("Failed to obtain status: " + err.Error())
		return nil
	}

	if attrs, err := client.CurrentSong(); err != nil {
		log.Error("Failed to obtain current song: " + err.Error())
		return nil
	} else {
		obj := reflect.ValueOf(&mpdFormatData).Elem()
		for key, val := range attrs {
			if f := obj.FieldByName(key); f.IsValid() {
				val = strings.TrimSpace(val)
				f.Set(reflect.ValueOf(val))
			}
		}
	}

	buffer := bytes.Buffer{}

	if err := m.template.Execute(&buffer, mpdFormatData); err != nil {
		log.Error("Failed to render mpd template: " + err.Error())
		return nil
	} else {
		mpdItem.Text = buffer.String()
	}

	item = mpdItem
	return
}

func CreateInstance(name string, config map[string]interface{}) (instance modules.ModuleInstance) {
	mpdInstance := MPDInstance{name: name}

	if v, ok := config["host_name"]; ok {
		mpdInstance.host_name = v.(string)
	} else {
		mpdInstance.host_name = "127.0.0.1"
	}

	if v, ok := config["port"]; ok {
		mpdInstance.port = v.(int)
	} else {
		mpdInstance.port = 6600
	}
	var format string
	if v, ok := config["format"]; ok {
		format = v.(string)
	} else {
		format = "[{{.Status}}] {{ .Artist }} - {{ .Song }}"
	}

	if tmpl, err := template.New(mpdInstance.name).Parse(format); err == nil {
		mpdInstance.template = tmpl
	} else {
		log.Error("Failed to parse template:" + format)
		instance = nil
		return
	}

	instance = modules.ModuleInstance(mpdInstance)
	return
}

var Module = modules.Module{
	Name:           "mpd",
	CreateInstance: CreateInstance,
}
