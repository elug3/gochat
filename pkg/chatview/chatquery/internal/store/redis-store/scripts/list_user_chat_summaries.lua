
local userId = KEYS[1]

local userChatsKey = "users:" .. userId .. ":chats"

local chatIds = redis.call("SMEMBERS", userChatsKey)
if not chatIds or #chatIds == 0 then
    return cjson.encode({
        chats = {},
        err = "",
    })
end

local summaries = {}

for _, chatId in ipairs(chatIds) do
    local metaKey = "chats:" .. chatId .. ":meta"
    local stateKey = "users:" .. userId .. ":chat:" .. chatId .. ":state"
    
    local meta = redis.call("HGETALL", metaKey)
    local state = redis.call("HGETALL", stateKey)

    local summaryMap = {}
    for i = 1, #meta, 2 do summaryMap[meta[i]] = meta[i + 1] end
    for i = 1, #state, 2 do summaryMap[state[i]] = state[i + 1] end

    local seq = tonumber(summaryMap["seq"])
    local lastReadSeq = tonumber(summaryMap["last_read_seq"])
    local unreadCount = seq - lastReadSeq
    
    local summary = {
        id = tonumber(chatId),
        name = summaryMap["name"],
        type = summaryMap["type"],
        last_message_at = tonumber(summaryMap["last_message_at"]),
        last_message = summaryMap["last_message"],
        unread_count = unreadCount,
        pinned = summaryMap["pinned"] == "1",
        muted = summaryMap["muted"] == "1",
    }
    table.insert(summaries, summary)
end

return cjson.encode({
    chats = summaries,
    err = "",
})

