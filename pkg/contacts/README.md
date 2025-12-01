# Contacts Service

Manages user profiles, groups, memberships, and personal contact lists for Gochat. Exposes an HTTP API, persists to SQLite, and emits/consumes events to keep other services in sync.

## Components

- HTTP server (`NewHttpServer`) built with Gin; routes defined in `handler.go`.
- Persistence via SQLite (`contacts.db` in `Options.SaveDir`); `NoSave` keeps everything in-memory.
- Event publisher to NATS when `NoEvent` is false; event listener subscribes to auth events.
- Optional icon generator (`identicon`) uploaded to S3-compatible storage when `NoIcons` is false.

## Configuration

Flags are wired in `ConfigureOptions` and consumed by `cmd/gochat-contacts_server`:

- `-H, --host` HTTP bind host (default `0.0.0.0`)
- `-p, --post` HTTP port (default `8080`)
- `-d, --data` directory for `contacts.db` (default `./`)
- `--no-save` disable persistence (in-memory DB)
- `--no-event` disable NATS publisher
- `--no-icons` skip icon generation/uploads
- `--nats-url` NATS server URL (env `NATS_URL` fallback, default `nats://localhost:4222`)
- `--s3-endpoint` S3/MinIO endpoint for profile icons (env `S3_ENDPOINT`, default `http://localhost:9000`)

Run locally:

```
go run ./cmd/gochat-contacts_server \
  --host 0.0.0.0 --post 8080 \
  --data ./data --nats-url nats://localhost:4222
```

## Events

Published (when mutations succeed and `pub` is configured):

- `contacts.group.created` (`GroupCreated`) after `CreateUserGroup`
- `contacts.group.deleted` (`GroupDeleted`) after `DeleteUserGroup`
- `contacts.member.joined` (`MemberJoined`) after `CreateUserGroup` and `Invite`
- `contacts.member.left` (`MemberLeft`) after `DeleteMember`

Consumed:

- `auth.registered` (`UserRegistered`) to auto-create a profile in `EventHandler.OnUserRegistered`.

## HTTP API

Base path: `/`

| Method | Path | Body | Description |
| --- | --- | --- | --- |
| GET | `/groups` | – | List groups (limit 50) |
| GET | `/groups/:group_id` | – | Get group by id |
| GET | `/can` | query `user_id`, `chat_id`, `target_id`, `action` | Permission check (actions: invite, send, read, leave, delete_member) |
| GET | `/user/:user_id/groups` | – | List groups the user belongs to |
| POST | `/user/:user_id/groups` | `{"name":"Group Name"}` | Create group and add creator as owner |
| GET | `/user/:user_id/groups/:group_id` | – | Get group if member |
| GET | `/user/:user_id/groups/:group_id/members` | – | List members if requester is in the group |
| POST | `/user/:user_id/groups/:group_id/members` | `{"user_id":123}` | Invite/add member (permission enforced) |
| DELETE | `/user/:user_id/groups/:group_id/members/:member_id` | – | Remove member or leave group |
| GET | `/user/:user_id/profile` | – | Get profile by user id |
| GET | `/profiles` | – | List profiles (limit 50) |
| POST | `/profiles` | `{"user_id":123,"name":"Alice"}` | Create profile and optional icon |
| DELETE | `/profiles/:user_id` | – | Delete profile (blocked if user owns groups) |
| GET | `/user/:user_id/contacts` | – | List contacts for user |
| POST | `/user/:user_id/contacts` | `{"target_id":456}` | Add user to contacts |
| GET | `/health` | – | Liveness check |

## Notes

- Transactions are scoped per request; most handlers open a store transaction, perform checks, then emit events after commit.
- `NoEvent` and `NoIcons` toggles let you run without NATS or S3 during local development.
