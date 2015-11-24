package modules


type CreateInstanceFunc func(name string, config map[string]interface{}) ModuleInstance
type RenderInstanceFunc func(instance ModuleInstance) (item Item, err error)

type Module struct {
	Name string
	CreateInstance CreateInstanceFunc
	RenderInstance RenderInstanceFunc
}

type Item interface {
	Marshal() []byte
}

type ModuleInstance interface {
	Name() string
	String() string
	Render() Item
}
