package anthropic

import "testing"

func TestDefaultAdapterFullURL(t *testing.T) {
	adapter := &DefaultAdapter{}

	tests := []struct {
		name    string
		baseURL string
		suffix  string
		want    string
	}{
		{
			name:    "no trailing slash",
			baseURL: "https://api.anthropic.com/v1",
			suffix:  "/messages",
			want:    "https://api.anthropic.com/v1/messages",
		},
		{
			name:    "trailing slash",
			baseURL: "https://api.anthropic.com/v1/",
			suffix:  "/messages",
			want:    "https://api.anthropic.com/v1/messages",
		},
		{
			name:    "multiple trailing slashes",
			baseURL: "https://api.anthropic.com/v1///",
			suffix:  "/messages",
			want:    "https://api.anthropic.com/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.fullURL(tt.baseURL, tt.suffix)
			if got != tt.want {
				t.Fatalf("fullURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
