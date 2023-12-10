package cache

import (
	"time"

	"github.com/dgraph-io/badger/v4"
)

func (b *BadgerCache) Has(str string) (bool, error) {
	_, err := b.Get(str)
	return err == nil, nil
}

func (b *BadgerCache) Get(str string) (interface{}, error) {
	var fromCache []byte

	err := b.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(str))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			fromCache = append([]byte{}, val...)
			return nil
		})
		return err
	})

	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(fromCache))
	if err != nil {
		return nil, err
	}

	item := decoded[str]
	return item, nil
}

func (b *BadgerCache) Set(str string, value interface{}, expires ...int) error {
	entry := Entry{str: value}
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	ttlDuration := time.Second
	if len(expires) > 0 {
		ttlDuration *= time.Duration(expires[0])
	}

	return b.Conn.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(str), encoded)
		if len(expires) > 0 {
			e = e.WithTTL(ttlDuration)
		}
		return txn.SetEntry(e)
	})
}

func (b *BadgerCache) Forget(str string) error {
	return b.Conn.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(str))
	})
}

func (b *BadgerCache) EmptyByMatch(str string) error {
	return b.emptyByMatch(str)
}

func (b *BadgerCache) Empty() error {
	return b.emptyByMatch("")
}

func (b *BadgerCache) emptyByMatch(str string) error {
	collectSize := 100000

	return b.Conn.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysCollected := 0

		for it.Seek([]byte(str)); it.ValidForPrefix([]byte(str)); it.Next() {
			key := it.Item().KeyCopy(nil)
			if err := txn.Delete(key); err != nil {
				return err
			}
			keysCollected++

			if keysCollected == collectSize {
				keysCollected = 0
				it.Seek(key) // Resume iteration from the last key
			}
		}

		return nil
	})
}
