package runtime

import "unsafe"

// https://github.com/torvalds/linux/blob/v5.5/arch/x86/entry/syscalls/syscall_64.tbl
const __x64_sys_write = 1
const __x64_sys_brk  = 12
const __x64_sys_exit = 60

func brk(addr uintptr) uintptr {
	var ret uintptr = Syscall(__x64_sys_brk, addr, 0, 0)
	return ret
}

func write(fd int, addr *byte, length int) {
	Syscall(__x64_sys_write, uintptr(fd), uintptr(unsafe.Pointer(addr)), uintptr(length))
}

func exit(code int) {
	Syscall(__x64_sys_exit, uintptr(code), 0 , 0)
}

