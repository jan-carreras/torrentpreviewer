package downloadPartials

type File struct {
	FileID int
	Start  int
	Length int
}

type CMD struct {
	ID    string
	Files []File
}
