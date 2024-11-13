package models

type Payload struct {
	Tenant    string `json:"tenant"`
	Namespace string `json:"namespace"`
	FileName  string `json:"filename"`
}

type SigningResponse struct {
	URL                string `json:"url"`
	EncryptedAesKey    string `json:"encrypted-aes-key"`
	EncryptedAesKeyURL string `json:"encrypted-aes-key-url"`
	AesKey             string `json:"aes-key"`
}
