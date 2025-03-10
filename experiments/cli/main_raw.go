/* NOTE(anton2929): inspired by https://gist.github.com/EddieIvan01/4449b64fc1eb597ffc2f317cfa7cc70c. */

package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"
)

/* From <sys/ioccom.h>. */
const (
	IOCPARM_SHIFT = 13                         /* number of bits for ioctl size */
	IOCPARM_MASK  = ((1 << IOCPARM_SHIFT) - 1) /* parameter length mask */

	IOC_OUT = 0x40000000 /* copy out parameters */
	IOC_IN  = 0x80000000 /* copy in parameters */
)

/* From <sys/ttycom.h>. */
const (
	TIOCGETA = uint((IOC_OUT) | ((unsafe.Sizeof(syscall.Termios{}) & IOCPARM_MASK) << 16) | ('t' << 8) | (19)) /* get termios struct */
	TIOCSETA = uint((IOC_IN) | ((unsafe.Sizeof(syscall.Termios{}) & IOCPARM_MASK) << 16) | ('t' << 8) | (20))  /* set termios struct */
)

func Ioctl(fd int, request uint, argp unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(request), uintptr(argp))
	if errno != 0 {
		return fmt.Errorf("ioctl failed with code %v", errno)
	}
	return nil
}

func GetTermios(fd int, termios *syscall.Termios) error {
	return Ioctl(fd, TIOCGETA, unsafe.Pointer(termios))
}

func SetTermios(fd int, termios *syscall.Termios) error {
	return Ioctl(fd, TIOCSETA, unsafe.Pointer(termios))
}

func MakeRaw(fd int) (*syscall.Termios, error) {
	var old syscall.Termios
	if err := GetTermios(fd, &old); err != nil {
		return nil, fmt.Errorf("failed to get previous termios: %v", err)
	}

	/* This attempts to replicate the behaviour documented for cfmakeraw in the termios(3) manpage. */
	var raw syscall.Termios
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8

	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	return &old, SetTermios(fd, &raw)
}

func Restore(fd int, old *syscall.Termios) error {
	return SetTermios(fd, old)
}

func main() {
	oldState, err := MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to switch terminal to RAW mode: %v", err)
	}
	defer Restore(int(os.Stdin.Fd()), oldState)

	if err := App(); err != nil {
		log.Printf("Failed to run application: %v", err)
	}
}
