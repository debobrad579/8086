bits 16

; Signed displacements
mov ax, [bx + di - 37]
mov [si - 300], cx
mov dx, [bx - 32]
mov cx, [bp + si - 128]
mov [bx + 256], dx

; Explicit sizes
mov [bp + di], byte 7
mov [di + 901], word 347
mov [bx + si], byte 25
mov [bp - 64], word 1024

; Direct address
mov bp, [5]
mov bx, [3458]
mov cx, [1024]
mov dx, [65535]

; Memory-to-accumulator test
mov ax, [2555]
mov ax, [16]
mov al, [255]
mov al, [16]
mov ax, [4096]
mov al, [128]

; Accumulator-to-memory test
mov [2554], ax
mov [15], ax
mov [255], al
mov [15], al
mov [4096], ax
mov [128], al
