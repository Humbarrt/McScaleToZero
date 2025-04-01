package queries

type scannable interface {
	Scan(dest ...any) error
}
