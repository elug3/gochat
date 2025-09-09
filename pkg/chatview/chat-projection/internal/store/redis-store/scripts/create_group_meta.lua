local chatId   = KEYS[1]
local name     = ARGV[1] or ""
local ts       = ARGV[2] or "0"

assert(chatId ~= "", "chatId required")

local metaKey = "chats:" .. chatId .. ":meta"

-- if redis.call("EXISTS", metaKey) == 1 then
--         return cjson.encode({
--                 added = 0,
--                 err = "chat: " .. chatId .. " already exists",
--         })
-- end

local added = redis.call("HSET", metaKey,
        "name", name,
        "lastMessage", "",
        "lastMessageAt", "0",
        "type", "group",
        "seq", "0"
)

return cjson.encode({
        added = added,
        err = nil
})
