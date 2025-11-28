# v0.9.3 (2025-11-16)

* *(unicodeurls,confusables,emojirunes,variationselector)* Updated to Unicode v17.
* *(dbutil)* Added option to log all queries without arguments.
* *(exmaps)* Added non-synchronous equivalent of `exsync.Set`.
* *(exslices)* Added utilities for deleting items by value.
* *(exslices)* Added non-synchronous `Stack` type.
* *(exstrings)* Added `LongestCommonPrefix`.
* *(shlex)* Added support for line continuations with backslashes.
* *(progver)* Fixed linkified version for tags.

# v0.9.2 (2025-10-16)

* *(progver)* Added program version calculation utility like the one used by
  mautrix bridges and Meowlnir.
* *(dbutil)* Added `sqlite3-fk-wal-fullsync` driver which is otherwise
  equivalent to `sqlite3-fk-wal`, but sets `PRAGMA synchronous=FULL` for better
  crash resistance.
* *(dbutil)* Added explicit error if comment prefix (`--`) isn't at the start of
  the line when using dialect filters with the `(lines commented)` modifier.
* *(exsync)* Added NewMapWithData, Clear, Len and CopyFrom methods for maps.
* *(exsync)* Added iterators for maps and sets.
* *(jsontime)* Changed `Unix*()` methods and `jsontime.U*Int()` functions to
  treat 0 and the zero `time.Time` value as the same.

# v0.9.1 (2025-09-16)

* *(dbutil)* Added general documentation.
* *(random)* Added `StringCharset` for generating a random string with a custom
  character set and `AppendSequence` for generating a random slice with a
  completely arbitrary types.
* *(exslices)* Added methods for deduplicating a slice by custom key.
* *(exsync)* Added `WaitTimeoutCtx` for waiting for an `Event` with both
  a timeout and a context.

# v0.9.0 (2025-08-16)

* Bumped minimum Go version to 1.24.
* **Breaking change *(exhttp)*** Refactored HandleErrors middleware to take raw
  response data instead of functions returning response data.
* *(requestlog)* Added option to recover and log panics.
* *(exhttp)* Added `syscall.EPIPE` to `IsNetworkError` checks.
* *(exsync)* Added `Notify` method for waking up all `Event` waiters without
  setting the flag. This is the atomic equivalent of `Set()` immediately
  followed by `Clear()`.
* *(exbytes)* Added `UnsafeString` method for converting a byte slice to a
  string without copying.
* *(exstrings)* Added `CollapseSpaces` to replace multiple sequential spaces
  with one.
* *(exstrings)* Added `PrefixByteRunLength` to count the number of occurrences
  of a given byte at the start of a string.
* *(base58)* Fixed panic when input contains non-ASCII characters.

# v0.8.8 (2025-06-16)

* *(requestlog)* Added option to log `X-Forwarded-For` header value.
* *(exstrings)* Added `LongestSequenceOfFunc` as a customizable version of
  `LongestSequenceOf`

# v0.8.7 (2025-05-16)

* *(jsonbytes)* Added utility for url-safe base64 to complement the existing
  standard unpadded base64 marshaling utility.
* *(exstrings)* Added `LongestSequenceOf` to find the longest sequence of a
  single character in a string.
* *(requestlog)* Implemented `Flush` in `CountingResponseWriter` to fix flushing
  HTTP response buffer when using request logging.
* *(exhttp)* Added utility for checking if a given error is a network error or
  an http2 stream error.

# v0.8.6 (2025-03-16)

* *(curl)* Added support for parsing cookies set using the `-b` flag, which
  recent versions of Chrome use.
* *(exstrings)* Added functions for hashing and constant time comparing strings
  without copying to a byte array.

# v0.8.5 (2025-02-16)

* Bumped minimum Go version to 1.23.
* *(dbutil)* Deprecated `NewRowIter` as it encourages bad error handling.
  `NewRowIterWithError` and `ConvertRowFn[T].NewRowIter` are recommended instead,
  as they support bundling an error inside the iterator.
* *(exslices)* Added utility to map and filter a slice in one go.
* *(confusable)* Fixed skeleton incorrectly including replacement characters
  for some input strings.
* *(exbytes)* Added utility that implements `io.Writer` for byte slices without
  resizing.
* *(glob)* Added `ToRegexPattern` helper which converts a glob to a regex
  without compiling it.

# v0.8.4 (2025-01-16)

* *(dbutil)* Added option to retry transaction begin calls.
* *(dbutil)* Added `QueryHelper.QueryManyIter` function to get a `RowIter`
  instead of pre-reading all rows into a list.
* *(jsontime)* Added utilities for durations.

# v0.8.3 (2024-12-16)

* *(exhttp)* Added global flag for disabling automatic CORS headers when using
  JSON response helper functions.

# v0.8.2 (2024-11-16)

* *(ffmpeg)* Added wrapper functions for `ffprobe`.
* *(emojirunes)* Added method to check if a string is only emojis.
* *(unicodeurls)* Updated data sheets used by emojirunes, variationselectors
  and other packages to Unicode 16.
