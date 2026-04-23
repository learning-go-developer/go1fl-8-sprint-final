package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
		return
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("database is not reachable: %v", err)
	}

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)", parcel.Client, parcel.Status, parcel.Address, parcel.CreatedAt)
	if err != nil {
		t.Fatalf("failed to add parcel: %v", err)
	}
	// 1. Get ID generate DB
	id, err := res.LastInsertId()
	if err != nil {
		// This error get data from driver
		t.Fatalf("failed to get last insert id: %v", err)
	}
	// 2. Check verify ID
	if id <= 0 {
		t.Fatal("error: database did not assign a valid ID")
	}
	// 3. Write get ID in structure
	parcel.Number = int(id)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	var pGet Parcel

	err = db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?", parcel.Number).Scan(
		&pGet.Number,
		&pGet.Client,
		&pGet.Status,
		&pGet.Address,
		&pGet.CreatedAt,
	)

	if err != nil {
		t.Fatalf("failed to get parcel: %v", err)
	}

	if pGet.Number != parcel.Number ||	
	pGet.Client != parcel.Client || 
	pGet.Status != parcel.Status || 
	pGet.Address != parcel.Address || 
	pGet.CreatedAt != parcel.CreatedAt {
		t.Fatal("error: retrieved parcel data does not match original")
	}
	
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	res, err := db.Exec("DELETE FROM parcel WHERE number = ?", parcel.Number)
    if err != nil {
        t.Fatalf("failed to delete parcel: %v", err)
    }

	rowsAffected, err := res.RowsAffected()
    if err != nil {
        t.Fatalf("could not get affected rows: %v", err)
    }

	if rowsAffected != 1 {
        t.Fatalf("no parcel found with number %d", parcel.Number)
    }

	var pDelete Parcel
	
	err = db.QueryRow("SELECT number FROM parcel WHERE number = ?", parcel.Number).Scan(&pDelete.Number)

	if err != sql.ErrNoRows {
		if err == nil {
			t.Fatal("error: parcel still exists in database after delete")
		} else {
			t.Fatalf("unexpected error while checking deletion: %v", err)
		}
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := // настройте подключение к БД

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := // настройте подключение к БД

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	// set status
	// обновите статус, убедитесь в отсутствии ошибки

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := // настройте подключение к БД

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
