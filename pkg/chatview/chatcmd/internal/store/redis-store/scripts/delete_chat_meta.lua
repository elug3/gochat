local chatId = KEYS[1]
local userId = KEYS[2]

assert(chatId ~= nil and chatId ~= "", "chatId required")
assert(userId ~= nil and userId ~= "", "userId required")

local metaKey = "chats:" .. chatId .. ":meta"
local userChatsKey = "users:" .. userId .. ":chats"

local deletedMeta = redis.call("DEL", metaKey)
local deletedUserChats = redis.call("DEL", userChatsKey)

return cjson.encode({
    deleted_meta = deletedMeta,
    deleted_user_chats = deletedUserChats,
    err = nil,
})
