# Notes

These are general notes about development around bubbletea and things that I still need to do / think about:

1. It seems `bubbletea` apps do not work well with go routines. When using error groups together with it to
implement the watcher, the `signal.NotifyContext` got affected by it since the app hijacks the input and `ctrl+c`
can't be handled correctly. Since it does not capture the keys, it never emits the SIGINT.
The hack around it was to call the `cancel` when the `tea.Quit` happens and handle `ctrl+c` within the tea app.
  - Fix was released in more recent versions with the introduction of `tea.Interrupt`

2. base16 colors can be used as a way to keep the app colours the same in any machine. See `tui` package.

3. Crossplane does not sadly expose its internals. Everything is in `internal/`.

4. The correct way to use `bubbletea` seems to be by using events. So I tried to refactor and try to expose
only read-only methods through the struct. Anything that would do a mutation should be an event and is handled
within the `events.go` file (might be called `handlers.go` in the future).


## New implementation for tree + search

- `renderTree` is becoming quite complex... Probably it should render the required "row" format into an slice in the Update method
  - Perhaps `tree` package should be responsible to convert tree into a slice of columns
  - `explorer.addNode` seems to already do some recursive job and could be the method that translates it to rows
- `tree` should be called `navigator`
  - `statusbar` should be removed
- Packages should be reorganised as `common`, `layout` and `app`

- In the `xplane` package, we should have something to output it
