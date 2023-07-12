package order

import "errors"

var ErrOrderCost = errors.New("invalid order cost")
var ErrOrderWeight = errors.New("invalid order weight")
var ErrOrderRegions = errors.New("invalid order regions")
var ErrZeroOrders = errors.New("zero orders")
var ErrCourierNotFound = errors.New("courier not found")
var ErrOrderNotFound = errors.New("order not found")
var ErrOrderNotAssigned = errors.New("order not found")
var ErrInvalidCompleteTime = errors.New("order complete time invalid")
var ErrOrderAlreadyDelivered = errors.New("order has already been delivered")
