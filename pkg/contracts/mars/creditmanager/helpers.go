package creditmanager

import "fmt"

func Contains(arr []Account, target int) bool {
	targetStr := fmt.Sprintf("%d", target) // Convert target to string
	for _, v := range arr {
		if v.ID == targetStr {
			return true
		}
	}
	return false
}
