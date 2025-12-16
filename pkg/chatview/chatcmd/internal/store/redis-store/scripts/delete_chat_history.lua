local userId = KEYS[1]
local chatId = KEYS[2]

assert(userId ~= nil and userId ~= "", "userId required")
assert(chatId ~= nil and chatId ~= "", "chatId required")

local userChatsKey = "users:" .. userId .. ":chats"
local stateKey = "users:" .. userId .. ":chat:" .. chatId .. ":state"

local removed = redis.call("SREM", userChatsKey, chatId)
local deletedState = redis.call("DEL", stateKey)

return cjson.encode({
    removed = removed,
    deleted_state = deletedState,
    err = nil,
})
