package executor

type scanner struct {
	target string

	selfInfo *CommandInfo

	found bool
}

func NewScanner(info *CommandInfo) *scanner {
	return &scanner{target: info.CurCmd, selfInfo: info, found: false}
}

func (s *scanner) Scan() bool {
	s.searchLowest(s.selfInfo)

	return s.found
}

func (s *scanner) searchLowest(info *CommandInfo) map[string]*CommandInfo {
	if s.found {
		return nil
	}

	if len(info.LowCmd) == 0 {
		// here is a leaf node, return the value
		return nil
	}

	for _, pInfo := range info.LowCmd {
		if pInfo.CurCmd == s.target {
			// we have found target node, directly finish.
			s.found = true
			break
		}

		pLow := s.searchLowest(pInfo)

		if len(pLow) == 0 {
			continue
		}

		info.TransitiveLow(pInfo)
	}

	return info.LowCmd
}
