package runner

import "strings"

func Chunk(cmd string) []string {
	if cmd == "" {
		return []string{}
	}

	tokens := strings.Fields(cmd)
	output := []string{}
	currentArg := ""
	for _, token := range tokens {
		if currentArg == "" {
			if !strings.HasPrefix(token, "'") && !strings.HasPrefix(token, `"`) {
				output = append(output, token)
			} else {
				currentArg = token[1:] + " "
			}
		} else {
			if strings.HasSuffix(token, "'") || strings.HasSuffix(token, `"`) {
				currentArg = currentArg + token[:len(token)-1]
				output = append(output, currentArg)
				currentArg = ""
			} else {
				currentArg = currentArg + token + " "
			}
		}
	}
	return output
}
