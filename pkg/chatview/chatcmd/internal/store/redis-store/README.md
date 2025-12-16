Redis Chat Store Schema
=======================

This store keeps lightweight chat metadata in Redis so the chatview service can show conversation lists quickly. All writes are done through Lua scripts to keep related updates atomic.

Key naming
----------
- `chats:{chatId}:meta` — hash storing chat-level metadata.
- `users:{userId}:chats` — set of chat IDs a user belongs to.
- `users:{userId}:chat:{chatId}:state` — hash storing per-user state for a chat.

Chat metadata (`chats:{chatId}:meta`)
-------------------------------------
Hash fields:
- `name` — chat/group name.
- `type` — chat type string; currently `"group"` is written by `create_group_meta.lua`.
- `last_message` — most recent message content saved via `update_last_message.lua`.
- `last_message_at` — Unix timestamp (string) of the last message.
- `seq` — monotonic sequence number incremented per message; initialized to `"0"`.

User chat membership (`users:{userId}:chats`)
---------------------------------------------
- A Redis set of chat IDs. Populated by `add_chat_to_user.lua` using `SADD`.

Per-user chat state (`users:{userId}:chat:{chatId}:state`)
----------------------------------------------------------
Hash fields initialized by `add_chat_to_user.lua`:
- `last_read_seq` — sequence number the user has read up to (starts at current chat `seq`).
- `pinned` — `"0"` or `"1"`.
- `muted` — `"0"` or `"1"`.

Lua scripts and flows
---------------------
- `create_group_meta.lua` (`CreateGroupChat`)  
  Creates `chats:{chatId}:meta` with base fields and `seq=0`. No-op if already present (guard commented out).
- `add_chat_to_user.lua` (`AddChatToUser`)  
  Adds the chat ID to `users:{userId}:chats`. If the per-user state hash does not exist, initializes `last_read_seq` to the chat’s current `seq`, with `pinned=0` and `muted=0`.
- `update_last_message.lua` (`UpdateLastMessage`)  
  Writes `last_message`, `last_message_at`, and increments `seq` atomically. Returns the new `seq`.
- `update_last_read_seq.lua` (`UpdateLastReadSeq`)  
  Sets `last_read_seq` in the user/chat state hash to the chat’s current `seq`, erroring if the chat meta hash is missing.
- `list_user_chat_summaries.lua` (unused in Go code)  
  Example read script that joins meta + state for each chat in `users:{userId}:chats` and returns JSON summaries.

Example
-------
1. `CreateGroupChat(chatId=42, name="General", timestamp=...)` creates `chats:42:meta` with `seq=0`.
2. `AddChatToUser(userId=7, chatId=42, ...)` adds `42` to `users:7:chats` and creates `users:7:chat:42:state` with `last_read_seq=0`.
3. `UpdateLastMessage(chatId=42, message="hi", timestamp=1700000000)` sets `last_message`, `last_message_at`, and bumps `seq` to `1`.
4. `UpdateLastReadSeq(userId=7, chatId=42)` updates `users:7:chat:42:state.last_read_seq` to `1`.

Operational notes
-----------------
- All keys are unversioned and use plain integers serialized as strings.
- Scripts currently do not delete keys; clean-up requires external logic.
- Sequence values (`seq`) are the primary ordering/index for unread calculations: `last_read_seq` <= `seq`.
