package redisstore

import (
	_ "embed"

	"github.com/redis/go-redis/v9"
)

//go:embed scripts/list_user_chat_summaries.lua
var listUserChatSummariesSrc string

var listUserChatSummariesScript = redis.NewScript(listUserChatSummariesSrc)
