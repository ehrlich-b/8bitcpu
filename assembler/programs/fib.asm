    LDAi $1
    OUT
    STA $32
    STA $31
loop:
    LDA $32
    LDB $31
    STA $31
    ADD
    OUT
    STA $32
    LDBi $140
    JGE end
    JMP loop
end:
    HLT
