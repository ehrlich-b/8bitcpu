
LDAi $0

loop:
    CALL add1
    OUT
    LDBi $10
    JGE end
    JMP loop

add1:
    LDBi $1
    ADD
    RET

end:
    HLT
