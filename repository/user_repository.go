package repository

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/Reskill-2022/volunteering/errors"
	"github.com/Reskill-2022/volunteering/model"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
)

const collectionName = "volunteers"

type UserRepository struct {
	logger  zerolog.Logger
	client1 *firestore.Client
	client2 *firestore.Client
}

var _ UserRepositoryInterface = (*UserRepository)(nil)

func NewUserRepository(logger zerolog.Logger) *UserRepository {
	r := &UserRepository{
		logger: logger,
	}

	r.client1 = getClient("./service-account-1.json")
	r.client2 = getClient("./service-account-2.json")

	return r
}

func getClient(saFile string) *firestore.Client {
	ctx := context.Background()

	sa := option.WithCredentialsFile(saFile)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	return client
}

func (u *UserRepository) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	u.logger.Debug().Msgf("Firestore: creating user with email: %s", user.Email)

	gotUser, err := u.GetUser(ctx, user.Email)
	if err == nil || gotUser != nil {
		return gotUser, nil
	}

	if _, err := u.client1.Collection(collectionName).Doc(user.Email).Set(ctx, user); err != nil {
		return nil, errors.From(err, "client1 failed to create user", 500)
	}

	if _, err := u.client2.Collection(collectionName).Doc(user.Email).Set(ctx, user); err != nil {
		return nil, errors.From(err, "client2 failed to create user", 500)
	}

	return &user, nil
}

func (u *UserRepository) UpdateUser(ctx context.Context, user model.User) (*model.User, error) {
	u.logger.Debug().Msgf("Firestore: updating user with email: %s", user.Email)

	updates := []firestore.Update{
		{Path: "state", Value: user.State},
		{Path: "organization", Value: user.Organization},
		{Path: "years_of_experience", Value: user.YearsOfExperience},
		{Path: "volunteer_areas", Value: user.VolunteerAreas},
		{Path: "is_underrepresented", Value: user.IsUnderrepresented},
		{Path: "is_convicted", Value: user.IsConvicted},
	}

	if _, err := u.client1.Collection(collectionName).Doc(user.Email).Update(ctx, updates); err != nil {
		return nil, errors.From(err, "client1 failed to update user data", 500)
	}

	if _, err := u.client2.Collection(collectionName).Doc(user.Email).Update(ctx, updates); err != nil {
		return nil, errors.From(err, "client2 failed to update user data", 500)
	}

	return &user, nil
}

func (u *UserRepository) GetUser(ctx context.Context, email string) (*model.User, error) {
	u.logger.Debug().Msgf("Firestore: getting user with email: %s", email)

	data, err := u.client1.Collection(collectionName).Doc(email).Get(ctx)
	if err != nil {
		return nil, errors.From(err, "User Account Not Found", 404)
	}

	user := model.User{}
	err = data.DataTo(&user)
	if err != nil {
		return nil, errors.From(err, "failed to bind user data", 500)
	}

	return &user, nil
}
