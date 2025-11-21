package bom

import "testing"

func TestFormat(t *testing.T) {
    tests := []struct {
        in   Component
        want string
    }{
        {Component{"libA", "1.0.0"}, "libA@1.0.0"},
        {Component{"libA", ""}, "libA"},
        {Component{"", "1.2.3"}, "@1.2.3"},
        {Component{"", ""}, ""},
    }

    for i, tt := range tests {
        if got := Format(tt.in); got != tt.want {
            t.Fatalf("case %d: got %q, want %q", i, got, tt.want)
        }
    }
}
