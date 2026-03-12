package mediarails

// Usage contains metering information for a generation request.
type Usage struct {
	// Unit is the billing unit (e.g., "characters", "seconds", "images", "gpu_seconds").
	Unit string

	// Quantity is the amount of units consumed.
	Quantity float64
}

// Cost calculates the total cost given a unit price.
func (u *Usage) Cost(unitPrice float64) float64 {
	if u == nil {
		return 0
	}
	return u.Quantity * unitPrice
}
