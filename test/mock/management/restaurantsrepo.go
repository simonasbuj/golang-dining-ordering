// Package management holds implementation of moclk restaurant and menu repos.
package management

import (
	"context"
	"errors"
	"golang-dining-ordering/services/management/dto"
	"time"

	"github.com/google/uuid"
)

var (
	errRepoFailed = errors.New("repo failed")
	// ErrUserIsNotManager is returned when a user attempts an action that requires restaurant manager privileges.
	ErrUserIsNotManager = errors.New("user is not a manager of this restaurant")
)

//nolint:gochecknoglobals
var (
	testUserID             = uuid.MustParse("67676767-6767-6767-6767-676767676767")
	testRestaurantID       = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantCurrency = "eur"
	testRestaurantAddress  = "Miško g. 7, Raudondvaris"
	testRestaurantName     = "Viskas Viename KO"
	testTableID            = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testTableName          = "table 01"
	testTableCapacity      = 4
	testDateTime           = time.Date(2025, time.December, 5, 19, 0, 0, 0, &time.Location{})
)

type mockRestaurantsRepo struct{}

// NewMockRestaurantsRepo creates mock restaurant repo.
func NewMockRestaurantsRepo() *mockRestaurantsRepo { //nolint:revive
	return &mockRestaurantsRepo{}
}

func (*mockRestaurantsRepo) CreateRestaurant(
	_ context.Context,
	reqDto *dto.CreateRestaurantDto,
) (*dto.CreateRestaurantDto, error) {
	if reqDto.Name != testRestaurantName {
		return nil, errRepoFailed
	}

	return &dto.CreateRestaurantDto{
		ID:       testRestaurantID,
		UserID:   testUserID,
		Name:     testRestaurantName,
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}, nil
}

func (*mockRestaurantsRepo) GetRestaurants(
	_ context.Context,
	reqDto *dto.GetRestaurantsReqDto,
) (*dto.GetRestaurantsRespDto, error) {
	if reqDto.Page == 69 { //nolint:mnd
		return nil, errRepoFailed
	}

	return &dto.GetRestaurantsRespDto{
		Page:  reqDto.Page,
		Limit: reqDto.Limit,
		Total: 1,
		Restaurants: []dto.RestaurantItemDto{
			{
				ID:        testRestaurantID,
				Name:      testRestaurantName,
				Address:   testRestaurantAddress,
				Currency:  testRestaurantCurrency,
				CreatedAt: testDateTime,
			},
		},
	}, nil
}

func (*mockRestaurantsRepo) GetRestaurantByID(
	_ context.Context,
	id uuid.UUID,
) (*dto.RestaurantItemDto, error) {
	if id != testRestaurantID {
		return nil, errRepoFailed
	}

	return &dto.RestaurantItemDto{
		ID:        testRestaurantID,
		Name:      testRestaurantName,
		Address:   testRestaurantAddress,
		Currency:  testRestaurantCurrency,
		CreatedAt: testDateTime,
	}, nil
}

func (*mockRestaurantsRepo) IsUserRestaurantManager(
	_ context.Context,
	userID, _ uuid.UUID,
) error {
	if userID != testUserID {
		return ErrUserIsNotManager
	}

	return nil
}

func (*mockRestaurantsRepo) UpdateRestaurant(
	_ context.Context,
	reqDto *dto.UpdateRestaurantRequestDto,
) (*dto.UpdateRestaurantResponseDto, error) {
	if reqDto.ID != testRestaurantID {
		return nil, errRepoFailed
	}

	return &dto.UpdateRestaurantResponseDto{
		ID:        testRestaurantID,
		Name:      testRestaurantName,
		Address:   testRestaurantAddress,
		Currency:  testRestaurantCurrency,
		CreatedAt: testDateTime,
		UpdatedAt: testDateTime,
		DeletedAt: testDateTime,
	}, nil
}

func (*mockRestaurantsRepo) CreateTable(
	_ context.Context,
	reqDto *dto.RestaurantTableDto,
) (*dto.RestaurantTableDto, error) {
	if reqDto.UserID != testUserID {
		return nil, ErrUserIsNotManager
	}

	if reqDto.RestaurantID != testRestaurantID {
		return nil, errRepoFailed
	}

	return &dto.RestaurantTableDto{
		ID:           testTableID,
		RestaurantID: testRestaurantID,
		UserID:       testUserID,
		Name:         testTableName,
		Capacity:     testTableCapacity,
	}, nil
}

func (*mockRestaurantsRepo) GetTables(
	_ context.Context,
	id uuid.UUID,
) ([]*dto.RestaurantTableDto, error) {
	if id != testRestaurantID {
		return nil, errRepoFailed
	}

	return []*dto.RestaurantTableDto{
		{
			ID:           testTableID,
			RestaurantID: testRestaurantID,
			UserID:       testUserID,
			Name:         testTableName,
			Capacity:     testTableCapacity,
		},
	}, nil
}
