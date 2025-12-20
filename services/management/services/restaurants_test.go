package services

import (
	"context"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/dto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	mock "golang-dining-ordering/test/mock/management"
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

type restaurantsServiceTestSuite struct {
	suite.Suite

	svc  *restaurantService
	user *authDto.TokenClaimsDto
}

func (suite *restaurantsServiceTestSuite) SetupSuite() {
	mockOrdersRepo := mock.NewMockRestaurantsRepo()
	suite.svc = NewRestaurantService(mockOrdersRepo)

	suite.user = &authDto.TokenClaimsDto{
		UserID: testUserID,
	}
}

func TestRestaurantsServiceTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(restaurantsServiceTestSuite))
}

func (suite *restaurantsServiceTestSuite) TestCreateRestaurant_Success() {
	reqDto := &dto.CreateRestaurantDto{
		ID:       testRestaurantID,
		UserID:   testUserID,
		Name:     testRestaurantName,
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}

	got, err := suite.svc.CreateRestaurant(context.Background(), reqDto)
	suite.Require().NoError(err)
	suite.Equal(reqDto, got)
}

func (suite *restaurantsServiceTestSuite) TestCreateRestaurant_Error() {
	reqDto := &dto.CreateRestaurantDto{
		ID:       testRestaurantID,
		UserID:   testUserID,
		Name:     "",
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}

	got, err := suite.svc.CreateRestaurant(context.Background(), reqDto)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *restaurantsServiceTestSuite) TestGetRestaurants_Success() {
	reqDto := &dto.GetRestaurantsReqDto{
		Page:  0,
		Limit: 0,
	}

	want := &dto.GetRestaurantsRespDto{
		Page:  1,
		Limit: 10,
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
	}

	got, err := suite.svc.GetRestaurants(context.Background(), reqDto)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *restaurantsServiceTestSuite) TestGetRestaurants_Error() {
	reqDto := &dto.GetRestaurantsReqDto{
		Page:  69,
		Limit: 0,
	}

	got, err := suite.svc.GetRestaurants(context.Background(), reqDto)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *restaurantsServiceTestSuite) TestGetRestaurantById_Success() {
	want := &dto.RestaurantItemDto{
		ID:        testRestaurantID,
		Name:      testRestaurantName,
		Address:   testRestaurantAddress,
		Currency:  testRestaurantCurrency,
		CreatedAt: testDateTime,
	}

	got, err := suite.svc.GetRestaurantByID(context.Background(), testRestaurantID)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *restaurantsServiceTestSuite) TestGetRestaurantById_Error() {
	tests := []struct {
		name         string
		restaurantID uuid.UUID
	}{
		{"invalid id", uuid.Nil},
		{"repo failed", uuid.Max},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(_ *testing.T) {
			got, err := suite.svc.GetRestaurantByID(context.Background(), tt.restaurantID)
			suite.Require().Error(err)
			suite.Nil(got)
		})
	}
}

func (suite *restaurantsServiceTestSuite) TestUpdateRestaurant_Success() {
	deleteFlag := true
	reqDto := &dto.UpdateRestaurantRequestDto{
		ID:         testRestaurantID,
		UserID:     testUserID,
		Name:       &testRestaurantName,
		Address:    &testRestaurantAddress,
		Currency:   &testRestaurantCurrency,
		DeleteFlag: &deleteFlag,
	}

	want := &dto.UpdateRestaurantResponseDto{
		ID:        testRestaurantID,
		Name:      testRestaurantName,
		Address:   testRestaurantAddress,
		Currency:  testRestaurantCurrency,
		CreatedAt: testDateTime,
		UpdatedAt: testDateTime,
		DeletedAt: testDateTime,
	}

	got, err := suite.svc.UpdateRestaurant(context.Background(), reqDto)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *restaurantsServiceTestSuite) TestUpdateRestaurant_Error() {
	tests := []struct {
		name         string
		restaurantID uuid.UUID
		userID       uuid.UUID
	}{
		{"user not a manager", testRestaurantID, uuid.Max},
		{"repo failed", uuid.Max, testUserID},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(_ *testing.T) {
			reqDto := &dto.UpdateRestaurantRequestDto{
				ID:         tt.restaurantID,
				UserID:     tt.userID,
				Name:       &testRestaurantName,
				Address:    &testRestaurantAddress,
				Currency:   &testRestaurantCurrency,
				DeleteFlag: nil,
			}

			got, err := suite.svc.UpdateRestaurant(context.Background(), reqDto)
			suite.Require().Error(err)
			suite.Nil(got)
		})
	}
}

func (suite *restaurantsServiceTestSuite) TestCreateTable_Success() {
	reqDto := &dto.RestaurantTableDto{
		ID:           testTableID,
		RestaurantID: testRestaurantID,
		UserID:       testUserID,
		Name:         testTableName,
		Capacity:     testTableCapacity,
	}

	got, err := suite.svc.CreateTable(context.Background(), reqDto)
	suite.Require().NoError(err)
	suite.Equal(reqDto, got)
}

func (suite *restaurantsServiceTestSuite) TestCreateTable_Error() {
	tests := []struct {
		name         string
		restaurantID uuid.UUID
		userID       uuid.UUID
	}{
		{"user not a manager", testRestaurantID, uuid.Max},
		{"repo failed", uuid.Nil, testUserID},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(_ *testing.T) {
			reqDto := &dto.RestaurantTableDto{
				ID:           testTableID,
				RestaurantID: tt.restaurantID,
				UserID:       tt.userID,
				Name:         testTableName,
				Capacity:     testTableCapacity,
			}

			got, err := suite.svc.CreateTable(context.Background(), reqDto)
			suite.Require().Error(err)
			suite.Nil(got)
		})
	}
}

func (suite *restaurantsServiceTestSuite) TestGetTables_Success() {
	want := []*dto.RestaurantTableDto{
		{
			ID:           testTableID,
			RestaurantID: testRestaurantID,
			UserID:       testUserID,
			Name:         testTableName,
			Capacity:     testTableCapacity,
		},
	}

	got, err := suite.svc.GetTables(context.Background(), testRestaurantID)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *restaurantsServiceTestSuite) TestGetTables_Error() {
	got, err := suite.svc.GetTables(context.Background(), uuid.Nil)
	suite.Require().Error(err)
	suite.Nil(got)
}
