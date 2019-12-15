package sm3

// #cgo LDFLAGS: -L../../librarys -lsmcryptokit -lcrypto
// #include "../../include/sm/smcryptokit.h"
import "C"
import "unsafe"

import (
	// "fmt"
	"hash"
)

type sm3Context struct {
	total  [2]uint32
	state  [8]uint32
	buffer [64]byte

	ipad [64]byte
	opad [64]byte
}

func New() hash.Hash {
	sm3ctx := &sm3Context{}
	sm3ctx.Reset()
	// fmt.Println("New SM3++++++++++++++++++++++")
	return sm3ctx
}

// func (sm3ctx *sm3Context) Free() {
// 	C.FreeSm3Ctx(unsafe.Pointer(sm3ctx.ctx))
// }

func (sm3ctx *sm3Context) Reset() {
	C.Sm3Starts(unsafe.Pointer(&sm3ctx.total[0]), C.int(len(sm3ctx.total)),
		unsafe.Pointer(&sm3ctx.state[0]), C.int(len(sm3ctx.state)))
}

func (sm3ctx *sm3Context) Size() int {
	return 32
}

func (sm3ctx *sm3Context) BlockSize() int {
	return 64
}

func (sm3ctx *sm3Context) Write(p []byte) (nn int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// C.Sm3Update(sm3ctx.ctx, unsafe.Pointer(&p[0]), C.int(len(p)))
	C.Sm3Update(unsafe.Pointer(&sm3ctx.total[0]), C.int(len(sm3ctx.total)),
		unsafe.Pointer(&sm3ctx.state[0]), C.int(len(sm3ctx.state)),
		unsafe.Pointer(&sm3ctx.buffer[0]), C.int(len(sm3ctx.buffer)),
		unsafe.Pointer(&p[0]), C.int(len(p)))
	return len(p), nil
}

func (sm3ctx *sm3Context) Sum(in []byte) []byte {
	h := make([]byte, 32, 32)
	total := sm3ctx.total
	state := sm3ctx.state
	buffer := sm3ctx.buffer
	// C.Sm3Finish(sm3ctx.ctx, unsafe.Pointer(&h[0]), C.int(32))
	C.Sm3Finish(unsafe.Pointer(&total[0]), C.int(len(total)),
		unsafe.Pointer(&state[0]), C.int(len(state)),
		unsafe.Pointer(&buffer[0]), C.int(len(buffer)),
		unsafe.Pointer(&h[0]), C.int(32))
	// C.FreeSm3Ctx(sm3ctx.ctx)
	// fmt.Printf("sm3 hash: %x\n", h)
	return append(in, h...)
}
