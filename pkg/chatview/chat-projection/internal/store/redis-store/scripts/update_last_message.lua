-- update_last_message increments the sequence number and updates last message info for a chat.
-- If chat metadata does not exist, return an error.

local chatId = KEYS[1]

local lastMessage = ARGV[1]
local lastMessageAt = ARGV[2]

assert(chatId ~= nil, "chatId is required")
assert(lastMessage ~= nil, "lastMessage is required")
assert(lastMessageAt ~= nil, "lastMessageAt is required")

local metaKey = "chats:" .. chatId .. ":meta"

-- if redis.call("EXISTS", metaKey) == 0 then
--     return cjson.encode({
--         seq = nil,
--         err = "chat: " .. chatId .. " does not exist",
--     })
-- end

redis.call("HSET", metaKey,
"lastMessage", lastMessage,
"lastMessageAt", lastMessageAt
)

local seq = redis.call("HINCRBY", metaKey, "seq", 1)

return cjson.encode({
    seq = seq,
    err = nil,
})