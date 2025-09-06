local userId = KEYS[1]
local chatId = KEYS[2]


assert(chatId ~= "", "chatId required")
assert(userId ~= "", "userId required")

local userChatsKey = "users:" .. userId .. ":chats" 
local metaKey = "chats:" .. chatId .. ":meta"
local stateKey = "users:" .. userId .. ":chat:" .. chatId .. ":state"


-- if redis.call("EXISTS", metaKey) == 0 then
--     return cjson.encode({
--         err = "chat:" .. chatId .. " does not exist"
--     })
-- end

local added = redis.call("SADD", userChatsKey, chatId)

local seq = redis.call("HGET", metaKey, "seq")
if not seq then
    seq = "0"
end

-- prevents user chat state from being overwritten if it already exists
if redis.call("EXISTS", stateKey) == 0 then
    redis.call("HSET", stateKey,
    "lastReadSeq", seq,
    "pinned", "0",
    "muted", "0")

    return cjson.encode({
        added = added,
        err = nil,
    })
end
