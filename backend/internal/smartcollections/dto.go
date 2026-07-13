package smartcollections

type CollectionResponse struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Count       int64  `json:"count"`
}
