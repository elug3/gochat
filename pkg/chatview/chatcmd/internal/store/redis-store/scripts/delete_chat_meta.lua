local chatId = KEYS[1]

assert(chatId ~= nil and chatId ~= "", "chatId required")

local metaKey = "chats:" .. chatId .. ":meta"

local deletedMeta = redis.call("DEL", metaKey)

local removedUserChats = 0
local cursor = "0"

repeat
    local scan = redis.call("SCAN", cursor, "MATCH", "users:*:chats", "COUNT", 100)
    cursor = scan[1]
    local keys = scan[2]

    for _, key in ipairs(keys) do
        local removed = redis.call("SREM", key, chatId)
        removedUserChats = removedUserChats + removed
    end
until cursor == "0"

local deletedStates = 0
cursor = "0"

repeat
    local scan = redis.call("SCAN", cursor, "MATCH", "users:*:chat:" .. chatId .. ":state", "COUNT", 100)
    cursor = scan[1]
    local keys = scan[2]

    for _, key in ipairs(keys) do
        deletedStates = deletedStates + redis.call("DEL", key)
    end
until cursor == "0"

return cjson.encode({
    deleted_meta = deletedMeta,
    removed_user_chats = removedUserChats,
    deleted_states = deletedStates,
    err = nil,
})
