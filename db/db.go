package db

type DB interface {
	Number() (int, error)
	Incr() (int, error)
	SetSettings(int, int) error
}
