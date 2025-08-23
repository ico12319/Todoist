package source

type Source interface {
	GetSource() string
	SetSource(source string)
}

type SqlSource struct {
	tableName string
}

func (s *SqlSource) GetSource() string {
	return s.tableName
}

func (s *SqlSource) SetSource(source string) {
	s.tableName = source
}
