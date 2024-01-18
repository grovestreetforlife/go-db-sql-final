package main

import (
	"database/sql"
	"log"
	"math/rand"
	"os"
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

var db *sql.DB

func TestMain(m *testing.M) {
	var err error
	db, err = sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Fatal("Can`t open db", err)
	}
	defer db.Close()
	os.Exit(m.Run())
}

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
	var err error
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, parcel.Number)

	// get
	p, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, parcel, p)

	// delete
	err = store.Delete(parcel.Number)
	require.NoError(t, err)

	_, err = store.Get(parcel.Number)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// set status
	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err)

	// check
	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusDelivered, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	store := NewParcelStore(db)
	// prepare
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
		assert.NotEmpty(t, id)

		// обновляем идентификатор у добавленной посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Equal(t, len(storedParcels), len(parcelMap))
	// check
	for _, parcel := range storedParcels {
		expectedParcel, ok := parcelMap[parcel.Number]
		if !ok {
			t.Fatal("Parcel not found in map")
		}
		require.NotEmpty(t, expectedParcel)
		require.Equal(t, expectedParcel, parcel)
	}
}
