package message

type MessageBody struct {
    Method  string          `json:"method"`
    Args    []interface{}   `json:"args"`
}