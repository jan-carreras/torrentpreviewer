package preview

import "errors"

func totalLength(files []FileInfo) (length int) {
	for _, f := range files {
		length += f.Length()
	}
	return length
}

func filesByID(files []FileInfo) map[int]*FileInfo {
	hash := make(map[int]*FileInfo)
	for i := 0; i < len(files); i++ {
		hash[files[i].ID()] = &files[i]
	}
	return hash
}

func validateFiles(files []FileInfo) error {
	for i := 0; i < len(files); i++ {
		if files[i].ID() != i {
			return errors.New("non correlative fileID. all ids should be sequential")
		}
	}
	return nil
}
