package ulasan

import "testing"

func TestR2StorageObjectKey(t *testing.T) {
	storage := &R2Storage{publicURL: "https://images.example.com"}
	tests := []struct {
		name     string
		photoURL string
		wantKey  string
		wantOK   bool
	}{
		{name: "public photo URL", photoURL: "https://images.example.com/ulasan/abc.jpg", wantKey: "ulasan/abc.jpg", wantOK: true},
		{name: "another host", photoURL: "https://other.example.com/ulasan/abc.jpg"},
		{name: "path traversal", photoURL: "https://images.example.com/ulasan/../secret"},
		{name: "empty key", photoURL: "https://images.example.com/"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotKey, gotOK := storage.ObjectKey(test.photoURL)
			if gotKey != test.wantKey || gotOK != test.wantOK {
				t.Fatalf("ObjectKey(%q) = (%q, %v), want (%q, %v)", test.photoURL, gotKey, gotOK, test.wantKey, test.wantOK)
			}
		})
	}
}
