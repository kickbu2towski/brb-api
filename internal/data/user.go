package data

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Avatar        string `json:"avatar"`
	Bio           string `json:"bio"`
}

type SearchUserResp struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Avatar         string `json:"avatar"`
	FollowingCount int    `json:"following_count"`
	FollowersCount int    `json:"followers_count"`
	IsFollowing    bool   `json:"is_following"`
	IsFriend       bool   `json:"is_friend"`
}

type BasicUserResp struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type Relation int

const (
	RelationFriends Relation = iota
	RelationFollowing
	RelationFollowers
)

type UserModel struct {
	Pool *pgxpool.Pool
}

func (m *UserModel) AddUser(ctx context.Context, u *User) error {
	stmt := `
		INSERT INTO users(id, username, email, email_verified, avatar) 
		VALUES
	    ($1, $2, $3, $4, $5) 
	  ON CONFLICT (id) DO UPDATE SET 
			username = excluded.username, 
			email = excluded.email, 
			email_verified = excluded.email_verified, 
			avatar = excluded.avatar
	  RETURNING id
	`
	args := []any{u.ID, u.Username, u.Email, u.EmailVerified, u.Avatar}
	_, err := m.Pool.Exec(ctx, stmt, args...)
	return err
}

func (m *UserModel) GetUserForToken(ctx context.Context, token, scope string) (*User, error) {
	hash := sha256.Sum256([]byte(token))

	stmt := `SELECT id, username, email, email_verified, avatar, bio FROM users u 
	 LEFT JOIN tokens t ON t.user_id = u.id WHERE t.hash = $1 AND scope = $2 AND t.expiry_time >= CURRENT_TIMESTAMP`

	var u User
	err := m.Pool.QueryRow(ctx, stmt, hash[:], scope).Scan(&u.ID, &u.Username, &u.Email, &u.EmailVerified, &u.Avatar, &u.Bio)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) GetUsers(ctx context.Context, userID, username string) ([]*SearchUserResp, error) {
	stmt := `
		SELECT
			u.id,
			u.username,
			u.avatar,
			COUNT(fr1.follower_id) AS followers_count,
			COUNT(fr2.following_id) AS following_count,
			EXISTS (
				SELECT 1 FROM follow_relations fr3
				WHERE fr3.follower_id = $1 AND fr3.following_id = u.id
			) AS is_following,
      EXISTS (
        SELECT 1 FROM follow_relations fr4
        WHERE fr4.follower_id = $1 AND fr4.following_id = u.id
      ) AND EXISTS (
        SELECT 1 FROM follow_relations fr5
        WHERE fr5.follower_id = u.id AND fr5.following_id = $1
      ) AS is_friend
		FROM
			users u
		LEFT JOIN
			follow_relations fr1 ON u.id = fr1.following_id
		LEFT JOIN
			follow_relations fr2 ON u.id = fr2.follower_id
    WHERE
	    u.username ilike $2 AND u.id <> $1
		GROUP BY
			u.id;
	`

	rows, err := m.Pool.Query(ctx, stmt, userID, fmt.Sprintf("%%%s%%", username))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*SearchUserResp, 0)
	for rows.Next() {
		var u SearchUserResp
		err := rows.Scan(&u.ID, &u.Username, &u.Avatar, &u.FollowersCount, &u.FollowingCount, &u.IsFollowing, &u.IsFriend)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	err = rows.Err()
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return users, nil
		}
		return nil, err
	}

	return users, nil
}

func (m *UserModel) FollowUser(ctx context.Context, followingID, followerID string) error {
	stmt := `
	 INSERT INTO follow_relations(following_id, follower_id) VALUES($1, $2)
	`
	_, err := m.Pool.Exec(ctx, stmt, followingID, followerID)
	return err
}

func (m *UserModel) UnfollowUser(ctx context.Context, followingID, followerID string) error {
	stmt := `
	 DELETE FROM follow_relations WHERE following_id = $1 AND follower_id = $2
	`
	_, err := m.Pool.Exec(ctx, stmt, followingID, followerID)
	return err
}

func (m *UserModel) GetUsersForRelation(ctx context.Context, relation Relation, userID string) ([]*BasicUserResp, error) {
	var stmt string

	switch relation {
	case RelationFriends:
		stmt = `
		SELECT u.id, u.username, u.avatar
		FROM users u
		JOIN follow_relations fr1 ON u.id = fr1.following_id
		JOIN follow_relations fr2 ON u.id = fr2.follower_id
	  AND fr2.following_id = fr1.follower_id
		WHERE fr1.follower_id = $1;
		`
	case RelationFollowing:
		stmt = `
		SELECT id, username, avatar FROM users u
		INNER JOIN follow_relations fr ON
		fr.follower_id = $1 AND fr.following_id = u.id
		`
	case RelationFollowers:
		stmt = `
		SELECT id, username, avatar FROM users u
		INNER JOIN follow_relations fr ON
		fr.follower_id = u.id AND fr.following_id = $1
		`
	}

	rows, err := m.Pool.Query(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*BasicUserResp{}
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

func (m *UserModel) GetUser(ctx context.Context, userID string) (*BasicUserResp, error) {
	stmt := `SELECT id, username, avatar FROM users WHERE id = $1`
	var u BasicUserResp
	err := m.Pool.QueryRow(ctx, stmt, userID).Scan(&u.ID, &u.Username, &u.Avatar)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
