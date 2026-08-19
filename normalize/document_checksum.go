package normalize

func validPIS(digits string) bool {
	if allBytesEqual(digits) {
		return false
	}
	weights := [10]int{3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i, weight := range weights {
		sum += int(digits[i]-'0') * weight
	}
	remainder := sum % 11
	dv := 0
	if remainder >= 2 {
		dv = 11 - remainder
	}
	return dv == int(digits[10]-'0')
}

func validCNH(digits string) bool {
	if allBytesEqual(digits) {
		return false
	}

	dsc := 0
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (9 - i)
	}
	dv1 := sum % 11
	if dv1 >= 10 {
		dv1 = 0
		dsc = 2
	}
	if dv1 != int(digits[9]-'0') {
		return false
	}

	sum = 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (i + 1)
	}
	dv2 := sum%11 - dsc
	if dv2 < 0 {
		dv2 += 11
	}
	if dv2 >= 10 {
		dv2 = 0
	}
	return dv2 == int(digits[10]-'0')
}

func validVoterID(digits string) bool {
	uf := digits[8:10]
	if uf < "01" || uf > "28" {
		return false
	}
	special := uf == "01" || uf == "02"

	sum := 0
	for i := 0; i < 8; i++ {
		sum += int(digits[i]-'0') * (i + 2)
	}
	dv1 := voterRemainder(sum, special)
	if dv1 != int(digits[10]-'0') {
		return false
	}

	sum = int(digits[8]-'0')*7 + int(digits[9]-'0')*8 + dv1*9
	return voterRemainder(sum, special) == int(digits[11]-'0')
}

func voterRemainder(sum int, specialUF bool) int {
	remainder := sum % 11
	if remainder == 10 {
		return 0
	}
	if remainder == 0 && specialUF {
		return 1
	}
	return remainder
}

func validCNS(digits string) bool {
	switch digits[0] {
	case '1', '2':
		return generateDefinitiveCNS(digits[:11]) == digits
	case '7', '8', '9':
		return weightedDigitSum(digits)%11 == 0
	default:
		return false
	}
}

func generateDefinitiveCNS(first11 string) string {
	total := weightedPrefixSum(first11, 11)
	dv := 11 - total%11
	if dv == 11 {
		dv = 0
	}
	if dv == 10 {
		total += 2
		dv = 11 - total%11
		if dv == 11 {
			dv = 0
		}
		return first11 + "001" + string(byte('0'+dv))
	}
	return first11 + "000" + string(byte('0'+dv))
}

func weightedDigitSum(digits string) int {
	return weightedPrefixSum(digits, len(digits))
}

func weightedPrefixSum(digits string, length int) int {
	sum := 0
	for i := 0; i < length; i++ {
		sum += int(digits[i]-'0') * (15 - i)
	}
	return sum
}
