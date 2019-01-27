package dummy

import (
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
)

type Module struct {
	value int
}

var myDev *Module = &Module{}

func init() {
	skeeter.RegisterModule("dummy", myDev)
}

func (m *Module) ModuleTest() {
	fmt.Println("dummy ModuleTest")
}
