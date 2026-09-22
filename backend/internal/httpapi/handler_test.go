package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Matyjash/PollAndGo/backend/internal/poll"
)

type fakeRepository struct {
	created    poll.CreateInput
	listedFor  string
	votePoll   string
	voteOption string
	voteVoter  string
}

func (f *fakeRepository) Create(_ context.Context, input poll.CreateInput) (poll.Poll, error) {
	f.created = input
	return poll.Poll{ID: "ABC123"}, nil
}
func (*fakeRepository) Get(context.Context, string, string) (poll.Poll, error) {
	return poll.Poll{}, poll.ErrNotFound
}
func (f *fakeRepository) ListForUser(_ context.Context, voterID string) ([]poll.Poll, error) {
	f.listedFor = voterID
	return []poll.Poll{{ID: "ABC123", Title: "Lunch?"}}, nil
}
func (f *fakeRepository) Vote(_ context.Context, pollID, optionID, voterID string) (poll.Poll, error) {
	f.votePoll, f.voteOption, f.voteVoter = pollID, optionID, voterID
	return poll.Poll{ID: pollID, UserVotedOption: &optionID}, nil
}

func TestCreatePoll(t *testing.T) {
	repo := &fakeRepository{}
	handler := New(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest("POST", "/api/polls", strings.NewReader(`{"title":" Lunch? ","options":[" Pizza ","Tacos"],"creatorId":"voter-1"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	if repo.created.Title != "Lunch?" || repo.created.Options[0] != "Pizza" {
		t.Fatalf("input was not normalized: %#v", repo.created)
	}
}

func TestCreatePollRejectsDuplicateOptions(t *testing.T) {
	handler := New(&fakeRepository{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest("POST", "/api/polls", strings.NewReader(`{"title":"Lunch?","options":["Pizza","pizza"],"creatorId":"voter-1"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestListPollsForUser(t *testing.T) {
	repo := &fakeRepository{}
	handler := New(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest("GET", "/api/polls?voterId=voter-1", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if repo.listedFor != "voter-1" {
		t.Fatalf("expected voter-1, got %q", repo.listedFor)
	}
}

func TestVote(t *testing.T) {
	repo := &fakeRepository{}
	handler := New(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest("POST", "/api/polls/ABC123/votes", strings.NewReader(`{"optionId":"option-1","voterId":"voter-1"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	if repo.votePoll != "ABC123" || repo.voteOption != "option-1" || repo.voteVoter != "voter-1" {
		t.Fatalf("unexpected vote arguments: %#v", repo)
	}
}
