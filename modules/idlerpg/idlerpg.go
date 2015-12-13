package idlerpg

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/andir/go3status/modules"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"text/template"
)

var log = logging.MustGetLogger("idlerpg")

type IRPGItem struct {
	Name   string `json:"name"`
	Text   string `json:"full_text"`
	Markup string `json:"markup"`
}

func (e IRPGItem) Marshal() (bytes []byte) {
	var err error
	if bytes, err = json.Marshal(e); err != nil {
		log.Error(err.Error())
	}
	return
}

type IRPGInstance struct {
	name        string
	uri         string
	player_name string
	template    *template.Template
	config      map[string]interface{}
}

func (i IRPGInstance) RefreshInterval() int {
	return 900
}

func (t IRPGInstance) Name() (n string) {
	n = t.name
	return
}

func (t IRPGInstance) Config() (m map[string]interface{}) {
	m = t.config
	return
}

func (t IRPGInstance) String() (s string) {
	s = t.Name()
	return
}

type Penalties struct {
	logout int
	quest  int
	quit   int
	kick   int
	part   int
	nick   int
	mesg   int
	total  int
}

type Items struct {
	weapon   int
	tunic    int
	shield   int
	leggings int
	ring     int
	gloves   int
	boots    int
	helm     int
	charm    int
	amulet   int
	total    int
}

type Player struct {
	Username   string    `xml:"username"`
	Isadmin    bool      `xml:"isadmin"`
	Level      int       `xml:"level"`
	Class      string    `xml:"class"`
	Ttl        int       `xml:"ttl"`
	Userhost   string    `xml:"userhost"`
	Online     bool      `xml:"online"`
	Totalidled int       `xml:"totalidled"`
	Xpos       int       `xml:"xpos"`
	Ypos       int       `xml:"ypos"`
	Penalties  Penalties `xml:"penalties"`
	Items      Items     `xml:"items"`
}

func (t IRPGInstance) downloadData() (p *Player) {
	player := new(Player)
	log.Debug("Downloading " + t.uri)
	if resp, err := http.Get(t.uri); err != nil {
		log.Fatal(err)
		p = nil
		return
	} else {
		b, _ := ioutil.ReadAll(resp.Body)
		xml.Unmarshal(b, &player)
		p = player
	}
	return
}

func (t IRPGInstance) Render() (i modules.Item) {

	item := IRPGItem{Name: t.name, Markup: "pango"}

	player := t.downloadData()
	if player == nil {
		return
	}

	var formatted bytes.Buffer
	t.template.Execute(&formatted, player)

	item.Text = formatted.String()

	i = modules.Item(item)
	return
}

func CreateInstance(name string, config map[string]interface{}) (moduleInstance modules.ModuleInstance) {
	i := IRPGInstance{
		name: name,
	}

	var base_uri string

	if v, ok := config["base_uri"]; ok {
		base_uri = v.(string)
	} else {
		base_uri = "http://irpg.bspar.org/xml.php?player="
	}

	if v, ok := config["player"]; ok {
		n := v.(string)
		i.player_name = n
	} else {
		log.Error("Missing player name")
		moduleInstance = nil
		return
	}
	i.uri = base_uri + i.player_name

	var format string
	if v, ok := config["format"]; ok {
		format = v.(string)
	} else {
		format = "irpg {{ .Username }}: <span color=\"{{ if .Online}}green{{ else }}red{{end}}\">{{.Level}}</span>"
	}

	if template, err := template.New(i.name).Parse(format); err == nil {
		i.template = template
	} else {
		log.Error("Failed to create template: " + err.Error())
		moduleInstance = nil
	}

	moduleInstance = i
	return
}

var Module = modules.Module{
	Name:           "idlerpg",
	CreateInstance: CreateInstance,
}
