.section .text
.type main.getNumber, @function
.globl main.getNumber
main.getNumber:
    movq $42, %rax
    retq
