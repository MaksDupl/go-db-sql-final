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
	// Выполняем вставку новой записи в таблицу parcel
	res, err := s.db.Exec(`
		INSERT INTO parcel (client, status, address, created_at)
		VALUES (?, ?, ?, ?)`,
		p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления записи: %w", err)
	}

	// Получаем идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения идентификатора последней записи: %w", err)
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// создаем объект Parcel
	p := Parcel{}

	// выполняем запрос для получения строки по номеру
	rows, err := s.db.Query(`
		SELECT client, status, address, created_at
		FROM parcel
		WHERE number = ?`, number)
	if err != nil {
		return Parcel{}, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	// Считываем данные из первой строки
	if rows.Next() {
		err := rows.Scan(&p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return Parcel{}, fmt.Errorf("ошибка сканирования данных: %w", err)
		}
	} else {
		return Parcel{}, fmt.Errorf("посылка с номером %d не найдена", number)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// выполняем запрос для получения строк по client
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	// создаем срез для хранения результата
	var res []Parcel

	// обрабатываем каждую строку
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных: %v", err)
		}
		res = append(res, p) // добавляем объект Parcel в срез
	}

	// проверяем на наличие ошибок в процессе обхода строк
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк: %w", err)
	}

	// возвращаем результат
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус для посылки с номером %d: %w", number, err)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	var status string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("посылка с номером %d не найдена", number)
		}
		return fmt.Errorf("ошибка получения данных для посылки с номером %d: %v", number, err)
	}

	if status != ParcelStatusRegistered {
		return fmt.Errorf("невозможно изменить адрес, статус посылки с номером %d не 'registered'", number)
	}

	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
	if err != nil {
		return fmt.Errorf("ошибка обновления адреса: %v", err)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	var status string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("посылка с номером %d не найдена", number)
		}
		return fmt.Errorf("ошибка получения данных для посылки с номером %d: %v", number, err)
	}

	if status != ParcelStatusRegistered {
		return fmt.Errorf("невозможно удалить строку, статус посылки с номером %d не 'registered'", number)
	}
	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
	if err != nil {
		return fmt.Errorf("ошибка удаления строки: %v", err)
	}

	return nil
}
