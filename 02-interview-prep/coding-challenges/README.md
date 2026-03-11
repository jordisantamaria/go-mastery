# Coding Challenges — Idiomatic Go Algorithms

Classic algorithm and data structure problems solved **idiomatically in Go**, not "Java translated to Go".

## Approach

In Go interviews, interviewers look for:

1. **Correct use of slices, maps, and the standard library**: don't reinvent what already exists.
2. **Clean and readable code**: Go prioritizes clarity over cleverness.
3. **Handling of edge cases**: nil inputs, empty slices, empty strings.
4. **Algorithmic complexity**: being able to explain Big-O for time and space.
5. **Idiomatic Go**: short naming, explicit error handling, useful zero values.

## Structure

- **`challenges.go`**: 10 functions with `panic("TODO")`. Your task is to implement them.
- **`challenges_test.go`**: complete tests with table-driven tests and edge cases. Run with `go test`.
- **`solutions.go.txt`**: complete solutions with explanations. Rename it to `solutions.go` and delete `challenges.go` to verify that the tests pass.

## Challenges

| # | Function | Description | Difficulty |
|---|---|---|---|
| 1 | `TwoSum` | Find two numbers that add up to a target | Easy |
| 2 | `ValidParentheses` | Verify balanced brackets | Easy |
| 3 | `MergeSortedArrays` | Merge two sorted arrays | Easy |
| 4 | `ReverseLinkedList` | Reverse a linked list | Easy |
| 5 | `LRUCache` | Implement an LRU cache with Get/Put | Medium |
| 6 | `MaxSubarraySum` | Subarray with maximum sum (Kadane) | Medium |
| 7 | `BinarySearch` | Classic binary search | Easy |
| 8 | `LevelOrderTraversal` | BFS on a binary tree | Medium |
| 9 | `IsAnagram` | Verify if two strings are anagrams | Easy |
| 10 | `TopKFrequent` | K most frequent elements | Medium |

## How to Use

```bash
# Run tests (they will fail until you implement the solutions)
cd 02-interview-prep/coding-challenges
go test -v

# See which tests fail
go test -v -run TestTwoSum

# Use the solutions to verify
cp solutions.go.txt solutions.go
rm challenges.go
go test -v
```

## Interview Tips

- **Start with edge cases**: what happens with empty input? nil? a single element?
- **Explain your approach before coding**: "I'm going to use a hash map for O(1) lookup..."
- **Mention the complexity**: "This is O(n) in time and O(n) in space."
- **Use the zero value**: in Go, the zero value of a `map[K]V` when accessing a non-existent key is the zero value of V. Take advantage of it.
- **Name well**: `i, j` for indices, `n` for size, `ok` for existence booleans.
