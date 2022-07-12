package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"

	"github.com/jackc/pgerrcode"

	"shortener/internal/handlers"
	"shortener/internal/storage"
	"shortener/internal/utils"
)

type Storage struct {
	DB *sql.DB
}

func (st *Storage) init() error {
	if _, err := st.DB.Exec(`
		create table if not exists uris (
			id serial,
			short varchar(7) not null,
			uri varchar(128) not null
		);
		create unique index if not exists uri on uris(short);
	`); err != nil {
		return err
	}
	if _, err := st.DB.Exec(`
		create table if not exists users (
			id serial,
			hash varchar(128) 
		)
	`); err != nil {
		return err
	}
	if _, err := st.DB.Exec(`
		create table if not exists uxu (
			id serial,
			uriid int,
			userid int
		)
	`); err != nil {
		return err
	}
	return nil
}

func New(db *sql.DB) (*Storage, error) {
	st := &Storage{
		DB: db,
	}
	if err := st.init(); err != nil {
		return nil, err
	}
	return st, nil
}

func (st *Storage) Ping(ctx context.Context) error {
	return st.DB.PingContext(ctx)
}

func (st *Storage) Get(short string) (string, bool) {
	var uri string
	err := st.DB.QueryRow(`
		select uri
		from uris t1
		where t1.short = $1
	`, short).Scan(&uri)
	if err != nil {
		log.Println(err)
		return uri, false
	}
	return uri, true
}

func (st *Storage) Users(base, hash string) ([]storage.Users, error) {
	u := make([]storage.Users, 0)
	rows, err := st.DB.Query(`
		select uri, short
		from uris t1
		inner join uxu t2
		on t1.id = t2.uriid
		inner join users t3
		on t2.userid = t3.id
		where hash = $1
	`, hash)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var uri, short string
		err := rows.Scan(&uri, &short)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		u = append(u, storage.Users{
			URI:   uri,
			Short: fmt.Sprintf("%s/%s", base, short),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return u, nil
}

func (st *Storage) Push(uri, hash string) (string, error) {
	short, err := utils.Shorty(st, uri) // shorty(Storage)
	if err != nil {
		return short, err
	}

	var userid sql.NullInt64
	if err := st.DB.QueryRow(`
		select id
		from users
		where hash = $1
	`, hash).Scan(&userid); err != nil && err != sql.ErrNoRows {
		log.Fatalln("select user id", err)
		return "", err
	}
	if !userid.Valid {
		if _, err := st.DB.Exec(`
			insert into users (hash) 
			values ($1);
		`, hash); err != nil {
			log.Println("insert user", err)
			return "", err
		}
		if err := st.DB.QueryRow(`
			select id
			from users
			where hash = $1
		`, hash).Scan(&userid); err != nil {
			log.Println("select user id", err)
			return "", err
		}
	}

	var uriid sql.NullInt64
	if _, err := st.DB.Exec(`
		insert into uris (short, uri) 
		values ($1, $2);
	`, short, uri); err != nil && err.(*pq.Error).Code == pgerrcode.UniqueViolation {
		log.Println("insert uri", err)
		return short, err
	} else if err != nil {
		log.Println("insert uri", err)
		return "", err
	}
	if err := st.DB.QueryRow(`
		select id
		from uris
		where uri = $1
	`, uri).Scan(&uriid); err != nil {
		log.Println("select uri id", err)
		return "", err
	}

	if err := st.DB.QueryRow(`
		select id
		from uxu
		where uriid = $1
		and userid = $2
	`, uriid, userid).Scan(&uriid); err == sql.ErrNoRows {
		if _, err := st.DB.Exec(`
			insert into uxu (uriid, userid)
			values($1, $2)
		`, uriid, userid); err != nil {
			log.Println("insert link", err)
			return "", err
		}
	}

	return short, nil
}

func (st *Storage) Batch(mm []*handlers.Batch) error {
	tx, err := st.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Fatalf("warn:> unexpected error on rollback tx: %s\n", err.Error())
		}
	}()

	stmt, err := tx.Prepare(`
		insert into uris (short, uri) 
		values ($1, $2);
	`)
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Fatalf("warn:> unexpected error on closing stmt: %s\n", err.Error())
		}
	}()

	for _, m := range mm {
		short, err := utils.Shorty(st, m.URI)
		if err != nil {
			return err
		}
		m.Short = short
		if _, err := stmt.Exec(m.Short, m.URI); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (st *Storage) Close() error {
	return st.DB.Close()
}
