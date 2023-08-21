package manager

import (
	"context"
	"database/sql"

	"github.com/sgykfjsm/golang-tidb-circleci-example/mydb"
)

type UserManager struct {
	db  *sql.DB
	q   *mydb.Queries
	ctx context.Context
}

func NewUserManager(db *sql.DB, ctx context.Context) UserManager {
	return UserManager{db, mydb.New(db), ctx}
}

func (u *UserManager) AddUser(user mydb.User) (int64, error) {
	res, err := u.q.AddUser(u.ctx, user.Name)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (u *UserManager) UpdateUser(user mydb.User) error {

	return u.q.UpdateUser(u.ctx, mydb.UpdateUserParams{
		ID:   user.ID,
		Name: user.Name,
	})
}

func (u *UserManager) GetUserByID(id int64) (mydb.User, error) {
	return u.q.GetUserByID(u.ctx, id)
}

func (u *UserManager) ListUsers() ([]mydb.User, error) {
	return u.q.ListUsers(u.ctx)
}
