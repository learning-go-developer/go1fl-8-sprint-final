package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"
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

// connectDB open connect with data base, checks its availability via Ping
// and registered automatic connection closure upon closure of the test.
// If error to connect or check, test breake (t.Fatalf).
func connectDB(t *testing.T, driver, dataSource string) *sql.DB {
	db, err := sql.Open(driver, dataSource)

	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("database is not reachable: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db := connectDB(t, "sqlite", "tracker.db")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("failed to add parcel: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id after store.Add")
	}
	parcel.Number = id

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	var pGet Parcel
	pGet, err = store.Get(parcel.Number)
	if err != nil {
		t.Fatalf("failed to get parcel: %v", err)
	}

	if pGet.Number != parcel.Number || pGet.Client != parcel.Client {
		t.Errorf("retrieved parcel mismatch: expected %+v, got %+v", parcel, pGet)
	}

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(parcel.Number)
	if err != nil {
		t.Fatalf("failed to delete parcel: %v", err)
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db := connectDB(t, "sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("failed to add parcel: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id after store.Add")
	}
	parcel.Number = id

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	if err != nil {
		t.Fatalf("failed to set address: %v", err)
	}

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	pGet, err := store.Get(parcel.Number)
	if err != nil {
		t.Fatalf("failed to get parcel after update: %v", err)
	}

	if pGet.Address != newAddress {
		t.Errorf("address mismatch: expected %s, got %s", newAddress, pGet.Address)
	}

	if pGet.Client != parcel.Client {
		t.Errorf("client changed unexpectedly: expected %d, got %d", parcel.Client, pGet.Client)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db := connectDB(t, "sqlite", "tracker.db")
	store := NewParcelStore(db) // создаем экземпляр стора
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("failed to add parcel: %v", err)
	}
	parcel.Number = id

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := "sent" // лучше использовать реалистичный статус
	err = store.SetStatus(parcel.Number, newStatus)
	if err != nil {
		t.Fatalf("failed to set status: %v", err)
	}

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	pGet, err := store.Get(parcel.Number)
	if err != nil {
		t.Fatalf("failed to get parcel: %v", err)
	}

	// Проверяем, что статус действительно изменился
	if pGet.Status != newStatus {
		t.Errorf("status mismatch: expected %s, got %s", newStatus, pGet.Status)
	}

	// Проверяем, что остальные важные данные не повредились при обновлении
	if pGet.Client != parcel.Client || pGet.Address != parcel.Address {
		t.Errorf("other fields mismatch after status update: expected %+v, got %+v", parcel, pGet)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db := connectDB(t, "sqlite", "tracker.db")
	store := NewParcelStore(db)

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
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatalf("failed to add parcel: %v", err)
		}

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Fatalf("failed to get parcels by client: %v", err)
	}

	if len(storedParcels) != len(parcels) {
		t.Fatalf("expected %d parcels, got %d", len(parcels), len(storedParcels))
	}

	// check
	for _, pGet := range storedParcels {
		original, ok := parcelMap[pGet.Number]
		if !ok {
			t.Fatalf("retrieved parcel %d not found in original map", pGet.Number)
		}

		if pGet.Client != original.Client {
			t.Errorf("parcel %d: client mismatch. Expected %d, got %d", pGet.Number, original.Client, pGet.Client)
		}

		if pGet.Address != original.Address {
			t.Errorf("parcel %d: address mismatch. Expected %s, got %s", pGet.Number, original.Address, pGet.Address)
		}

		if pGet.Status != original.Status {
			t.Errorf("parcel %d: status mismatch. Expected %s, got %s", pGet.Number, original.Status, pGet.Status)
		}

		if pGet.CreatedAt != original.CreatedAt {
			t.Errorf("parcel %d: created_at mismatch. Expected %v, got %v", pGet.Number, original.CreatedAt, pGet.CreatedAt)
		}
	}
}
