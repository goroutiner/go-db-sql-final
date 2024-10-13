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
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.Error(t, err)
	} 
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	rowParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, rowParcel.Number, id)
	require.Equal(t, rowParcel.Client, parcel.Client)
	require.Equal(t, rowParcel.Status, parcel.Status)
	require.Equal(t, rowParcel.Address, parcel.Address)
	require.Equal(t, rowParcel.CreatedAt, parcel.CreatedAt)
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.NotEmpty(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.Error(t, err)
	} 
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	service := NewParcelService(store)
	
	err = service.ChangeAddress(id, newAddress)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	rowParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, rowParcel.Address, newAddress)

	// delete
	// удаляем тестовую посылку
	err = store.Delete(id)
	require.NoError(t, err)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.Error(t, err)
	} 
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	service := NewParcelService(store)
	
	err = service.NextStatus(id)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	rowParcel, err := store.Get(id)
	require.NoError(t, err)
	require.NotEqual(t, parcel.Status, rowParcel.Status)

	// delete
	// удаляем тестовую посылку
	err = store.Delete(id)
	require.NoError(t, err)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.Error(t, err)
	} 
	defer db.Close()

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
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id) 

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))
	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		if parcelInMap, ok := parcelMap[parcel.Number]; !ok {
			require.Fail(t, "Посылка не найдена! Посылки с идентификатором = %d нет в parcelMap", parcel.Number)
		} else {
			require.Equal(t, parcel.Number, parcelInMap.Number)
			require.Equal(t, parcel.Client, parcelInMap.Client)
			require.Equal(t, parcel.Address, parcelInMap.Address)
			require.Equal(t, parcel.CreatedAt, parcelInMap.CreatedAt)
			require.Equal(t, parcel.Status, parcelInMap.Status)
		}
	}

	// delete
	// удаляем тестовые посылки 
	_, err = store.db.Exec("DELETE FROM parcel")
	require.NoError(t, err)
}
