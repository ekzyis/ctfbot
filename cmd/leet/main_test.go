package main

import "testing"

func TestLeet(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"test", "7357"},
		{"secret", "53cr37"},
		{"you found the secret", "y0u f0und 7h3 53cr37"},
		{"Capture The Flag", "C4p7ur3 7h3 F149"}, // T->7 (case-insensitive); unmapped C/F keep case
		{"gg", "99"},
		{"xyz", "xyz"},         // no substitutable letters
		{"a-b_c.d!", "4-8_c.d!"}, // punctuation passes through
		{"", ""},
	}
	for _, tc := range cases {
		if got := leet(tc.in); got != tc.want {
			t.Errorf("leet(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
