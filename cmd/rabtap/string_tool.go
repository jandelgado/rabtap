package main

// given an array of bools flags[] and an array of strings list[] of the same
// size, return an array of strings s[i] where b[i] is true
func filterStringList(flags []bool, list []string) []string {
	result := []string{}
	for i, s := range list {
		if flags[i] {
			result = append(result, s)
		}
	}
	return result
}
