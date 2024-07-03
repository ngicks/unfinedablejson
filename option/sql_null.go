package option

import (
	"database/sql"
	"database/sql/driver"
)

var (
	_ sql.Scanner   = (*SqlNull[any])(nil)
	_ driver.Valuer = SqlNull[any]{}
)

type SqlNull[T any] struct {
	Option[T]
}

func (n *SqlNull[T]) Scan(src any) error {
	var null sql.Null[T]
	err := null.Scan(src)
	if err != nil {
		return err
	}
	n.Option = FromSqlNull(null)
	return nil
}

func (n SqlNull[T]) Value() (driver.Value, error) {
	return n.Option.SqlNull().Value()
}
