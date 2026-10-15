# debug

Diagnostics utilities. Not for production use.

## Tracer

Prefixed, indented prints for call-graph visibility. `Enter` increases
indentation; the returned func reverses it on `defer`. Lines are atomic
across goroutines (global mutex around each write).

```go
var trace = debug.NewTracer("fetcher")

func (f *Fetcher) Next() (*Doc, error) {
    defer trace.Enter("Fetcher.Next", f)()
    trace.Println("acquiring iterator prefix=%s", f.prefix)

    doc, err := f.read()
    if err != nil {
        trace.Println("read FAILED: %v", err)
        return nil, err
    }
    trace.Println("read OK docID=%s", doc.ID)
    return doc, nil
}
```

Output (note the indentation tracks the call stack, source-location is on):

```text
fetcher:      > Fetcher.Next (14000af1b90)        | fetcher.go:42
fetcher:          acquiring iterator prefix=/col/1 | fetcher.go:43
fetcher:          > Fetcher.read (14000af1b90)    | fetcher.go:88
fetcher:              loaded 312 bytes            | fetcher.go:101
fetcher:          read OK docID=bae-123           | fetcher.go:50
```

Disable a tracer (`Disable()`) to silence one subsystem without recompiling.
Globals: `debug.PrefixWidth` (default 12), `debug.ShowSourceLocation` (default true).

## Timeline

Process-wide ordered log of cross-goroutine events. Each `Log` records
elapsed-since-first-call, an actor tag, and the message; output is buffered
until `Render()`. Use when temporal ordering across goroutines is the thing
you're trying to understand — Tracer prints atomically per line but can't
preserve event order across racing writers.

```go
func TestPushlogGrantRace(t *testing.T) {
    debug.DefaultTimeline.Enable()
    t.Cleanup(func() {
        fmt.Print(debug.DefaultTimeline.Render())
        debug.DefaultTimeline.Disable()
    })

    // Production code calls debug.DefaultTimeline.Log(...) from
    // sender + receiver goroutines.
    runScenario(t)
}
```

Output, after one failing run of the demess pushlog/grant-race scenario:

```text
T+    0ms  TEST    enabled
T+ 4094ms  N-recv  pushlog arrived (isReplicator=false)
T+ 4094ms  N-recv  -> SourceHub CheckDocAccess (start)
T+ 4097ms  N-recv  <- SourceHub answer: peerHasAccess=false
T+ 4097ms  N-recv  gate=NO  (silent drop, no log, no retry)
T+ 6122ms  N-recv  pushlog arrived
T+ 6124ms  N-recv  gate=YES -> syncDAG
T+ 6124ms  N-send  hasAccess request
T+ 6125ms  N-send  -> SourceHub CheckDocAccess (start)
T+ 6127ms  N-send  <- SourceHub answer: peerHasAccess=true (serve block)
T+ 6127ms  N-recv  fetch OK
```

The 2-second silent drop and the ~3ms sender ACP check are immediately
legible — neither would be obvious in interleaved per-line tracer output.

## ResourceTracker

Tracks acquire/release pairs to catch leaks. One `Track` per acquire,
one `Untrack` per release, with a description that shows up in failure
messages.

```go
var iters = debug.NewResourceTracker("documentFetcher.iter")

func openIterator(prefix string) *Iterator {
    it := newIter(prefix)
    iters.Track(it, "prefix="+prefix)
    return it
}

func (it *Iterator) Close() error {
    iters.Untrack(it)
    return it.inner.Close()
}
```

### Asserting no leaks at end of test

```go
func TestFetcher_ReleasesIterators(t *testing.T) {
    f := newFetcher(...)
    require.NoError(t, f.Run(ctx))
    iters.AssertEmpty(t)   // t.Errorf with the description of each leak
}
```

A failure looks like:

```text
fetcher_test.go:42: ResourceTracker[documentFetcher.iter]: 1 resource(s) still tracked:
    - prefix=/col/1/index/age (addr: 14000138240)
```

### Isolating a leak from a specific block

When other code in the same process holds legitimate long-lived resources,
`Count() > 0` isn't a leak signal. Snapshot before, `AddedSince` after —
the result is exactly what the suspect block left behind:

```go
before := iters.Snapshot()       // pre-existing legitimate work captured

runSuspectBlock(t)

for addr, desc := range iters.AddedSince(before) {
    t.Errorf("leak introduced by suspect block: %s (addr=%x)", desc, addr)
}
```

`StillHeld(snap)` is the inverse — resources from `snap` that haven't been
released — useful for "this should have been closed by now" assertions.

Methods: `Track`, `Untrack`, `Count`, `Remaining`, `Snapshot`, `AddedSince`,
`StillHeld`, `Clear`, `AssertEmpty`, `AssertCount`.
