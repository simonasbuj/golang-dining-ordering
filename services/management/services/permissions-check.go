package services

import (
	"context"
	"errors"
	"fmt"
	"golang-dining-ordering/services/management/repository"

	"github.com/google/uuid"
)

// ErrUserIsNotManager is returned when a user attempts an action that requires restaurant manager privileges.
var ErrUserIsNotManager = errors.New("user is not a manager of this restaurant")

func isUserRestaurantManager(
	ctx context.Context,
	userID uuid.UUID,
	restaurantID uuid.UUID,
	repo repository.RestaurantRepository,
) error {
	err := repo.IsUserRestaurantManager(ctx, userID, restaurantID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUserIsNotManager, err)
	}

	return nil
}
