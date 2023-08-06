package data

import "github.com/jackc/pgx/v5/pgxpool"

type Models struct {
	Users    UserModel
	Messages MessageModel
	Tokens   TokenModel
	DMs      DMModel
}

func NewModels(pool *pgxpool.Pool) *Models {
	return &Models{
		Users: UserModel{
			Pool: pool,
		},
		Tokens: TokenModel{
			Pool: pool,
		},
		Messages: MessageModel{
			Pool: pool,
		},
		DMs: DMModel{
			Pool: pool,
		},
	}
}
