# watchport

`watchport` is a small CLI that checks HTTP(S) URLs in parallel and reports their status and response time.

## Build

Requires Go 1.27 or newer.

```sh
go build -o watchport .
```

## Usage

```sh
./watchport -timeout 3s https://example.com https://api.github.com
```

`-timeout` sets the maximum wait for each URL. It defaults to 5 seconds. A `2xx` response is successful; network errors, timeouts, and other HTTP statuses are reported as failures.

Example output:

```text
OK https://example.com HTTP 200 (142ms)
OK https://api.github.com HTTP 200 (218ms)
```

The process exits with code `0` when every check succeeds, `1` when a check fails, and `2` for invalid arguments.

## Development

```sh
go test ./...
go vet ./...
```
