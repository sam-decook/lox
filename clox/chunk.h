#ifndef clox_chunk_h
#define clox_chunk_h

#include "common.h"
#include "value.h"

typedef enum {
    OP_CONSTANT,
    OP_NIL,
    OP_TRUE,
    OP_FALSE,
    OP_POP,
    OP_GET_LOCAL,
    OP_SET_LOCAL,
    OP_GET_GLOBAL,
    OP_DEFINE_GLOBAL,
    OP_SET_GLOBAL,
    OP_GET_UPVALUE,
    OP_SET_UPVALUE,
    OP_GET_PROPERTY,
    OP_SET_PROPERTY,
    OP_EQUAL,
    OP_GREATER,
    OP_LESS,
    OP_ADD,
    OP_SUBTRACT,
    OP_MULTIPLY,
    OP_DIVIDE,
    OP_NOT,
    OP_NEGATE,
    OP_PRINT,
    OP_JUMP,
    OP_JUMP_IF_FALSE,
    OP_LOOP,
    OP_CALL,
    OP_INVOKE,
    OP_CLOSURE,
    OP_CLOSE_UPVALUE,
    OP_RETURN,
    OP_CLASS,
    OP_METHOD,
} OpCode;

typedef struct {
    int count;
    int capacity;
    uint8_t* code;
    int* lines;
    ValueArray constants;
} Chunk;

void initChunk(Chunk *chunk);
void writeChunk(Chunk* chunk, uint8_t byte, int line);
int addConstant(Chunk* chunk, Value value);
void freeChunk(Chunk* chunk);

#endif

/* Challenges:
 * Instead of storing the line number for each instruction (many instructions
 * will have the same line number), develop a compressed encoding and
 * implement a getLine() function. Sacrifice a bit of speed during decomp for
 * less memory usage overall.
 * - array of first instruction number to use a line
 * - array of that line number (wait, necessary?)
 *   - no for one-file programs (index + 1 = line number)
 * - binary search the first array (heck, a linear scan might be fast enough)
 *
 * OP_CONSTANT uses one byte for its operand, allowing only 256 constants. That
 * should be enough for most programs, but not all. Make a new OP_CONSTANT_LONG
 * instruction that stores it in 3 bytes (24 bits, 4 bytes in total).
 *
 * What other binary operators could we eliminate to make our bytecode simpler?
 *
 * Conversely, we can speed up our VM by adding more instructions, e.g. a
 * dedicated greater_than_or_equal.
 */
