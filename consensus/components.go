package consensus

//======================================================
//             request components
//======================================================

type logCert struct {
	id uint64
	assignedLogs map[uint64]bool
}

//======================================================
//             request components
//======================================================

type msgID struct {
	author   uint64
	sequence uint64
}
