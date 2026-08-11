package logic

// IsKnownSensorKind restricts the sensor "kind" field to a fixed vocabulary.
func IsKnownSensorKind(kind string) bool {
	switch kind {
	case "temperature", "humidity", "pressure", "motion":
		return true
	default:
		return false
	}
}

// IsPercentage validates that a signal-quality score falls within 0-100.
func IsPercentage(n int32) bool {
	return n >= 0 && n <= 100
}
