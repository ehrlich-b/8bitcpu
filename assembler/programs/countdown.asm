    LDAi $5
    OUT
subtract:
    LDBi $1
    SUB
    OUT
    JZ end
    JMP subtract
end:
    HLT
