# tmp

Scratch files and one-off helper programs live here.

This directory is a nested Go module so repo-root `go test ./...` ignores ad hoc `package main` files in `tmp/`.

Run helpers from this directory by passing the helper file you created:

```bash
go run ./your-helper.go
```
