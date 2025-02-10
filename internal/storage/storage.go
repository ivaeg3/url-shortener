package storage

type Storage interface {
	Save(url string) (string, error)
	Get(shortURL string) (string, error)
}
