<h1 align="center">
  crossplane-explorer
</h1>

<p align="center">
  🧰 Enhanced Crossplane explorer
</p>

![screenshot](./screenshot.png)

`crossplane trace` is a very handy tool, but it is not very interactive and requires a few extra
hops to properly debug its traced objects. This tool aims on closing this gap by providing
an interactive tracing explorer based on the tracer output.

## ✨ Features

### Trace

- ✨ Expanded details at a glance, with highlight colouring for possible issues
- 📖 Show YAML objects from the explorer, with no need to do it separately in kubectl
- 📖 Clean object YAMLs without `managedFields` (useful on apply, not as much on describe/get)
- 📋 Copy full qualified objects names straight from UI (API group + Kind + name)
- ♻️ Automatic refresh

### Upcoming

- Allow mutating resource annotations (pause, finaliser)

## 📀 Install

### Linux and Windows

[Check the releases section](https://github.com/brunoluiz/crossplane-explorer/releases) for more information details.

### MacOS

```
brew install brunoluiz/tap/crossplane-explorer
```

### Other

Use `go install` to install it

```
go install github.com/brunoluiz/crossplane-explorer@latest
```

## ⚙️ Usage

You must have `crossplane` installed, since this application can run with any crossplane CLI version.

```
crossplane-explorer trace -n namespace Object/hello-world
```

## 🧾 To-do

- Add kubectl describe on `Enter` or `d` press
- Open issue around issues on colour rendering on tables bubbles (reason why I had to fork)
- Open issue at crossplane so people can use their tracing parsers
- Add search capability to the `viewer`
