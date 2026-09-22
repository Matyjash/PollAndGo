package poll

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(context.Context, CreateInput) (Poll, error)
	Get(context.Context, string, string) (Poll, error)
	ListForUser(context.Context, string) ([]Poll, error)
	Vote(context.Context, string, string, string) (Poll, error)
}

type Store struct{ db *pgxpool.Pool }

func NewStore(db *pgxpool.Pool) *Store { return &Store{db: db} }

func (s *Store) Create(ctx context.Context, input CreateInput) (Poll, error) {
	for attempts := 0; attempts < 4; attempts++ {
		id, err := randomID(6)
		if err != nil {
			return Poll{}, err
		}

		tx, err := s.db.Begin(ctx)
		if err != nil {
			return Poll{}, err
		}
		var created Poll
		err = tx.QueryRow(ctx,
			`INSERT INTO polls (id, title, creator_id) VALUES ($1, $2, $3) RETURNING id, title, created_at`,
			id, input.Title, input.CreatorID,
		).Scan(&created.ID, &created.Title, &created.CreatedAt)
		if err != nil {
			_ = tx.Rollback(ctx)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			return Poll{}, err
		}

		for position, label := range input.Options {
			optionID, err := uuid()
			if err != nil {
				_ = tx.Rollback(ctx)
				return Poll{}, err
			}
			if _, err = tx.Exec(ctx,
				`INSERT INTO poll_options (id, poll_id, label, position) VALUES ($1, $2, $3, $4)`,
				optionID, id, label, position,
			); err != nil {
				_ = tx.Rollback(ctx)
				return Poll{}, err
			}
			created.Options = append(created.Options, Option{ID: optionID, Label: label})
		}
		if err = tx.Commit(ctx); err != nil {
			return Poll{}, err
		}
		created.CreatedByUser = true
		return created, nil
	}
	return Poll{}, fmt.Errorf("could not allocate poll id")
}

func (s *Store) Get(ctx context.Context, id, voterID string) (Poll, error) {
	var result Poll
	err := s.db.QueryRow(ctx, `
		SELECT id, title, created_at, COALESCE(creator_id = $2::varchar(100), false)
		FROM polls WHERE id = $1`, strings.ToUpper(id), voterID,
	).Scan(&result.ID, &result.Title, &result.CreatedAt, &result.CreatedByUser)
	if errors.Is(err, pgx.ErrNoRows) {
		return Poll{}, ErrNotFound
	}
	if err != nil {
		return Poll{}, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT o.id, o.label, COUNT(v.option_id)::int
		FROM poll_options o
		LEFT JOIN votes v ON v.option_id = o.id
		WHERE o.poll_id = $1
		GROUP BY o.id, o.label, o.position
		ORDER BY o.position`, result.ID)
	if err != nil {
		return Poll{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var option Option
		if err := rows.Scan(&option.ID, &option.Label, &option.VoteCount); err != nil {
			return Poll{}, err
		}
		result.Options = append(result.Options, option)
	}
	if err := rows.Err(); err != nil {
		return Poll{}, err
	}

	if voterID != "" {
		var optionID string
		err = s.db.QueryRow(ctx,
			`SELECT option_id FROM votes WHERE poll_id = $1 AND voter_id = $2`, result.ID, voterID,
		).Scan(&optionID)
		if err == nil {
			result.UserVotedOption = &optionID
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return Poll{}, err
		}
	}
	return result, nil
}

func (s *Store) ListForUser(ctx context.Context, voterID string) ([]Poll, error) {
	rows, err := s.db.Query(ctx, `
		SELECT p.id
		FROM polls p
		WHERE p.creator_id = $1::varchar(100)
		   OR EXISTS (SELECT 1 FROM votes v WHERE v.poll_id = p.id AND v.voter_id = $1::varchar(100))
		ORDER BY p.created_at DESC`, voterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	results := make([]Poll, 0, len(ids))
	for _, id := range ids {
		result, err := s.Get(ctx, id, voterID)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Store) Vote(ctx context.Context, pollID, optionID, voterID string) (Poll, error) {
	command, err := s.db.Exec(ctx, `
		INSERT INTO votes (poll_id, option_id, voter_id)
		SELECT o.poll_id, o.id, $3::varchar(100)
		FROM poll_options o
		WHERE o.id = $2::varchar(36) AND o.poll_id = $1::varchar(12)`,
		strings.ToUpper(pollID), optionID, voterID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Poll{}, ErrAlreadyVoted
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return Poll{}, ErrNotFound
		}
		return Poll{}, err
	}
	if command.RowsAffected() == 0 {
		return Poll{}, ErrInvalidOption
	}
	return s.Get(ctx, pollID, voterID)
}

func randomID(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(buf)), nil
}

func uuid() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:]), nil
}
