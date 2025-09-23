local chatId   = KEYS[1]
local name     = ARGV[1]
local ts       = ARGV[2]

assert(chatId ~= "", "chatId required")
assert(name ~= "", "name required")

local metaKey = "chats:" .. chatId .. ":meta"

-- if redis.call("EXISTS", metaKey) == 1 then
--         return cjson.encode({
--                 added = 0,
--                 err = "chat: " .. chatId .. " already exists",
--         })
-- end

local added = redis.call("HSET", metaKey,
        "name", name,
        "last_message", "",
        "last_message_at", "0",
        "type", "group",
        "seq", "0"
)

return cjson.encode({
        added = added,
        err = nil
})
