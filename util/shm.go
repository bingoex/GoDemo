package main

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	KIpcCreate = 1 << 9
	KIpcExcl   = 2 << 9
	kShmSize   = 4096 * 100
)

func attachShm(key, size int) (addr *uint8, err error) {
	var (
		shmid    uintptr
		shmAddr  uintptr
		sysErr   syscall.Errno
		shmExist bool = true
	)

	shmid, _, sysErr = syscall.RawSyscall(syscall.SYS_SHMGET,
		uintptr(key), uintptr(size), uintptr(0666))
	if sysErr != 0 {
		shmExist = false
		fmt.Println("No exist")
		shmid, _, sysErr = syscall.RawSyscall(syscall.SYS_SHMGET,
			uintptr(key), uintptr(size), uintptr(0666|KIpcCreate))
		if sysErr != 0 {
			return nil, errors.New(sysErr.Error())
		}
	}

	shmAddr, _, sysErr = syscall.RawSyscall(syscall.SYS_SHMAT, uintptr(shmid), 0, 0)
	if sysErr != 0 {
		return nil, errors.New(sysErr.Error())
	}

	addr = (*uint8)(unsafe.Pointer(shmAddr))

	// memset to zero
	if false == shmExist {
		c := (*[kShmSize]uint8)(unsafe.Pointer(shmAddr))
		for i := 0; i < size; i++ {
			c[i] = 0x0
		}
	}

	return addr, nil
}

func failOnErr(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	_, err := attachShm(0x92431, kShmSize)
	failOnErr(err)
	for {
	}
}
