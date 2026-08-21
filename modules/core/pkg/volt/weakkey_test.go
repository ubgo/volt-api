package volt

import "testing"

func TestIsWeakOrDefaultKey(t *testing.T) {
	cases := []struct {
		name string
		key  string
		weak bool
	}{
		{"empty", "", true},
		{"blank", "   ", true},
		{"default pkl placeholder", "change-me-encryption-key-32ch", true},
		{"sample pkl placeholder", "xxxx", true},
		{"real 32-byte key", "Zk9xQ2pLm7vR4tWnB3yC8sD1fG6hJ0aE", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsWeakOrDefaultKey(c.key); got != c.weak {
				t.Errorf("IsWeakOrDefaultKey(%q) = %v, want %v", c.key, got, c.weak)
			}
		})
	}
}
