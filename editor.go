package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]byte
	Ispeed uint32
	Ospeed uint32
}

var origTermios *Termios

func TcSetAttr(fd uintptr, termios *Termios) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS+1), uintptr(unsafe.Pointer(termios))); err != 0 {
		return err
	}
	return nil
}

func TcGetAttr(fd uintptr) *Termios {
	var termios = &Termios{}
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(termios))); err != 0 {
		log.Fatalf("Problem getting termial attributes: %s\n", err)
	}
	return termios
}

func enableRawMode() {
	origTermios := TcGetAttr(os.Stdin.Fd())
	var raw Termios
	raw = *origTermios
	raw.Iflag &^= syscall.BRKINT | syscall.IXON | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN+1] = 0
	raw.Cc[syscall.VTIME+1] = 1
	if e := TcSetAttr(os.Stdin.Fd(), &raw); e != nil {
		log.Fatalf("Problem enabling raw mode: %s\n", e)
	}
}

func disableRawMode() {
	fmt.Fprintf(os.Stderr, "Enter disableRawmode\n")
	if e := TcSetAttr(os.Stdin.Fd(), origTermios); e != nil {
		log.Fatalf("Problem disabling raw mode: %s\n", e)
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()
	buffer := make([]byte, 1)
	var cc int
	var err error
	for cc, err = os.Stdin.Read(buffer);
		buffer[0] != 'q' && cc >= 0;
		cc, err = os.Stdin.Read(buffer) {
		if buffer[0] > 20 && buffer[0] < 0x7f {
			fmt.Printf("%3d  %c  %d\r\n", buffer[0], buffer[0], cc)
		} else {
			fmt.Printf("%3d      %d\r\n", buffer[0], cc)
		}
		buffer[0] = 0
	}
	if err != nil {
		disableRawMode()
		log.Fatal(err)
	}
}
