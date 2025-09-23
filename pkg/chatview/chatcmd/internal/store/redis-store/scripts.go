package redisstore

import (
	_ "embed"

	"github.com/redis/go-redis/v9"
)

//go:embed scripts/create_group_meta.lua
var createGroupMetaSrc string

//go:embed scripts/update_last_message.lua
var updateLastMessageSrc string

//go:embed scripts/add_chat_to_user.lua
var addChatToUserSrc string

//go:embed scripts/update_last_read_seq.lua
var updateLastReadSeqSrc string

var (
	createGroupMetaScript   = redis.NewScript(createGroupMetaSrc)
	updateLastMessageScript = redis.NewScript(updateLastMessageSrc)
	addChatToUserScript     = redis.NewScript(addChatToUserSrc)
	updateLastReadSeqScript = redis.NewScript(updateLastReadSeqSrc)
)
