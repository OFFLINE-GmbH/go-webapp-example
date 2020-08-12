package fs

import "os"

func EnsureDir(path string) error {
	isExisting, err := Exists(path)
	if err != nil {
		return err
	}
	if !isExisting {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
