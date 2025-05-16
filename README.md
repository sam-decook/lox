# lox
Implementation of the Lox language, following [Crafting Interpreters](https://craftinginterpreters.com).

An archived version can be found in the `codecrafters` folder as it was when I finished the
challenge - after roughly 8 months of somewhat sporadic effort (July 9 24 - Mar 7 25)!
And of course they came out with the Classes and Inheritance extensions right after I
finished. Those were finished on May 15 25.

There is still a lot of work I would like to put into it. First and foremost: testing! I got
tripped up a few times when I made sweeping changes. I don't know if I should take the tests from
CodeCrafters, or use the test suite included with the book. Also on the list
 - refactor to use separate modules
 - put Tokens in the AST to have line numbers in the error messages
 - improve error message functions and make them consistent.

I will likely just work on the test suite, because I plan to follow the book's bytecode
interpreter, to get some experience with writing C, and then port that to Zig to try out that
language. So one test suite for all three projects would be very useful.
