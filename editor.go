package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"unsafe"
)

/*** data ***/
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

/*** terminal ***/

func die(err error) {
	disableRawMode()
	io.WriteString(os.Stdout, "\x1b[2J");
	io.WriteString(os.Stdout, "\x1b[H");
	log.Fatal(err)
}

func TcSetAttr(fd uintptr, termios *Termios) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(/*TCSETS*/0x5402+1), uintptr(unsafe.Pointer(termios))); err != 0 {
		return err
	}
	return nil
}

func TcGetAttr(fd uintptr) *Termios {
	var termios = &Termios{}
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, /*TCGETS*/0x5401, uintptr(unsafe.Pointer(termios))); err != 0 {
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

func editorReadKey() byte {
	var buffer [1]byte
	var cc int
	var err error
	for cc, err = os.Stdin.Read(buffer[:]); cc != 1; cc, err = os.Stdin.Read(buffer[:]) {
	}
	if err != nil {
		die(err)
	}
	return buffer[0]
}

/*** input ***/

func editorProcessKeypress() {
	c := editorReadKey()
	switch c {
	case 'q' & 0x1f:
		io.WriteString(os.Stdout, "\x1b[2J");
		io.WriteString(os.Stdout, "\x1b[H");
		disableRawMode()
		os.Exit(0)
	}
}

/*** output ***/
func editorRefleshScreen() {
	io.WriteString(os.Stdout, "\x1b[2J");
	io.WriteString(os.Stdout, "\x1b[H");
}

/*** init ***/
func main() {
	enableRawMode()
	defer disableRawMode()

	for {
		editorRefleshScreen()
		editorProcessKeypress()
	}
}
