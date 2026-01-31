package medialibrary

type TAggregateType string

type MediaType struct {
	MimeTypes  []string
	Extensions []string
}

const (
	TypeImage        = "image"
	TypeImageVector  = "image_vector"
	TypePDF          = "pdf"
	TypeAudio        = "audio"
	TypeVideo        = "video"
	TypeArchive      = "archive"
	TypeDocument     = "document"
	TypeSpreadsheet  = "spreadsheet"
	TypePresentation = "presentation"
)

var mediaTypes = map[string]MediaType{
	TypeImage: {
		MimeTypes:  []string{"image/jpeg", "image/png", "image/gif"},
		Extensions: []string{"jpg", "jpeg", "png", "gif"},
	},
	TypeImageVector: {
		MimeTypes:  []string{"image/svg+xml"},
		Extensions: []string{"svg"},
	},
	TypePDF: {
		MimeTypes:  []string{"application/pdf"},
		Extensions: []string{"pdf"},
	},
	TypeAudio: {
		MimeTypes:  []string{"audio/aac", "audio/ogg", "audio/mpeg", "audio/mp3", "audio/wav"},
		Extensions: []string{"aac", "ogg", "oga", "mp3", "wav"},
	},
	TypeVideo: {
		MimeTypes:  []string{"video/mp4", "video/mpeg", "video/ogg", "video/webm"},
		Extensions: []string{"mp4", "m4v", "mov", "ogv", "webm"},
	},
	TypeArchive: {
		MimeTypes:  []string{"application/zip", "application/x-compressed-zip", "multipart/x-zip"},
		Extensions: []string{"zip"},
	},
	TypeDocument: {
		MimeTypes:  []string{"text/plain", "application/plain", "text/xml", "text/json", "application/json", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		Extensions: []string{"doc", "docx", "txt", "text", "xml", "json"},
	},
	TypeSpreadsheet: {
		MimeTypes:  []string{"application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		Extensions: []string{"xls", "xlsx"},
	},
	TypePresentation: {
		MimeTypes:  []string{"application/vnd.ms-powerpoint", "application/vnd.openxmlformats-officedocument.presentationml.presentation", "application/vnd.openxmlformats-officedocument.presentationml.slideshow"},
		Extensions: []string{"ppt", "pptx", "ppsx"},
	},
}

type AggregateType struct {
	mediaTypes map[string]MediaType
}

// func (a AggregateType) GetType(mimeType string) string {

// }
