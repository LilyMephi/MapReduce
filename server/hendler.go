package main

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	serverpb "mapreduce/proto"
)

func transform(op serverpb.Operation, payload string) (string, error) {
	switch op {
	case serverpb.Operation_OP_UPPER:
		return strings.ToUpper(payload), nil
	case serverpb.Operation_OP_LOWER:
		return strings.ToLower(payload), nil
	case serverpb.Operation_OP_REVERSE:
		r := []rune(payload)
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return string(r), nil
	case serverpb.Operation_OP_BASE64:
		return base64.StdEncoding.EncodeToString([]byte(payload)), nil
	case serverpb.Operation_OP_SHA256:
		sum := sha256.Sum256([]byte(payload))
		return base64.StdEncoding.EncodeToString(sum[:]), nil
	case serverpb.Operation_OP_COUNT_WORDS:
		return itoa(int64(len(strings.Fields(payload)))), nil
	default:
		return "", errors.New("unsupported operation")
	}
}

func itoa(n int64) string {
	// простой int64 → string без strconv для наглядности, можно strconv.FormatInt
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func runTask(req *serverpb.TaskRequest) *serverpb.TaskResponse {
	start := time.Now()
	resp := &serverpb.TaskResponse{RequestId: req.GetRequestId()}

	if req.GetPayload() == "" {
		resp.Ok = false
		resp.Error = "empty payload"
		resp.ProcessingNs = time.Since(start).Nanoseconds()
		return resp
	}

	out, err := transform(req.GetOperation(), req.GetPayload())
	resp.ProcessingNs = time.Since(start).Nanoseconds()
	if err != nil {
		resp.Ok = false
		resp.Error = err.Error()
		return resp
	}
	resp.Ok = true
	resp.Result = out
	return resp
}