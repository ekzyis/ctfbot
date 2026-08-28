package client

import (
	"errors"
	"fmt"
	"strings"
)

// minimal bech32 (BIP-173) decoder, enough to decode NIP-19 nsec keys.

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var bech32Generator = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

func bech32Polymod(values []int) int {
	chk := 1
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				chk ^= bech32Generator[i]
			}
		}
	}
	return chk
}

func bech32HrpExpand(hrp string) []int {
	out := make([]int, 0, len(hrp)*2+1)
	for _, c := range hrp {
		out = append(out, int(c)>>5)
	}
	out = append(out, 0)
	for _, c := range hrp {
		out = append(out, int(c)&31)
	}
	return out
}

// bech32Decode returns the human-readable part and the 5-bit data (checksum
// stripped) of a bech32 string, verifying the checksum.
func bech32Decode(s string) (string, []int, error) {
	if s != strings.ToLower(s) && s != strings.ToUpper(s) {
		return "", nil, errors.New("mixed-case string")
	}
	s = strings.ToLower(s)

	pos := strings.LastIndex(s, "1")
	if pos < 1 || pos+7 > len(s) {
		return "", nil, errors.New("invalid separator position")
	}

	hrp := s[:pos]
	data := make([]int, 0, len(s)-pos-1)
	for _, c := range s[pos+1:] {
		idx := strings.IndexRune(bech32Charset, c)
		if idx < 0 {
			return "", nil, fmt.Errorf("invalid character %q", c)
		}
		data = append(data, idx)
	}

	if bech32Polymod(append(bech32HrpExpand(hrp), data...)) != 1 {
		return "", nil, errors.New("invalid checksum")
	}

	return hrp, data[:len(data)-6], nil
}

// bech32ConvertBits regroups bit-values from one base to another (e.g. 5-bit
// bech32 groups to 8-bit bytes).
func bech32ConvertBits(data []int, from, to uint, pad bool) ([]byte, error) {
	acc := 0
	bits := uint(0)
	maxv := (1 << to) - 1
	out := make([]byte, 0, len(data)*int(from)/int(to)+1)
	for _, value := range data {
		if value < 0 || value>>from != 0 {
			return nil, errors.New("value out of range")
		}
		acc = (acc << from) | value
		bits += from
		for bits >= to {
			bits -= to
			out = append(out, byte((acc>>bits)&maxv))
		}
	}
	if pad {
		if bits > 0 {
			out = append(out, byte((acc<<(to-bits))&maxv))
		}
	} else if bits >= from || ((acc<<(to-bits))&maxv) != 0 {
		return nil, errors.New("invalid padding")
	}
	return out, nil
}
