package slot

func compareSlots(slot1, slot2 []uint64) bool {
	for index, val1 := range slot1 {
		val2 := slot2[index]

		if val1 != val2 {
			return false
		}
	}

	return true
}
