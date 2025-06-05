#!/bin/zsh

echo "Compiling..."
clang -o clox_interpreter *.c -Wall -Wextra
echo
echo "Running..."
./clox_interpreter $1
