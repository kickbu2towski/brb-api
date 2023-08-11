package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DM struct {
	id           int
	Participants []string
}

type DMModel struct {
	Pool *pgxpool.Pool
}

func (m *DMModel) GetDMListForUser(ctx context.Context, userID int) ([]*BasicUserResp, error) {
	stmt := `
    SELECT id, username, avatar FROM users u
    WHERE u.id in 
      (SELECT p.participants[1] FROM
        (select array_remove(participants, $1) as participants FROM dms d2
          WHERE d2.id in (
            select d.id from dms d
            join messages m on
              m.dm_id = d.id
              and $1 = ANY(d.participants)
            group by d.id)
         ) as 
      p)
  `

	users := make([]*BasicUserResp, 0)
	rows, err := m.Pool.Query(ctx, stmt, userID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return users, nil
		default:
			return nil, err
		}
	}

	for rows.Next() {
		var user BasicUserResp
		err := rows.Scan(&user.ID, &user.Username, &user.Avatar)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m *DMModel) GetDMForParticipants(ctx context.Context, participants []int) (int, error) {
	stmt := `SELECT id FROM dms WHERE participants <@ $1`
	var (
		dmID      int
		isInitial bool
	)

	err := m.Pool.QueryRow(ctx, stmt, participants).Scan(&dmID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			isInitial = true
		default:
			return dmID, err
		}
	}

	if !isInitial {
		return dmID, nil
	}

	stmt = `INSERT INTO dms(participants) VALUES($1) RETURNING id`
	err = m.Pool.QueryRow(ctx, stmt, participants).Scan(&dmID)
	return dmID, nil
}
