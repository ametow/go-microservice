package courier

import "errors"

// courier error types
var ErrCourierNotFound = errors.New("courier not found")
var ErrOrderNotFound = errors.New("order not found")
var ErrCourierBadType = errors.New("invalid courier type")
var ErrCourierBadRegions = errors.New("invalid regions")
var ErrCourierBadWorkingHours = errors.New("invalid working hours")
var ErrZeroLengthCouriers = errors.New("zero length couriers")
