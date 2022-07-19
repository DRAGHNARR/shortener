package storage

type Users struct {
	URI   string `json:"original_url"`
	Short string `json:"short_url"`
	ID    string `json:"ID,omitempty"`
}

type URIsItem struct {
	URI     string
	Deleted bool
}
