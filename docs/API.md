# Poll & Go API

The backend listens on port `8080` by default and serves JSON responses. Error responses have the following shape:

```json
{
  "error": "error description"
}
```

## Health check

### `GET /healthz`

Returns `200 OK` when the HTTP service is running.

## Polls

### `POST /api/polls`

Creates a poll. A poll must have a title and between 2 and 20 unique options.

```json
{
  "title": "Where should we have lunch?",
  "options": ["Pizza", "Tacos"],
  "creatorId": "anonymous-browser-id"
}
```

Returns the created poll with `201 Created`.

### `GET /api/polls?voterId={voterId}`

Lists polls created by or voted in by the supplied anonymous browser ID, ordered from newest to oldest.

### `GET /api/polls/{pollId}?voterId={voterId}`

Returns a poll, its options, current vote totals, and the option selected by this visitor when applicable.

## Votes

### `POST /api/polls/{pollId}/votes`

Casts one vote for an option.

```json
{
  "optionId": "option-id",
  "voterId": "anonymous-browser-id"
}
```

Returns the updated poll with `201 Created`. The database enforces one vote per `(poll_id, voter_id)`, so repeated or concurrent submissions cannot create additional votes.
