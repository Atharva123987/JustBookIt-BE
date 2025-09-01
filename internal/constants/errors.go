package constants

import "errors"

var (
	ErrSeatNotFound   = errors.New("seat not found")
	ErrSeatReserved   = errors.New("seat already reserved")
	ErrInvalidSeatNum = errors.New("invalid seat number")
)

var DomainErrorsMapped = map[error]string{
	ErrSeatNotFound:   "Seat not found",
	ErrSeatReserved:   "Seat already reserved",
	ErrInvalidSeatNum: "Invalid seat number",
}
