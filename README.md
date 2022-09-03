# Fake assembly

This is a fake assembly language. It is used for educational purpose only.

To run this program you need **Golang** installed in your machine. Then run:

```sh
$ go run . ./examples/a1.asm
```

The target file name must end with `.asm`.

## How to access memory

We limit the memory to have only 1024 slots. Every slot is initialized with 0.

### 1. Constants

To use constants you just need to use a number without any prefix. Like:

```
123
321
```

Are all constants

### 2. Variables

A variable is a direct access to memory, you can do that by just adding the prefix `$` and an index for the position that you want to access. Suppose that your memory look like this [4,3,6,7,8] so:

```
$1 # Here you are accessing the memory indexed by 1 so of this is 3
$0 # Here you are accessing the memory indexed by 0 so of this is 4
```

### 3. Reference

Reference is an indirect access to memory, where a variable will define the slot position that you want to access. Again suppose that the memory look like this [2,3,8,7] so:

```
&1 # Here you use the memory at '1'(value == 3) as an index, similar to memory[memory[1]] = 7
&0 # Here you use the memory at '0'(value == 2) as an index, similar to memory[memory[0]] = 8
```

## Instructions

### 1. Operation

Write the result of the operation `{C, $, &}1 {-, +, *, /} {C, $, &}2` to memory `{$, &}0`.

```
$0 = 232 + &12
&1 = $12 / 2
```

### 2. Label

To define a label you can do that by just

```
{label}:
```

Replace `{label}` with an appropriate value.

### 3. To

Change the code flow if the predicate is true the pattern is `to {label} if {C, $, &}1 {==, !=, >, <, >=, <=} {C, $, &}1 {&&, ||} ...`.

```
to main if 1 == 1

some_label:
  $0 = $0 + 1
  to main if &1 <= &2 || $1 > 2

main:
  $0 = $0 - 1
  to some_label if $0 > 1 && $0 < 12
```

### 5. Write

Print a value. Suppose that your memory looks like this `[1,2,3,4,5]`.

```
print 1 # Will print `$ 1`
print $1 # Will print `$ [ 1 ] 2`
print &1 # Will print `$ [ 1 -> 2 ] 3`
```

### 6. Read

Read a value from the input file. Will jump to the label provided if unable to read(optional).

```
read $0
read $0 end
read $0 end

end:
  write $0
```
