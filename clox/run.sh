#!/bin/zsh

clang -o clox_interpreter *.c -Wall -Wextra | less
./clox_interpreter $1
