#include "textflag.h"

// func asmCopy(dst, src unsafe.Pointer, size uintptr)
TEXT ·asmCopy(SB), NOSPLIT, $0-24
    MOVQ dst+0(FP), DI
    MOVQ src+8(FP), SI
    MOVQ size+16(FP), CX

    // 使用 REP MOVSB 指令
    REP; MOVSB

    RET
 