package normalize

func validBrazilAreaCode(areaCode string) bool {
	if len(areaCode) != 2 {
		return false
	}

	first := areaCode[0]
	second := areaCode[1]
	switch first {
	case '1':
		return second >= '1' && second <= '9'
	case '2':
		return second == '1' || second == '2' || second == '4' || second == '7' || second == '8'
	case '3':
		return (second >= '1' && second <= '5') || second == '7' || second == '8'
	case '4':
		return second >= '1' && second <= '9'
	case '5':
		return second == '1' || (second >= '3' && second <= '5')
	case '6':
		return second >= '1' && second <= '9'
	case '7':
		return second == '1' || (second >= '3' && second <= '5') || second == '7' || second == '9'
	case '8', '9':
		return second >= '1' && second <= '9'
	default:
		return false
	}
}
