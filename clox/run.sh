#!/bin/zsh

clang -o clox_interpreter *.c -Wall -Wextra -Wno-unused-parameter | less
./clox_interpreter $1
