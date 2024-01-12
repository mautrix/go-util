# v0.3.0 (unreleased)

* **Breaking change *(dbutil)*** Removed all non-context methods.
* *(dbutil)* Added query helper to reduce boilerplate with executing database
  queries and scanning results.
* *(exsync)* Added generic `Set` utility that wraps a valueless map with a mutex.
* *(exerrors)* Added `Must` helper to turn `(T, error)` returns into `T` or panic.
* *(ffmpeg)* Added `Supported` and `SetPath` for checking if ffmpeg is available
  and overriding the binary path respectively.

# v0.2.1 (2023-11-16)

* *(dbutil)* Fixed read-only db close error not including actual error message.

# v0.2.0 (2023-10-16)

* *(jsontime)* Added helpers for unix microseconds and nanoseconds, as well as
  alternative structs that parse JSON strings instead of ints (all precisions).
* *(exzerolog)* Added generic helpers to generate `*zerolog.Array`s out of slices.
* *(exslices)* Added helpers for finding the difference between two slices.
  * `Diff` is a generic implementation using maps which works with any
    `comparable` types (i.e. types that have the equality operator `==` defined).
  * `SortedDiff` is a more efficient implementation which can take any types
     (using the help of a `compare` function), but the input must be sorted and
     shouldn't have duplicates.

# v0.1.0 (2023-09-16)

Initial release
