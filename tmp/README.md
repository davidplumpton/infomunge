# tmp

Scratch files and one-off helper programs live here.

This directory is a nested Go module so repo-root `go test ./...` ignores ad hoc `package main` files in `tmp/`.

Run helpers from this directory, for example:

```bash
go run ./inspect.go
```
