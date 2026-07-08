package model

type UploadResponse struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
}
