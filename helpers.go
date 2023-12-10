package medego

import (
	"crypto/rand"
	"os"
)

const (
	randomString = "abcdefghijklmnopqrstuvwxyzABCDEFGQRSTUVWXYZ0123456789_+"
)

func (c *Medego) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomString)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

func (c *Medego) CreateDirIfNotExists(dir string) error {
	const mode = 0755
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Medego) CreateFileIfNotExists(path string) error {
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}
	return nil
}
