package shared

import "fmt"

var (
	VoteUidAuidLockKey = "vote.lock#%v#%v#"
)

func GetVoteUidAuidLockKey(uid, auid int64) string {
	return fmt.Sprintf(VoteUidAuidLockKey, uid, auid)
}
