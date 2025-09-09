
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
    local stateKey = "users:" .. userId .. ":chats:" .. chatId .. ":state"
    
    local meta = redis.call("HGETALL", metaKey)
    local state = redis.call("HGETALL", stateKey)

    local metaMap = {}
    for i = 1, #meta, 2 do metaMap[meta[i]] = meta[i + 1] end
    for i = 1, #state, 2 do metaMap[state[i]] = state[i + 1] end

    local seq = tonumber(metaMap["seq"] or "0")
    local lastReadSeq = tonumber(metaMap["last_read_seq"] or "0")
    
    local summary = {
        chatId = chatId,
        name = metaMap["name"],
        lastMessageAt = metaMap["lastMessageAt"],
        lastMessage = metaMap["lastMessage"],
        pinned = metaMap["pinned"] == "1",
        muted = metaMap["muted"] == "1",
    }
    table.insert(summaries, summary)
end

return cjson.encode({
    chats = summaries,
    err = "",
})

