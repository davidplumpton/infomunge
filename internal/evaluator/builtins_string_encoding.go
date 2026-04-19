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
func callBuiltinToBase64(args []Value, e *ast.CallExpr) (Value, error) {
	text, err := requireOneStringArg(args, "toBase64", e)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.EncodeToString([]byte(text)), nil
}

// callBuiltinFromBase64 implements the fromBase64(string) function.
func callBuiltinFromBase64(args []Value, e *ast.CallExpr) (Value, error) {
	encoded, err := requireOneStringArg(args, "fromBase64", e)
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
func callBuiltinHash(args []Value, e *ast.CallExpr) (Value, error) {
	strs, err := assertStringArgs(args, 2, "hash", e)
	if err != nil {
		return nil, err
	}
	text, algorithm := strs[0], strs[1]

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
func callBuiltinToHex(args []Value, e *ast.CallExpr) (Value, error) {
	text, err := requireOneStringArg(args, "toHex", e)
	if err != nil {
		return nil, err
	}
	return hex.EncodeToString([]byte(text)), nil
}

// callBuiltinFromHex implements the fromHex(string) function.
func callBuiltinFromHex(args []Value, e *ast.CallExpr) (Value, error) {
	encoded, err := requireOneStringArg(args, "fromHex", e)
	if err != nil {
		return nil, err
	}

	decoded, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("fromHex expects a valid hex string: %s", err), e.Pos())
	}

	return string(decoded), nil
}
