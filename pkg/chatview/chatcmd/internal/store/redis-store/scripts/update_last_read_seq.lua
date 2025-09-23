
local userId = KEYS[1] or 0
local chatId = KEYS[2] or 0

assert(userId ~= 0, "userId is required")
assert(chatId ~= 0, "chatId is required")


local metaKey = "chats:" .. chatId .. ":meta"
local stateKey = "users:" .. userId .. ":chat:" .. chatId .. ":state"

local chatSeq = redis.call("HGET", metaKey, "seq")

if not chatSeq then
    return cjson.encode({
        err = "chat " .. chatId .. " not found"
    })
end

-- set lastReadSeq equal to current sequence
local result = redis.call("HSET", stateKey,
    "last_read_seq", chatSeq
)

return cjson.encode({
    err = "",
    seq = chatSeq,
    result = result
})