package poll

import "time"

type Poll struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Options         []Option  `json:"options"`
	CreatedAt       time.Time `json:"createdAt"`
	UserVotedOption *string   `json:"userVotedOptionId,omitempty"`
	CreatedByUser   bool      `json:"createdByUser"`
}

type Option struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	VoteCount int    `json:"voteCount"`
}

type CreateInput struct {
	Title     string   `json:"title"`
	Options   []string `json:"options"`
	CreatorID string   `json:"creatorId"`
}

var ErrNotFound = errorString("poll not found")
var ErrAlreadyVoted = errorString("this user has already voted")
var ErrInvalidOption = errorString("option does not belong to this poll")

type errorString string

func (e errorString) Error() string { return string(e) }
