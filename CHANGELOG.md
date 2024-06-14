# v0.4.2 (2024-04-16)

* *(dbutil)* Added utility for building mass insert queries.
* *(dbutil)* Added utility for using reflect to build a RowIter.

# v0.4.1 (2024-03-16)

* *(exfmt)* Added utility for converting HTTP requests to `curl` commands.
* *(exmime)* Added hardcoded extension override for `audio/mp4` -> `.m4a`.
* *(dbutil)* Added `UnixPtr`, `UnixMilliPtr` and `ConvertedPtr` helpers for
  converting `time.Time` into `*int64` so that zero times are nil and other
  times are unix.
* *(dbutil)* Added `UntypedNil` utility for avoiding typed nils, and `JSONPtr`
  for wrapping a struct in the existing `JSON` utility using `UntypedNil`.
* *(dbutil)* Added periodic logs to `DoTxn` if the transaction takes more than
  5 seconds.

# v0.4.0 (2024-02-16)

* *(jsonbytes)* Added utilities for en/decoding byte slices as unpadded base64.
* *(jsontime)* Fixed serialization of Unix(Micro|Nano)String types.
* *(exzerolog)* Added helper function for setting sensible zerolog globals
  such as CallerMarshalFunc, default loggers and better level colors.
* *(dbutil)* Added helper for wrapping a raw slice in a RowIter.
  * This is useful for interfaces that return RowIters to allow implementing
    the interface without SQL.
  * The RowIter interface may be moved to a separate package in the future to
    further separate it from SQL databases.
* *(dbutil)* Added helper for converting RowIter to map.

# v0.3.0 (2024-01-16)

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
