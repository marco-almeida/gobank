package storage

type Storage interface {
	Get(int) error
}