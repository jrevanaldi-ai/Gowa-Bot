package helper

import "time"

const aiMsgIDsCacheKey = "ai_msgids_"
const aiReplyTTL = 30 * time.Minute

func RegisterAIMessage(cache *Cache, chatJID, msgID string) {
	if cache == nil || msgID == "" {
		return
	}

	key := aiMsgIDsCacheKey + chatJID

	var ids map[string]bool
	if val, found := cache.Get(key); found {
		if existing, ok := val.(map[string]bool); ok {
			ids = existing
		}
	}
	if ids == nil {
		ids = make(map[string]bool)
	}

	ids[msgID] = true
	cache.Set(key, ids, aiReplyTTL)
}

func IsAIReply(cache *Cache, chatJID, replyMsgID string) bool {
	if cache == nil || replyMsgID == "" {
		return false
	}

	val, found := cache.Get(aiMsgIDsCacheKey + chatJID)
	if !found {
		return false
	}

	ids, ok := val.(map[string]bool)
	if !ok {
		return false
	}

	return ids[replyMsgID]
}