* *(dbutil)* Added support for mass inserts with no static parameters.

# v0.8.1 (2024-10-16)

* **Breaking change *(lottie)*** Improved interface to take a destination file
  name rather than returning bytes. The method was internally using a file
  anyway, so forcing reading it into memory was a waste.
* *(ffmpeg)* Added `ConvertPathWithDestination` to specify destination file
  manually.
* *(exhttp)* Added utility for applying middlewares to any HTTP handler.
* *(exfmt)* Made duration formatting more customizable.
* *(dbutil)* Changed table existence checks during schema upgrades to properly
  return errors instead of panicking.
* *(dbutil)* Fixed `sqlite-fkey-off` transaction mode.

# v0.8.0 (2024-09-16)

* *(dbutil)* Changed litestream package to allow importing as no-op even when
  cgo is disabled.
* *(ptr)* Added `NonZero` and `NonDefault` helpers to get nil if the value is
  zero/default or a pointer to the value otherwise.
* *(ffmpeg)* Fixed files not being removed if conversion fails.
* *(pblite)* Added pblite (protobuf as JSON arrays) en/decoder.
* *(exhttp)* Added utilities for JSON responses, CORS headers and other things.
* *(glob)* Added utility for parsing Matrix globs into efficient matchers, with
  a fallback to regex for more complicated patterns.
* *(exsync)* Added `Size`, `Pop`, `ReplaceAll` and `AsList` for `Set`.
* *(variationselector)* Fixed plain numbers being emojified by `Add`.

# v0.7.0 (2024-08-16)

* Bumped minimum Go version to 1.22.
* *(curl)* Added `Parse` function to parse a curl command exported from browser
  devtools.
* *(exfmt)* Moved `FormatCurl` to `curl` package.
* *(exslices)* Added `DeduplicateUnsorted` utility for deduplicating items in a
  list while preserving order.
* *(exsync)* Deprecated `ReturnableOnce` in favor of the standard library's
  [`sync.OnceValues`].
* *(exsync)* Added `Event` which works similar to Python's [`asyncio.Event`].
* *(confusable)* Added implementation of confusable detection from [UTS #39].
* *(dbutil)* Added deadlock detection option which panics if a database call is
  made without the appropriate transaction context in a goroutine which
  previously entered a database transaction.

[UTS #39]: https://www.unicode.org/reports/tr39/#Confusable_Detection
[`sync.OnceValues`]: https://pkg.go.dev/sync#OnceValues
[`asyncio.Event`]: https://docs.python.org/3/library/asyncio-sync.html#asyncio.Event

# v0.6.0 (2024-07-16)

* *(dbutil)* Added `-- transaction: sqlite-fkey-off` mode to upgrades, which
  allows safer upgrades that disable foreign keys without having to disable
  transactions entirely.
  * **Breaking change:** `UpgradeTable.Register` now takes a TxnMode instead
    of a bool as the 5th parameter.
  * **Breaking change:** `Database.DoTxn` now takes `*dbutil.TxnOptions`
    instead of `*sql.TxOptions`. `nil` is still allowed and the existing fields
    are still supported, but there's a new field too.
  * **Breaking change:** `Database.Conn` was renamed to `Execable` to avoid
    confusion with the new `AcquireConn` method (`Execable` just returns the
    current transaction or database, `AcquireConn` acquires a connection for
    exclusive use).
* *(dbutil)* Added finalizer for RowIter to panic if the rows aren't iterated.
* *(dbutil)* Changed `QueryOne` to return a zero value (usually nil) if the
  `Scan` method returns an error.
* *(progress)* Implemented `io.Seeker` in `Reader`.
* *(ptr)* Added new utilities for creating pointers to arbitrary values, as
  well as safely dereferencing pointers and making shallow clones.
* *(exslices)* Added functions to cast slices to different types.
* *(gnuzip)* Added wrappers for gzip that operate on `[]byte`s instead of
  `io.Reader`/`Writer`s.
* *(lottie)* Added wrapper for [lottieconverter] similar to ffmpeg.
* *(variationselector)* Fixed edge cases where `Add` and `FullyQualify`
  produced invalid output.

[lottieconverter]: https://github.com/sot-tech/LottieConverter

# v0.5.0 (2024-06-16)

* **Breaking change *(configupgrade)*** Changed `Helper` into an interface.
* *(configupgrade)* Added `ProxyHelper` that prepends a given path to all calls
  to a target `Helper`.
* *(dbutil)* Added support for notating line filters as `(line commented)` to
  indicate that they should be uncommented when the filter matches.
* *(dbutil)* Prevented accidentally using the transaction of another database
  connection by mixing contexts.
* *(fallocate)* Added utility for allocating file space on disk.
  Currently compatible with Linux (including Android) and macOS.
* *(requestlog)* Added utility for HTTP access logging.
* *(progress)* Added `io.Reader` and `io.Writer` wrappers that support
  monitoring progress of the reading/writing.

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
