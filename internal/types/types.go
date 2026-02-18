package types

import "time"

type Asset struct {
	UUID             string    `json:"uuid"`
	OriginalFilename string    `json:"original_filename"`
	OriginalFilesize int64     `json:"original_filesize"`
	Date             time.Time `json:"date"`
	IsMovie          bool      `json:"ismovie"`
	LocalPath        string    `json:"local_path,omitempty"`
}

type ImmichAsset struct {
	OriginalFileName string `json:"originalFileName"`
	FileSizeInByte   int64  `json:"fileSizeInByte"`
	ExifInfo         struct {
		FileSizeInByte int64 `json:"fileSizeInByte"`
	} `json:"exifInfo"`
	DateTimeOriginal string `json:"dateTimeOriginal"`
	Checksum         string `json:"checksum"`
}
