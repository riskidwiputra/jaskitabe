package ulasan

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePublicListParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		query    string
		param    string
		fallback int
		min      int
		max      int
		want     int
		wantErr  bool
	}{
		{name: "default", query: "", param: "page", fallback: 1, min: 1, max: 100, want: 1},
		{name: "valid", query: "?rating=5", param: "rating", fallback: 0, min: 0, max: 5, want: 5},
		{name: "below minimum", query: "?page=0", param: "page", fallback: 1, min: 1, max: 100, wantErr: true},
		{name: "above maximum", query: "?limit=25", param: "limit", fallback: 9, min: 1, max: 24, wantErr: true},
		{name: "not an integer", query: "?page=abc", param: "page", fallback: 1, min: 1, max: 100, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest("GET", "/api/public/ulasan"+test.query, nil)

			got, err := parsePublicListParam(context, test.param, test.fallback, test.min, test.max)
			if (err != nil) != test.wantErr {
				t.Fatalf("parsePublicListParam() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Fatalf("parsePublicListParam() = %d, want %d", got, test.want)
			}
		})
	}
}
