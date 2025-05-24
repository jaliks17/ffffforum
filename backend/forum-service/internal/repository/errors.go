package repository

import "errors"

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrPermissionDenied = errors.New("permission denied")
) 