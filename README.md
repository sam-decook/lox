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

## clox
I did not write a test suite. I've never written a big project in C, so I just pretty much copy-pasted his bytecode VM, which meant I didn't need to worry about testing. Most likely at least. Also, I tried to use the test suite included in the repository and couldn't get it to work. If I remember correctly from a few months ago, there were issues with Dart, and possibly dependencies.

## Testing
I wrote a simple testing framework that compares the output of a reference implementation of clox to the output of your implementation. It alerted me to a few bugs in my implementation, I made a few mistakes while copying over the code.

It skips the tests in the benchmark folder since they print out the running time and it does not handle the tests in the scanning suite appropriately.

Also, this would be a great opportunity to use Go's concurrency to speed up testing.

| Implementation | Passed | Failed | Speed |
|---|---|---|---|
| codecrafters final | 129 | 125 | 70.6% |
| codecrafters final + `-no-fail-stderr` | 236 | 18 | 70.6% |
| clox | 254 | 0 | 99.5% |
| clox + `-O3` | 254 | 0 | 100.3% |
