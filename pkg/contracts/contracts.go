package contracts

type Resource struct {
	Url   string
	Depth int
}

// Key returns a stable deduplication key based solely on the URL,
// ignoring Depth so the same URL at different depths is treated as one message.
func (r Resource) Key() string {
	return r.Url
}
