package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(`INSERT INTO parcel (client, status, address, created_at) 
                           VALUES (?, ?, ?, ?)`,
		p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert parcel: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}

	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?", number)

	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return Parcel{}, fmt.Errorf("parcel with number %d not found: %w", number, err)
		}
		return Parcel{}, fmt.Errorf("failed to get parcel: %w", err)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, fmt.Errorf("failed to query parcels by client: %w", err)
	}
	defer rows.Close()

	var res []Parcel

	for rows.Next() {
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parcel row: %w", err)
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	res, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return fmt.Errorf("failed to update status for parcel %d: %w", number, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no parcel found with number %d", number)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	res, err := s.db.Exec("UPDATE parcel SET address = ? WHERE number = ? AND status = ?",
		address, number, ParcelStatusRegistered)

	if err != nil {
		return fmt.Errorf("failed to update address for parcel %d: %w", number, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("could not update address: parcel %d not found or status is not 'registered'", number)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec("DELETE FROM parcel WHERE number = ? AND status = ?", number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("failed to delete parcel %d: %w", number, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("could not delete: parcel %d not found or it is already in transit", number)
	}

	return nil
}
