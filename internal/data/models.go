package data

import "github.com/jackc/pgx/v5/pgxpool"

type Models struct {
	Users UserModel
	Tokens TokenModel
}

func NewModels(pool *pgxpool.Pool) *Models {
	return &Models{
		Users: UserModel{
			Pool: pool,
		},
		Tokens: TokenModel{
			Pool: pool,
		},
	}
}
