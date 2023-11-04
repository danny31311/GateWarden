package db

import (
	"GateWarden/internals/app/models"
	"context"
	"errors"

	"fmt"
	"github.com/georgysavva/scany/pgxscan"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type CarsStorage struct {
	databasePool *pgxpool.Pool
}

type userCar struct {
	UserId       int64 `db:"userid"`
	Name         string
	Rank         string
	CarId        int64 `db:"carid""`
	Brand        string
	Colour       string
	LicensePlate string
}

func convertJoinedQueryToCar(input userCar) models.Car {
	return models.Car{
		Id:           input.CarId,
		Colour:       input.Colour,
		Brand:        input.Brand,
		LicensePlate: input.LicensePlate,
		Owner: models.User{
			Id:   input.UserId,
			Name: input.Name,
			Rank: input.Rank,
		},
	}
}

func NewCarStorage(pool *pgxpool.Pool) *CarsStorage {
	storage := new(CarsStorage)
	storage.databasePool = pool
	return storage
}

func (storage *CarsStorage) GetCarsList(userIdFilter int64, brandFilter string, colourFilter string, licenseFilter string) []models.Car {
	query := "SELECT users.id as userid, users.name, users.rank, c.id as carid, c.brand,c.colour, c.license_plate FROM users join cars c on users.id = c.user_id WHERE 1=1"
	placeholderNum := 1
	args := make([]interface{}, 0)
	if userIdFilter != 0 {
		query += fmt.Sprintf(" AND users.ID = $%d", placeholderNum)
		args = append(args, userIdFilter)
		placeholderNum++
	}
	if brandFilter != "" {
		query += fmt.Sprintf(" AND brand ILIKE $%d", placeholderNum)
		args = append(args, fmt.Sprintf("%%%s%%", brandFilter))
		placeholderNum++
	}
	if colourFilter != "" {
		query += fmt.Sprintf(" AND colour ILIKE $%d", placeholderNum)
		args = append(args, fmt.Sprintf("%%%s%%", colourFilter))
	}
	if licenseFilter != "" {
		query += fmt.Sprintf(" AND license_plate ILIKE $%d", placeholderNum)
		args = append(args, fmt.Sprintf("%%%s%%", licenseFilter))
		placeholderNum++
	}
	var dbResult []userCar
	err := pgxscan.Select(context.Background(), storage.databasePool, &dbResult, query, args...)
	if err != nil {
		logrus.Errorln(err)
	}
	result := make([]models.Car, len(dbResult))
	for idx, dbEntity := range dbResult {
		result[idx] = convertJoinedQueryToCar(dbEntity)
	}
	return result
}

func (storage *CarsStorage) GetCarById(id int64) models.Car {
	query := "SELECT users.id as userid, users.name, users.rank, c.id as carid, c.brand, c.colour, c.license_plate FROM users join cars c on users.id = c.user_id WHERE c.id = $1"
	var result []userCar
	err := pgxscan.Select(context.Background(), storage.databasePool, &result, query, id)
	if err != nil {
		logrus.Errorln(err)
		return models.Car{}
	}
	if len(result) == 0 {
		return models.Car{}
	}
	return convertJoinedQueryToCar(result[0])
}

func (storage *CarsStorage) CreateCar(car models.Car) error {
	ctx := context.Background()
	tx, err := storage.databasePool.Begin(ctx)
	defer func() {
		err = tx.Rollback(context.Background())
		if err != nil {
			logrus.Errorln(err)
		}
	}()

	query := "SELECT id FROM users WHERE ID = $1"

	id := -1

	err = pgxscan.Get(ctx, tx, &id, query, car.Owner.Id)
	if err != nil {
		logrus.Errorln(err)
		err = tx.Rollback(context.Background())
		if err != nil {
			logrus.Errorln(err)
		}
		return err
	}
	if id == -1 {
		return errors.New("user not found")
	}
	insertQuery := "INSERT INTO cars(user_id, colour, brand, license_plate) VALUES ($1, $2, $3, $4)"
	_, err = tx.Exec(context.Background(), insertQuery, car.Owner.Id, car.Colour, car.Brand, car.LicensePlate)

	if err != nil {
		logrus.Errorln(err)
		err = tx.Rollback(context.Background())
		if err != nil {
			logrus.Errorln(err)
		}
		return err
	}
	err = tx.Commit(context.Background())
	if err != nil {
		logrus.Errorln(err)
	}
	return err

}
