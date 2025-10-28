# ymldiff

[![Test](https://github.com/consi/ymldiff/actions/workflows/test.yml/badge.svg)](https://github.com/consi/ymldiff/actions/workflows/test.yml)
[![Release](https://github.com/consi/ymldiff/actions/workflows/release.yml/badge.svg)](https://github.com/consi/ymldiff/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/consi/ymldiff)](https://goreportcard.com/report/github.com/consi/ymldiff)

**ymldiff** is an intelligent YAML comparison tool that goes beyond simple text diffs. It understands YAML structure and provides meaningful, colored output showing additions, deletions, and modifications. Greatly improves life when you comparing Kubernetes manifests, helm charts, etc.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap consi/ymldiff https://github.com/consi/ymldiff
brew install ymldiff
```

### Latest version using Go

```bash
go install github.com/consi/ymldiff@latest
```

### Download pre-built binary

Download the latest release from the [releases page](https://github.com/consi/ymldiff/releases).

## Usage

```bash
# Basic comparison
ymldiff old.yaml new.yaml

# Compare without showing comments
ymldiff -c config1.yaml config2.yaml
ymldiff --disable-comments config1.yaml config2.yaml

# Compare without document separator comments
ymldiff -d config1.yaml config2.yaml

# Compare without colors (for piping to files or logs)
ymldiff -n config1.yaml config2.yaml

# Combine multiple options (short flags can be combined)
ymldiff -cd config1.yaml config2.yaml
ymldiff -cdn config1.yaml config2.yaml
```

### Example output:
```
$ ./ymldiff -cdn old.yaml new.yaml
---
~ .data.something: true → false

---
~ .spec.template.spec.containers[zookeeper].env[ZOO_HEAP_SIZE].value: 8192 → 7192
```

## License

MIT License

Copyright (c) 2025 Marek Wajdzik

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
