package processors

import (
	"GateWarden/internals/app/db"
	"GateWarden/internals/app/models"
	"errors"
)

type CarsProcessor struct {
	storage *db.CarsStorage
}

func NewCarsProcessor(storage *db.CarsStorage) *CarsProcessor {
	processor := new(CarsProcessor)
	processor.storage = storage
	return processor
}

func (processor *CarsProcessor) CreateCar(car models.Car) error {
	if car.Colour == "" {
		return errors.New("colour should not be empty")
	}
	if car.Brand == "" {
		return errors.New("brand should not be empty")
	}
	if car.LicensePlate == "" {
		return errors.New("license plate should not be empty")
	}
	if car.Owner.Id <= 0 {
		return errors.New("owner id shall be filled")
	}
	return processor.storage.CreateCar(car)
}

func (processor *CarsProcessor) FindCar(id int64) (models.Car, error) {
	car := processor.storage.GetCarById(id)
	if car.Id != id {
		return car, errors.New("car not found")
	}
	return car, nil
}

func (processor *CarsProcessor) ListCars(userId int64, brandFilter string, colourFilter string, licenseFilter string) ([]models.Car, error) {
	return processor.storage.GetCarsList(userId, brandFilter, colourFilter, licenseFilter), nil
}
