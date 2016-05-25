package guard

const (
	maxHoleOffset = 64
)

func maxInt(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
