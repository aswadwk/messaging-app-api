package utils

import (
	"aswadwk/messaging-task-go/dto"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func QueryPaginate(db *gorm.DB, model any, result any, page, perPage int, scopes ...func(*gorm.DB) *gorm.DB) (dto.QueryResponse, error) {
	var response dto.QueryResponse
	var total int64

	// Create a new query with the model and apply all scopes
	query := db.Model(model)
	for _, scope := range scopes {
		query = scope(query)
	}

	// Count the total number of records with scopes applied
	if err := query.Count(&total).Error; err != nil {
		return dto.QueryResponse{}, fmt.Errorf("failed to count records: %w", err)
	}

	// Set default values
	if perPage <= 0 {
		perPage = 10
	}
	if page <= 0 {
		page = 1
	}

	// Calculate pagination values
	lastPage := (int(total) + perPage - 1) / perPage
	if lastPage == 0 {
		lastPage = 1
	}

	// Handle empty result case early
	if total == 0 {
		return dto.QueryResponse{
			Total:    0,
			PerPage:  perPage,
			CurPage:  page,
			LastPage: lastPage,
			Data:     result,
		}, nil
	}

	// Fetch the records for the current page
	offset := (page - 1) * perPage
	if err := query.
		Limit(perPage).
		Offset(offset).
		Find(result).
		Error; err != nil {
		return dto.QueryResponse{}, fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch records")
	}

	response = dto.QueryResponse{
		Total:    int(total),
		PerPage:  perPage,
		CurPage:  page,
		LastPage: lastPage,
		Data:     result,
	}

	return response, nil
}

func SearchByName(query dto.QueryDto) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if query.Search != "" {
			return db.Where("name LIKE ?", "%"+query.Search+"%")
		}
		return db
	}
}

func SearchBy(field, value string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(field+" LIKE ?", "%"+value+"%")
	}
}

func OrderBy(field, order string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if order == "desc" {
			return db.Order(field + " DESC")
		}
		return db.Order(field + " ASC")
	}
}

// MapSlice maps a slice of any struct to another slice of any struct using a mapper function.
// dstSlicePtr harus pointer ke slice tujuan, srcSlice adalah slice sumber, mapper adalah fungsi mapping.
func MapSlice(dstSlicePtr any, srcSlice any, mapper func(src any) any) {
	srcVal := reflect.ValueOf(srcSlice)
	dstVal := reflect.ValueOf(dstSlicePtr).Elem()

	for i := 0; i < srcVal.Len(); i++ {
		mapped := mapper(srcVal.Index(i).Interface())
		dstVal.Set(reflect.Append(dstVal, reflect.ValueOf(mapped)))
	}
}
