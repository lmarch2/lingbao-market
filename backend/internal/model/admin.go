package model

type FeedbackRequest struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

type ResolveFeedbackRequest struct {
	Action string `json:"action"`
}

type FeedbackMessage struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Reason       string `json:"reason"`
	Reporter     string `json:"reporter"`
	CreatedAt    int64  `json:"createdAt"`
	Resolved     bool   `json:"resolved"`
	ResolvedAt   int64  `json:"resolvedAt,omitempty"`
	ResolvedBy   string `json:"resolvedBy,omitempty"`
	Action       string `json:"action,omitempty"`
	RemovedTime  int64  `json:"removedTime,omitempty"`
	RemovedPrice int64  `json:"removedPrice,omitempty"`
}

type AdminLogEntry struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Actor     string            `json:"actor"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}
