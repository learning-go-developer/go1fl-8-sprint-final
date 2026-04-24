package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

// connectDB open connect with data base, checks its availability via Ping
// and registered automatic connection closure upon closure of the test.
// If error to connect or check, test breake (t.Fatalf).
func connectDB(t *testing.T, driver, dataSource string) *sql.DB {
	db, err := sql.Open(driver, dataSource)

	require.NoError(t, err, "failed to open database")
	require.NoError(t, db.Ping(), "database is not reachable")

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
	require.NoError(t, err)
	require.NotEqual(t, 0, id)
	parcel.Number = id
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	var pGet Parcel
	pGet, err = store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, parcel, pGet)
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(parcel.Number)
	require.NoError(t, err)
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
	require.NoError(t, err)
	require.NotEqual(t, 0, id)
	parcel.Number = id
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	pGet, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, newAddress, pGet.Address)
	assert.Equal(t, parcel.Client, pGet.Client)
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
	require.NoError(t, err)
	parcel.Number = id
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := "sent" // лучше использовать реалистичный статус
	err = store.SetStatus(parcel.Number, newStatus)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	pGet, err := store.Get(parcel.Number)
	require.NoError(t, err)
	// Проверяем, что статус действительно изменился
	require.Equal(t, newStatus, pGet.Status)
	// Проверяем, что остальные важные данные не повредились при обновлении
	assert.Equal(t, parcel.Client, pGet.Client, "client mismatch")
	assert.Equal(t, parcel.Address, pGet.Address, "address mismatch")
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
		require.NoError(t, err)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels), "count of parcels mismatch")

	// check
	for _, pGet := range storedParcels {
		original, ok := parcelMap[pGet.Number]
		require.True(t, ok, "parcel %d not found in original map", pGet.Number)
		assert.Equal(t, original, pGet, "mismatch in parcel %d", pGet.Number)
	}
}
