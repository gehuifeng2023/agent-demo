package model

type Source struct {
	File     string `json:"file"`
	ChunkID  string `json:"chunk_id"`
	Content  string `json:"content"`
	Position int    `json:"position"`
}
