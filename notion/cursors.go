package notion

import "errors"

// sad code, no generic, me copy-paste sad

type PageCursor struct {
	GetNext func()(PageCursor, error)
	Current []Page

	Remaining bool
}

type BlockCursor struct {
	GetNext func()(BlockCursor, error)
	Current []Block

	Remaining bool
}

type DatabaseCursor struct {
	GetNext func()(DatabaseCursor, error)
	Current []Database

	Remaining bool
}

func (cursor *BlockCursor) Next() error {
	if !cursor.Remaining {
		cursor.Current = []Block{ }
		return errors.New("end of list")
	}

	next, err := cursor.GetNext()
	if err != nil { return err }

	*cursor = next

	return nil
}

func (cursor *BlockCursor) ReadAll() []Block {
	all := cursor.Current

	for cursor.Remaining {
		err := cursor.Next()
		if err != nil { return all }

		all = append(all, cursor.Current...)
	}

	return all
}

func (cursor *PageCursor) Next() error {
	if !cursor.Remaining {
		cursor.Current = []Page{ }
		return errors.New("end of list")
	}

	next, err := cursor.GetNext()
	if err != nil { return err }

	*cursor = next

	return nil
}

func (cursor *PageCursor) ReadAll() []Page {
	all := cursor.Current

	for cursor.Remaining {
		err := cursor.Next()
		if err != nil { return all }

		all = append(all, cursor.Current...)
	}

	return all
}

func (cursor *DatabaseCursor) Next() error {
	if !cursor.Remaining {
		cursor.Current = []Database{ }
		return errors.New("end of list")
	}

	next, err := cursor.GetNext()
	if err != nil { return err }

	*cursor = next

	return nil
}

func (cursor *DatabaseCursor) ReadAll() []Database {
	all := cursor.Current

	for cursor.Remaining {
		err := cursor.Next()
		if err != nil { return all }

		all = append(all, cursor.Current...)
	}

	return all
}
