package load

import (
	"fmt"
	"syscall"
)

func main() {
	usage := Rusage{}
	if err := syscall.Getrusage(0, usage); err != nil {
		fmt.Println(err.Error())
		return
	}

}
