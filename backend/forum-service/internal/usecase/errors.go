package usecase

import "errors"

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrForbidden       = errors.New("permission denied")
) 