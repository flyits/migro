package schema

// IndexType represents the type of database index
type IndexType int

const (
	IndexTypeIndex IndexType = iota
	IndexTypeUnique
	IndexTypePrimary
	IndexTypeFulltext
)

// Index represents a database index definition
type Index struct {
	Name    string
	Type    IndexType
	Columns []string
}

// NewIndex creates a new index with the given columns
func NewIndex(columns ...string) *Index {
	return &Index{
		Type:    IndexTypeIndex,
		Columns: columns,
	}
}

// Named sets the name of the index
func (i *Index) Named(name string) *Index {
	i.Name = name
	return i
}

// Unique makes this a unique index
func (i *Index) Unique() *Index {
	i.Type = IndexTypeUnique
	return i
}

// Primary makes this a primary key index
func (i *Index) Primary() *Index {
	i.Type = IndexTypePrimary
	return i
}

// Fulltext makes this a fulltext index (MySQL only)
func (i *Index) Fulltext() *Index {
	i.Type = IndexTypeFulltext
	return i
}
