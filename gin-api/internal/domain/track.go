package domain

type Track struct {
	ID       uint   `gorm:"primaryKey"`
	Title    string `json:"title"`
	BPM      int    `json:"bpm"`
	Key      string `json:"key"`
	FilePath string `json:"file_path"`
}
