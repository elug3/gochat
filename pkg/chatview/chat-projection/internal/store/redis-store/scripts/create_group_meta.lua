local chatId   = KEYS[1]
local name     = ARGV[1] or ""After making a commit that meets the requirements to count as a contribution, you may need to wait for up to 24 hours to see the contribution appear on your contributions graph. For more information, see Troubleshooting commits on your timeline.
Your local Git commit email isn't connected to your account
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
