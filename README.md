# envelope

Turn environment variables into a YAML document. Use it as a command that
prints YAML to stdout, or as a Go library that hands you the parsed tree.

## Motivation

Containers configure well through environment variables and badly through
files: a variable can be set in a Dockerfile, a compose file, a Kubernetes
manifest or a cluster manager UI, while a config file has to be baked into the
image or mounted from somewhere. But applications want structured config —
nested sections, lists — and an environment is flat.

envelope treats the variable name as a path into a YAML document, so the flat
environment can describe an arbitrarily nested structure. It is the same idea
[container-supervisor](https://github.com/BaseCrusher/container-supervisor) uses
for its `SUPERVISOR_`-prefixed overrides, pulled out into its own tool and
extended with lists.

## Usage

```
envelope [-prefix PREFIX] > config.yml
```

- `-prefix` — only convert variables whose name starts with this prefix. The
  prefix is stripped from the resulting key. Without it every variable in the
  environment is converted, including `PATH` and `HOME`, so in practice you
  always want one.

```sh
export APP_LOGLEVEL=info
export APP_PROCESSES__DB__PATH=/bin/postgres
export APP_PROCESSES__DB__PORT=5432
export APP_PROCESSES__DB__ARGUMENTS__0=--verbose
export APP_PROCESSES__DB__ARGUMENTS__1='--limit 10'

envelope -prefix APP_
```

```yaml
LOGLEVEL: info
PROCESSES:
    DB:
        ARGUMENTS:
            - --verbose
            - --limit 10
        PATH: /bin/postgres
        PORT: 5432
```

Keys keep the case they were written in. Environment variables are
conventionally uppercase, so the output is too; write the variable in the case
your consumer expects (`APP_processes__db__path`) if that matters.

## Naming

The variable name, minus the prefix, is a path. Segments are separated by `__`
(double underscore).

| Name | Result |
|---|---|
| `A__B__C=x` | nested mappings — `A: {B: {C: x}}` |
| `A__0=x`, `A__1=y` | a list — a segment that is **all digits** is an index |
| `A__0__PORT=80` | a list of mappings |
| `A__0__0=x` | nested lists |
| `ADDRESS_LINE_1=x` | a single `_` is an ordinary character, this is one key |
| `A___80__TLS=true` | escape — a leading `_` is stripped and the rest is always a name, so this is the key `"80"`, not index 80 |

Indexes are sorted numerically, so `__2` comes before `__10`. Gaps are closed:
`__0` and `__7` alone produce a two-element list.

## Values

Values are parsed as YAML scalars, so the types come out right:

| Variable | YAML |
|---|---|
| `PORT=8080` | `PORT: 8080` (int) |
| `DEBUG=true` | `DEBUG: true` (bool) |
| `NOTE=~` | `NOTE: null` |
| `ARGS=[]` | `ARGS: []` (empty list) |
| `ENV={}` | `ENV: {}` (empty mapping) |
| `PORT='"8080"'` | `PORT: "8080"` (forced string) |
| `DESC=key: value` | `DESC: 'key: value'` (stays a string) |

A value that would parse as a *non-empty* list or mapping is kept as a string,
so connection strings and anything else containing YAML punctuation survive
untouched. `[]` and `{}` are honoured because they are the only way to express
an empty collection. Quote a value to force it to stay a string.

## Errors

The command exits non-zero and prints the offending variable when the
environment does not describe a valid document:

- `APP_DB=x` together with `APP_DB__PORT=5432` — a key cannot be both a value
  and a block.
- `APP_ARGS__0=a` together with `APP_ARGS__NAME=b` — a block cannot mix list
  indexes and names.
- `APP_A____B=x` — an empty path segment. A key that itself starts with `_`
  cannot be escaped.

## As a library

```
go get github.com/BaseCrusher/envelope
```

`Marshal` gives you the document as YAML bytes, `Parse` gives you the tree
(`map[string]any`, or `[]any` if the top level is a list) so you can merge it
with config from elsewhere before decoding.

```go
package main

import (
	"fmt"
	"os"

	"github.com/BaseCrusher/envelope"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel string   `yaml:"LOGLEVEL"`
	Args     []string `yaml:"ARGS"`
}

func main() {
	out, err := envelope.Marshal(os.Environ(), "APP_")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var cfg Config
	if err := yaml.Unmarshal(out, &cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", cfg)
}
```

Both functions take the environment as a `[]string` in `KEY=VALUE` form rather
than reading it themselves, so tests can pass a fixed slice and callers can feed
in variables from somewhere other than the process environment.

## In a container

The binary is static and depends on nothing, so it can generate the config file
in a build stage and never ship in the final image:

```dockerfile
FROM alpine AS config
COPY envelope /usr/local/bin/
ARG APP_LOGLEVEL=info
RUN envelope -prefix APP_ > /config.yml

FROM gcr.io/distroless/static-debian12
COPY --from=config /config.yml /etc/app/config.yml
```

Or ship it and run it at startup, so the environment of the running container
decides the config:

```sh
envelope -prefix APP_ > /etc/app/config.yml && exec /bin/app
```

## Not supported

YAML features with no place to live in a variable name or value:

- anchors, aliases and merge keys (`<<`) — no way to reference another node
- explicit tags (`!!binary`, `!MyType`) — quoting covers the only common case,
  `!!str`
- multiple documents (`---`) and comments
- mapping keys that are not strings — the escape produces the *string* `"80"`,
  not the integer `80`, so a document keyed by ints, bools or floats is out of
  reach

## Building

```sh
go build ./cmd/envelope
go test ./...
```
