package modules

type Module interface {
	InitModule(cbs []byte) error
	Run()
	Cleanup() error
}
