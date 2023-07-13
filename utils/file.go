package utils

import (
	"fmt"
	"strings"
)

func FixShellScriptVariables(script string, variableValues map[string]string) (string, error) {
	// Split the script into lines
	lines := strings.Split(script, "\n")

	// Iterate over each line and variable value pair
	for i, line := range lines {
		for variableName, newValue := range variableValues {
			if strings.HasPrefix(line, "export "+variableName+"=") {
				// Replace the existing variable assignment with the new value
				lines[i] = "export " + variableName + "=" + newValue
				fmt.Println(lines[i])
				break
			}
		}
	}

	// Join the lines back into a single script
	updatedScript := strings.Join(lines, "\n")

	return updatedScript, nil
}
