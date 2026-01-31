package evaluator

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"go/ast"
	"strings"
)

// callBuiltinToBase64 implements the toBase64(string) function.
func callBuiltinToBase64(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("toBase64 requires exactly 1 argument: string", e.Pos())
	}

	text, err := assertStringArg(args[0], 0, "toBase64", e)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.EncodeToString([]byte(text)), nil
}

// callBuiltinFromBase64 implements the fromBase64(string) function.
func callBuiltinFromBase64(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("fromBase64 requires exactly 1 argument: string", e.Pos())
	}

	encoded, err := assertStringArg(args[0], 0, "fromBase64", e)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("fromBase64 expects a valid base64 string: %s", err), e.Pos())
	}

	return string(decoded), nil
}

// callBuiltinHash implements the hash(string, algorithm) function.
func callBuiltinHash(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("hash requires exactly 2 arguments: string, algorithm", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "hash", e)
	if err != nil {
		return nil, err
	}

	algorithm, err := assertStringArg(args[1], 2, "hash", e)
	if err != nil {
		return nil, err
	}

	normalized := strings.ToUpper(algorithm)
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")

	switch normalized {
	case "MD5":
		sum := md5.Sum([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "SHA1":
		sum := sha1.Sum([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "SHA256":
		sum := sha256.Sum256([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "SHA384":
		sum := sha512.Sum384([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "SHA512":
		sum := sha512.Sum512([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	default:
		return nil, newPosError("hash expects algorithm to be one of: MD5, SHA-1, SHA-256, SHA-384, SHA-512", e.Pos())
	}
}

// callBuiltinToHex implements the toHex(string) function.
func callBuiltinToHex(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("toHex requires exactly 1 argument: string", e.Pos())
	}

	text, err := assertStringArg(args[0], 0, "toHex", e)
	if err != nil {
		return nil, err
	}

	return hex.EncodeToString([]byte(text)), nil
}

// callBuiltinFromHex implements the fromHex(string) function.
func callBuiltinFromHex(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("fromHex requires exactly 1 argument: string", e.Pos())
	}

	encoded, err := assertStringArg(args[0], 0, "fromHex", e)
	if err != nil {
		return nil, err
	}

	decoded, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("fromHex expects a valid hex string: %s", err), e.Pos())
	}

	return string(decoded), nil
}
