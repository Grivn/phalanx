package scanner

import (
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/types"
)

type scanner struct {
	target string

	selfInfo *types.CommandInfo

	found bool
}

func NewScanner(info *types.CommandInfo) api.CondorcetScanner {
	return &scanner{target: info.Digest, selfInfo: info, found: false}
}

func (s *scanner) HasCyclic() bool {
	s.searchLowest(s.selfInfo)

	return s.found
}

func (s *scanner) searchLowest(info *types.CommandInfo) map[string]*types.CommandInfo {
	if s.found {
		return nil
	}

	if len(info.LowCmd) == 0 {
		// here is a leaf node, return the value
		return nil
	}

	for _, pInfo := range info.LowCmd {
		if pInfo.Digest == s.target {
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
